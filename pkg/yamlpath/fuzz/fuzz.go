/*
 * Copyright 2020 VMware, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package fuzz

import "github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"

// Fuzz allows go-fuzz to drive the lexer/parser.
func Fuzz(data []byte) int {
	p, err := yamlpath.NewPath(string(data))
	if err != nil {
		if p != nil {
			panic("Path != nil on error")
		}
		return 0
	}
	return 1
}
