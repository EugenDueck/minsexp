package minsexp

import (
	"github.com/shopspring/decimal"
	"testing"
)
import "github.com/stretchr/testify/require"

func TestSimpleForms(t *testing.T) {
	for inputForm, expectedOutput := range map[string]interface{}{
		"()": nil,
	} {
		readExpr, idx, err := Read(inputForm, 0)
		require.Nil(t, err, inputForm)
		require.Equal(t, idx, len(inputForm), inputForm)

		printed := Print(readExpr)
		require.Equal(t, inputForm, printed)

		evalledExpr, err := Eval(nil, nil, readExpr)
		require.Nil(t, err, inputForm)
		require.Equal(t, evalledExpr, expectedOutput, inputForm)
	}
}

func TestBindLexicalScope(t *testing.T) {
	readExpr, idx, err := Read("a", 0)
	require.Nil(t, err)
	require.Equal(t, 1, idx)
	require.Equal(t, Symbol("a"), readExpr)

	expectedOutput := 5
	evalledExpr, err := Eval(nil, []map[string]interface{}{{"a": expectedOutput}}, readExpr)
	require.Nil(t, err)
	require.Equal(t, expectedOutput, evalledExpr)
}

func TestBindEnv(t *testing.T) {
	form := "(+ 1 a)"
	readExpr, idx, err := Read(form, 0)
	require.Nil(t, err)
	require.Equal(t, len(form), idx)
	require.NotNil(t, readExpr)

	printed := Print(readExpr)
	require.Equal(t, form, printed)

	aVal := decimal.NewFromFloat(5)
	expectedOutput := aVal.Add(decimal.NewFromFloat(1))
	env := make(map[string]interface{}, 2)
	//for k, v := range StdEnv {
	//	env[k] = v
	//}
	env["+"] = StdEnv["+"]
	env["a"] = aVal
	evalledExpr, err := Eval(env, nil, readExpr)
	require.Nil(t, err)
	require.Equal(t, expectedOutput, evalledExpr)
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
		expectedExprOut, err := decimal.NewFromString(v)
		require.Nil(t, err, v)
		exprOut := testRead(t, v, expectedExprOut, v)
		testEval(t, exprOut, expectedExprOut, v)
	}
}

func TestMalformedNumbers(t *testing.T) {
	for _, v := range []string{"0.", "1.", "-1.", "+2."} {
		expr, idx, err := Read(v, 0)
		require.Nil(t, expr, v)
		require.NotNil(t, err, v)
		require.Equal(t, len(v)-1, idx, v)
	}
}

func TestSymbols(t *testing.T) {
	for _, v := range []string{"+", "-", "a", "and", "if", "nil", ":asdf", "let*"} {
		testRead(t, v, Symbol(v), v)
	}
}

func testRead(t *testing.T, in string, expectedOut interface{}, msgAndArgs ...interface{}) interface{} {
	expr, idx, err := Read(in, 0)
	require.Nil(t, err, msgAndArgs...)
	require.Equal(t, len(in), idx, msgAndArgs...)
	require.Equal(t, expectedOut, expr, msgAndArgs...)
	return expr

}

func testEval(t *testing.T, exprIn interface{}, expectedOut interface{}, msgAndArgs ...interface{}) interface{} {
	exprOut, err := Eval(nil, nil, exprIn)
	require.Nil(t, err, msgAndArgs...)
	require.Equal(t, expectedOut, exprOut, msgAndArgs...)
	return exprOut
}
