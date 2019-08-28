# minsexp

A (not quite) minimal golang library that reads, evaluates and prints s-expressions. The environment for evaluation can be customized so that minsexp can be used as an embedded language e.g. for a rules engine.

## Sample usage
```
form := "(+ 1 a)"
a := 5
sexp, idx, err := minsexp.Read(form, 0)    // +++++++++++ read form
if err != nil {
  env := make(map[string]interface{}, 2)   // create env map
  env["+"] = StdEnv["+"]                   // borrow "+" fn from standard env
  env["a"] = decimal.NewFromFloat(a)       // bind value (numbers have to be decimals)
  out, err := minsexp.Eval(env, nil, sexp) // +++++++++++ eval sexp
  if err != nil {
    fmt.Println(out)                       // prints "6"
  }
}
```
## Features

- there is one built-in special form: `let`, which binds sequentially (like CL's `let*`)
- many other basic special forms (like `and`, `or`, `if`) and functions (like `=`, `not=`, `+`, `<=`) can be used
  - many basic functions are still missing (string concatenation etc)
- there are `get` and `set` functions that get or set (public) struct fields
- functions, special forms and variables share a single namespace
- numbers are of type `github.com/shopspring/decimal.Decimal`
- no support for macros

## Symbols of the core library

### symbols
- nil
- true
- false

### special forms
- let (built-in)
- and
- or
- if

### functions
- not
- =
- not=
- compare
- <=
- <
- \>=
- \>
- \+
- \-
- \*
- /
- get
- set
