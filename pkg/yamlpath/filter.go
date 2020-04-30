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
		return compareChildren(parseTree, greaterThan)

	case lexemeFilterGreaterThanOrEqual:
		return compareChildren(parseTree, greaterThanOrEqual)

	case lexemeFilterLessThan:
		return compareChildren(parseTree, lessThan)

	case lexemeFilterLessThanOrEqual:
		return compareChildren(parseTree, lessThanOrEqual)

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

	default:
		return never
	}
}

func never(*yaml.Node) bool {
	return false
}

func compareChildren(parseTree *filterNode, accept func(int) bool) func(*yaml.Node) bool {
	lhs := parseTree.children[0]
	rhs := parseTree.children[1]
	if lhs == nil || rhs == nil {
		return never
	}
	if isItemFilter(lhs) {
		if isNumericLiteral(rhs) {
			lhsPath := filterAtPath(lhs)
			return func(node *yaml.Node) (result bool) {
				defer func() {
					if p := recover(); p != nil {
						result = false
					}
				}()
				match := false
				for _, n := range lhsPath(node) {
					if !accept(compare(n, rhs)) {
						return false
					}
					match = true
				}
				return match
			}
		} else if isItemFilter(rhs) {
			lhsPath := filterAtPath(lhs)
			rhsPath := filterAtPath(rhs)
			return func(node *yaml.Node) (result bool) {
				defer func() {
					if p := recover(); p != nil {
						result = false
					}
				}()
				match := false
				for _, m := range lhsPath(node) {
					for _, n := range rhsPath(node) {
						if !accept(compareNodes(m, n)) {
							return false
						}
						match = true
					}
				}
				return match
			}
		}
	} else if isNumericLiteral(lhs) {
		if isItemFilter(rhs) {
			rhsPath := filterAtPath(rhs)
			return func(node *yaml.Node) (result bool) {
				defer func() {
					if p := recover(); p != nil {
						result = false
					}
				}()
				match := false
				for _, n := range rhsPath(node) {
					if !accept(-compare(n, lhs)) {
						return false
					}
					match = true
				}
				return match
			}
		} else if isNumericLiteral(rhs) {
			return func(node *yaml.Node) bool {
				return accept(compareLiterals(lhs, rhs))
			}
		}
	}
	return never
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

func greaterThan(c int) bool {
	return c > 0
}

func greaterThanOrEqual(c int) bool {
	return c >= 0
}

func lessThan(c int) bool {
	return c < 0
}

func lessThanOrEqual(c int) bool {
	return c <= 0
}

func isItemFilter(n *filterNode) bool {
	return n.lexeme.typ == lexemeFilterAt // TODO: add root too
}

func isNumericLiteral(n *filterNode) bool {
	return n.lexeme.typ == lexemeFilterFloatLiteral || n.lexeme.typ == lexemeFilterIntegerLiteral
}

func compare(lhs *yaml.Node, rhs *filterNode) int {
	if isNumericLiteral(rhs) {
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

func compareNodes(lhs *yaml.Node, rhs *yaml.Node) int {
	lhsFloat, err := strconv.ParseFloat(lhs.Value, 64)
	if err != nil {
		panic(err)
	}
	rhsFloat, err := strconv.ParseFloat(rhs.Value, 64)
	if err != nil {
		panic(err)
	}
	return compareFloat64(lhsFloat, rhsFloat)
}

func compareLiterals(lhs *filterNode, rhs *filterNode) int {
	if isNumericLiteral(lhs) && isNumericLiteral(rhs) {
		rhsFloat, err := strconv.ParseFloat(rhs.lexeme.val, 64)
		if err != nil {
			panic(err)
		}
		lhsFloat, err := strconv.ParseFloat(lhs.lexeme.val, 64)
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
