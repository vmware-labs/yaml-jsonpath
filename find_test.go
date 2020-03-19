package yamlpath_test

import (
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
`
	var n yaml.Node

	err := yaml.Unmarshal([]byte(y), &n)
	require.NoError(t, err)

	cases := []struct {
		name            string
		path            string
		expected        []*yaml.Node
		expectedPathErr string
	}{
		{
			name:            "identity",
			path:            "",
			expected:        []*yaml.Node{&n},
			expectedPathErr: "",
		},
		{
			name:            "root",
			path:            "$",
			expected:        []*yaml.Node{n.Content[0]},
			expectedPathErr: "",
		},
		{
			name:            "dot child",
			path:            "$.store",
			expected:        []*yaml.Node{n.Content[0].Content[1]},
			expectedPathErr: "",
		},
		{
			name:            "dot child of child",
			path:            "$.store.book",
			expected:        []*yaml.Node{n.Content[0].Content[1].Content[1]},
			expectedPathErr: "",
		},
		{
			name:            "bracket child",
			path:            "$['store']",
			expected:        []*yaml.Node{n.Content[0].Content[1]},
			expectedPathErr: "",
		},
		{
			name:            "bracket child of child",
			path:            "$['store']['book']",
			expected:        []*yaml.Node{n.Content[0].Content[1].Content[1]},
			expectedPathErr: "",
		},
		{
			name:            "bracket child unmatched",
			path:            "$['store",
			expected:        []*yaml.Node{n.Content[0].Content[1]},
			expectedPathErr: "unmatched ['",
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

			actual, err := yamlpath.Find(&n, p)
			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}
