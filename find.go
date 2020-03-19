package yamlpath

import "gopkg.in/yaml.v3"

// Find locates all the nodes belonging to a specified node which match the given path.
func Find(node *yaml.Node, path Path) ([]*yaml.Node, error) {
	return path(node)
}
