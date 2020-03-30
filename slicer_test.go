/*
 * Copyright 2020 Go YAML Path Authors
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
			name:        "negative step",
			index:       "::-1",
			length:      4,
			expected:    []int{3, 2, 1, 0},
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
			name:        "negative range",
			index:       "-1:-3",
			length:      10,
			expected:    []int{9, 8},
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
			name:        "deeper negative range",
			index:       "-2:-4",
			length:      10,
			expected:    []int{8, 7},
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
			name:        "too many colons",
			index:       "1:2:3:4",
			length:      10,
			expectedErr: "malformed array index",
		},
		{
			name:        "non-integer array index",
			index:       "1:2:a",
			length:      10,
			expectedErr: "non-integer array index",
		},
	}

	for _, tc := range cases {
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
}
