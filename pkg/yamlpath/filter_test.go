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
	"gopkg.in/yaml.v3"
)

func TestNewFilter(t *testing.T) {
	cases := []struct {
		name      string
		filter    string
		parseTree *filterNode
		yamlDoc   string
		match     bool
		focus     bool // if true, run only tests with focus set to true
	}{
		{
			name:      "no lexemes",
			filter:    "",
			parseTree: nil,
			yamlDoc:   "",
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
		// TODO: parentheses
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
			var n yaml.Node
			err := yaml.Unmarshal([]byte(tc.yamlDoc), &n)
			require.NoError(t, err)

			parseTree := parseFilterString(tc.filter)
			match := newFilter(parseTree)(&n)
			require.Equal(t, tc.match, match)
		})
	}

	if focussed {
		t.Fatalf("testcase(s) still focussed")
	}
}

func parseFilterString(filter string) *filterNode {
	path := fmt.Sprintf("$[?(%s)]", filter)
	lexer := lex("Path lexer", path)

	lexemes := []lexeme{}
	for {
		lexeme := lexer.nextLexeme()
		if lexeme.typ == lexemeEOF {
			break
		}
		lexemes = append(lexemes, lexeme)
	}

	return newFilterNode(lexemes[2 : len(lexemes)-2])
}
