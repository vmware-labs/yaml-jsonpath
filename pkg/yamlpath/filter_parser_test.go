/*
 * Copyright 2020 VMware, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFilterNode(t *testing.T) {
	cases := []struct {
		name     string
		lexemes  []lexeme
		expected *filterNode
		focus    bool // if true, run only tests with focus set to true
	}{
		{
			name:     "no lexemes",
			lexemes:  []lexeme{},
			expected: nil,
		},
		{
			name: "integer literal",
			lexemes: []lexeme{
				{typ: lexemeFilterIntegerLiteral, val: "1"},
			},
			expected: &filterNode{
				lexeme:   lexeme{typ: lexemeFilterIntegerLiteral, val: "1"},
				subpath:  []lexeme{},
				children: []*filterNode{},
			},
		},
		{
			name: "string literal",
			lexemes: []lexeme{
				{typ: lexemeFilterStringLiteral, val: "a"},
			},
			expected: &filterNode{
				lexeme:   lexeme{typ: lexemeFilterStringLiteral, val: "a"},
				subpath:  []lexeme{},
				children: []*filterNode{},
			},
		},
		{
			name: "boolean literal",
			lexemes: []lexeme{
				{typ: lexemeFilterBooleanLiteral, val: "true"},
			},
			expected: &filterNode{
				lexeme:   lexeme{typ: lexemeFilterBooleanLiteral, val: "true"},
				subpath:  []lexeme{},
				children: []*filterNode{},
			},
		},
		{
			name: "null literal",
			lexemes: []lexeme{
				{typ: lexemeFilterNullLiteral, val: "null"},
			},
			expected: &filterNode{
				lexeme:   lexeme{typ: lexemeFilterNullLiteral, val: "null"},
				subpath:  []lexeme{},
				children: []*filterNode{},
			},
		},
		{
			name: "existence filter",
			lexemes: []lexeme{
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
			},
			expected: &filterNode{
				lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
				subpath: []lexeme{
					{typ: lexemeDotChild, val: ".child"},
				},
				children: []*filterNode{},
			},
		},
		{
			name: "numeric comparison filter, path to literal",
			lexemes: []lexeme{
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterGreaterThan, val: ">"},
				subpath: []lexeme{},
				children: []*filterNode{
					{
						lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
						subpath: []lexeme{
							{typ: lexemeDotChild, val: ".child"},
						},
						children: []*filterNode{},
					},
					{
						lexeme:   lexeme{typ: lexemeFilterIntegerLiteral, val: "1"},
						subpath:  []lexeme{},
						children: []*filterNode{},
					},
				},
			},
		},
		{
			name: "numeric comparison filter, root path to literal",
			lexemes: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterGreaterThan, val: ">"},
				subpath: []lexeme{},
				children: []*filterNode{
					{
						lexeme: lexeme{typ: lexemeRoot, val: "$"},
						subpath: []lexeme{
							{typ: lexemeDotChild, val: ".child"},
						},
						children: []*filterNode{},
					},
					{
						lexeme:   lexeme{typ: lexemeFilterIntegerLiteral, val: "1"},
						subpath:  []lexeme{},
						children: []*filterNode{},
					},
				},
			},
		},
		{
			name: "numeric comparison filter, path to path",
			lexemes: []lexeme{
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child1"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child2"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterGreaterThan, val: ">"},
				subpath: []lexeme{},
				children: []*filterNode{
					{
						lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
						subpath: []lexeme{
							{typ: lexemeDotChild, val: ".child1"},
						},
						children: []*filterNode{},
					},
					{
						lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
						subpath: []lexeme{
							{typ: lexemeDotChild, val: ".child2"},
						},
						children: []*filterNode{},
					},
				},
			},
		},
		{
			name: "existence || existence filter",
			lexemes: []lexeme{
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".a"},
				{typ: lexemeFilterOr, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".b"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterOr, val: "||"},
				subpath: []lexeme{},
				children: []*filterNode{
					{
						lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
						subpath: []lexeme{
							{typ: lexemeDotChild, val: ".a"},
						},
						children: []*filterNode{},
					},
					{
						lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
						subpath: []lexeme{
							{typ: lexemeDotChild, val: ".b"},
						},
						children: []*filterNode{},
					},
				},
			},
		},
		{
			name: "existence || existence filter with bracket children",
			lexemes: []lexeme{
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeBracketChild, val: "['a']"},
				{typ: lexemeFilterOr, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeBracketChild, val: "['b']"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterOr, val: "||"},
				subpath: []lexeme{},
				children: []*filterNode{
					{
						lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
						subpath: []lexeme{
							{typ: lexemeBracketChild, val: "['a']"},
						},
						children: []*filterNode{},
					},
					{
						lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
						subpath: []lexeme{
							{typ: lexemeBracketChild, val: "['b']"},
						},
						children: []*filterNode{},
					},
				},
			},
		},
		{
			name: "comparison || existence filter",
			lexemes: []lexeme{
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".a"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
				{typ: lexemeFilterOr, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".b"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterOr, val: "||"},
				subpath: []lexeme{},
				children: []*filterNode{
					{
						lexeme:  lexeme{typ: lexemeFilterGreaterThan, val: ">"},
						subpath: []lexeme{},
						children: []*filterNode{
							{
								lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
								subpath: []lexeme{
									{typ: lexemeDotChild, val: ".a"},
								},
								children: []*filterNode{},
							},
							{
								lexeme:   lexeme{typ: lexemeFilterIntegerLiteral, val: "1"},
								subpath:  []lexeme{},
								children: []*filterNode{},
							},
						},
					},
					{
						lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
						subpath: []lexeme{
							{typ: lexemeDotChild, val: ".b"},
						},
						children: []*filterNode{},
					},
				},
			},
		},
		{
			name: "existence || comparison filter",
			lexemes: []lexeme{
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".a"},
				{typ: lexemeFilterOr, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".b"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterOr, val: "||"},
				subpath: []lexeme{},
				children: []*filterNode{
					{
						lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
						subpath: []lexeme{
							{typ: lexemeDotChild, val: ".a"},
						},
						children: []*filterNode{},
					},
					{
						lexeme:  lexeme{typ: lexemeFilterGreaterThan, val: ">"},
						subpath: []lexeme{},
						children: []*filterNode{
							{
								lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
								subpath: []lexeme{
									{typ: lexemeDotChild, val: ".b"},
								},
								children: []*filterNode{},
							},
							{
								lexeme:   lexeme{typ: lexemeFilterIntegerLiteral, val: "1"},
								subpath:  []lexeme{},
								children: []*filterNode{},
							},
						},
					},
				},
			},
		},
		{
			name: "comparison || comparison filter",
			lexemes: []lexeme{
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".a"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
				{typ: lexemeFilterOr, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".b"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeFilterIntegerLiteral, val: "2"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterOr, val: "||"},
				subpath: []lexeme{},
				children: []*filterNode{
					{
						lexeme:  lexeme{typ: lexemeFilterGreaterThan, val: ">"},
						subpath: []lexeme{},
						children: []*filterNode{
							{
								lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
								subpath: []lexeme{
									{typ: lexemeDotChild, val: ".a"},
								},
								children: []*filterNode{},
							},
							{
								lexeme:   lexeme{typ: lexemeFilterIntegerLiteral, val: "1"},
								subpath:  []lexeme{},
								children: []*filterNode{},
							},
						},
					},
					{
						lexeme:  lexeme{typ: lexemeFilterGreaterThan, val: ">"},
						subpath: []lexeme{},
						children: []*filterNode{
							{
								lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
								subpath: []lexeme{
									{typ: lexemeDotChild, val: ".b"},
								},
								children: []*filterNode{},
							},
							{
								lexeme:   lexeme{typ: lexemeFilterIntegerLiteral, val: "2"},
								subpath:  []lexeme{},
								children: []*filterNode{},
							},
						},
					},
				},
			},
		},
		{
			name: "existence || existence && existence filter",
			lexemes: []lexeme{
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".a"},
				{typ: lexemeFilterOr, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".b"},
				{typ: lexemeFilterAnd, val: "&&"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".c"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterOr, val: "||"},
				subpath: []lexeme{},
				children: []*filterNode{
					{
						lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
						subpath: []lexeme{
							{typ: lexemeDotChild, val: ".a"},
						},
						children: []*filterNode{},
					},
					{
						lexeme:  lexeme{typ: lexemeFilterAnd, val: "&&"},
						subpath: []lexeme{},
						children: []*filterNode{
							{
								lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
								subpath: []lexeme{
									{typ: lexemeDotChild, val: ".b"},
								},
								children: []*filterNode{},
							},
							{
								lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
								subpath: []lexeme{
									{typ: lexemeDotChild, val: ".c"},
								},
								children: []*filterNode{},
							},
						},
					},
				},
			},
		},
		{
			name: "existence && existence || existence filter",
			lexemes: []lexeme{
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".a"},
				{typ: lexemeFilterAnd, val: "&&"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".b"},
				{typ: lexemeFilterOr, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".c"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterOr, val: "||"},
				subpath: []lexeme{},
				children: []*filterNode{
					{
						lexeme:  lexeme{typ: lexemeFilterAnd, val: "&&"},
						subpath: []lexeme{},
						children: []*filterNode{
							{
								lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
								subpath: []lexeme{
									{typ: lexemeDotChild, val: ".a"},
								},
								children: []*filterNode{},
							},
							{
								lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
								subpath: []lexeme{
									{typ: lexemeDotChild, val: ".b"},
								},
								children: []*filterNode{},
							},
						},
					},
					{
						lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
						subpath: []lexeme{
							{typ: lexemeDotChild, val: ".c"},
						},
						children: []*filterNode{},
					},
				},
			},
		},
		{
			name: "existence filter in parentheses",
			lexemes: []lexeme{
				{typ: lexemeFilterOpenBracket, val: "("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterCloseBracket, val: ")"},
			},
			expected: &filterNode{
				lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
				subpath: []lexeme{
					{typ: lexemeDotChild, val: ".child"},
				},
				children: []*filterNode{},
			},
		},
		{
			name: "nested filter (edge case)",
			lexemes: []lexeme{
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".y"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".z"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeDotChild, val: ".w"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeFilterIntegerLiteral, val: "2"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterEquality, val: "=="},
				subpath: []lexeme{},
				children: []*filterNode{
					{
						lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
						subpath: []lexeme{
							{typ: lexemeDotChild, val: ".y"},
							{typ: lexemeFilterBegin, val: "[?("},
							{typ: lexemeFilterAt, val: "@"},
							{typ: lexemeDotChild, val: ".z"},
							{typ: lexemeFilterEquality, val: "=="},
							{typ: lexemeFilterIntegerLiteral, val: "1"},
							{typ: lexemeFilterEnd, val: ")]"},
							{typ: lexemeDotChild, val: ".w"},
						},
						children: []*filterNode{},
					},
					{
						lexeme:   lexeme{typ: lexemeFilterIntegerLiteral, val: "2"},
						subpath:  []lexeme{},
						children: []*filterNode{},
					},
				},
			},
		},
		{
			name: "nested filter involving root (edge case)",
			lexemes: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".y"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".z"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeDotChild, val: ".w"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeFilterIntegerLiteral, val: "2"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterEquality, val: "=="},
				subpath: []lexeme{},
				children: []*filterNode{
					{
						lexeme: lexeme{typ: lexemeRoot, val: "$"},
						subpath: []lexeme{
							{typ: lexemeDotChild, val: ".y"},
							{typ: lexemeFilterBegin, val: "[?("},
							{typ: lexemeRoot, val: "$"},
							{typ: lexemeDotChild, val: ".z"},
							{typ: lexemeFilterEquality, val: "=="},
							{typ: lexemeFilterIntegerLiteral, val: "1"},
							{typ: lexemeFilterEnd, val: ")]"},
							{typ: lexemeDotChild, val: ".w"},
						},
						children: []*filterNode{},
					},
					{
						lexeme:   lexeme{typ: lexemeFilterIntegerLiteral, val: "2"},
						subpath:  []lexeme{},
						children: []*filterNode{},
					},
				},
			},
		},
		{
			name: "negated existence filter",
			lexemes: []lexeme{
				{typ: lexemeFilterNot, val: "!"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterNot, val: "!"},
				subpath: []lexeme{},
				children: []*filterNode{
					{
						lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
						subpath: []lexeme{
							{typ: lexemeDotChild, val: ".child"},
						},
						children: []*filterNode{},
					},
				},
			},
		},
		{
			name: "negated numeric comparison filter",
			lexemes: []lexeme{
				{typ: lexemeFilterNot, val: "!"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterNot, val: "!"},
				subpath: []lexeme{},
				children: []*filterNode{
					{
						lexeme:  lexeme{typ: lexemeFilterGreaterThan, val: ">"},
						subpath: []lexeme{},
						children: []*filterNode{
							{
								lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
								subpath: []lexeme{
									{typ: lexemeDotChild, val: ".child"},
								},
								children: []*filterNode{},
							},
							{
								lexeme:   lexeme{typ: lexemeFilterIntegerLiteral, val: "1"},
								subpath:  []lexeme{},
								children: []*filterNode{},
							},
						},
					},
				},
			},
		},
		{
			name: "negated parentheses",
			lexemes: []lexeme{
				{typ: lexemeFilterNot, val: "!"},
				{typ: lexemeFilterOpenBracket, val: "("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterCloseBracket, val: ")"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterNot, val: "!"},
				subpath: []lexeme{}, children: []*filterNode{
					{
						lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
						subpath: []lexeme{
							{typ: lexemeDotChild, val: ".child"},
						},
						children: []*filterNode{},
					},
				},
			},
		},
		{
			name: "regular expression match filter on path",
			lexemes: []lexeme{
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterMatchesRegularExpression, val: "=~"},
				{typ: lexemeFilterRegularExpressionLiteral, val: "/.*/"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterMatchesRegularExpression, val: "=~"},
				subpath: []lexeme{},
				children: []*filterNode{
					{
						lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
						subpath: []lexeme{
							{typ: lexemeDotChild, val: ".child"},
						},
						children: []*filterNode{},
					},
					{
						lexeme:   lexeme{typ: lexemeFilterRegularExpressionLiteral, val: "/.*/"},
						subpath:  []lexeme{},
						children: []*filterNode{},
					},
				},
			},
		},
		{
			name: "incomplete term (edge case, garbage in garbage out)",
			lexemes: []lexeme{
				{typ: lexemeFilterAt, val: "@"},
			},
			expected: &filterNode{
				lexeme:   lexeme{typ: lexemeFilterAt, val: "@"},
				subpath:  []lexeme{},
				children: []*filterNode{},
			},
		},
		{
			name: "unclosed parentheses (edge case, garbage in garbage out)",
			lexemes: []lexeme{
				{typ: lexemeFilterOpenBracket, val: "("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
			},
			expected: &filterNode{
				lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
				subpath: []lexeme{
					{typ: lexemeDotChild, val: ".child"},
				},
				children: []*filterNode{},
			},
		},
		{
			name: "unexpected end of nested filter (edge case, garbage in garbage out)",
			lexemes: []lexeme{
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".y"},
				{typ: lexemeFilterEnd, val: ")]"},
			},
			expected: &filterNode{
				lexeme: lexeme{typ: lexemeFilterAt, val: "@"},
				subpath: []lexeme{
					{typ: lexemeDotChild, val: ".y"},
				},
				children: []*filterNode{},
			},
		},
		{
			name: "unexpected close bracket (edge case, garbage in garbage out)",
			lexemes: []lexeme{
				{typ: lexemeFilterCloseBracket, val: ")"},
			},
			expected: nil,
		},
	}

	focussed := false
	for _, tc := range cases {
		if tc.focus {
			focussed = true
			break
		}
	}

	for _, tc := range cases {
		if focussed && !tc.focus {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			actual := newFilterNode(tc.lexemes)
			if focussed {
				// sometimes easier to read this than a diff
				fmt.Println("Expected:")
				fmt.Println(tc.expected.String())
				fmt.Println("Actual:")
				fmt.Println(actual.String())
			}
			require.Equal(t, tc.expected, actual)
		})
	}

	if focussed {
		t.Fatalf("testcase(s) still focussed")
	}
}

// String is only used by tests
func (n *filterNode) String() string {
	return "---\n" + n.indentedString(0) + "\n---\n"
}

func (n *filterNode) indentedString(indent int) string {
	i := strings.Repeat("    ", indent)
	s := n.lexeme.val
	for _, l := range n.subpath {
		s += l.val
	}
	c := ""
	for _, child := range n.children {
		c += "\n" + child.indentedString(indent+1)
	}
	return fmt.Sprintf("%s%s%s", i, s, c)
}
