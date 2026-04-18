/*
 * Copyright 2020 VMware, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import (
	"errors"
	"iter"
	"slices"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// Path is a compiled YAML path expression.
type Path struct {
	f func(node, root *yaml.Node) iter.Seq[*yaml.Node]
}

// Find applies the Path to a YAML node and returns the addresses of the subnodes which match the Path.
func (p *Path) Find(node *yaml.Node) ([]*yaml.Node, error) {
	return p.find(node, node), nil // currently, errors are not possible
}

func (p *Path) find(node, root *yaml.Node) []*yaml.Node {
	return slices.Collect(p.f(node, root))
}

// NewPath constructs a Path from a string expression.
func NewPath(path string) (*Path, error) {
	return newPath(lex("Path lexer", path))
}

func newPath(l *lexer) (*Path, error) {
	lx := l.nextLexeme()

	switch lx.typ {

	case lexemeError:
		return nil, errors.New(lx.val)

	case lexemeIdentity, lexemeEOF:
		return new(identity), nil

	case lexemeRoot:
		subPath, err := newPath(l)
		if err != nil {
			return nil, err
		}
		return new(func(node, root *yaml.Node) iter.Seq[*yaml.Node] {
			if node.Kind == yaml.DocumentNode {
				node = node.Content[0]
			}
			return compose(fromNode(node), subPath, root)
		}), nil

	case lexemeRecursiveDescent:
		subPath, err := newPath(l)
		if err != nil {
			return nil, err
		}
		childName := strings.TrimPrefix(lx.val, "..")
		switch childName {
		case "*":
			// includes all nodes, not just mapping nodes
			return new(func(node, root *yaml.Node) iter.Seq[*yaml.Node] {
				return compose(recurseNode(node), allChildrenThen(subPath), root)
			}), nil

		case "":
			return new(func(node, root *yaml.Node) iter.Seq[*yaml.Node] {
				return compose(recurseNode(node), subPath, root)
			}), nil

		default:
			return new(func(node, root *yaml.Node) iter.Seq[*yaml.Node] {
				return compose(recurseNode(node), childThen(childName, subPath), root)
			}), nil
		}

	case lexemeDotChild:
		subPath, err := newPath(l)
		if err != nil {
			return nil, err
		}
		childName := strings.TrimPrefix(lx.val, ".")

		return childThen(childName, subPath), nil

	case lexemeUndottedChild:
		subPath, err := newPath(l)
		if err != nil {
			return nil, err
		}

		return childThen(lx.val, subPath), nil

	case lexemeBracketChild:
		subPath, err := newPath(l)
		if err != nil {
			return nil, err
		}
		childNames := strings.TrimSpace(lx.val)
		childNames = strings.TrimSuffix(strings.TrimPrefix(childNames, "["), "]")
		childNames = strings.TrimSpace(childNames)
		return bracketChildThen(childNames, subPath), nil

	case lexemeArraySubscript:
		subPath, err := newPath(l)
		if err != nil {
			return nil, err
		}
		subscript := strings.TrimSuffix(strings.TrimPrefix(lx.val, "["), "]")
		return arraySubscriptThen(subscript, subPath), nil

	case lexemeFilterBegin, lexemeRecursiveFilterBegin:
		var recursive bool

		if lx.typ == lexemeRecursiveFilterBegin {
			recursive = true
		}
		filterLexemes := []lexeme{}
		filterNestingLevel := 1
	f:
		for {
			lx := l.nextLexeme()
			switch lx.typ {
			case lexemeFilterBegin:
				filterNestingLevel++
			case lexemeFilterEnd:
				filterNestingLevel--
				if filterNestingLevel == 0 {
					break f
				}
			case lexemeError:
				return nil, errors.New(lx.val)

			case lexemeEOF:
				// should never happen as lexer should have detected an error
				return nil, errors.New("missing end of filter")
			}
			filterLexemes = append(filterLexemes, lx)
		}

		subPath, err := newPath(l)
		if err != nil {
			return nil, err
		}
		if recursive {
			return recursiveFilterThen(filterLexemes, subPath), nil
		}
		return filterThen(filterLexemes, subPath), nil
	case lexemePropertyName:
		subPath, err := newPath(l)
		if err != nil {
			return nil, err
		}
		childName := strings.TrimPrefix(lx.val, ".")
		childName = strings.TrimSuffix(childName, propertyName)
		return propertyNameChildThen(childName, subPath), nil
	case lexemeBracketPropertyName:
		subPath, err := newPath(l)
		if err != nil {
			return nil, err
		}
		childNames := strings.TrimSpace(lx.val)
		childNames = strings.TrimSuffix(childNames, propertyName)
		childNames = strings.TrimSuffix(strings.TrimPrefix(childNames, "["), "]")
		childNames = strings.TrimSpace(childNames)
		return propertyNameBracketChildThen(childNames, subPath), nil
	case lexemeArraySubscriptPropertyName:
		subPath, err := newPath(l)
		if err != nil {
			return nil, err
		}
		subscript := strings.TrimSuffix(strings.TrimPrefix(lx.val, "["), "]~")
		return propertyNameArraySubscriptThen(subscript, subPath), nil
	}

	return nil, errors.New("invalid path syntax")
}

func identity(node, root *yaml.Node) iter.Seq[*yaml.Node] {
	if node.Kind == 0 {
		return fromNodes()
	}
	return fromNode(node)
}

func empty(node, root *yaml.Node) iter.Seq[*yaml.Node] {
	return fromNodes()
}

func compose(i iter.Seq[*yaml.Node], p *Path, root *yaml.Node) iter.Seq[*yaml.Node] {
	its := []iter.Seq[*yaml.Node]{}
	for a := range i {
		its = append(its, p.f(a, root))
	}
	return fromIterators(its...)
}

func new(f func(node, root *yaml.Node) iter.Seq[*yaml.Node]) *Path {
	return &Path{f: f}
}

func propertyNameChildThen(childName string, p *Path) *Path {
	childName = unescape(childName)

	return new(func(node, root *yaml.Node) iter.Seq[*yaml.Node] {
		if node.Kind != yaml.MappingNode {
			return empty(node, root)
		}
		for i, n := range node.Content {
			if i%2 == 0 && n.Value == childName {
				return compose(fromNode(node.Content[i]), p, root)
			}
		}
		return empty(node, root)
	})
}

func propertyNameBracketChildThen(childNames string, p *Path) *Path {
	unquotedChildren := bracketChildNames(childNames)

	return new(func(node, root *yaml.Node) iter.Seq[*yaml.Node] {
		if node.Kind != yaml.MappingNode {
			return empty(node, root)
		}
		its := []iter.Seq[*yaml.Node]{}
		for _, childName := range unquotedChildren {
			for i, n := range node.Content {
				if i%2 == 0 && n.Value == childName {
					its = append(its, fromNode(node.Content[i]))
				}
			}
		}
		return compose(fromIterators(its...), p, root)
	})
}

func propertyNameArraySubscriptThen(subscript string, p *Path) *Path {
	return new(func(node, root *yaml.Node) iter.Seq[*yaml.Node] {
		if node.Kind == yaml.MappingNode && subscript == "*" {
			its := []iter.Seq[*yaml.Node]{}
			for i, n := range node.Content {
				if i%2 != 0 {
					continue // skip child values
				}
				its = append(its, compose(fromNode(n), p, root))
			}
			return fromIterators(its...)
		}
		return empty(node, root)
	})
}

func childThen(childName string, p *Path) *Path {
	if childName == "*" {
		return allChildrenThen(p)
	}
	childName = unescape(childName)

	return new(func(node, root *yaml.Node) iter.Seq[*yaml.Node] {
		if node.Kind != yaml.MappingNode {
			return empty(node, root)
		}
		for i, n := range node.Content {
			if i%2 == 0 && n.Value == childName {
				return compose(fromNode(node.Content[i+1]), p, root)
			}
		}
		return empty(node, root)
	})
}

func bracketChildNames(childNames string) []string {
	s := strings.Split(childNames, ",")
	// reconstitute child names with embedded commas
	children := []string{}
	accum := ""
	for _, c := range s {
		if balanced(c, '\'') && balanced(c, '"') {
			if accum != "" {
				accum += "," + c
			} else {
				children = append(children, c)
				accum = ""
			}
		} else {
			if accum == "" {
				accum = c
			} else {
				accum += "," + c
				children = append(children, accum)
				accum = ""
			}
		}
	}
	if accum != "" {
		children = append(children, accum)
	}

	unquotedChildren := []string{}
	for _, c := range children {
		c = strings.TrimSpace(c)
		if strings.HasPrefix(c, "'") {
			c = strings.TrimSuffix(strings.TrimPrefix(c, "'"), "'")
		} else {
			c = strings.TrimSuffix(strings.TrimPrefix(c, `"`), `"`)
		}
		c = unescape(c)
		unquotedChildren = append(unquotedChildren, c)
	}
	return unquotedChildren
}

func balanced(c string, q rune) bool {
	bal := true
	prev := eof
	for i := 0; i < len(c); {
		rune, width := utf8.DecodeRuneInString(c[i:])
		i += width
		if rune == q {
			if i > 0 && prev == '\\' {
				prev = rune
				continue
			}
			bal = !bal
		}
		prev = rune
	}
	return bal
}

func bracketChildThen(childNames string, p *Path) *Path {
	unquotedChildren := bracketChildNames(childNames)

	return new(func(node, root *yaml.Node) iter.Seq[*yaml.Node] {
		if node.Kind != yaml.MappingNode {
			return empty(node, root)
		}
		its := []iter.Seq[*yaml.Node]{}
		for _, childName := range unquotedChildren {
			for i, n := range node.Content {
				if i%2 == 0 && n.Value == childName {
					its = append(its, fromNode(node.Content[i+1]))
				}
			}
		}
		return compose(fromIterators(its...), p, root)
	})
}

func unescape(raw string) string {
	esc := ""
	escaped := false
	for i := 0; i < len(raw); {
		rune, width := utf8.DecodeRuneInString(raw[i:])
		i += width
		if rune == '\\' {
			if escaped {
				esc += string(rune)
			}
			escaped = !escaped
			continue
		}
		escaped = false
		esc += string(rune)
	}

	return esc
}

func allChildrenThen(p *Path) *Path {
	return new(func(node, root *yaml.Node) iter.Seq[*yaml.Node] {
		switch node.Kind {
		case yaml.MappingNode:
			its := []iter.Seq[*yaml.Node]{}
			for i, n := range node.Content {
				if i%2 == 0 {
					continue // skip child names
				}
				its = append(its, compose(fromNode(n), p, root))
			}
			return fromIterators(its...)

		case yaml.SequenceNode:
			its := []iter.Seq[*yaml.Node]{}
			for i := range node.Content {
				its = append(its, compose(fromNode(node.Content[i]), p, root))
			}
			return fromIterators(its...)

		default:
			return empty(node, root)
		}
	})
}

func arraySubscriptThen(subscript string, p *Path) *Path {
	return new(func(node, root *yaml.Node) iter.Seq[*yaml.Node] {
		if node.Kind == yaml.MappingNode && subscript == "*" {
			its := []iter.Seq[*yaml.Node]{}
			for i, n := range node.Content {
				if i%2 == 0 {
					continue // skip child names
				}
				its = append(its, compose(fromNode(n), p, root))
			}
			return fromIterators(its...)
		}
		if node.Kind != yaml.SequenceNode {
			return empty(node, root)
		}

		slice, err := slice(subscript, len(node.Content))
		if err != nil {
			panic(err) // should not happen, lexer should have detected errors
		}

		its := []iter.Seq[*yaml.Node]{}
		for _, s := range slice {
			if s >= 0 && s < len(node.Content) {
				its = append(its, compose(fromNode(node.Content[s]), p, root))
			}
		}
		return fromIterators(its...)
	})
}

func filterThen(filterLexemes []lexeme, p *Path) *Path {
	filter := newFilter(newFilterNode(filterLexemes))
	return new(func(node, root *yaml.Node) iter.Seq[*yaml.Node] {
		its := []iter.Seq[*yaml.Node]{}
		if node.Kind == yaml.SequenceNode {
			for _, c := range node.Content {
				if filter(c, root) {
					its = append(its, compose(fromNode(c), p, root))
				}
			}
		} else {
			if filter(node, root) {
				its = append(its, compose(fromNode(node), p, root))
			}
		}
		return fromIterators(its...)
	})
}

func recursiveFilterThen(filterLexemes []lexeme, p *Path) *Path {
	filter := newFilter(newFilterNode(filterLexemes))
	return new(func(node, root *yaml.Node) iter.Seq[*yaml.Node] {
		its := []iter.Seq[*yaml.Node]{}

		if filter(node, root) {
			its = append(its, compose(fromNode(node), p, root))
		}
		return fromIterators(its...)
	})
}

func fromIterators(its ...iter.Seq[*yaml.Node]) iter.Seq[*yaml.Node] {
	return func(yield func(*yaml.Node) bool) {
		for _, seq := range its {
			seq(yield)
		}
	}
}

func fromNode(node *yaml.Node) iter.Seq[*yaml.Node] {
	return fromNodes(node)
}

func fromNodes(nodes ...*yaml.Node) iter.Seq[*yaml.Node] {
	return slices.Values(nodes)
}

func recurseNode(node *yaml.Node) iter.Seq[*yaml.Node] {
	return func(yield func(*yaml.Node) bool) {
		for _, n := range node.Content {
			seq := recurseNode(n)
			seq(yield)
		}
		if !yield(node) {
			return
		}
	}
}
