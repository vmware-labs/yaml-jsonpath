/*
 * Copyright 2020 Go YAML Path Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath_test

import (
	"bytes"
	"fmt"
	"log"

	"github.com/glyn/go-yamlpath/pkg/yamlpath"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/yaml.v3"
)

// Example uses a Path to find certain nodes and replace their content. Unlike a global change, it avoids false positives.
func Example() {
	y := `---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sample-deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
      - name: nginy
        image: nginy
        ports:
        - containerPort: 81
`
	var n yaml.Node

	err := yaml.Unmarshal([]byte(y), &n)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}

	p, err := yamlpath.NewPath("$.spec.template.spec.containers[*].image")
	if err != nil {
		log.Fatalf("cannot create path: %v", err)
	}

	q := p.Find(&n)

	for _, i := range q {
		i.Value = "example.com/user/" + i.Value
	}

	var buf bytes.Buffer
	e := yaml.NewEncoder(&buf)
	defer e.Close()
	e.SetIndent(2)

	err = e.Encode(&n)
	if err != nil {
		log.Fatalf("cannot marshal node: %v", err)
	}

	z := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: sample-deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: example.com/user/nginx
        ports:
        - containerPort: 80
      - name: nginy
        image: example.com/user/nginy
        ports:
        - containerPort: 81
`
	if buf.String() == z {
		fmt.Printf("success")
	} else {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(buf.String(), z, false)
		fmt.Println(diffs)
	}

	// Output: success
}
