/*
 * Copyright 2020 Go YAML Path Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import (
	"errors"
	"strings"

	"github.com/dprotaso/go-yit"
	"gopkg.in/yaml.v3"
)

// Path is a compiled YAML path expression.
type Path struct {
	f func(*yaml.Node) yit.Iterator
}

// Find applies the Path to a YAML node and returns the addresses of the subnodes which match the Path.
func (p *Path) Find(node *yaml.Node) []*yaml.Node {
	return p.f(node).ToArray()
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
			return new(empty), err
		}
		return new(func(node *yaml.Node) yit.Iterator {
			if node.Kind != yaml.DocumentNode {
				return empty(node)
			}
			return compose(yit.FromNode(node.Content[0]), subPath)
		}), nil

	case lexemeRecursiveDescent:
		subPath, err := newPath(l)
		if err != nil {
			return new(empty), err
		}
		childName := strings.TrimPrefix(lx.val, "..")
		if childName == "*" { // includes all nodes, not just mapping nodes
			return new(func(node *yaml.Node) yit.Iterator {
				return compose(yit.FromNode(node).RecurseNodes(), subPath)
			}), nil
		}
		return new(func(node *yaml.Node) yit.Iterator {
			return compose(yit.FromNode(node).RecurseNodes(), childThen(childName, subPath))
		}), nil

	case lexemeDotChild:
		subPath, err := newPath(l)
		if err != nil {
			return new(empty), err
		}
		childName := strings.TrimPrefix(lx.val, ".")
		return childThen(childName, subPath), nil

	case lexemeBracketChild:
		subPath, err := newPath(l)
		if err != nil {
			return new(empty), err
		}
		childNames := strings.TrimSuffix(strings.TrimPrefix(lx.val, "['"), "']")
		return childrenThen(childNames, subPath), nil

	case lexemeArraySubscript:
		subPath, err := newPath(l)
		if err != nil {
			return new(empty), err
		}
		subscript := strings.TrimSuffix(strings.TrimPrefix(lx.val, "["), "]")
		return arraySubscriptThen(subscript, subPath), nil

	case lexemeBracketFilter:
		filterLexemes := []lexeme{}
	f:
		for {
			lx := l.nextLexeme()
			switch lx.typ {
			case lexemeFilterBracket:
				break f
			case lexemeEOF:
				// should never happen as lexer should have detected an error
				return new(empty), errors.New("missing end of filter")
			}
			filterLexemes = append(filterLexemes, lx)
		}

		subPath, err := newPath(l)
		if err != nil {
			return new(empty), err
		}
		return filterThen(filterLexemes, subPath), nil
	}

	return new(empty), errors.New("invalid path syntax")
}

func identity(node *yaml.Node) yit.Iterator {
	return yit.FromNode(node)
}

func empty(*yaml.Node) yit.Iterator {
	return yit.FromNodes()
}

func compose(i yit.Iterator, p *Path) yit.Iterator {
	its := []yit.Iterator{}
	for a, ok := i(); ok; a, ok = i() {
		its = append(its, p.f(a))
	}
	return yit.FromIterators(its...)
}

func new(f func(node *yaml.Node) yit.Iterator) *Path {
	return &Path{f: f}
}

func childrenThen(childNames string, p *Path) *Path {
	c := strings.SplitN(childNames, ".", 2)
	if len(c) == 2 {
		return childThen(c[0], childrenThen(c[1], p))
	}
	return childThen(c[0], p)
}

func childThen(childName string, p *Path) *Path {
	if childName == "*" {
		return allChildrenThen(p)
	}
	return new(func(node *yaml.Node) yit.Iterator {
		if node.Kind != yaml.MappingNode {
			return empty(node)
		}
		for i, n := range node.Content {
			if n.Value == childName {
				return compose(yit.FromNode(node.Content[i+1]), p)
			}
		}
		return empty(node)
	})
}

func allChildrenThen(p *Path) *Path {
	return new(func(node *yaml.Node) yit.Iterator {
		if node.Kind != yaml.MappingNode {
			return empty(node)
		}
		its := []yit.Iterator{}
		for _, n := range node.Content {
			its = append(its, compose(yit.FromNode(n), p))
		}
		return yit.FromIterators(its...)
	})
}

func arraySubscriptThen(subscript string, p *Path) *Path {
	return new(func(node *yaml.Node) yit.Iterator {
		if node.Kind != yaml.SequenceNode {
			return empty(node)
		}

		slice, err := slice(subscript, len(node.Content))
		if err != nil {
			panic(err) // should not happen, lexer should have detected errors
		}

		its := []yit.Iterator{}
		for _, s := range slice {
			its = append(its, compose(yit.FromNode(node.Content[s]), p))

		}
		return yit.FromIterators(its...)
	})
}

func filterThen(filterLexemes []lexeme, p *Path) *Path {
	filter := parseFilter(filterLexemes)
	return new(func(node *yaml.Node) yit.Iterator {
		if node.Kind != yaml.SequenceNode {
			panic("not implemented")
		}

		its := []yit.Iterator{}
		for _, c := range node.Content {
			if filter(c) {
				its = append(its, compose(yit.FromNode(c), p))
			}
		}
		return yit.FromIterators(its...)
	})
}

func parseFilter(filterLexemes []lexeme) func(*yaml.Node) bool {
	return func(*yaml.Node) bool {
		return true
	}
}
