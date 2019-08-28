package minsexp

import (
	"github.com/shopspring/decimal"
	"testing"
)
import "github.com/stretchr/testify/require"

func TestForms(t *testing.T) {
	for inputForm, expectedOutput := range map[string]interface{}{
		"(not false)":          true,
		"(not nil)":            true,
		"(not true)":           false,
		"(not 1)":              false,
		"(not \"a\")":          false,
		"(= (not false) true)": true,
		"(= (not true) false)": true,
		"(= 1)":                true,
		"(= 1 1)":              true,
		"(= 1 1 1)":            true,
		"(= 1 1 2)":            false,
		"(= 1 2)":              false,
		"(= 1 \"1\")":          false,
		"(not= 1)":             false,
		"(not= 1 1)":           false,
		"(not= 1 1 1)":         false,
		"(not= 1 1 2)":         true,
		"(not= 1 2)":           true,
		"(not= 1 \"1\")":       true,
		"(+)":                  decimal.Zero,
		"(+ 1 2)":              decimal.NewFromFloat(3),
		"(- 1)":                decimal.NewFromFloat(-1),
		"(- 1 2)":              decimal.NewFromFloat(-1),
		"(*)":                  decimal.NewFromFloat(1),
		"(* 1 2)":              decimal.NewFromFloat(2),
		"(* 1 2 3)":            decimal.NewFromFloat(6),
		"(/ 1)":                decimal.NewFromFloat(1),
		"(/ 1 2)":              decimal.NewFromFloat(0.5),
		"(/ 1 2 3)":            decimal.NewFromFloat(0.5).Div(decimal.NewFromFloat(3)),
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

func TestLet(t *testing.T) {
	for inputForm, expectedOutput := range map[string]interface{}{
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

type testType string

const (
	testA testType = "a"
	testB testType = "b"
)

type testStruct struct {
	Price decimal.Decimal
	Test  testType
}

func TestGet(t *testing.T) {
	for inputForm, expectedOutput := range map[string]interface{}{
		"(get obj \"Price\")": decimal.NewFromFloat(1.2345),
		"(get obj \"Test\")":  testA,
	} {
		readSexp, idx, err := Read(inputForm, 0)
		require.Nil(t, err, inputForm)
		require.Equal(t, idx, len(inputForm), inputForm)

		printed := Print(readSexp)
		require.Equal(t, inputForm, printed)

		strukt := testStruct{decimal.NewFromFloat(1.2345), testA}
		evalledSexp, err := Eval(StdEnv, []map[string]interface{}{{"obj": &strukt}}, readSexp)
		require.Nil(t, err, inputForm)
		if decV, ok := expectedOutput.(decimal.Decimal); ok {
			if ok {
				if decV.Cmp(evalledSexp.(decimal.Decimal)) != 0 {
					require.Equal(t, expectedOutput, evalledSexp, inputForm)
				}
			}
		} else {
			require.Equal(t, expectedOutput, evalledSexp, inputForm)
		}
	}
}

func TestSet(t *testing.T) {
	for inputForm, expectedOutput := range map[string]interface{}{
		"(set obj \"Price\" 1.23456)":                 &testStruct{decimal.NewFromFloat(1.23456), testA},
		"(get (set obj \"Price\" 1.23456) \"Price\")": decimal.NewFromFloat(1.23456),
		"(set obj \"Test\" \"b\")":                    &testStruct{decimal.NewFromFloat(1234), testB},
		"(get (set obj \"Test\" \"b\") \"Test\")":     testB,
	} {
		readSexp, idx, err := Read(inputForm, 0)
		require.Nil(t, err, inputForm)
		require.Equal(t, idx, len(inputForm), inputForm)

		printed := Print(readSexp)
		require.Equal(t, inputForm, printed)

		strukt := testStruct{decimal.NewFromFloat(1234), testA}
		evalledSexp, err := Eval(StdEnv, []map[string]interface{}{{"obj": &strukt}}, readSexp)
		require.Nil(t, err, inputForm)
		require.Equal(t, expectedOutput, evalledSexp, inputForm)
	}
}
