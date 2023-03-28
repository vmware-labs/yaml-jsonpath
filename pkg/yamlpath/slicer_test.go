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

func TestSlicer(t *testing.T) {
	cases := []struct {
		name        string
		index       string
		length      int
		expected    []int
		expectedErr string
		focus       bool // if true, run only tests with focus set to true
	}{
		{
			name:        "index",
			index:       "3",
			length:      10,
			expected:    []int{3},
			expectedErr: "",
		},
		{
			name:        "range",
			index:       "1:3",
			length:      10,
			expected:    []int{1, 2},
			expectedErr: "",
		},
		{
			name:        "range with step",
			index:       "1:6:2",
			length:      10,
			expected:    []int{1, 3, 5},
			expectedErr: "",
		},
		{
			name:        "wildcard",
			index:       "*",
			length:      4,
			expected:    []int{0, 1, 2, 3},
			expectedErr: "",
		},
		{
			name:        "range with everything omitted, short form",
			index:       ":",
			length:      4,
			expected:    []int{0, 1, 2, 3},
			expectedErr: "",
		}, {
			name:        "range with everything omitted, long form",
			index:       "::",
			length:      4,
			expected:    []int{0, 1, 2, 3},
			expectedErr: "",
		},
		{
			name:        "range with start omitted",
			index:       ":2",
			length:      10,
			expected:    []int{0, 1},
			expectedErr: "",
		},
		{
			name:        "range with start and end omitted",
			index:       "::2",
			length:      10,
			expected:    []int{0, 2, 4, 6, 8},
			expectedErr: "",
		},
		{
			name:        "last index",
			index:       "-1",
			length:      4,
			expected:    []int{3},
			expectedErr: "",
		},
		{
			name:        "overflowed index",
			index:       "4",
			length:      4,
			expected:    []int{},
			expectedErr: "",
		},
		{
			name:        "underflowed index",
			index:       "-5",
			length:      4,
			expected:    []int{},
			expectedErr: "",
		},
		{
			name:        "negative step with default start and end",
			index:       "::-1",
			length:      4,
			expected:    []int{3, 2, 1, 0},
			expectedErr: "",
		},
		{
			name:        "negative step with default start",
			index:       ":0:-1",
			length:      4,
			expected:    []int{3, 2, 1},
			expectedErr: "",
		},
		{
			name:        "negative step with default end",
			index:       "2::-1",
			length:      4,
			expected:    []int{2, 1, 0},
			expectedErr: "",
		},
		{
			name:        "larger negative step",
			index:       "::-2",
			length:      4,
			expected:    []int{3, 1},
			expectedErr: "",
		},
		{
			name:        "negative range with default step",
			index:       "-1:-3",
			length:      10,
			expected:    []int{},
			expectedErr: "",
		},
		{
			name:        "negative range with negative step",
			index:       "-1:-3:-1",
			length:      10,
			expected:    []int{9, 8},
			expectedErr: "",
		},
		{
			name:        "negative range with larger negative step",
			index:       "-1:-6:-2",
			length:      10,
			expected:    []int{9, 7, 5},
			expectedErr: "",
		},
		{
			name:        "larger negative range with larger negative step",
			index:       "-1:-7:-2",
			length:      10,
			expected:    []int{9, 7, 5},
			expectedErr: "",
		},
		{
			name:        "negative from, positive to",
			index:       "-5:7",
			length:      10,
			expected:    []int{5, 6},
			expectedErr: "",
		},
		{
			name:        "negative from",
			index:       "-2:",
			length:      10,
			expected:    []int{8, 9},
			expectedErr: "",
		},
		{
			name:        "positive from, negative to",
			index:       "1:-1",
			length:      10,
			expected:    []int{1, 2, 3, 4, 5, 6, 7, 8},
			expectedErr: "",
		},
		{
			name:        "negative from, positive to, negative step",
			index:       "-1:1:-1",
			length:      10,
			expected:    []int{9, 8, 7, 6, 5, 4, 3, 2},
			expectedErr: "",
		},
		{
			name:        "positive from, negative to, negative step",
			index:       "7:-5:-1",
			length:      10,
			expected:    []int{7, 6},
			expectedErr: "",
		},
		{
			name:        "too many colons",
			index:       "1:2:3:4",
			length:      10,
			expectedErr: "malformed array index, too many colons",
		},
		{
			name:        "non-integer array index",
			index:       "1:2:a",
			length:      10,
			expectedErr: "non-integer array index",
		},
		{
			name:        "zero step",
			index:       "1:2:0",
			length:      10,
			expectedErr: "array index step value must be non-zero",
		},
		{
			name:     "empty range",
			index:    "2:2",
			length:   10,
			expected: []int{},
		},
		{
			name:     "union",
			index:    "0,2",
			length:   10,
			expected: []int{0, 2},
		},
		{
			name:     "union with whitespace",
			index:    " 0 , 1 ",
			length:   10,
			expected: []int{0, 1},
		},
		{
			name:        "union with duplicated results (deviation from comparison project consensus)",
			index:       "1,0,1",
			length:      3,
			expected:    []int{1, 0, 1},
			expectedErr: "",
		},
		{
			name:        "union with wildcard and index",
			index:       "*,1",
			length:      3,
			expectedErr: "error in union member 0: wildcard cannot be used in union",
		},
		{
			name:     "default indices with empty array",
			index:    ":",
			length:   0,
			expected: []int{},
		},
		{
			name:     "negative step with empty array",
			index:    "::-1",
			length:   0,
			expected: []int{},
		},
		{
			name:        "empty string",
			index:       "",
			length:      10,
			expectedErr: "array index missing",
		},
		{
			name:     "maximal range with positive step",
			index:    "0:10",
			length:   10,
			expected: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name:     "maximal range with negative step",
			index:    "9:0:-1",
			length:   10,
			expected: []int{9, 8, 7, 6, 5, 4, 3, 2, 1},
		},
		{
			name:     "excessively large to value",
			index:    "2:113667776004",
			length:   10,
			expected: []int{2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name:     "excessively small from value",
			index:    "-113667776004:1",
			length:   10,
			expected: []int{0},
		},
		{
			name:     "excessively large from value with negative step",
			index:    "113667776004:0:-1",
			length:   10,
			expected: []int{9, 8, 7, 6, 5, 4, 3, 2, 1},
		},

		{
			name:     "excessively small to value with negative step",
			index:    "3:-113667776004:-1",
			length:   10,
			expected: []int{3, 2, 1, 0},
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
		actual, err := slice(tc.index, tc.length)
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
				return
			}
			require.Equal(t, tc.expected, actual)
		})
	}

	if focussed {
		t.Fatalf("testcase(s) still focussed")
	}
}
