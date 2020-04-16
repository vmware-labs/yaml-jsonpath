/*
 * Copyright 2020 Go YAML Path Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLexer(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		expected []lexeme
	}{
		{
			name: "identity",
			path: "",
			expected: []lexeme{
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "root",
			path: "$",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "dot child",
			path: "$.child",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "dot child with implicit root",
			path: ".child",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"}, // synthetic
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "dot child with no name",
			path: "$.",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeError, val: "child name missing after ."},
			},
		},
		{
			name: "dot child with trailing dot",
			path: "$.child.",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeError, val: "child name missing after ."},
			},
		},
		{
			name: "dot child of dot child",
			path: "$.child1.child2",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child1"},
				{typ: lexemeDotChild, val: ".child2"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "dot child with array subscript",
			path: "$.child[*]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeArraySubscript, val: "[*]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "dot child with malformed array subscript",
			path: "$.child[1:2:3:4]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeError, val: "invalid array index, too many colons: [1:2:3:4]"},
			},
		},
		{
			name: "dot child with non-integer array subscript",
			path: "$.child[1:2:a]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeError, val: "invalid array index containing non-integer value: [1:2:a]"},
			},
		},
		{
			name: "bracket child",
			path: "$['child']",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketChild, val: "['child']"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "bracket child with no name",
			path: "$['']",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeError, val: "child name missing from ['']"},
			},
		},
		{
			name: "bracket child of bracket child",
			path: "$['child1']['child2']",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketChild, val: "['child1']"},
				{typ: lexemeBracketChild, val: "['child2']"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "bracket dotted child",
			path: "$['child1.child2']",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketChild, val: "['child1.child2']"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "bracket child with array subscript",
			path: "$['child'][*]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketChild, val: "['child']"},
				{typ: lexemeArraySubscript, val: "[*]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "bracket child with malformed array subscript",
			path: "$['child'][1:2:3:4]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketChild, val: "['child']"},
				{typ: lexemeError, val: "invalid array index, too many colons: [1:2:3:4]"},
			},
		},
		{
			name: "bracket child with non-integer array subscript",
			path: "$['child'][1:2:a]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketChild, val: "['child']"},
				{typ: lexemeError, val: "invalid array index containing non-integer value: [1:2:a]"},
			},
		},
		{
			name: "bracket child of dot child",
			path: "$.child1['child2']",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child1"},
				{typ: lexemeBracketChild, val: "['child2']"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "dot child of bracket child",
			path: "$['child1'].child2",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketChild, val: "['child1']"},
				{typ: lexemeDotChild, val: ".child2"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "recursive descent",
			path: "$..child",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeRecursiveDescent, val: "..child"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "recursive descent of dot child",
			path: "$.child1..child2",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child1"},
				{typ: lexemeRecursiveDescent, val: "..child2"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "recursive descent of bracket child",
			path: "$['child1']..child2",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketChild, val: "['child1']"},
				{typ: lexemeRecursiveDescent, val: "..child2"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "repeated recursive descent",
			path: "$..child1..child2",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeRecursiveDescent, val: "..child1"},
				{typ: lexemeRecursiveDescent, val: "..child2"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "recursive descent with dot child",
			path: "$..child1.child2",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeRecursiveDescent, val: "..child1"},
				{typ: lexemeDotChild, val: ".child2"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "recursive descent with bracket child",
			path: "$..child1['child2']",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeRecursiveDescent, val: "..child1"},
				{typ: lexemeBracketChild, val: "['child2']"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "recursive descent with missing name",
			path: "$..",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeError, val: "child name missing after .."},
			},
		},
		{
			name: "wildcarded children",
			path: "$.*",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".*"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "simple filter",
			path: "$[?(@.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketFilter, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterBracket, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "simple filter with leading whitespace",
			path: "$[?( @.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketFilter, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterBracket, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "simple filter with trailing whitespace",
			path: "$[?( @.child )]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketFilter, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterBracket, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "simple filter with bracket",
			path: "$[?((@.child))]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketFilter, val: "[?("},
				{typ: lexemeFilterOpenBracket, val: "("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterCloseBracket, val: ")"},
				{typ: lexemeFilterBracket, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "simple filter with bracket with extra whitespace",
			path: "$[?( ( @.child ) )]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketFilter, val: "[?("},
				{typ: lexemeFilterOpenBracket, val: "("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterCloseBracket, val: ")"},
				{typ: lexemeFilterBracket, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "simple filter with more complex subpath",
			path: "$[?((@.child[0]))]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketFilter, val: "[?("},
				{typ: lexemeFilterOpenBracket, val: "("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeArraySubscript, val: "[0]"},
				{typ: lexemeFilterCloseBracket, val: ")"},
				{typ: lexemeFilterBracket, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "simple negative filter",
			path: "$[?(!@.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketFilter, val: "[?("},
				{typ: lexemeFilterNot, val: "!"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterBracket, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "simple negative filter with extra whitespace",
			path: "$[?( ! @.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketFilter, val: "[?("},
				{typ: lexemeFilterNot, val: "!"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterBracket, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "simple filter with root expression",
			path: "$[?($.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketFilter, val: "[?("},
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterBracket, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter conjunction",
			path: "$[?(@.child&&@.other)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketFilter, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterConjunction, val: "&&"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".other"},
				{typ: lexemeFilterBracket, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter invalid leading conjunction",
			path: "$[?(&&",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketFilter, val: "[?("},
				{typ: lexemeError, val: "missing first operand for binary operator &&"},
			},
		},
		{
			name: "filter conjunction with extra whitespace",
			path: "$[?(@.child && @.other)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketFilter, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterConjunction, val: "&&"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".other"},
				{typ: lexemeFilterBracket, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter disjunction",
			path: "$[?(@.child||@.other)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketFilter, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterDisjunction, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".other"},
				{typ: lexemeFilterBracket, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter invalid leading disjunction",
			path: "$[?(||",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketFilter, val: "[?("},
				{typ: lexemeError, val: "missing first operand for binary operator ||"},
			},
		},
		{
			name: "filter disjunction with extra whitespace",
			path: "$[?(@.child || @.other)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketFilter, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterDisjunction, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".other"},
				{typ: lexemeFilterBracket, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := lex("test", tc.path)
			actual := []lexeme{}
			for {
				lexeme := l.nextLexeme()
				if lexeme.typ == lexemeEOF {
					break
				}
				actual = append(actual, lexeme)
			}
			require.Equal(t, tc.expected, actual)
		})
	}
}
