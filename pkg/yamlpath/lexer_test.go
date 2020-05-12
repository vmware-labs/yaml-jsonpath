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
		focus    bool // if true, run only tests with focus set to true
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
			name: "unmatched closing bracket",
			path: ")",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeError, val: `syntax error at position 0, following ""`},
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
				{typ: lexemeError, val: `child name missing at position 2, following "$."`},
			},
		},
		{
			name: "dot child with missing dot",
			path: "$a",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeError, val: `invalid path syntax at position 1, following "$"`},
			},
		},
		{
			name: "dot child with trailing dot",
			path: "$.child.",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeError, val: `child name missing at position 8, following ".child."`},
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
				{typ: lexemeError, val: "invalid array index, too many colons: [1:2:3:4] before position 16"},
			},
		},
		{
			name: "dot child with non-integer array subscript",
			path: "$.child[1:2:a]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeError, val: "invalid array index containing non-integer value: [1:2:a] before position 14"},
			},
		},
		{
			name: "dot child with unclosed array subscript",
			path: "$.child[*",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeError, val: `unmatched [ at position 9, following ".child[*"`},
			},
		},
		{
			name: "dot child with missing array subscript",
			path: "$.child[]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeError, val: "subscript missing from [] before position 9"},
			},
		},
		{
			name: "dot child with embedded space",
			path: "$.child more",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeError, val: `invalid character ' ' at position 7, following ".child"`},
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
				{typ: lexemeError, val: "child name missing from [''] before position 5"},
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
				{typ: lexemeError, val: "invalid array index, too many colons: [1:2:3:4] before position 19"},
			},
		},
		{
			name: "bracket child with non-integer array subscript",
			path: "$['child'][1:2:a]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketChild, val: "['child']"},
				{typ: lexemeError, val: "invalid array index containing non-integer value: [1:2:a] before position 17"},
			},
		},
		{
			name: "bracket child with unclosed array subscript",
			path: "$['child'][*",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketChild, val: "['child']"},
				{typ: lexemeError, val: `unmatched [ at position 12, following "['child'][*"`},
			},
		},
		{
			name: "bracket child with missing array subscript",
			path: "$['child'][]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketChild, val: "['child']"},
				{typ: lexemeError, val: "subscript missing from [] before position 12"},
			},
		},
		{
			name: "bracket child followed by space",
			path: "$['child'] ",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketChild, val: "['child']"},
				{typ: lexemeError, val: `invalid character ' ' at position 10, following "['child']"`},
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
				{typ: lexemeError, val: "invalid array index, too many colons: [1:2:3:4] before position 19"},
			},
		},
		{
			name: "bracket child with non-integer array subscript",
			path: "$['child'][1:2:a]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeBracketChild, val: "['child']"},
				{typ: lexemeError, val: "invalid array index containing non-integer value: [1:2:a] before position 17"},
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
				{typ: lexemeError, val: `child name missing at position 3, following "$.."`},
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
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "simple filter with leading whitespace",
			path: "$[?( @.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "simple filter with trailing whitespace",
			path: "$[?( @.child )]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "simple filter with bracket",
			path: "$[?((@.child))]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterOpenBracket, val: "("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterCloseBracket, val: ")"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "simple filter with bracket with extra whitespace",
			path: "$[?( ( @.child ) )]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterOpenBracket, val: "("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterCloseBracket, val: ")"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "simple filter with more complex subpath",
			path: "$[?((@.child[0]))]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterOpenBracket, val: "("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeArraySubscript, val: "[0]"},
				{typ: lexemeFilterCloseBracket, val: ")"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "missing filter ",
			path: "$[?()]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeError, val: `missing filter at position 4, following "[?("`},
			},
		},
		{
			name: "unclosed filter",
			path: "$[?(",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeError, val: `invalid filter syntax at position 4, following "[?("`},
			},
		},
		{
			name: "filter with missing operator",
			path: "$[?(@.child @.other)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeError, val: `invalid filter expression at position 12, following ".child "`},
			},
		},
		{
			name: "filter with malformed term",
			path: "$[?([)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeError, val: `invalid filter syntax at position 4, following "[?("`},
			},
		},
		{
			name: "filter with misplaced open bracket",
			path: "$[?(@.child ()]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeError, val: `invalid filter expression at position 12, following ".child "`},
			},
		},
		{
			name: "simple negative filter",
			path: "$[?(!@.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterNot, val: "!"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "misplaced filter negation",
			path: "$[?(@.child !@.other)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeError, val: `invalid filter expression at position 12, following ".child "`},
			},
		},
		{
			name: "simple negative filter with extra whitespace",
			path: "$[?( ! @.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterNot, val: "!"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "simple filter with root expression",
			path: "$[?($.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter integer equality, literal on the right",
			path: "$[?(@.child==1)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter integer equality with invalid literal",
			path: "$[?(@.child==-)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeError, val: `invalid integer literal "-": invalid syntax before position 14`},
			},
		},
		{
			name: "filter integer equality with integer literal which is too large",
			path: "$[?(@.child==9223372036854775808)]", // 2**63, too large for signed 64-bit integer
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeError, val: `invalid integer literal "9223372036854775808": value out of range before position 32`},
			},
		},
		{
			name: "filter integer equality with invalid float literal",
			path: "$[?(@.child==1.2.3)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeError, val: `invalid float literal "1.2.3": invalid syntax before position 18`},
			},
		},
		{
			name: "filter integer equality with invalid string literal",
			path: "$[?(@.child=='x)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeError, val: `unmatched string delimiter "'" at position 13, following "=="`},
			},
		},
		{
			name: "filter integer equality, literal on the left",
			path: "$[?(1==@.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter float equality, literal on the left",
			path: "$[?(1.5==@.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterFloatLiteral, val: "1.5"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter equality with missing left hand value",
			path: "$[?(==@.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeError, val: `missing first operand for binary operator == at position 4, following "[?("`},
			},
		},
		{
			name: "filter equality with missing left hand value inside bracket",
			path: "$[?((==@.child))]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterOpenBracket, val: "("},
				{typ: lexemeError, val: `missing first operand for binary operator == at position 5, following "("`},
			},
		},
		{
			name: "filter equality with missing right hand value",
			path: "$[?(@.child==)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeError, val: `invalid filter term at position 13, following "=="`},
			},
		},
		{
			name: "filter integer equality, root path on the right",
			path: "$[?(@.child==$.x)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".x"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter integer equality, root path on the left",
			path: "$[?($.x==@.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".x"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter string equality, literal on the right",
			path: "$[?(@.child=='x')]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeFilterStringLiteral, val: "'x'"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter string equality, literal on the left",
			path: "$[?('x'==@.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterStringLiteral, val: "'x'"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter string equality, literal on the left with unmatched string delimiter",
			path: "$[?('x==@.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeError, val: `unmatched string delimiter "'" at position 4, following "[?("`},
			},
		},
		{
			name: "filter string equality with unmatched string delimiter",
			path: "$[?(@.child=='x)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeError, val: `unmatched string delimiter "'" at position 13, following "=="`},
			},
		},
		{
			name: "filter integer inequality, literal on the right",
			path: "$[?(@.child!=1)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterInequality, val: "!="},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter inequality with missing left hand operator",
			path: "$[?(!=1)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeError, val: `missing first operand for binary operator != at position 4, following "[?("`},
			},
		},
		{
			name: "filter equality with missing right hand value",
			path: "$[?(@.child!=)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterInequality, val: "!="},
				{typ: lexemeError, val: `invalid filter term at position 13, following "!="`},
			},
		},
		{
			name: "filter greater than, integer literal on the right",
			path: "$[?(@.child>1)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter greater than, decimal literal on the right",
			path: "$[?(@.child> 1.5)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeFilterFloatLiteral, val: "1.5"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter greater than, path to path",
			path: "$[?(@.child1>@.child2)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child1"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child2"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter greater than with left hand operand missing",
			path: "$[?(>1)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeError, val: `missing first operand for binary operator > at position 4, following "[?("`},
			},
		},
		{
			name: "filter greater than with missing right hand value",
			path: "$[?(@.child>)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeError, val: `invalid filter term at position 12, following ">"`},
			},
		},
		{
			name: "filter greater than, string on the right",
			path: "$[?(@.child>'x')]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeError, val: `strings cannot be compared using > at position 12, following ">"`},
			},
		},
		{
			name: "filter greater than, string on the left",
			path: "$[?('x'>@.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterStringLiteral, val: "'x'"},
				{typ: lexemeError, val: `strings cannot be compared using > at position 7, following "'x'"`},
			},
		},
		{
			name: "filter greater than or equal, integer literal on the right",
			path: "$[?(@.child>=1)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterGreaterThanOrEqual, val: ">="},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter greater than or equal, decimal literal on the right",
			path: "$[?(@.child>=1.5)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterGreaterThanOrEqual, val: ">="},
				{typ: lexemeFilterFloatLiteral, val: "1.5"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter greater than or equal with left hand operand missing",
			path: "$[?(>=1)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeError, val: `missing first operand for binary operator >= at position 4, following "[?("`},
			},
		},
		{
			name: "filter greater than or equal with missing right hand value",
			path: "$[?(@.child>=)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterGreaterThanOrEqual, val: ">="},
				{typ: lexemeError, val: `invalid filter term at position 13, following ">="`},
			},
		},
		{
			name: "filter greater than or equal, string on the right",
			path: "$[?(@.child>='x')]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterGreaterThanOrEqual, val: ">="},
				{typ: lexemeError, val: `strings cannot be compared using >= at position 13, following ">="`},
			},
		},
		{
			name: "filter greater than or equal, string on the left",
			path: "$[?('x'>=@.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterStringLiteral, val: "'x'"},
				{typ: lexemeError, val: `strings cannot be compared using >= at position 7, following "'x'"`},
			},
		},
		{
			name: "filter less than, integer literal on the right",
			path: "$[?(@.child<1)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterLessThan, val: "<"},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter less than, decimal literal on the right",
			path: "$[?(@.child< 1.5)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterLessThan, val: "<"},
				{typ: lexemeFilterFloatLiteral, val: "1.5"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter less than with left hand operand missing",
			path: "$[?(<1)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeError, val: `missing first operand for binary operator < at position 4, following "[?("`},
			},
		},
		{
			name: "filter less than with missing right hand value",
			path: "$[?(@.child<)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterLessThan, val: "<"},
				{typ: lexemeError, val: `invalid filter term at position 12, following "<"`},
			},
		},
		{
			name: "filter less than, string on the right",
			path: "$[?(@.child<'x')]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterLessThan, val: "<"},
				{typ: lexemeError, val: `strings cannot be compared using < at position 12, following "<"`},
			},
		},
		{
			name: "filter less than, string on the left",
			path: "$[?('x'<@.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterStringLiteral, val: "'x'"},
				{typ: lexemeError, val: `strings cannot be compared using < at position 7, following "'x'"`},
			},
		},
		{
			name: "filter less than or equal, integer literal on the right",
			path: "$[?(@.child<=1)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterLessThanOrEqual, val: "<="},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter less than or equal, decimal literal on the right",
			path: "$[?(@.child<=1.5)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterLessThanOrEqual, val: "<="},
				{typ: lexemeFilterFloatLiteral, val: "1.5"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter less than or equal with left hand operand missing",
			path: "$[?(<=1)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeError, val: `missing first operand for binary operator <= at position 4, following "[?("`},
			},
		},
		{
			name: "filter less than or equal with missing right hand value",
			path: "$[?(@.child<=)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterLessThanOrEqual, val: "<="},
				{typ: lexemeError, val: `invalid filter term at position 13, following "<="`},
			},
		},
		{
			name: "filter less than or equal, string on the right",
			path: "$[?(@.child<='x')]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterLessThanOrEqual, val: "<="},
				{typ: lexemeError, val: `strings cannot be compared using <= at position 13, following "<="`},
			},
		},
		{
			name: "filter less than or equal, string on the left",
			path: "$[?('x'<=@.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterStringLiteral, val: "'x'"},
				{typ: lexemeError, val: `strings cannot be compared using <= at position 7, following "'x'"`},
			},
		},
		{
			name: "filter conjunction",
			path: "$[?(@.child&&@.other)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterAnd, val: "&&"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".other"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter conjunction with literals and whitespace",
			path: "$[?(@.child == 'x' && -9 == @.other)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeFilterStringLiteral, val: "'x'"},
				{typ: lexemeFilterAnd, val: "&&"},
				{typ: lexemeFilterIntegerLiteral, val: "-9"},
				{typ: lexemeFilterEquality, val: "=="},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".other"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter conjunction with bracket children",
			path: "$[?(@['child'][*]&&@['other'])]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeBracketChild, val: "['child']"},
				{typ: lexemeArraySubscript, val: "[*]"},
				{typ: lexemeFilterAnd, val: "&&"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeBracketChild, val: "['other']"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter invalid leading conjunction",
			path: "$[?(&&",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeError, val: `missing first operand for binary operator && at position 4, following "[?("`},
			},
		},
		{
			name: "filter conjunction with extra whitespace",
			path: "$[?(@.child && @.other)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterAnd, val: "&&"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".other"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter disjunction",
			path: "$[?(@.child||@.other)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterOr, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".other"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter invalid leading disjunction",
			path: "$[?(||",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeError, val: `missing first operand for binary operator || at position 4, following "[?("`},
			},
		},
		{
			name: "filter disjunction with extra whitespace",
			path: "$[?(@.child || @.other)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterOr, val: "||"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".other"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "simple filter of child",
			path: "$.child[?(@.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter with missing end",
			path: "$[?(@.child",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeError, val: `missing end of filter at position 11, following ".child"`},
			},
		},
		{
			name: "nested filter (edge case)",
			path: "$[?(@.y[?(@.z)])]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".y"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".z"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter negation",
			path: "$[?(!@.child)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterNot, val: "!"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter negation of comparison (edge case)",
			path: "$[?(!@.child>1)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterNot, val: "!"},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterGreaterThan, val: ">"},
				{typ: lexemeFilterIntegerLiteral, val: "1"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter negation of bracket",
			path: "$[?(!(@.child))]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterNot, val: "!"},
				{typ: lexemeFilterOpenBracket, val: "("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterCloseBracket, val: ")"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter regular expression",
			path: "$[?(@.child=~/.*/)]",
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterMatchesRegularExpression, val: "=~"},
				{typ: lexemeFilterRegularExpressionLiteral, val: "/.*/"},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter regular expression with escaped /",
			path: `$[?(@.child=~/\/.*/)]`,
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterMatchesRegularExpression, val: "=~"},
				{typ: lexemeFilterRegularExpressionLiteral, val: `/\/.*/`},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: `filter regular expression with escaped \`,
			path: `$[?(@.child=~/\\/)]`,
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterMatchesRegularExpression, val: "=~"},
				{typ: lexemeFilterRegularExpressionLiteral, val: `/\\/`},
				{typ: lexemeFilterEnd, val: ")]"},
				{typ: lexemeIdentity, val: ""},
			},
		},
		{
			name: "filter regular expression with missing leading /",
			path: `$[?(@.child=~.*/)]`,
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterMatchesRegularExpression, val: "=~"},
				{typ: lexemeError, val: `regular expression does not start with / at position 13, following "=~"`},
			},
		},
		{
			name: "filter regular expression with missing trailing /",
			path: `$[?(@.child=~/.*)]`,
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterMatchesRegularExpression, val: "=~"},
				{typ: lexemeError, val: `unmatched regular expression delimiter / at position 13, following "=~"`},
			},
		},
		{
			name: "filter regular expression to match string literal",
			path: `$[?('x'=~/.*/)]`,
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterStringLiteral, val: "'x'"},
				{typ: lexemeError, val: `literal cannot be matched using =~ at position 7, following "'x'"`},
			},
		},
		{
			name: "filter regular expression to match integer literal",
			path: `$[?(0=~/.*/)]`,
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterIntegerLiteral, val: "0"},
				{typ: lexemeError, val: `literal cannot be matched using =~ at position 5, following "0"`},
			},
		},
		{
			name: "filter regular expression to match float literal",
			path: `$[?(.1=~/.*/)]`,
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterFloatLiteral, val: ".1"},
				{typ: lexemeError, val: `literal cannot be matched using =~ at position 6, following ".1"`},
			},
		},
		{
			name: "filter invalid regular expression",
			path: `$[?(@.child=~/(.*/)]`,
			expected: []lexeme{
				{typ: lexemeRoot, val: "$"},
				{typ: lexemeFilterBegin, val: "[?("},
				{typ: lexemeFilterAt, val: "@"},
				{typ: lexemeDotChild, val: ".child"},
				{typ: lexemeFilterMatchesRegularExpression, val: "=~"},
				{typ: lexemeError, val: "invalid regular expression at position 13, following \"=~\": error parsing regexp: missing closing ): `(.*`"},
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

	if focussed {
		t.Fatalf("testcase(s) still focussed")
	}
}

func TestLexemeTypeComparators(t *testing.T) {
	cases := []struct {
		name        string
		lexemeType  lexemeType
		comparisons map[comparison]bool // can't compare functions, so need to test the function behaviour
		expectPanic bool
		focus       bool // if true, run only tests with focus set to true
	}{
		{
			name:       "equal",
			lexemeType: lexemeFilterEquality,
			comparisons: map[comparison]bool{
				compareLessThan:     false,
				compareEqual:        true,
				compareGreaterThan:  false,
				compareIncomparable: false,
			},
		},
		{
			name:       "notEqual",
			lexemeType: lexemeFilterInequality,
			comparisons: map[comparison]bool{
				compareLessThan:     true,
				compareEqual:        false,
				compareGreaterThan:  true,
				compareIncomparable: true,
			},
		},
		{
			name:       "greaterThan",
			lexemeType: lexemeFilterGreaterThan,
			comparisons: map[comparison]bool{
				compareLessThan:     false,
				compareEqual:        false,
				compareGreaterThan:  true,
				compareIncomparable: false,
			},
		},
		{
			name:       "greaterThanOrEqual",
			lexemeType: lexemeFilterGreaterThanOrEqual,
			comparisons: map[comparison]bool{
				compareLessThan:     false,
				compareEqual:        true,
				compareGreaterThan:  true,
				compareIncomparable: false,
			},
		},
		{
			name:       "lessThan",
			lexemeType: lexemeFilterLessThan,
			comparisons: map[comparison]bool{
				compareLessThan:     true,
				compareEqual:        false,
				compareGreaterThan:  false,
				compareIncomparable: false,
			},
		},
		{
			name:       "lessThanOrEqual",
			lexemeType: lexemeFilterLessThanOrEqual,
			comparisons: map[comparison]bool{
				compareLessThan:     true,
				compareEqual:        true,
				compareGreaterThan:  false,
				compareIncomparable: false,
			},
		},
		{
			name:       "non-comparator lexeme",
			lexemeType: lexemeEOF,
			comparisons: map[comparison]bool{
				compareLessThan:     false,
				compareEqual:        false,
				compareGreaterThan:  false,
				compareIncomparable: false,
			},
			expectPanic: true,
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
			panicked := func() (panicked bool) {
				defer func() {
					if recover() != nil {
						panicked = true
					}
				}()
				for comparison, result := range tc.comparisons {
					require.Equal(t, result, tc.lexemeType.comparator()(comparison), "%v", comparison)
				}
				return false
			}()
			require.Equal(t, tc.expectPanic, panicked, "panic")
		})
	}

	if focussed {
		t.Fatalf("testcase(s) still focussed")
	}
}
