package minsexp

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"strings"
	"testing"
)
import "github.com/stretchr/testify/require"

func TestDontPanicInEval(t *testing.T) {
	for inputForm, expectedErr := range map[string]string{
		`(panic)`:         "panic!",
		`(panic "oh no")`: "minsexp: oh no",
	} {
		panicFn := func(args []interface{}) (interface{}, error) {
			if len(args) == 1 {
				panic(args[0])
			} else {
				panic(errors.New("panic!"))
			}
		}
		lexicalScopes := []map[string]interface{}{{"panic": panicFn}}

		readSexp, idx, err := Read(inputForm, 0)
		require.Nil(t, err, inputForm)
		require.Equal(t, idx, len(inputForm), inputForm)

		printed := Print(readSexp)
		require.Equal(t, inputForm, printed)

		evalledSexp, err := Eval(StdEnv, lexicalScopes, readSexp)
		require.Nil(t, evalledSexp, inputForm)
		require.NotNil(t, err, inputForm)
		require.Equal(t, expectedErr, err.Error(), inputForm)
		stackTracer, ok := err.(stackTracer)
		require.True(t, ok, inputForm)
		require.True(t, len(stackTracer.StackTrace()) >= 5, inputForm)
		topOfStack := fmt.Sprintf("%+s", stackTracer.StackTrace()[0])
		require.Equal(t, 0, strings.Index(topOfStack, "github.com/EugenDueck/minsexp.Eval"), inputForm)
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func TestSimpleForms(t *testing.T) {
	for inputForm, expectedOutput := range map[string]interface{}{
		"()": nil,
	} {
		readSexp, idx, err := Read(inputForm, 0)
		require.Nil(t, err, inputForm)
		require.Equal(t, idx, len(inputForm), inputForm)

		printed := Print(readSexp)
		require.Equal(t, inputForm, printed)

		evalledSexp, err := Eval(nil, nil, readSexp)
		require.Nil(t, err, inputForm)
		require.Equal(t, evalledSexp, expectedOutput, inputForm)
	}
}

func TestBindLexicalScope(t *testing.T) {
	expectedOutput := 5
	evalledSexp, err := ReadEval([]map[string]interface{}{{"a": expectedOutput}}, "a")
	require.Nil(t, err)
	require.Equal(t, expectedOutput, evalledSexp)
}

func TestBindEnv(t *testing.T) {
	form := "(+ 1 a)"
	readSexp, idx, err := Read(form, 0)
	require.Nil(t, err)
	require.Equal(t, len(form), idx)
	require.NotNil(t, readSexp)

	printed := Print(readSexp)
	require.Equal(t, form, printed)

	aVal := decimal.NewFromFloat(5)
	expectedOutput := aVal.Add(decimal.NewFromFloat(1))
	env := make(map[string]interface{}, 2)
	//for k, v := range StdEnv {
	//	env[k] = v
	//}
	env["+"] = StdEnv["+"]
	env["a"] = aVal
	evalledSexp, err := Eval(env, nil, readSexp)
	require.Nil(t, err)
	require.Equal(t, expectedOutput, evalledSexp)
}

func TestNumbers(t *testing.T) {
	for _, v := range []string{"0", "1", "1.2", "-1.2", "+1", "+1.2",
		// max float64
		"179769313486232000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		// max float64 * 10
		"1797693134862320000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		// min float64
		"-179769313486232000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		// min float64 * 10
		"-1797693134862320000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
	} {
		expectedSexpOut, err := decimal.NewFromString(v)
		require.Nil(t, err, v)
		testRead(t, v, expectedSexpOut, v)
		testReadFully(t, v, expectedSexpOut, v)
	}
}

func TestMalformedNumbers(t *testing.T) {
	for _, v := range []string{"0.", "1.", "-1.", "+2."} {
		sexp, idx, err := Read(v, 0)
		require.Nil(t, sexp, v)
		require.NotNil(t, err, v)
		require.Equal(t, len(v)-1, idx, v)
	}
}

func TestSymbols(t *testing.T) {
	for _, v := range []string{"+", "-", "a", "and", "if", "nil", ":asdf", "let*"} {
		testRead(t, v, Symbol(v), v)
		testReadFully(t, v, Symbol(v), v)
	}
}

func testRead(t *testing.T, in string, expectedOut interface{}, msgAndArgs ...interface{}) interface{} {
	sexp, idx, err := Read(in, 0)
	require.Nil(t, err, msgAndArgs...)
	require.Equal(t, len(in), idx, msgAndArgs...)
	require.Equal(t, expectedOut, sexp, msgAndArgs...)
	return sexp
}

func testReadFully(t *testing.T, in string, expectedOut interface{}, msgAndArgs ...interface{}) interface{} {
	sexp, err := ReadFully(in)
	require.Nil(t, err, msgAndArgs...)
	require.Equal(t, expectedOut, sexp, msgAndArgs...)
	return sexp
}

func TestLet(t *testing.T) {
	for inputForm, expectedOutput := range map[string]interface{}{
		"(let a 0 a)":                     decimal.Zero,
		"(let a 1 a 0 a)":                 decimal.Zero,
		"(let a 1 (let a 0 a))":           decimal.Zero,
		"(let a 0 (do (let a 1 a) a))":    decimal.Zero,
		"(let 1)":                         decimal.NewFromFloat(1),
		"(let (+))":                       decimal.Zero,
		"(let a 1 (+))":                   decimal.Zero,
		"(let a 2 (+ 1 a))":               decimal.NewFromFloat(3),
		"(let a 1 (- a))":                 decimal.NewFromFloat(-1),
		"(let a 2 (- 1 a))":               decimal.NewFromFloat(-1),
		"(let a 1 (*))":                   decimal.NewFromFloat(1),
		"(let a 1 (* a 2))":               decimal.NewFromFloat(2),
		"(let a 3 (* 1 2 a))":             decimal.NewFromFloat(6),
		"(let a 1 (/ a))":                 decimal.NewFromFloat(1),
		"(let a 1 (/ a 2))":               decimal.NewFromFloat(0.5),
		"(let a 2 (/ 1 a 3))":             decimal.NewFromFloat(0.5).Div(decimal.NewFromFloat(3)),
		"(let a 1 b 2 (+))":               decimal.Zero,
		"(let a 1 b 2 (+ a b))":           decimal.NewFromFloat(3),
		"(let a 1 b 2 (- a))":             decimal.NewFromFloat(-1),
		"(let a 1 b 2 (- a b))":           decimal.NewFromFloat(-1),
		"(let a 1 b 2 (*))":               decimal.NewFromFloat(1),
		"(let a 1 b 2 (* a b))":           decimal.NewFromFloat(2),
		"(let a 1 b 2 (* a b 3))":         decimal.NewFromFloat(6),
		"(let a 1 b 2 (/ a))":             decimal.NewFromFloat(1),
		"(let a 1 b 2 (/ a b))":           decimal.NewFromFloat(0.5),
		"(let a 2 b 3 (/ 1 a b))":         decimal.NewFromFloat(0.5).Div(decimal.NewFromFloat(3)),
		"(let a 0 b 0 a 1 b 2 (+))":       decimal.Zero,
		"(let a 0 b 0 a 1 b 2 (+ a b))":   decimal.NewFromFloat(3),
		"(let a 0 b 0 a 1 b 2 (- a))":     decimal.NewFromFloat(-1),
		"(let a 0 b 0 a 1 b 2 (- a b))":   decimal.NewFromFloat(-1),
		"(let a 0 b 0 a 1 b 2 (*))":       decimal.NewFromFloat(1),
		"(let a 0 b 0 a 1 b 2 (* a b))":   decimal.NewFromFloat(2),
		"(let a 0 b 0 a 1 b 2 (* a b 3))": decimal.NewFromFloat(6),
		"(let a 0 b 0 a 1 b 2 (/ a))":     decimal.NewFromFloat(1),
		"(let a 0 b 0 a 1 b 2 (/ a b))":   decimal.NewFromFloat(0.5),
		"(let a 0 b 0 a 2 b 3 (/ 1 a b))": decimal.NewFromFloat(0.5).Div(decimal.NewFromFloat(3)),
	} {
		readSexp, idx, err := Read(inputForm, 0)
		require.Nil(t, err, inputForm)
		require.Equal(t, idx, len(inputForm), inputForm)

		printed := Print(readSexp)
		require.Equal(t, inputForm, printed)

		evalledSexp, err := Eval(StdEnv, nil, readSexp)
		require.Nil(t, err, inputForm)
		if decV, ok := expectedOutput.(decimal.Decimal); ok {
			if ok {
				require.Zero(t, decV.Cmp(evalledSexp.(decimal.Decimal)), inputForm)
			}
		} else {
			require.Equal(t, expectedOutput, evalledSexp, inputForm)
		}
	}
}

func TestLetFails(t *testing.T) {
	for _, inputForm := range []string{
		`(let)`,
		`(let a 1)`,
		`(let a 1 b 2)`,
		`(let a 1 b)`,
	} {
		readSexp, idx, err := Read(inputForm, 0)
		require.Nil(t, err, inputForm)
		require.Equal(t, idx, len(inputForm), inputForm)

		printed := Print(readSexp)
		require.Equal(t, inputForm, printed)

		evalledSexp, err := Eval(nil, nil, readSexp)
		require.NotNil(t, err, inputForm)
		require.Nil(t, evalledSexp, inputForm)
	}
}
