# Fuzz testing

This uses [go-fuzz](https://github.com/dvyukov/go-fuzz).

## Initial setup

This shouldn't need repeating, but just in case: follow the instructions in the `go-fuzz` [README](https://github.com/dvyukov/go-fuzz).

The initial corpus was generated using commands such as:
```
cd pkg/yamlpath/fuzz/corpus
grep 'path:' ../../lexer_test.go | grep -o '".*"' | sed 's/^"//' | sed 's/"$//' | awk '1==1{close("lexer_test"i);x="lexer_test"++i;}{print > x}'
grep 'selector:' ../../../../test/testdata/regression_suite.yaml | grep -o '".*"' | sed 's/^"//' | sed 's/"$//' | awk '1==1{close("regression_suite"i);x="regression_suite"++i;}{print > x}'
```

## Fuzzing

Again follow the instructions in the `go-fuzz` [README](https://github.com/dvyukov/go-fuzz):
```
cd pkg/yamlpath/fuzz
go-fuzz-build
go-fuzz -procs 20
```
You can increase the parallelism level (`-procs`), but if you choose too large a value, fuzzing may fail with `too many open files` (especially on macOS).

It's worth stopping (using `ctrl-C`) and restarting fuzzing every so often as this seems to increase the corpus and coverage more quickly.
See [Why does restarting/recompiling often generate corpus?](https://github.com/dvyukov/go-fuzz/issues/257)

Feel free to contribute new corpus, or other improvements, by pull request as usual. *Please don't check in the changes to `go.mod` and `go.sum` caused by `go-fuzz-build` as these aren't otherwise needed.*

## Entertainment

I used [watchman](https://facebook.github.io/watchman/) to print out new corpus as it's found:
```
cd pkg/yamlpath/fuzz/corpus
watchman watch $PWD
watchman -- trigger $PWD buildme '*' -- cat
tail -f /usr/local/var/run/watchman/*/log
```

You're log location may vary - see [stack overflow](https://stackoverflow.com/questions/27723367/watchman-where-is-the-default-log-file).