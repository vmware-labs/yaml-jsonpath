# Go YAML Path

YAML node searching using path notation.

## YAML Path Syntax

Valid paths conform to the following BNF syntax.

```
<path> ::= <identity> | <root> <subpath>
<identity> := ""                                         ; the current node
<root> ::= "$"                                           ; the root node of a document
<subpath> ::= <identity> | <child> <subpath> |
              <recursive descent> <subpath>

<child> ::= <dot child> | <bracket child>
<dot child> ::= "." <child name> | ".*"                  ; child named <child name> or all children
<bracket child> ::= "['" <child name> "']"               ; child named <child name>

<recursive descent> ::= ".." <child name>                ; all the descendants named <child name>
```
