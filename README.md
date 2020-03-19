# Go YAML Path

YAML node searching using path notation.

## YAML Path Syntax

Valid paths conform to the following BNF syntax.

```
<path> ::= <identity> | <root> <subpath>
<identity> := """                                        ; the current node
<root> ::= "$"                                           ; the root node of a document
<subpath> ::= <identity> | <child> <subpath>

<child> ::= <dot child> | <bracket child>
<dot child> ::= "." <child name>                         ; a child of <path> named <child name>
<bracket child> ::= <path> "['" <child name> "']"        ; a child of <path> named <child name>
```
