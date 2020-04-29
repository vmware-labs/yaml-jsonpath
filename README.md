# Go YAML Path

[![GoDoc](https://godoc.org/github.com/glyn/go-yamlpath?status.svg)](https://godoc.org/github.com/glyn/go-yamlpath)
[![Go Report Card](https://goreportcard.com/badge/glyn/go-yamlpath)](https://goreportcard.com/report/glyn/go-yamlpath)

[YAML node](https://godoc.org/gopkg.in/yaml.v3#Node) searching using path notation.

## Syntax

Valid paths are strings conforming to the following BNF syntax.

```
<path> ::= <identity> | <root> <subpath> | <subpath>
<identity> := ""                                         ; the current node
<root> ::= "$"                                           ; the root node of a document
<subpath> ::= <identity> | <child> <subpath> |
              <child> <array access> <subpath> |
              <recursive descent> <subpath>

<child> ::= <dot child> | <bracket child>
<dot child> ::= "." <child name> | ".*"                  ; child named <child name> or all children
<bracket child> ::= "['" <child name> "']"               ; child named <child name>

<recursive descent> ::= ".." <child name>                ; all the descendants named <child name>

<array access> ::= "[" <index> "]" | "[" <filter> "]"    ; zero or more elements of a sequence
<index> ::= <integer> | <range> | "*"                    ; specific index, range of indices, or all indices
<range> ::= <integer> ":" <integer> |                    ; start (inclusive) to end (exclusive)
            <integer> ":" <integer> ":" <integer>        ; start (inclusive) to end (exclusive) by step

<filter> ::= "?(" <filter expr> ")"
<filter expr> ::= <filter conjunction> [ "||" <filter conjunction ]* ; disjunction
<filter conjunction> := <basic filter> [ "&&" <basic filter> ]* ; conjunction (binds more tightly than ||)
<basic filter> ::= <filter existence> |
                   "!" <filter existence> |              ; TODO: negation (should this apply to brackets too?)
                   <filter term> == <filter term> |      ; equality
                   <filter term> != <filter term> |      ; inequality
                   <filter term> > <filter term> |       ; numeric greater than
                   <filter term> >= <filter term> |      ; numeric greater than or equal to
                   <filter term> < <filter term> |       ; TODO: numeric less than
                   <filter term> <= <filter term> |      ; TODO: numeric less than or equal to
                   <filter term> =~ <regular expr>  |    ; TODO: matches regular expression
                   "(" <filter expr> ")"                 ; TODO: bracketing
<filter term> ::= "@" <subpath> |                        ; item relative to element being processed
                  "$" <subpath>                          ; item relative to root node of a document
                  <filter literal>
<filter existence> ::= "@" <subpath>                     ; item, relative to element being processed, exists
                       "$" <subpath>                     ; item, relative to root node of a document, exists
<regular expr> := "/" <string> "/"                       ; regular expression <<<<<<<<< TBD
<filter literal> ::= <integer> |                         ; positive or negative decimal integer
                     <floating point number> |           ; floating point number
                     "'" <string without '> "'"          ; string enclosed in single quotes
```

The `NewPath` function parses a string path and returns a corresponding value of the `Path` type and
an error indicating whether parsing succeeded or failed. 

## Semantics

The `Path` type's `Find` method takes a YAML node and returns a slice of descendents of the input node which match the Path. Each matching node appears at least once in the slice (but _may_ appear more than once).
If there are no matches, an empty slice is returned.

A path is logically a series of matchers. To start with, the first matcher is applied to a slice consisting of just the node which was input to the `Find` method. Each matcher is applied in turn to the slice of nodes found so far and the results are combined into a single slice, which then passes to the next matcher, and so on. If a matcher produces an
empty slice, then each subsequent matcher also produces an empty slice and the `Find` method returns an empty slice.

The following matchers, with corresponding concrete syntax, are supported. See the BNF syntax above for details of
the concrete syntax.

### Identity: empty string

This matches all the nodes in the input slice which therefore become the nodes of the matcher's output slice.
The identity matcher defines the behaviour of a path consisting of the empty string and is the only way
of terminating the `<subpath>` production in the BNF syntax.

### Root: `$`

This matches the root node of the input YAML node. This matcher may be specified only at the start of the path. It is optional and, if omitted, the root node is matched before the rest of the path is applied. The output slice consists of just the root node.

### Child: `.childname` or `['childname']`

This matches the children with the given name of all the mapping nodes in the input slice. The output slice consists of all those children. The given name may be a single child name (no periods) or a series of single child names separated by periods. Non-mapping nodes in the input slice are not matched.

Although either form `.childname` or `['childname']` accepts a child name with embedded spaces, the 
`['childname']` form may be more convenient in some situations.

### Recursive Descent: `..childname` or `..*`

A matcher of the form `..childname` selects all the descendents of the nodes in the input slice (including those nodes) with the given name (using the same rules as the child matcher). The output slice consists of all the matching descendents.

A matcher of the form `..*` selects all the descendents of the nodes in the input slice (including those nodes).

### Array Subscript: `[integer]`, `[start:end]`, `[start:end:step]`, or `[*]`

This matches subsequences of all the sequence nodes in the input slice. Non-sequence nodes in the
input slice are not matched.

A matcher of the form `[integer]` selects the corresponding node in each sequence node, with `0` meaning the first node in the sequence, `1` the second node, and so on. A special index of `-1` selects the last node in each sequence.

A matcher of the form `[start:end]` or `[start:end:step]` selects the corresponding nodes in each sequence node starting from the start of the range (inclusive) to the end of the range (exclusive) with an optional step value (which defaults to `1`). A step value of `-1` may be used to step backwards from the end of the sequence to the
start.

A matcher of the form `[*]` selects all the nodes in each sequence node.

## References

The following sources inspired the syntax and semantics of Go YAML Path:

* [JSONPath - XPath for JSON](https://goessner.net/articles/JsonPath/) by Stefan Goessner
* [JSONPath Support](https://kubernetes.io/docs/reference/kubectl/jsonpath/) in the Kubernetes Reference documentation
* [JSONPath User Guide](https://unofficial-kubernetes.readthedocs.io/en/latest/user-guide/jsonpath/) in the Unofficial Kubernetes documentation
* [JSONPath Syntax](https://support.smartbear.com/alertsite/docs/monitors/api/endpoint/jsonpath.html) in the SmartBear AlertSite documentation