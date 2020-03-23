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

	"github.com/glyn/go-yamlpath"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/yaml.v3"
)

func Example() {
	// TODO: change the example once array indexing is implemented
	y := `---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sample-deployment
spec:
  temphack:
    name: blah
    image: blah
  template:
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
`
	var n yaml.Node

	err := yaml.Unmarshal([]byte(y), &n)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}

	// p, err := yamlpath.NewPath("$.spec.template.spec.containers[0].image")
	p, err := yamlpath.NewPath("$.spec.temphack.image")
	if err != nil {
		log.Fatalf("cannot create path: %v", err)
	}

	q, err := p(&n)
	if err != nil || len(q) != 1 {
		log.Fatalf("path failed: %v", err)
	}

	q[0].Value = "asdf"

	var buf bytes.Buffer
	e := yaml.NewEncoder(&buf)
	defer e.Close()
	e.SetIndent(2)

	// d, err := yaml.Marshal(&n)
	err = e.Encode(&n)
	if err != nil {
		log.Fatalf("cannot marshal node: %v", err)
	}

	z := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: sample-deployment
spec:
  temphack:
    name: blah
    image: asdf
  template:
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
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
