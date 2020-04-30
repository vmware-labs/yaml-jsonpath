/*
 * Copyright 2020 Go YAML Path Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import (
	"fmt"
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
			name: "literal",
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
				{typ: lexemeFilterDisjunction, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".b"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterDisjunction, val: "||"},
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
			name: "comparison || existence filter",
			lexemes: []lexeme{
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".a"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
				{typ: lexemeFilterDisjunction, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".b"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterDisjunction, val: "||"},
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
				{typ: lexemeFilterDisjunction, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".b"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterDisjunction, val: "||"},
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
				{typ: lexemeFilterDisjunction, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".b"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeFilterIntegerLiteral, val: "2"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterDisjunction, val: "||"},
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
				{typ: lexemeFilterDisjunction, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".b"},
				{typ: lexemeFilterConjunction, val: "&&"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".c"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterDisjunction, val: "||"},
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
						lexeme:  lexeme{typ: lexemeFilterConjunction, val: "&&"},
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
				{typ: lexemeFilterConjunction, val: "&&"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".b"},
				{typ: lexemeFilterDisjunction, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".c"},
			},
			expected: &filterNode{
				lexeme:  lexeme{typ: lexemeFilterDisjunction, val: "||"},
				subpath: []lexeme{},
				children: []*filterNode{
					{
						lexeme:  lexeme{typ: lexemeFilterConjunction, val: "&&"},
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
				{typ: lexemeBracketFilter, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".z"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
				{typ: lexemeFilterBracket, val: ")]"},
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
							{typ: lexemeBracketFilter, val: "[?("},
							{typ: lexemeFilterAt, val: "@"},
							{typ: lexemeDotChild, val: ".z"},
							{typ: lexemeFilterEquality, val: "=="},
							{typ: lexemeFilterIntegerLiteral, val: "1"},
							{typ: lexemeFilterBracket, val: ")]"},
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
