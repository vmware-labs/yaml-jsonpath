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

func TestComparators(t *testing.T) {
	cases := []struct {
		name        string
		comparison  comparison
		comparator  comparator
		comparisons map[comparison]bool
		focus       bool // if true, run only tests with focus set to true
	}{
		{
			name:       "equal",
			comparator: equal,
			comparisons: map[comparison]bool{
				compareLessThan:     false,
				compareEqual:        true,
				compareGreaterThan:  false,
				compareIncomparable: false,
			},
		},
		{
			name:       "notEqual",
			comparator: notEqual,
			comparisons: map[comparison]bool{
				compareLessThan:     true,
				compareEqual:        false,
				compareGreaterThan:  true,
				compareIncomparable: true,
			},
		},
		{
			name:       "greaterThan",
			comparator: greaterThan,
			comparisons: map[comparison]bool{
				compareLessThan:     false,
				compareEqual:        false,
				compareGreaterThan:  true,
				compareIncomparable: false,
			},
		},
		{
			name:       "greaterThanOrEqual",
			comparator: greaterThanOrEqual,
			comparisons: map[comparison]bool{
				compareLessThan:     false,
				compareEqual:        true,
				compareGreaterThan:  true,
				compareIncomparable: false,
			},
		},
		{
			name:       "lessThan",
			comparator: lessThan,
			comparisons: map[comparison]bool{
				compareLessThan:     true,
				compareEqual:        false,
				compareGreaterThan:  false,
				compareIncomparable: false,
			},
		},
		{
			name:       "lessThanOrEqual",
			comparator: lessThanOrEqual,
			comparisons: map[comparison]bool{
				compareLessThan:     true,
				compareEqual:        true,
				compareGreaterThan:  false,
				compareIncomparable: false,
			},
		},
		{
			name:       "string equal",
			comparator: equal,
			comparisons: map[comparison]bool{
				compareStrings("a", "a"): true,
				compareStrings("a", "b"): false,
			},
		},
		{
			name:       "string not equal",
			comparator: notEqual,
			comparisons: map[comparison]bool{
				compareStrings("a", "a"): false,
				compareStrings("a", "b"): true,
			},
		},
		{
			name:       "float64 equal",
			comparator: equal,
			comparisons: map[comparison]bool{
				compareFloat64(1.1, 1.1): true,
				compareFloat64(1.1, 1.2): false,
			},
		},
		{
			name:       "float64 not equal",
			comparator: notEqual,
			comparisons: map[comparison]bool{
				compareFloat64(1.1, 1.1): false,
				compareFloat64(1.1, 1.2): true,
			},
		},
		{
			name:       "float64 greater than",
			comparator: greaterThan,
			comparisons: map[comparison]bool{
				compareFloat64(1.1, 1.2): false,
				compareFloat64(1.1, 1.1): false,
				compareFloat64(1.2, 1.1): true,
			},
		},
		{
			name:       "float64 greater than or equal",
			comparator: greaterThanOrEqual,
			comparisons: map[comparison]bool{
				compareFloat64(1.1, 1.2): false,
				compareFloat64(1.1, 1.1): true,
				compareFloat64(1.2, 1.1): true,
			},
		},
		{
			name:       "float64 less than",
			comparator: lessThan,
			comparisons: map[comparison]bool{
				compareFloat64(1.1, 1.2): true,
				compareFloat64(1.1, 1.1): false,
				compareFloat64(1.2, 1.1): false,
			},
		},
		{
			name:       "float64 less than or equal",
			comparator: lessThanOrEqual,
			comparisons: map[comparison]bool{
				compareFloat64(1.1, 1.2): true,
				compareFloat64(1.1, 1.1): true,
				compareFloat64(1.2, 1.1): false,
			},
		},
		{
			name:       "node values equal",
			comparator: equal,
			comparisons: map[comparison]bool{
				compareNodeValues("a", "a"):   true,
				compareNodeValues("a", "b"):   false,
				compareNodeValues("1.0", "1"): true,
				compareNodeValues("1.0", "a"): false,
				compareNodeValues("a", "1.0"): false,
			},
		},
		{
			name:       "node values not equal",
			comparator: notEqual,
			comparisons: map[comparison]bool{
				compareNodeValues("a", "a"):   false,
				compareNodeValues("a", "b"):   true,
				compareNodeValues("1.0", "1"): false,
				compareNodeValues("1.0", "a"): true,
				compareNodeValues("a", "1.0"): true,
			},
		},
		{
			name:       "node values greater than",
			comparator: greaterThan,
			comparisons: map[comparison]bool{
				compareNodeValues("1.1", "1.2"): false,
				compareNodeValues("1.1", "1.1"): false,
				compareNodeValues("1.2", "1.1"): true,
				compareNodeValues("a", "a"):     false, // should be excluded by lexer
				compareNodeValues("1.0", "a"):   false, // should be excluded by lexer
				compareNodeValues("a", "1.0"):   false, // should be excluded by lexer
			},
		},
		{
			name:       "node values greater than or equal",
			comparator: greaterThanOrEqual,
			comparisons: map[comparison]bool{
				compareNodeValues("1.1", "1.2"): false,
				compareNodeValues("1.1", "1.1"): true,
				compareNodeValues("1.2", "1.1"): true,
				compareNodeValues("a", "a"):     true,  // should be excluded by lexer
				compareNodeValues("1.0", "a"):   false, // should be excluded by lexer
				compareNodeValues("a", "1.0"):   false, // should be excluded by lexer
			},
		},
		{
			name:       "node values less than",
			comparator: lessThan,
			comparisons: map[comparison]bool{
				compareNodeValues("1.1", "1.2"): true,
				compareNodeValues("1.1", "1.1"): false,
				compareNodeValues("1.2", "1.1"): false,
				compareNodeValues("a", "a"):     false, // should be excluded by lexer
				compareNodeValues("1.0", "a"):   false, // should be excluded by lexer
				compareNodeValues("a", "1.0"):   false, // should be excluded by lexer
			},
		},
		{
			name:       "node values less than or equal",
			comparator: lessThanOrEqual,
			comparisons: map[comparison]bool{
				compareNodeValues("1.1", "1.2"): true,
				compareNodeValues("1.1", "1.1"): true,
				compareNodeValues("1.2", "1.1"): false,
				compareNodeValues("a", "a"):     true,  // should be excluded by lexer
				compareNodeValues("1.0", "a"):   false, // should be excluded by lexer
				compareNodeValues("a", "1.0"):   false, // should be excluded by lexer
			},
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
			for comparison, result := range tc.comparisons {
				require.Equal(t, result, tc.comparator(comparison), "%v", comparison)
			}
		})
	}

	if focussed {
		t.Fatalf("testcase(s) still focussed")
	}
}
