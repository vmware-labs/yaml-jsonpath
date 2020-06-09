/*
 * Copyright 2020 VMware, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"io/ioutil"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

func TestRegressionSuite(t *testing.T) {
	y, err := ioutil.ReadFile("./testdata/regression_suite.yaml")
	if err != nil {
		t.Error(err)
	}

	var suite regressionSuite

	if err = yaml.Unmarshal(y, &suite); err != nil {
		t.Fatal(err)
	}

	focussed := false
	for _, tc := range suite.Testcases {
		if tc.Focus {
			focussed = true
			break
		}
	}

	tests, passed, consensus := 0, 0, 0
	for _, tc := range suite.Testcases {
		if focussed && !tc.Focus {
			continue
		}
		if tc.Exclude && !tc.Focus {
			continue
		}
		tests++
		if tc.Consensus.Kind > 0 {
			consensus++
		}
		if pass := t.Run(tc.Name, func(t *testing.T) {
			defer func() {
				p := recover()
				if p != nil {
					// fail on panic regardless of whether there is a consensus
					t.Fatalf("Panicked on test: %s: %v", tc.Name, p)
				}
			}()

			path, err := yamlpath.NewPath(tc.Selector)
			// if there is a consensus, check that the returned error agrees with it
			if tc.Consensus.Kind > 0 {
				if tc.Consensus.Value == "NOT_SUPPORTED" {
					require.Error(t, err, "NewPath allowed selector not supported by the consensus: %s, test: %s", tc.Selector, tc.Name)
				} else {
					require.NoError(t, err, "NewPath failed with selector: %s, test: %s", tc.Selector, tc.Name)
				}
			}
			if err != nil {
				require.Nil(t, path)
				return
			}
			require.NotNil(t, path)

			results, err := path.Find(&tc.Document)
			// if there is a consensus, check we agree with it
			if tc.Consensus.Kind > 0 {
				require.NoError(t, err, "Find failed with selector: %s, test: %s", tc.Selector, tc.Name)

				sanitise(tc.Consensus.Content)
				sanitise(results)
				if !tc.Ordered {
					require.ElementsMatch(t, tc.Consensus.Content, results, "Disagreed with consensus, selector: %s, test: %s", tc.Selector, tc.Name)
				} else {
					require.Equal(t, tc.Consensus.Content, results, "Disagreed with consensus, selector: %s, test: %s", tc.Selector, tc.Name)
				}
			}
		}); pass {
			passed++
		}
	}

	t.Logf("%d passed and %d failed of %d tests of which %d had consensus", passed, tests-passed, tests, consensus)

	if focussed {
		t.Fatalf("testcase(s) still focussed")
	}
}

// clear line and column numbers and sort objects by key
func sanitise(nodes []*yaml.Node) {
	for _, n := range nodes {
		n.Line = 0
		n.Column = 0
		if n.Kind == yaml.MappingNode {
			keys := []string{}
			keyNodes := make(map[string]*yaml.Node)
			valueNodes := make(map[string]*yaml.Node)
			for i, m := range n.Content {
				if i%2 != 0 {
					continue // skip non child names
				}
				keys = append(keys, m.Value)
				keyNodes[m.Value] = m
				valueNodes[m.Value] = n.Content[i+1]
			}

			sort.Strings(keys)

			for i, key := range keys {
				n.Content[2*i] = keyNodes[key]
				n.Content[2*i+1] = valueNodes[key]
			}
		}
		sanitise(n.Content)
	}
}

type testcase struct {
	Name      string `yaml:"id"`
	Selector  string
	Document  yaml.Node
	Ordered   bool
	Consensus yaml.Node
	Focus     bool // if true, run only tests with focus set to true
	Exclude   bool // if true, do not run this test unless it is focussed
}

type regressionSuite struct {
	Testcases []testcase `yaml:"queries"`
}
