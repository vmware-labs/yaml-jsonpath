/*
 * Copyright 2020 Go YAML Path Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import (
	"errors"
	"strconv"
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
	lexeme := l.nextLexeme()

	switch lexeme.typ {

	case lexemeError:
		return nil, errors.New(lexeme.val)

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
		childName := strings.TrimPrefix(lexeme.val, "..")
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
		childName := strings.TrimPrefix(lexeme.val, ".")
		return childThen(childName, subPath), nil

	case lexemeBracketChild:
		subPath, err := newPath(l)
		if err != nil {
			return new(empty), err
		}
		childNames := strings.TrimSuffix(strings.TrimPrefix(lexeme.val, "['"), "']")
		return childrenThen(childNames, subPath), nil

	case lexemeArraySubscript:
		subPath, err := newPath(l)
		if err != nil {
			return new(empty), err
		}
		subscript := strings.TrimSuffix(strings.TrimPrefix(lexeme.val, "["), "]")
		return arraySubscriptThen(subscript, subPath), nil
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
		from := 0
		step := 1
		var to int
		if subscript == "*" {
			to = len(node.Content)
		} else {
			sliceParms := strings.Split(subscript, ":")
			if len(sliceParms) > 3 {
				// need to prevent this in the lexer
				panic("malformed array subscript")
			}
			p := []int{}
			for i, s := range sliceParms {
				if i == 0 && s == "" {
					p = append(p, 0)
					continue
				}
				if i == 1 && s == "" {
					p = append(p, len(node.Content))
					continue
				}
				if i == 2 && s == "" {
					p = append(p, 1)
					continue
				}
				n, err := strconv.Atoi(s)
				if err != nil {
					// need to prevent this in the lexer
					panic("non-integer array subscript index")
				}
				p = append(p, n)
			}
			from = p[0]
			if from == -1 { // TODO: from < -1
				from = len(node.Content) - 1
			}
			to = from + 1
			if len(p) >= 2 {
				to = p[1]
			}
			if len(p) == 3 {
				step = p[2]
			}
		}
		its := []yit.Iterator{}
		if step > 0 {
			for i := from; i < to; i += step {
				its = append(its, compose(yit.FromNode(node.Content[i]), p))
			}
		} else if step < 0 {
			for i := to - 1; i >= from; i += step {
				its = append(its, compose(yit.FromNode(node.Content[i]), p))
			}
		}
		return yit.FromIterators(its...)
	})
}
