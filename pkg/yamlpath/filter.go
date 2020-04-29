/*
 * Copyright 2020 Go YAML Path Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

func newFilter(parseTree *filterNode) func(*yaml.Node) bool {
	if parseTree == nil {
		return never
	}

	switch parseTree.lexeme.typ {
	case lexemeFilterAt:
		path := filterAtPath(parseTree)
		return func(node *yaml.Node) bool {
			return len(path(node)) > 0
		}

	case lexemeFilterGreaterThan:
		lhs := parseTree.children[0]
		rhs := parseTree.children[1]
		// FIXME: do not assume lhs is lexemeFilterAt and rhs is a numeric literal
		lhsPath := filterAtPath(lhs)
		return func(node *yaml.Node) bool {
			for _, n := range lhsPath(node) {
				// if n <= rhs return false
				if compare(n, rhs) < 1 {
					return false
				}
			}
			return true
		}

	case lexemeFilterEquality:
		lhs := parseTree.children[0]
		rhs := parseTree.children[1]
		// FIXME: do not assume lhs is lexemeFilterAt and rhs is a literal
		lhsPath := filterAtPath(lhs)
		return func(node *yaml.Node) bool {
			for _, n := range lhsPath(node) {
				if rhs.lexeme.typ == lexemeFilterStringLiteral {
					if stripFilterStringLiteral(rhs.lexeme) != n.Value {
						return false
					}
				} else if compare(n, rhs) != 0 {
					return false
				}
			}
			return true
		}

	case lexemeFilterInequality:
		lhs := parseTree.children[0]
		rhs := parseTree.children[1]
		// FIXME: do not assume lhs is lexemeFilterAt and rhs is a literal
		lhsPath := filterAtPath(lhs)
		return func(node *yaml.Node) bool {
			for _, n := range lhsPath(node) {
				if rhs.lexeme.typ == lexemeFilterStringLiteral {
					if stripFilterStringLiteral(rhs.lexeme) == n.Value {
						return false
					}
				} else if compare(n, rhs) == 0 {
					return false
				}
			}
			return true
		}

	case lexemeFilterDisjunction:
		f1 := newFilter(parseTree.children[0])
		f2 := newFilter(parseTree.children[1])
		return func(node *yaml.Node) bool {
			return f1(node) || f2(node)
		}

	case lexemeFilterConjunction:
		f1 := newFilter(parseTree.children[0])
		f2 := newFilter(parseTree.children[1])
		return func(node *yaml.Node) bool {
			return f1(node) && f2(node)
		}
	}

	panic("not implemented")
}

func never(*yaml.Node) bool {
	return false
}

func filterAtPath(parseTree *filterNode) func(*yaml.Node) []*yaml.Node {
	subpath := ""
	for _, lexeme := range parseTree.subpath {
		subpath += lexeme.val
	}
	path, err := NewPath(subpath)
	if err != nil {
		return func(*yaml.Node) []*yaml.Node {
			return []*yaml.Node{}
		}
	}
	return path.Find
}

func compare(lhs *yaml.Node, rhs *filterNode) int {
	if rhs.lexeme.typ == lexemeFilterFloatLiteral || rhs.lexeme.typ == lexemeFilterIntegerLiteral {
		rhsFloat, err := strconv.ParseFloat(rhs.lexeme.val, 64)
		if err != nil {
			panic(err)
		}
		lhsFloat, err := strconv.ParseFloat(lhs.Value, 64)
		if err != nil {
			panic(err)
		}
		return compareFloat64(lhsFloat, rhsFloat)
	}
	panic("not implemented")
}

func compareFloat64(lhs, rhs float64) int {
	if lhs < rhs {
		return -1
	}
	if lhs > rhs {
		return 1
	}
	return 0
}

func stripFilterStringLiteral(l lexeme) string {
	if l.typ != lexemeFilterStringLiteral {
		panic(fmt.Sprintf("%#v", l))
	}
	return l.val[1 : len(l.val)-1]
}
