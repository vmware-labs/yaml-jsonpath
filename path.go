/*
 * Copyright 2020 Go YAML Path Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import (
	"errors"
	"strings"

	"github.com/dprotaso/go-yit"
	"gopkg.in/yaml.v3"
)

// Path is a compiled YAML path expression.
type Path struct {
	f func(*yaml.Node) yit.Iterator
}

// Find applies the Path to a YAML node and returns the addresses of the subnodes which match the Path.
func (p *Path) Find(node *yaml.Node) []*yaml.Node {
	return p.f(node).ToArray()
}

// NewPath constructs a Path from a string expression.
func NewPath(path string) (*Path, error) {
	// identity
	if path == "" {
		return new(identity), nil
	}

	// root node
	if strings.HasPrefix(path, "$") {
		suffix := strings.TrimPrefix(path, "$")
		subPath, err := NewPath(suffix)
		if err != nil {
			return new(empty), err
		}
		return new(func(node *yaml.Node) yit.Iterator {
			if node.Kind != yaml.DocumentNode {
				return empty(node)
			}
			return compose(yit.FromNode(node.Content[0]), subPath)
		}), nil
	}

	// recursive descent
	if strings.HasPrefix(path, "..") {
		suffix := strings.TrimPrefix(path, "..")

		i := strings.IndexAny(suffix, ".[")
		var (
			childName string
			subPath   *Path
		)
		if i >= 0 {
			childName = suffix[:i]
			var err error
			subPath, err = NewPath(suffix[i:])
			if err != nil {
				return new(empty), err
			}
		} else {
			childName = suffix
			subPath = new(identity)
		}
		if childName == "" {
			return new(empty), errors.New("missing child name")
		}

		return new(func(node *yaml.Node) yit.Iterator {
			return compose(yit.FromNode(node).RecurseNodes(), childThen(childName, subPath))
		}), nil
	}

	// dot child
	if strings.HasPrefix(path, ".") {
		suffix := strings.TrimPrefix(path, ".")

		i := strings.IndexAny(suffix, ".[")
		var (
			childName string
			subPath   *Path
		)
		if i >= 0 {
			childName = suffix[:i]
			var err error
			subPath, err = NewPath(suffix[i:])
			if err != nil {
				return new(empty), err
			}
		} else {
			childName = suffix
			subPath = new(identity)
		}
		if childName == "" {
			return new(empty), errors.New("missing child name")
		}
		return childThen(childName, subPath), nil
	}

	// bracket child
	if strings.HasPrefix(path, "['") {
		suffix := strings.TrimPrefix(path, "['")

		tail := strings.SplitN(suffix, "']", 2)
		if len(tail) != 2 {
			return new(empty), errors.New("unmatched ['")
		}
		var err error
		subPath, err := NewPath(tail[1])
		if err != nil {
			return new(empty), err
		}
		childName := tail[0]
		if childName == "" {
			return new(empty), errors.New("missing child name")
		}
		return childThen(childName, subPath), nil

	}

	return new(empty), errors.New("invalid path syntax")
}

func identity(node *yaml.Node) yit.Iterator {
	return yit.FromNode(node)
}

func empty(*yaml.Node) yit.Iterator {
	return yit.FromNodes()
}

func compose(i yit.Iterator, p *Path) yit.Iterator {
	its := []yit.Iterator{}
	for a, ok := i(); ok; a, ok = i() {
		its = append(its, p.f(a))
	}
	return yit.FromIterators(its...)
}

func new(f func(node *yaml.Node) yit.Iterator) *Path {
	return &Path{f: f}
}

func childThen(childName string, p *Path) *Path {
	return new(func(node *yaml.Node) yit.Iterator {
		if node.Kind != yaml.MappingNode {
			return empty(node)
		}
		for i, n := range node.Content {
			if n.Value == childName {
				j := yit.FromNode(node.Content[i+1])
				return compose(j, p)
			}
		}
		return empty(node)
	})
}
