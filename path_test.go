/*
 * Copyright 2020 Go YAML Path Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath_test

import (
	"bytes"
	"testing"

	"github.com/glyn/go-yamlpath"
	"github.com/stretchr/testify/require"
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
`
	var n yaml.Node

	err := yaml.Unmarshal([]byte(y), &n)
	require.NoError(t, err)

	cases := []struct {
		name            string
		path            string
		expectedStrings []string
		expectedPathErr string
	}{
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
			name:            "dot child with no name",
			path:            "$.",
			expectedPathErr: "child name missing after .",
		},
		{
			name:            "dot child with trailing dot",
			path:            "$.store.",
			expectedPathErr: "child name missing after .",
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
			name: "dot child with embedded space",
			path: "$.store.feather duster.price",
			expectedStrings: []string{
				"9.95\n",
			},
			expectedPathErr: "",
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
			name:            "bracket child with no name",
			path:            "$['']",
			expectedPathErr: "child name missing from ['']",
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
			name: "bracket dotted child",
			path: "$['store.book']",
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
			name: "bracket child with embedded wildcard",
			path: "$['store.*.color']",
			expectedStrings: []string{
				"red\n",
			},
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
			name:            "bracket child unmatched",
			path:            "$['store",
			expectedPathErr: "unmatched ['",
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
				"color: red\nprice: 19.95\n",
				"color\n",
				"red\n",
				"price\n",
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
			expectedPathErr: "child name missing after ..",
		},
		{
			name: "dot wildcarded children",
			path: "$.store.bicycle.*",
			expectedStrings: []string{
				"color\n",
				"red\n",
				"price\n",
				"19.95\n",
			},
			expectedPathErr: "",
		},
		{
			name: "bracketed wildcarded children",
			path: "$['store.bicycle.*']",
			expectedStrings: []string{
				"color\n",
				"red\n",
				"price\n",
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
`},
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
`},
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
`},
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
`},
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
`},
			expectedPathErr: "",
		}, {
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
`},
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
`},
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
`},
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
`},
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
`},
			expectedPathErr: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := yamlpath.NewPath(tc.path)
			if tc.expectedPathErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedPathErr)
				return
			}

			actual := p.Find(&n)

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
}
