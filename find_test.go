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
		name        string
		path        string
		expected    []*yaml.Node
		expectedErr string
	}{
		{
			name:        "identity",
			path:        "",
			expected:    []*yaml.Node{&n},
			expectedErr: "",
		},
		{
			name:        "root",
			path:        "$",
			expected:    []*yaml.Node{n.Content[0]},
			expectedErr: "",
		},
		{
			name:        "dot child",
			path:        "$.store",
			expected:    []*yaml.Node{n.Content[0].Content[1]},
			expectedErr: "",
		},
		{
			name:        "dot child of child",
			path:        "$.store.book",
			expected:    []*yaml.Node{n.Content[0].Content[1].Content[1]},
			expectedErr: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := yamlpath.NewPath(tc.path)
			require.NoError(t, err)

			actual, err := yamlpath.Find(&n, p)
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
			require.Equal(t, tc.expected, actual)
		})
	}
}
