/*
 * Copyright 2020 Go YAML Path Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import "strconv"

type comparison int

const (
	compareLessThan comparison = iota
	compareEqual
	compareGreaterThan
	compareIncomparable
)

func (c comparison) invertOrdering() comparison {
	switch c {
	case compareLessThan:
		return compareGreaterThan
	case compareGreaterThan:
		return compareLessThan
	default:
		return c
	}
}

type orderingOperator string

const (
	operatorLessThan           orderingOperator = "<"
	operatorLessThanOrEqual    orderingOperator = "<="
	operatorGreaterThan        orderingOperator = ">"
	operatorGreaterThanOrEqual orderingOperator = ">="
)

func (o orderingOperator) String() string {
	return string(o)
}

type comparator func(comparison) bool

func equal(c comparison) bool {
	return c == compareEqual
}

func notEqual(c comparison) bool {
	return c != compareEqual
}

func greaterThan(c comparison) bool {
	return c == compareGreaterThan
}

func greaterThanOrEqual(c comparison) bool {
	return c == compareGreaterThan || c == compareEqual
}

func lessThan(c comparison) bool {
	return c == compareLessThan
}

func lessThanOrEqual(c comparison) bool {
	return c == compareLessThan || c == compareEqual
}

func compareStrings(a, b string) comparison {
	if a == b {
		return compareEqual
	}
	return compareIncomparable
}

func compareFloat64(lhs, rhs float64) comparison {
	if lhs < rhs {
		return compareLessThan
	}
	if lhs > rhs {
		return compareGreaterThan
	}
	return compareEqual
}

func compareNodeValues(lhs string, rhs string) comparison {
	numeric := true
	lhsFloat, err := strconv.ParseFloat(lhs, 64)
	if err != nil {
		numeric = false
	}
	rhsFloat, err := strconv.ParseFloat(rhs, 64)
	if err != nil {
		numeric = false
	}
	if numeric {
		return compareFloat64(lhsFloat, rhsFloat)
	}
	return compareStrings(lhs, rhs)
}
