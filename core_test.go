package minsexp

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"testing"
)
import "github.com/stretchr/testify/require"

func TestDontPanicInEval(t *testing.T) {
	panicIdx := 0
	for inputForm, expectedErr := range map[string]interface{}{
		`(panic)`:         errors.New("minsexp: panic! 1"),
		`(panic "oh no")`: errors.New("panic! 2: oh no"),
	} {
		panicIdx++
		panicFn := func(args []interface{}) (interface{}, error) {
			if len(args) == 1 {
				panic(errors.New(fmt.Sprintf("panic! %v: %v", panicIdx, args[0])))
			} else {
				panic(fmt.Sprintf("panic! %v", panicIdx))
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
		require.Equal(t, expectedErr, err, inputForm)
	}
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
	readSexp, idx, err := Read("a", 0)
	require.Nil(t, err)
	require.Equal(t, 1, idx)
	require.Equal(t, Symbol("a"), readSexp)

	expectedOutput := 5
	evalledSexp, err := Eval(nil, []map[string]interface{}{{"a": expectedOutput}}, readSexp)
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
