# Go YAML Path

YAML node searching using path notation.

## YAML Path Syntax

Valid paths conform to the following BNF syntax.

```
<path> ::= <identity> | <root> <subpath> | <subpath>     ; <subpath> on its own is semantically equivalent to <root> <subpath>
<identity> := ""                                         ; the current node
<root> ::= "$"                                           ; the root node of a document
<subpath> ::= <identity> | <child> <subpath> |
              <child> <array subscript> <subpath> |
              <recursive descent> <subpath>

<child> ::= <dot child> | <bracket child>
<dot child> ::= "." <child name> | ".*"                  ; child named <child name> or all children
<bracket child> ::= "['" <child name> "']"               ; child named <child name>

<recursive descent> ::= ".." <child name>                ; all the descendants named <child name>

<array subscript> ::= "[" <index> "]"                    ; zero or more elements of an array
<index> ::= <integer> | <range> | "*"                    ; specific index, range of indices, or all indices
<range> ::= <integer> ":" <integer> |                    ; start to end
            <integer> ":" <integer> ":" <integer>        ; start to end by step
```
