/*
 * Copyright 2020 VMware, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewFilter(t *testing.T) {
	cases := []struct {
		name      string
		filter    string
		parseTree *filterNode
		yamlDoc   string
		rootDoc   string
		match     bool
		focus     bool // if true, run only tests with focus set to true
	}{
		{
			name:      "no lexemes",
			filter:    "",
			parseTree: nil,
			yamlDoc:   "",
			rootDoc:   "",
			match:     false,
		},
		{
			name:   "existence filter, match",
			filter: "@.category",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: true,
		},
		{
			name:   "existence filter, no match",
			filter: "@.nosuch",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: false,
		},
		{
			name:   "numeric comparison filter, match",
			filter: "@.price>8.90",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: true,
		},
		{
			name:   "numeric comparison filter, no match",
			filter: "@.price>9",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: false,
		},
		{
			name:   "numeric comparison filter, match",
			filter: "@.price>=8.95",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: true,
		},
		{
			name:   "numeric comparison filter, no match",
			filter: "@.price>=9",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: false,
		},
		{
			name:   "numeric comparison filter, match",
			filter: "@.price<8.96",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: true,
		},
		{
			name:   "numeric comparison filter, no match",
			filter: "@.price<8",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: false,
		},
		{
			name:   "numeric comparison filter, match",
			filter: "@.price<=8.95",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: true,
		},
		{
			name:   "numeric comparison filter, no match",
			filter: "@.price<=8",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: false,
		},
		{
			name:   "numeric comparison filter, match",
			filter: "8.90<@.price",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: true,
		},
		{
			name:   "numeric comparison filter, no match",
			filter: "9<@.price",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: false,
		},
		{
			name:   "numeric comparison filter, match",
			filter: "8.95<=@.price",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: true,
		},
		{
			name:   "numeric comparison filter, no match",
			filter: "9<=@.price",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: false,
		},
		{
			name:   "numeric comparison filter, match",
			filter: "8.96>@.price",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: true,
		},
		{
			name:   "numeric comparison filter, no match",
			filter: "8>@.price",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: false,
		},
		{
			name:   "numeric comparison filter, match",
			filter: "8.95>=@.price",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: true,
		},
		{
			name:   "numeric comparison filter, no match",
			filter: "8>=@.price",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: false,
		},
		{
			name:   "numeric comparison filter, path to path, match",
			filter: "@.x<@.y",
			yamlDoc: `---
x: 1
y: 2
`,
			match: true,
		},
		{
			// When a filter path does not match, it produces an empty set of nodes.
			// Comparison against an empty set does not match even it matches every element of the set.
			name:   "numeric comparison filter, not found path to literal, no match",
			filter: "@.x>=9",
			yamlDoc: `---
category: reference
`,
			match: false,
		},
		{
			// When a filter path does not match, it produces an empty set of nodes.
			// Comparison against an empty set does not match even it matches every element of the set.
			name:   "numeric comparison filter, literal to not found path, no match",
			filter: "1<@.x",
			yamlDoc: `---
category: reference
`,
			match: false,
		},
		{
			// When a filter path does not match, it produces an empty set of nodes.
			// Comparison against an empty set does not match even it matches every element of the set.
			name:   "numeric comparison filter, path to not found path, match",
			filter: "@.x<@.y",
			yamlDoc: `---
x: 1
`,
			match: false,
		},
		{
			name:   "numeric comparison filter, path to path, match",
			filter: "@.x<@.y && @.y==@.z && @.y==@.w",
			yamlDoc: `---
x: 1.1
y: 2
z: 2.0
w: 02
`,
			match: true,
		},
		{
			name:   "numeric comparison filter, path to path, no match",
			filter: "@.x>@.y",
			yamlDoc: `---
x: 1
y: 2
`,
			match: false,
		},
		{
			name:    "numeric comparison filter, literal to literal, match",
			filter:  "8>=7",
			yamlDoc: "",
			match:   true,
		},
		{
			name:    "numeric comparison filter, literal to literal, no match",
			filter:  "8<7",
			yamlDoc: "",
			match:   false,
		},
		{
			name:   "numeric comparison filter, multiple, match",
			filter: "@.price[*]>8.90",
			yamlDoc: `---
price: [9,9.5]
`,
			match: true,
		},
		{
			name:   "numeric comparison filter, multiple, no match",
			filter: "@.price[*]>8.90",
			yamlDoc: `---
price: [8,9,9.5]
`,
			match: false,
		},
		{
			name:   "numeric comparison filter, path to path, single to multiple, match",
			filter: "@.x<@.y[*]",
			yamlDoc: `---
x: 1
y: [2,3]
`,
			match: true,
		},
		{
			name:   "numeric comparison filter, path to path, single to empty set, match",
			filter: "@.x<@.y[*]",
			yamlDoc: `---
x: 1
y: []
`,
			match: false,
		},
		{
			name:   "numeric comparison filter, path to path, single to multiple, no match",
			filter: "@.x<@.y[*]",
			yamlDoc: `---
x: 4
y: [2,3]
`,
			match: false,
		},
		{
			name:   "numeric comparison filter, path to path, multiple to multiple, match",
			filter: "@.x[*]<@.y[*]",
			yamlDoc: `---
x: [0,1]
y: [2,3]
`,
			match: true,
		},
		{
			name:   "numeric comparison filter, path to path, multiple to multiple, no match",
			filter: "@.x[*]<@.y[*]",
			yamlDoc: `---
x: [0,2]
y: [2,3]
`,
			match: false,
		},
		{
			name:   "numeric comparison filter, path to invalid path, no match",
			filter: "@.x<@.y", // panics
			yamlDoc: `---
x: 4
y: [2,3]
`,
			match: false,
		},
		{
			name:   "numeric comparison filter, literal to invalid path, no match",
			filter: "1<@.y", // panics
			yamlDoc: `---
y: [2,3]
`,
			match: false,
		},
		{
			name:   "numeric comparison filter, invalid path to literal, no match",
			filter: "@.y>1", // panics
			yamlDoc: `---
y: [2,3]
`,
			match: false,
		},
		{
			// this testcase relies on an artifice of the test framework to test an edge case
			// which would normally not be reached because the lexer returns an error
			name:   "numeric comparison filter, integer to string, no match",
			filter: "1>'x'", // produces filter parse tree with nil child
			yamlDoc: `---
y: [2,3]
`,
			match: false,
		},
		{
			name:   "string comparison filter, path to path, match",
			filter: "@.x==@.y && @x==@z",
			yamlDoc: `---
x: 'a'
y: "a"
z: a
`,
			match: true,
		},
		{
			name:   "string comparison filter, path to path, no match",
			filter: "@.x==@.y",
			yamlDoc: `---
x: a
y: b
`,
			match: false,
		},
		{
			name:    "comparison filter, string literal to numeric literal, no match",
			filter:  "'x'==7",
			yamlDoc: "",
			match:   false,
		},
		{
			name:    "comparison filter, numeric literal to string literal, no match",
			filter:  "7=='x'",
			yamlDoc: "",
			match:   false,
		},
		{
			name:   "existence || existence filter",
			filter: "@.a || @.b",
			yamlDoc: `---
a: x
`,
			match: true,
		},
		{
			name:   "existence || existence filter",
			filter: "@.a || @.b",
			yamlDoc: `---
b: x
`,
			match: true,
		},
		{
			name:   "existence || existence filter",
			filter: "@.a || @.b",
			yamlDoc: `---
c: x
`,
			match: false,
		},
		{
			name:   "comparison || existence filter",
			filter: "@.a>1 || @.b",
			yamlDoc: `---
a: 0
`,
			match: false,
		},
		{
			name:   "comparison || existence filter",
			filter: "@.a>1 || @.b",
			yamlDoc: `---
a: 2
`,
			match: true,
		},
		{
			name:   "comparison || existence filter",
			filter: "@.a>1 || @.b",
			yamlDoc: `---
b: x
`,
			match: true,
		},
		{
			name:   "existence || existence && existence filter",
			filter: "@.a || @.b && @.c",
			yamlDoc: `---
a: x
`,
			match: true,
		},
		{
			name:   "existence || existence && existence filter",
			filter: "@.a || @.b && @.c",
			yamlDoc: `---
b: x
`,
			match: false,
		},
		{
			name:   "existence || existence && existence filter",
			filter: "@.a || @.b && @.c",
			yamlDoc: `---
c: x
`,
			match: false,
		},
		{
			name:   "existence || existence && existence filter",
			filter: "@.a || @.b && @.c",
			yamlDoc: `---
b: x
c: x
`,
			match: true,
		},
		{
			// test just a single case of parentheses as these do not end up in the parse tree
			name:   "(existence || existence) && existence filter",
			filter: "(@.a || @.b) && @.c",
			yamlDoc: `---
a: x
`,
			match: false,
		},
		{
			name:   "nested filter (edge case), match",
			filter: "@.y[?(@.z==1)].w==2",
			yamlDoc: `---
y:
- z: 1
  w: 2	
`,
			match: true,
		},
		{
			name:   "nested filter (edge case), no match",
			filter: "@.y[?(@.z==5)].w==2",
			yamlDoc: `---
y:
- z: 1
  w: 2
`,
			match: false,
		},
		{
			name:   "nested filter (edge case), no match",
			filter: "@.y[?(@.z==1)].w==4",
			yamlDoc: `---
y:
- z: 1
  w: 2
`,
			match: false,
		},
		{
			name:   "filter involving root on right, match",
			filter: "@.price==$.price",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			rootDoc: `---
price: 8.95
`,
			match: true,
		},
		{
			name:   "filter involving root on left, match",
			filter: "$.price==@.price",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			rootDoc: `---
price: 8.95
`,
			match: true,
		},
		{
			name:   "negated existence filter, no match",
			filter: "!@.category",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: false,
		},
		{
			name:   "negated existence filter, match",
			filter: "!@.nosuch",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: true,
		},
		{
			name:   "negated parentheses",
			filter: "!(@.a) && @.c",
			yamlDoc: `---
c: x
`,
			match: true,
		},
		{
			name:   "regular expression filter at path, match",
			filter: "@.category=~/ref.*ce/",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: true,
		},
		{
			name:   "regular expression filter at path, no match",
			filter: "@.category=~/.*x/",
			yamlDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: false,
		},
		{
			name:   "regular expression filter root path, match",
			filter: "$.category=~/ref.*ce/",
			rootDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: true,
		},
		{
			name:   "regular expression filter root path, no match",
			filter: "$.category=~/.*x/",
			rootDoc: `---
category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			match: false,
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
			n := unmarshalDoc(t, tc.yamlDoc)
			root := unmarshalDoc(t, tc.rootDoc)

			parseTree := parseFilterString(tc.filter)
			match := newFilter(parseTree)(n, root)
			require.Equal(t, tc.match, match)
		})
	}

	if focussed {
		t.Fatalf("testcase(s) still focussed")
	}
}

func unmarshalDoc(t *testing.T, doc string) *yaml.Node {
	var n yaml.Node
	err := yaml.Unmarshal([]byte(doc), &n)
	require.NoError(t, err)
	return &n
}

func parseFilterString(filter string) *filterNode {
	path := fmt.Sprintf("$[?(%s)]", filter)
	lexer := lex("Path lexer", path)

	lexemes := []lexeme{}
	for {
		lexeme := lexer.nextLexeme()
		if lexeme.typ == lexemeError {
			return newFilterNode(lexemes[2:])
		}
		if lexeme.typ == lexemeEOF {
			break
		}
		lexemes = append(lexemes, lexeme)
	}

	return newFilterNode(lexemes[2 : len(lexemes)-2])
}
