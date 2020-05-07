/*
 * Copyright 2020 Go YAML Path Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import (
	"regexp"

	"gopkg.in/yaml.v3"
)

type filter func(node, root *yaml.Node) bool

func newFilter(n *filterNode) filter {
	if n == nil {
		return never
	}

	switch n.lexeme.typ {
	case lexemeFilterAt, lexemeRoot:
		path := pathFilterScanner(n)
		return func(node, root *yaml.Node) bool {
			return len(path(node, root)) > 0
		}

	case lexemeFilterEquality, lexemeFilterInequality,
		lexemeFilterGreaterThan, lexemeFilterGreaterThanOrEqual,
		lexemeFilterLessThan, lexemeFilterLessThanOrEqual:
		return comparisonFilter(n)

	case lexemeFilterMatchesRegularExpression:
		return matchRegularExpression(n)

	case lexemeFilterNot:
		f := newFilter(n.children[0])
		return func(node, root *yaml.Node) bool {
			return !f(node, root)
		}

	case lexemeFilterOr:
		f1 := newFilter(n.children[0])
		f2 := newFilter(n.children[1])
		return func(node, root *yaml.Node) bool {
			return f1(node, root) || f2(node, root)
		}

	case lexemeFilterAnd:
		f1 := newFilter(n.children[0])
		f2 := newFilter(n.children[1])
		return func(node, root *yaml.Node) bool {
			return f1(node, root) && f2(node, root)
		}

	default:
		return never
	}
}

func never(node, root *yaml.Node) bool {
	return false
}

func comparisonFilter(n *filterNode) filter {
	return nodeToFilter(n, func(l, r string) bool {
		return n.lexeme.comparator()(compareNodeValues(l, r))
	})
}

func nodeToFilter(n *filterNode, accept func(string, string) bool) filter {
	lhsPath := newFilterScanner(n.children[0])
	rhsPath := newFilterScanner(n.children[1])
	return func(node, root *yaml.Node) (result bool) {
		match := false
		for _, l := range lhsPath(node, root) {
			for _, r := range rhsPath(node, root) {
				if !accept(l, r) {
					return false
				}
				match = true
			}
		}
		return match
	}
}

type filterScanner func(*yaml.Node, *yaml.Node) []string

func emptyScanner(*yaml.Node, *yaml.Node) []string {
	return []string{}
}

func newFilterScanner(n *filterNode) filterScanner {
	switch {
	case n == nil:
		return emptyScanner

	case n.isItemFilter():
		return pathFilterScanner(n)

	case n.isLiteral():
		return literalFilterScanner(n)

	default:
		return emptyScanner
	}
}

func pathFilterScanner(n *filterNode) filterScanner {
	var at bool
	switch n.lexeme.typ {
	case lexemeFilterAt:
		at = true
	case lexemeRoot:
		at = false
	default:
		panic("false precondition")
	}
	subpath := ""
	for _, lexeme := range n.subpath {
		subpath += lexeme.val
	}
	path, err := NewPath(subpath)
	if err != nil {
		return func(node, root *yaml.Node) []string {
			return []string{}
		}
	}
	return func(node, root *yaml.Node) []string {
		if at {
			return values(path.Find(node))
		}
		return values(path.Find(root))
	}
}

func values(nodes []*yaml.Node) []string {
	v := []string{}
	for _, n := range nodes {
		v = append(v, n.Value)
	}
	return v
}

func literalFilterScanner(n *filterNode) filterScanner {
	v := n.lexeme.literalValue()
	return func(node, root *yaml.Node) []string {
		return []string{v}
	}
}

func matchRegularExpression(parseTree *filterNode) filter {
	return nodeToFilter(parseTree, stringMatchesRegularExpression)
}

func stringMatchesRegularExpression(s, expr string) bool {
	re, _ := regexp.Compile(expr) // regex already compiled during lexing
	return re.Match([]byte(s))
}
