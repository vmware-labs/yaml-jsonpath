# Go YAML Path

Search and replace YAML using path notation.

## YAML Path Syntax

Valid paths conform to the following BNF syntax.

```
<path> ::= <identity> | <root> | <child>
<identity> := """                                  ; the current node
<root> ::= "$"                                     ; the root node of a document
<child> ::= <dot child> | <bracket child>
<dot child> ::= <path> "." <name>                  ; a child of <path> named <name>
```