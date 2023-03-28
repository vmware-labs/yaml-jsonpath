/*
 * Copyright 2020 VMware, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

func TestFind(t *testing.T) {
	y := `---
store:
  book:
  - category: reference
    author: Nigel Rees
    title: Sayings of the Century
    price: 8.95
  - category: fiction
    author: Evelyn Waugh
    title: Sword of Honour
    price: 12.99
  - category: fiction
    author: Herman Melville
    title: Moby Dick
    isbn: 0-553-21311-3
    price: 8.99
  - category: fiction
    author: J. R. R. Tolkien
    title: The Lord of the Rings
    isbn: 0-395-19395-8
    price: 22.99
  bicycle:
    color: red
    price: 19.95
  feather duster:
    price: 9.95
x:
  - y:
    - z: 1
      w: 2
  - y:
    - z: 3
      w: 4
test~: hello world
test: this is a test
`
	var n yaml.Node

	err := yaml.Unmarshal([]byte(y), &n)
	require.NoError(t, err)

	cases := []struct {
		name            string
		path            string
		expectedStrings []string
		expectedPathErr string
		focus           bool // if true, run only tests with focus set to true
	}{
		{
			name: "property names",
			path: "$.store~",
			expectedStrings: []string{
				`store
`,
			},
			expectedPathErr: "",
		},
		{
			name: "property names bracket child",
			path: "$.store['book']~",
			expectedStrings: []string{
				`book
`,
			},
			expectedPathErr: "",
		},
		{
			name: "property names bracket children",
			path: "$.store.book[0]['category','author']~",
			expectedStrings: []string{
				`category
`,
				`author
`,
			},
			expectedPathErr: "",
		},
		{
			name: "property names arraysubscript",
			path: "$.store.book[0][*]~",
			expectedStrings: []string{
				`category
`,
				`author
`,
				`title
`,
				`price
`,
			},
			expectedPathErr: "",
		},
		{
			name: "property names bracket child with ~ in name",
			path: "$['test~']~",
			expectedStrings: []string{
				`test~
`,
			},
			expectedPathErr: "",
		},
		{
			name: "dotted child with ~ in name",
			path: "$.test~",
			expectedStrings: []string{
				`test
`,
			},
			expectedPathErr: "",
		},
		{
			name: "identity",
			path: "",
			expectedStrings: []string{`store:
  book:
  - category: reference
    author: Nigel Rees
    title: Sayings of the Century
    price: 8.95
  - category: fiction
    author: Evelyn Waugh
    title: Sword of Honour
    price: 12.99
  - category: fiction
    author: Herman Melville
    title: Moby Dick
    isbn: 0-553-21311-3
    price: 8.99
  - category: fiction
    author: J. R. R. Tolkien
    title: The Lord of the Rings
    isbn: 0-395-19395-8
    price: 22.99
  bicycle:
    color: red
    price: 19.95
  feather duster:
    price: 9.95
x:
- y:
  - z: 1
    w: 2
- y:
  - z: 3
    w: 4
test~: hello world
test: this is a test
`},
			expectedPathErr: "",
		},
		{
			name: "root",
			path: "$",
			expectedStrings: []string{`store:
  book:
  - category: reference
    author: Nigel Rees
    title: Sayings of the Century
    price: 8.95
  - category: fiction
    author: Evelyn Waugh
    title: Sword of Honour
    price: 12.99
  - category: fiction
    author: Herman Melville
    title: Moby Dick
    isbn: 0-553-21311-3
    price: 8.99
  - category: fiction
    author: J. R. R. Tolkien
    title: The Lord of the Rings
    isbn: 0-395-19395-8
    price: 22.99
  bicycle:
    color: red
    price: 19.95
  feather duster:
    price: 9.95
x:
- y:
  - z: 1
    w: 2
- y:
  - z: 3
    w: 4
test~: hello world
test: this is a test
`},
			expectedPathErr: "",
		},
		{
			name: "dot child",
			path: "$.store",
			expectedStrings: []string{`book:
- category: reference
  author: Nigel Rees
  title: Sayings of the Century
  price: 8.95
- category: fiction
  author: Evelyn Waugh
  title: Sword of Honour
  price: 12.99
- category: fiction
  author: Herman Melville
  title: Moby Dick
  isbn: 0-553-21311-3
  price: 8.99
- category: fiction
  author: J. R. R. Tolkien
  title: The Lord of the Rings
  isbn: 0-395-19395-8
  price: 22.99
bicycle:
  color: red
  price: 19.95
feather duster:
  price: 9.95
`},
			expectedPathErr: "",
		},
		{
			name: "dot child with implicit root",
			path: ".store",
			expectedStrings: []string{`book:
- category: reference
  author: Nigel Rees
  title: Sayings of the Century
  price: 8.95
- category: fiction
  author: Evelyn Waugh
  title: Sword of Honour
  price: 12.99
- category: fiction
  author: Herman Melville
  title: Moby Dick
  isbn: 0-553-21311-3
  price: 8.99
- category: fiction
  author: J. R. R. Tolkien
  title: The Lord of the Rings
  isbn: 0-395-19395-8
  price: 22.99
bicycle:
  color: red
  price: 19.95
feather duster:
  price: 9.95
`},
			expectedPathErr: "",
		},
		{
			name: "undotted child with implicit root",
			path: "store",
			expectedStrings: []string{`book:
- category: reference
  author: Nigel Rees
  title: Sayings of the Century
  price: 8.95
- category: fiction
  author: Evelyn Waugh
  title: Sword of Honour
  price: 12.99
- category: fiction
  author: Herman Melville
  title: Moby Dick
  isbn: 0-553-21311-3
  price: 8.99
- category: fiction
  author: J. R. R. Tolkien
  title: The Lord of the Rings
  isbn: 0-395-19395-8
  price: 22.99
bicycle:
  color: red
  price: 19.95
feather duster:
  price: 9.95
`},
			expectedPathErr: "",
		},
		{
			name: "undotted all children with implicit root",
			path: "*",
			expectedStrings: []string{
				`book:
- category: reference
  author: Nigel Rees
  title: Sayings of the Century
  price: 8.95
- category: fiction
  author: Evelyn Waugh
  title: Sword of Honour
  price: 12.99
- category: fiction
  author: Herman Melville
  title: Moby Dick
  isbn: 0-553-21311-3
  price: 8.99
- category: fiction
  author: J. R. R. Tolkien
  title: The Lord of the Rings
  isbn: 0-395-19395-8
  price: 22.99
bicycle:
  color: red
  price: 19.95
feather duster:
  price: 9.95
`,
				`- y:
  - z: 1
    w: 2
- y:
  - z: 3
    w: 4
`,
				`hello world
`,
				`this is a test
`,
			},
			expectedPathErr: "",
		},
		{
			name:            "dot child with no name",
			path:            "$.",
			expectedPathErr: `child name missing at position 2, following "$."`,
		},
		{
			name:            "dot child with trailing dot",
			path:            "$.store.",
			expectedPathErr: `child name missing at position 8, following ".store."`,
		},
		{
			name: "dot child of dot child",
			path: "$.store.book",
			expectedStrings: []string{`- category: reference
  author: Nigel Rees
  title: Sayings of the Century
  price: 8.95
- category: fiction
  author: Evelyn Waugh
  title: Sword of Honour
  price: 12.99
- category: fiction
  author: Herman Melville
  title: Moby Dick
  isbn: 0-553-21311-3
  price: 8.99
- category: fiction
  author: J. R. R. Tolkien
  title: The Lord of the Rings
  isbn: 0-395-19395-8
  price: 22.99
`},
			expectedPathErr: "",
		},
		{
			name: "dot child with embedded wildcard",
			path: "$.store.*.color",
			expectedStrings: []string{
				"red\n",
			},
			expectedPathErr: "",
		},
		{
			name:            "dot child with embedded space",
			path:            "$.store.feather duster.price",
			expectedPathErr: `invalid character ' ' at position 15, following ".feather"`,
		},
		{
			name: "bracket child",
			path: "$['store']",
			expectedStrings: []string{`book:
- category: reference
  author: Nigel Rees
  title: Sayings of the Century
  price: 8.95
- category: fiction
  author: Evelyn Waugh
  title: Sword of Honour
  price: 12.99
- category: fiction
  author: Herman Melville
  title: Moby Dick
  isbn: 0-553-21311-3
  price: 8.99
- category: fiction
  author: J. R. R. Tolkien
  title: The Lord of the Rings
  isbn: 0-395-19395-8
  price: 22.99
bicycle:
  color: red
  price: 19.95
feather duster:
  price: 9.95
`},
			expectedPathErr: "",
		},
		{
			name: "bracket child with double quotes",
			path: `$["store"]`,
			expectedStrings: []string{`book:
- category: reference
  author: Nigel Rees
  title: Sayings of the Century
  price: 8.95
- category: fiction
  author: Evelyn Waugh
  title: Sword of Honour
  price: 12.99
- category: fiction
  author: Herman Melville
  title: Moby Dick
  isbn: 0-553-21311-3
  price: 8.99
- category: fiction
  author: J. R. R. Tolkien
  title: The Lord of the Rings
  isbn: 0-395-19395-8
  price: 22.99
bicycle:
  color: red
  price: 19.95
feather duster:
  price: 9.95
`},
			expectedPathErr: "",
		},
		{
			name: "bracket child of bracket child",
			path: "$['store']['book']",
			expectedStrings: []string{`- category: reference
  author: Nigel Rees
  title: Sayings of the Century
  price: 8.95
- category: fiction
  author: Evelyn Waugh
  title: Sword of Honour
  price: 12.99
- category: fiction
  author: Herman Melville
  title: Moby Dick
  isbn: 0-553-21311-3
  price: 8.99
- category: fiction
  author: J. R. R. Tolkien
  title: The Lord of the Rings
  isbn: 0-395-19395-8
  price: 22.99
`},
			expectedPathErr: "",
		},
		{
			name:            "bracket dotted child",
			path:            "$['store.book']",
			expectedStrings: []string{},
			expectedPathErr: "",
		},
		{
			name: "bracket child with embedded space",
			path: "$.store['feather duster'].price",
			expectedStrings: []string{
				"9.95\n",
			},
			expectedPathErr: "",
		},
		{
			name: "bracket child of dot child",
			path: "$.store['book']",
			expectedStrings: []string{`- category: reference
  author: Nigel Rees
  title: Sayings of the Century
  price: 8.95
- category: fiction
  author: Evelyn Waugh
  title: Sword of Honour
  price: 12.99
- category: fiction
  author: Herman Melville
  title: Moby Dick
  isbn: 0-553-21311-3
  price: 8.99
- category: fiction
  author: J. R. R. Tolkien
  title: The Lord of the Rings
  isbn: 0-395-19395-8
  price: 22.99
`},
			expectedPathErr: "",
		},
		{
			name: "dot child of bracket child",
			path: "$['store'].book",
			expectedStrings: []string{`- category: reference
  author: Nigel Rees
  title: Sayings of the Century
  price: 8.95
- category: fiction
  author: Evelyn Waugh
  title: Sword of Honour
  price: 12.99
- category: fiction
  author: Herman Melville
  title: Moby Dick
  isbn: 0-553-21311-3
  price: 8.99
- category: fiction
  author: J. R. R. Tolkien
  title: The Lord of the Rings
  isbn: 0-395-19395-8
  price: 22.99
`},
			expectedPathErr: "",
		},
		{
			name:            "unclosed bracket child",
			path:            "$['store",
			expectedPathErr: `unmatched "'" at position 8, following "$['store"`,
		},
		{
			name: "recursive descent",
			path: "$..price",
			expectedStrings: []string{
				"8.95\n",
				"12.99\n",
				"8.99\n",
				"22.99\n",
				"19.95\n",
				"9.95\n",
			},
			expectedPathErr: "",
		},
		{
			name: "recursive descent of dot child",
			path: "$.store.book..price",
			expectedStrings: []string{
				"8.95\n",
				"12.99\n",
				"8.99\n",
				"22.99\n",
			},
			expectedPathErr: "",
		},
		{
			name: "recursive descent of child starting with undotted implicit root",
			path: "store.book..price",
			expectedStrings: []string{
				"8.95\n",
				"12.99\n",
				"8.99\n",
				"22.99\n",
			},
			expectedPathErr: "",
		},
		{
			name: "recursive descent of bracket child",
			path: "$['store']['book']..price",
			expectedStrings: []string{
				"8.95\n",
				"12.99\n",
				"8.99\n",
				"22.99\n",
			},
			expectedPathErr: "",
		},
		{
			name: "recursive descent with wildcard",
			path: "$.store.bicycle..*",
			expectedStrings: []string{
				"red\n",
				"19.95\n",
			},
			expectedPathErr: "",
		},
		{
			name: "repeated recursive descent",
			path: "$..book..price",
			expectedStrings: []string{
				"8.95\n",
				"12.99\n",
				"8.99\n",
				"22.99\n",
			},
			expectedPathErr: "",
		},
		{
			name: "recursive descent with dot child",
			path: "$..bicycle.color",
			expectedStrings: []string{
				"red\n",
			},
			expectedPathErr: "",
		},
		{
			name: "recursive descent with bracket child",
			path: "$..bicycle['color']",
			expectedStrings: []string{
				"red\n",
			},
			expectedPathErr: "",
		},
		{
			name:            "recursive descent with missing name",
			path:            "$..",
			expectedPathErr: `child name or array access or filter missing after recursive descent at position 3, following "$.."`,
		},
		{
			name: "dot wildcarded children",
			path: "$.store.bicycle.*",
			expectedStrings: []string{
				"red\n",
				"19.95\n",
			},
			expectedPathErr: "",
		},
		{
			name: "array subscript wildcard",
			path: "$.store.book[*]",
			expectedStrings: []string{
				`category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
				`category: fiction
author: Evelyn Waugh
title: Sword of Honour
price: 12.99
`,
				`category: fiction
author: Herman Melville
title: Moby Dick
isbn: 0-553-21311-3
price: 8.99
`,
				`category: fiction
author: J. R. R. Tolkien
title: The Lord of the Rings
isbn: 0-395-19395-8
price: 22.99
`,
			},
			expectedPathErr: "",
		},
		{
			name: "array subscript single",
			path: "$.store.book[0]",
			expectedStrings: []string{
				`category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			},
			expectedPathErr: "",
		},
		{
			name: "array subscript from:to",
			path: "$.store.book[1:3]",
			expectedStrings: []string{
				`category: fiction
author: Evelyn Waugh
title: Sword of Honour
price: 12.99
`,
				`category: fiction
author: Herman Melville
title: Moby Dick
isbn: 0-553-21311-3
price: 8.99
`,
			},
			expectedPathErr: "",
		},
		{
			name: "array subscript from:to:step",
			path: "$.store.book[0:3:2]",
			expectedStrings: []string{
				`category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
				`category: fiction
author: Herman Melville
title: Moby Dick
isbn: 0-553-21311-3
price: 8.99
`,
			},
			expectedPathErr: "",
		},
		{
			name: "array subscript :to",
			path: "$.store.book[:2]",
			expectedStrings: []string{
				`category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
				`category: fiction
author: Evelyn Waugh
title: Sword of Honour
price: 12.99
`,
			},
			expectedPathErr: "",
		},
		{
			name: "array subscript ::step",
			path: "$.store.book[::2]",
			expectedStrings: []string{
				`category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
				`category: fiction
author: Herman Melville
title: Moby Dick
isbn: 0-553-21311-3
price: 8.99
`,
			},
			expectedPathErr: "",
		},
		{
			name: "array subscript from:to:",
			path: "$.store.book[1:3:]",
			expectedStrings: []string{
				`category: fiction
author: Evelyn Waugh
title: Sword of Honour
price: 12.99
`,
				`category: fiction
author: Herman Melville
title: Moby Dick
isbn: 0-553-21311-3
price: 8.99
`,
			},
			expectedPathErr: "",
		},
		{
			name: "array subscript ::",
			path: "$.store.book[::]",
			expectedStrings: []string{
				`category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
				`category: fiction
author: Evelyn Waugh
title: Sword of Honour
price: 12.99
`,
				`category: fiction
author: Herman Melville
title: Moby Dick
isbn: 0-553-21311-3
price: 8.99
`,
				`category: fiction
author: J. R. R. Tolkien
title: The Lord of the Rings
isbn: 0-395-19395-8
price: 22.99
`,
			},
			expectedPathErr: "",
		},
		{
			name: "array subscript ::-1",
			path: "$.store.book[::-1]",
			expectedStrings: []string{
				`category: fiction
author: J. R. R. Tolkien
title: The Lord of the Rings
isbn: 0-395-19395-8
price: 22.99
`,
				`category: fiction
author: Herman Melville
title: Moby Dick
isbn: 0-553-21311-3
price: 8.99
`,
				`category: fiction
author: Evelyn Waugh
title: Sword of Honour
price: 12.99
`,
				`category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			},
			expectedPathErr: "",
		},
		{
			name: "array subscript -3:-1",
			path: "$.store.book[-3:-1]",
			expectedStrings: []string{
				`category: fiction
author: Evelyn Waugh
title: Sword of Honour
price: 12.99
`,
				`category: fiction
author: Herman Melville
title: Moby Dick
isbn: 0-553-21311-3
price: 8.99
`,
			},
			expectedPathErr: "",
		},
		{
			name: "array subscript -1:",
			path: "$.store.book[-1:]",
			expectedStrings: []string{
				`category: fiction
author: J. R. R. Tolkien
title: The Lord of the Rings
isbn: 0-395-19395-8
price: 22.99
`,
			},
			expectedPathErr: "",
		},
		{
			name:            "missing array subscript",
			path:            "$.store.book[]",
			expectedStrings: []string{},
			expectedPathErr: "subscript missing from [] before position 14",
		},
		{
			name:            "malformed array subscript",
			path:            "$.store.book[::0]",
			expectedStrings: []string{},
			expectedPathErr: "invalid array index [::0] before position 17: array index step value must be non-zero",
		},
		{
			name:            "array subscript out of bounds",
			path:            "$.store.book[99]",
			expectedStrings: []string{},
			expectedPathErr: "",
		},
		{
			name: "filter >",
			path: "$.store.book[?(@.price > 8.98)]",
			expectedStrings: []string{
				`category: fiction
author: Evelyn Waugh
title: Sword of Honour
price: 12.99
`,
				`category: fiction
author: Herman Melville
title: Moby Dick
isbn: 0-553-21311-3
price: 8.99
`,
				`category: fiction
author: J. R. R. Tolkien
title: The Lord of the Rings
isbn: 0-395-19395-8
price: 22.99
`,
			},
			expectedPathErr: "",
		},
		{
			name: "filter ==",
			path: "$.store.book[?(@.category == 'reference')]",
			expectedStrings: []string{
				`category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			},
			expectedPathErr: "",
		},
		{
			name: "filter == with bracket child",
			path: "$.store.book[?(@.category == 'reference')]",
			expectedStrings: []string{
				`category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			},
			expectedPathErr: "",
		},
		{
			name: "filter !=",
			path: "$.store.book[?(@.category != 'fiction')]",
			expectedStrings: []string{
				`category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
			},
			expectedPathErr: "",
		},
		{
			name: "filter involving root",
			path: "$.store.book[?(@.price > $.store.bicycle.price)]",
			expectedStrings: []string{`category: fiction
author: J. R. R. Tolkien
title: The Lord of the Rings
isbn: 0-395-19395-8
price: 22.99
`},
			expectedPathErr: "",
		},
		{
			name: "nested filter (edge case)",
			path: "$.x[?(@.y[?(@.z==1)].w==2)]",
			expectedStrings: []string{
				`y:
- z: 1
  w: 2
`,
			},
			expectedPathErr: "",
		},
		{
			name: "negated filter",
			path: "$.store.book[?(!@.isbn)]",
			expectedStrings: []string{
				`category: reference
author: Nigel Rees
title: Sayings of the Century
price: 8.95
`,
				`category: fiction
author: Evelyn Waugh
title: Sword of Honour
price: 12.99
`,
			},
			expectedPathErr: "",
		},
		{
			name: "map filter",
			path: `$.store.bicycle[?(@.color == "red")]`,
			expectedStrings: []string{
				`color: red
price: 19.95
`,
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
			p, err := yamlpath.NewPath(tc.path)
			if err != nil {
				require.Nil(t, p)
			}
			if tc.expectedPathErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedPathErr)
				return
			}

			actual, err := p.Find(&n)
			require.NoError(t, err)

			actualStrings := []string{}
			for _, a := range actual {
				var buf bytes.Buffer
				e := yaml.NewEncoder(&buf)
				e.SetIndent(2)

				err = e.Encode(a)
				require.NoError(t, err)
				e.Close()
				actualStrings = append(actualStrings, buf.String())
			}

			require.Equal(t, tc.expectedStrings, actualStrings)
		})
	}

	if focussed {
		t.Fatalf("testcase(s) still focussed")
	}
}

func TestFindOtherDocuments(t *testing.T) {
	cases := []struct {
		name            string
		input           string
		path            string
		expectedStrings []string
		expectedPathErr string
		focus           bool // if true, run only tests with focus set to true
	}{
		{
			name:            "empty document",
			expectedStrings: []string{},
		},
		{
			name: "document with values matching keys",
			input: `c: a
a: b`,
			path:            ".a",
			expectedStrings: []string{"b\n"},
		},
		{
			name: "document with top-level array",
			input: `- c: a
- a: b`,
			path:            "$[0]",
			expectedStrings: []string{"c: a\n"},
		},
		{
			name: "document with top-level array, .*",
			input: `- c: a
- a: b`,
			path:            "$.*",
			expectedStrings: []string{"c: a\n", "a: b\n"},
		},
		{
			name: "document with top-level array, filter with double-quoted string literal",
			input: `- c: a
- a: b`,
			path:            `$[?(@.c=="a")]`,
			expectedStrings: []string{"c: a\n"},
		},
		{
			name: "union with keys",
			input: `key: value
another: entry`,
			path:            `$['key','another']`,
			expectedStrings: []string{"value\n", "entry\n"},
		},
		{
			name: "bracket child with quoted union literal",
			input: `",": value
another: entry`,
			path:            `$[',']`,
			expectedStrings: []string{"value\n"},
		},
		{
			name:            "array access after recursive descent",
			input:           `{"k": [{"key": "some value"}, {"key": 42}], "kk": [[{"key": 100}, {"key": 200}, {"key": 300}], [{"key": 400}, {"key": 500}, {"key": 600}]], "key": [0, 1]}`,
			path:            `$..[1].key`,
			expectedStrings: []string{"42\n", "200\n", "500\n"},
		},
		{
			name:            "filter after recursive descent",
			input:           `{"k": [{"key": "some value"}, {"key": 42}], "kk": [[{"key": 100}, {"key": 200}, {"key": 300}], [{"key": 400}, {"key": 500}, {"key": 600}]], "key": [0, 1]}`,
			path:            `$..[?(@.key>=500)]`,
			expectedStrings: []string{"{\"key\": 500}\n", "{\"key\": 600}\n"},
		},
		{
			name:            "union with wildcard and numbers (deviation from comparison project consensus)",
			input:           `["a","b","c"]`,
			path:            `$[*,1,0,*]`,
			expectedPathErr: `invalid array index [*,1,0,*] before position 10: error in union member 0: wildcard cannot be used in union`,
		},
		{
			name:            "special characters in bracket child name",
			input:           `{":@.\"$,*'\\": 42}`,
			path:            `$[':@."$,*\'\\']`,
			expectedStrings: []string{"42\n"},
		},
		{
			name:  "filter with boolean value comparison",
			input: `[{"a":true, "b": 1}, {"a":"true", "b": 2}]`,
			path:  `$[?(@.a==true)]`,
			expectedStrings: []string{`{"a": true, "b": 1}
`},
		},
		{
			name:  "filter with null value comparison",
			input: `[{"a":null, "b": 1}, {"a":"null", "b": 2}]`,
			path:  `$[?(@.a==null)]`,
			expectedStrings: []string{`{"a": null, "b": 1}
`},
		},
		{
			name:  "filter with integer that appears in string",
			input: `[{"a": "42", "b": 1}]`,
			path:  `$[?(@.a!=42)]`,
			expectedStrings: []string{`{"a": "42", "b": 1}
`},
		},
		{
			name:            "filter involving value of current node",
			input:           `[0,42,100]`,
			path:            `$[?(@>=42)]`,
			expectedStrings: []string{"42\n", "100\n"},
		},
		{
			name:            "filter with fractional float",
			input:           `[0,-4.2,100]`,
			path:            `$[?(@==-42E-1)]`,
			expectedStrings: []string{"-4.2\n"},
		},
		{
			name:            "filter with boolean predicate",
			input:           `[0]`,
			path:            `$[?(true)]`,
			expectedStrings: []string{"0\n"},
		},
		{
			name:            "relaxed spelling of true, false, and null literals", // See https://yaml.org/spec/1.2/spec.html#id2805071
			input:           `[FALSE, False, false, fAlse, TRUE, True, true, tRue, NULL, Null, null, nUll]`,
			path:            `$[?(@==false || @==true || @==null)]`,
			expectedStrings: []string{"FALSE\n", "False\n", "false\n", "TRUE\n", "True\n", "true\n", "NULL\n", "Null\n", "null\n"},
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
			var n yaml.Node
			err := yaml.Unmarshal([]byte(tc.input), &n)
			require.NoError(t, err)

			p, err := yamlpath.NewPath(tc.path)
			if tc.expectedPathErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedPathErr)
				return
			}

			actual, err := p.Find(&n)
			require.NoError(t, err)

			actualStrings := []string{}
			for _, a := range actual {
				var buf bytes.Buffer
				e := yaml.NewEncoder(&buf)
				e.SetIndent(2)

				err = e.Encode(a)
				require.NoError(t, err)
				e.Close()
				actualStrings = append(actualStrings, buf.String())
			}

			require.Equal(t, tc.expectedStrings, actualStrings)
		})
	}

	if focussed {
		t.Fatalf("testcase(s) still focussed")
	}
}
