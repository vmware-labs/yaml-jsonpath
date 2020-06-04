## Fuzz testing

This uses [go-fuzz](https://github.com/dvyukov/go-fuzz).

## Initial setup

The shouldn't need repeating, but just in case follow the instructions in the go-fuzz [README](https://github.com/dvyukov/go-fuzz).

The initial corpus was generated using commands such as:
```
cd pkg/yamlpath/fuzz/corpus
grep 'path:' ../../lexer_test.go | grep -o '".*"' | sed 's/^"//' | sed 's/"$//' | awk '1==1{close("lexer_test"i);x="lexer_test"++i;}{print > x}'
grep 'selector:' ../../../../test/testdata/regression_suite.yaml | grep -o '".*"' | sed 's/^"//' | sed 's/"$//' | awk '1==1{close("regression_suite"i);x="regression_suite"++i;}{print > x}'
```