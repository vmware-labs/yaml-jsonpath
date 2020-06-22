/*
 * Copyright 2020 VMware, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBracketChildNames(t *testing.T) {
	cases := []struct {
		name            string
		input           string
		expectedStrings []string
		focus           bool // if true, run only tests with focus set to true
	}{
		{
			name:            "single child",
			input:           "'child'",
			expectedStrings: []string{"child"},
		},
		{
			name:            "double quoted child",
			input:           `"child"`,
			expectedStrings: []string{"child"},
		},
		{
			name:            "multiple children",
			input:           "'child1','child2'",
			expectedStrings: []string{"child1", "child2"},
		},
		{
			name:            "mixed quoted children",
			input:           `"child1",'child2'`,
			expectedStrings: []string{"child1", "child2"},
		},
		{
			name:            "child with ' escaped",
			input:           "'Bob\\'s'",
			expectedStrings: []string{"Bob's"},
		},
		{
			name:            `child with " escaped`,
			input:           `'Bob\"s'`,
			expectedStrings: []string{`Bob"s`},
		},
		{
			name:            `single quoted child with escaped and unescaped quotes`,
			input:           `'\'\\"\"'`,
			expectedStrings: []string{`'\""`},
		},
		{
			name:            `double quoted child with escaped and unescaped quotes`,
			input:           `"\"\\'\'"`,
			expectedStrings: []string{`"\''`},
		},
		{
			name:            `single quoted child with special characters`,
			input:           `':@."$,*\'\\'`,
			expectedStrings: []string{`:@."$,*'\`},
		},
		{
			name:            `double quoted child with special characters`,
			input:           `":@.\"$,*'\\"`,
			expectedStrings: []string{`:@."$,*'\`},
		},
		{
			name:            "child with union delimiter",
			input:           "','",
			expectedStrings: []string{","},
		},
		{
			name:            "children with union delimiters",
			input:           `',',","`,
			expectedStrings: []string{",", ","},
		},
		{
			name:            "child with union delimiters",
			input:           "',,'",
			expectedStrings: []string{",,"},
		},
		{
			name:            "child with union delimiters and escapes",
			input:           "'\\',\\',\\''",
			expectedStrings: []string{"',','"},
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
			actualStrings := bracketChildNames(tc.input)
			require.Equal(t, tc.expectedStrings, actualStrings)
		})
	}

	if focussed {
		t.Fatalf("testcase(s) still focussed")
	}
}
