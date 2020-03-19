package yamlpath

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

// Path is a compiled YAML path expression.
type Path func(*yaml.Node) ([]*yaml.Node, error)

func identity(node *yaml.Node) ([]*yaml.Node, error) {
	return []*yaml.Node{node}, nil
}

func bad(*yaml.Node) ([]*yaml.Node, error) {
	return []*yaml.Node{}, errors.New("invalid path")
}

// NewPath constructs a Path from a string expression.
func NewPath(path string) (Path, error) {
	if path == "" {
		return identity, nil
	}

	if strings.HasPrefix(path, "$") {
		tail := strings.SplitN(path, "$", 2)
		t, err := NewPath(tail[1])
		if err != nil {
			return bad, err
		}
		return func(node *yaml.Node) ([]*yaml.Node, error) {
			if node.Kind != yaml.DocumentNode {
				return []*yaml.Node{}, errors.New("not a document node, so no root object")
			}
			return t(node.Content[0])
		}, nil
	}

	if strings.HasPrefix(path, ".") {
		tail := strings.SplitN(path, ".", 3)
		var t Path
		if len(tail) == 3 {
			var err error
			t, err = NewPath("." + tail[2])
			if err != nil {
				return bad, err
			}
		} else {
			t = identity
		}
		return func(node *yaml.Node) ([]*yaml.Node, error) {
			if node.Kind != yaml.MappingNode {
				return []*yaml.Node{}, errors.New("not a mapping node, so no children")
			}
			for i, n := range node.Content {
				if n.Value == tail[1] {
					return t(node.Content[i+1])
				}
			}
			return []*yaml.Node{}, errors.New("not found")
		}, nil
	}
	panic("not implemented!")
}
