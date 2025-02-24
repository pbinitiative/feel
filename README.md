# FEEL.go

The interpreter of the FEEL language(Friendly Enough Expression
Language) in go, FEEL is broadly used in DMN and BPMN to provide rule
engine and script support, the FEEL.go module can be imported into
other go projects or used as command line executable as FEEL
interpreter.

## Credits

This library is based on work of @superisaac/FEEL.go and many thanks and credits go to him.

## Build
* run `make build` to build feel interpreter bin/feel
* run `make test` to run testing

## Use in golang applications
```golang
import (
  feel "github.com/pbinitiative/feel"
)

res, err := feel.EvalString("5 + 7")
```

## Examples, with using the CLI tool
```shell

% bin/feel -c '"hello " + "world"'
"hello world"

% bin/feel -c '(function(a, b) a + b)(5, 8)'
13

# dump AST tree instead of evaluating the script
% bin/feel -c 'bind("a", 5); if a > 3 then "larger" else "smaller"' -ast
(explist (call bind ["a", 5]) (if (> a 3) "larger"  "smaller"))

% bin/feel -c 'some x in [3, 4, 8, 9] satisfies x % 2 = 0'
4

% bin/feel -c 'every x in [3, 4, 8, 9] satisfies x % 2 = 0'
[
  4,
  8
]
```
