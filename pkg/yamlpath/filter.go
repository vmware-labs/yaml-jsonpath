/*
 * Copyright 2020 Go YAML Path Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type filter func(node, root *yaml.Node) bool

func newFilter(parseTree *filterNode) filter {
	if parseTree == nil {
		return never
	}

	switch parseTree.lexeme.typ {
	case lexemeFilterAt, lexemeRoot:
		path := filterPath(parseTree)
		return func(node, root *yaml.Node) bool {
			return len(path(node, root)) > 0
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
		return compareChildren(parseTree, equal)

	case lexemeFilterInequality:
		return compareChildren(parseTree, notEqual)

	case lexemeFilterMatchesRegularExpression:
		return matchRegularExpression(parseTree)

	case lexemeFilterNot:
		f := newFilter(parseTree.children[0])
		return func(node, root *yaml.Node) bool {
			return !f(node, root)
		}

	case lexemeFilterOr:
		f1 := newFilter(parseTree.children[0])
		f2 := newFilter(parseTree.children[1])
		return func(node, root *yaml.Node) bool {
			return f1(node, root) || f2(node, root)
		}

	case lexemeFilterAnd:
		f1 := newFilter(parseTree.children[0])
		f2 := newFilter(parseTree.children[1])
		return func(node, root *yaml.Node) bool {
			return f1(node, root) && f2(node, root)
		}

	default:
		return never
	}
}

var _ filter = never

func never(node, root *yaml.Node) bool {
	return false
}

func compareChildren(parseTree *filterNode, accept func(int) bool) filter {
	lhs := parseTree.children[0]
	rhs := parseTree.children[1]
	if lhs == nil || rhs == nil {
		return never
	}
	if isItemFilter(lhs) {
		switch {
		case isLiteral(rhs):
			lhsPath := filterPath(lhs)
			return func(node, root *yaml.Node) (result bool) {
				defer func() {
					if p := recover(); p != nil {
						result = false
					}
				}()
				match := false
				for _, n := range lhsPath(node, root) {
					if !accept(compareNodeToLiteral(n, rhs)) {
						return false
					}
					match = true
				}
				return match
			}

		case isItemFilter(rhs):
			lhsPath := filterPath(lhs)
			rhsPath := filterPath(rhs)
			return func(node, root *yaml.Node) (result bool) {
				defer func() {
					if p := recover(); p != nil {
						result = false
					}
				}()
				match := false
				for _, m := range lhsPath(node, root) {
					for _, n := range rhsPath(node, root) {
						if !accept(compareNodes(m, n)) {
							return false
						}
						match = true
					}
				}
				return match
			}

		default:
			panic("missing case")
		}

	} else if isLiteral(lhs) {
		switch {
		case isItemFilter(rhs):
			rhsPath := filterPath(rhs)
			return func(node, root *yaml.Node) (result bool) {
				defer func() {
					if p := recover(); p != nil {
						result = false
					}
				}()
				match := false
				for _, n := range rhsPath(node, root) {
					if !accept(-compareNodeToLiteral(n, lhs)) {
						return false
					}
					match = true
				}
				return match
			}

		case isNumericLiteral(rhs):
			return func(node, root *yaml.Node) bool {
				return accept(compareNumericLiterals(lhs, rhs))
			}

		case isStringLiteral(rhs):
			return func(node, root *yaml.Node) bool {
				return accept(compareStringLiterals(lhs, rhs))
			}

		default:
			panic("missing case")
		}
	}
	return never
}

func filterPath(parseTree *filterNode) func(*yaml.Node, *yaml.Node) []*yaml.Node {
	var at bool
	switch parseTree.lexeme.typ {
	case lexemeFilterAt:
		at = true
	case lexemeRoot:
		at = false
	default:
		panic("false precondition")
	}
	subpath := ""
	for _, lexeme := range parseTree.subpath {
		subpath += lexeme.val
	}
	path, err := NewPath(subpath)
	if err != nil {
		return func(node, root *yaml.Node) []*yaml.Node {
			return []*yaml.Node{}
		}
	}
	return func(node, root *yaml.Node) []*yaml.Node {
		if at {
			return path.Find(node)
		}
		return path.Find(root)
	}
}

func equal(c int) bool {
	return c == 0
}

func notEqual(c int) bool {
	return c != 0
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
	return n.lexeme.typ == lexemeFilterAt || n.lexeme.typ == lexemeRoot
}

func isLiteral(n *filterNode) bool {
	return isStringLiteral(n) || isNumericLiteral(n)
}

func isStringLiteral(n *filterNode) bool {
	return n.lexeme.typ == lexemeFilterStringLiteral
}

func isNumericLiteral(n *filterNode) bool {
	return n.lexeme.typ == lexemeFilterFloatLiteral || n.lexeme.typ == lexemeFilterIntegerLiteral
}

func compareNodeToLiteral(lhs *yaml.Node, rhs *filterNode) int {
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
	} else if isStringLiteral(rhs) {
		if lhs.Value == stripFilterStringLiteral(rhs.lexeme) {
			return 0
		}
		return 1 // any non-zero value equally valid
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

func compareNumericLiterals(lhs *filterNode, rhs *filterNode) int {
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

func compareStringLiterals(lhs *filterNode, rhs *filterNode) int {
	if isStringLiteral(lhs) && isStringLiteral(rhs) {
		if lhs.lexeme.val == rhs.lexeme.val {
			return 0
		}
		return 1 // any non-zero value is enough
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

func matchRegularExpression(parseTree *filterNode) filter {
	lhs := parseTree.children[0]
	rhs := parseTree.children[1]
	if lhs == nil || rhs == nil {
		return never
	}
	if isItemFilter(lhs) {
		lhsPath := filterPath(lhs)
		return func(node, root *yaml.Node) (result bool) {
			match := false
			for _, n := range lhsPath(node, root) {
				if !stringMatchesRegularExpression(n.Value, rhs) {
					return false
				}
				match = true
			}
			return match
		}
	} else if isStringLiteral(lhs) {
		return func(node, root *yaml.Node) (result bool) {
			return stringMatchesRegularExpression(stripFilterStringLiteral(lhs.lexeme), rhs)
		}
	}
	return never
}

func stringMatchesRegularExpression(s string, rhs *filterNode) bool {
	re, _ := regex(rhs.lexeme.val) // regex already compiled during lexing
	return re.Match([]byte(s))
}

func regex(rawRegex string) (*regexp.Regexp, error) {
	re := strings.ReplaceAll(rawRegex[1:len(rawRegex)-1], `\/`, `/`)
	return regexp.Compile(re)
}
