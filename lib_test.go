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
		"(>= 1 1)":             true,
		"(> 2 1)":              true,
		"(<= 1 1)":             true,
		"(< 2 1)":              false,
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

// run:                    go test -bench=. -benchtime 5s -cpuprofile=cpu.out
// create profiling graph: go tool pprof cpu.out
//                         web
func BenchmarkRead(b *testing.B) {
	forms := map[string]interface{}{
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
	}
	for n := 0; n < b.N; n++ {
		for inputForm := range forms {
			_, _, _ = Read(inputForm, 0)
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
	Price        decimal.Decimal
	PricePtr1    *decimal.Decimal
	PricePtr2    *decimal.Decimal
	TestType     testType
	TestTypePtr1 *testType
	TestTypePtr2 *testType
}

func TestGet(t *testing.T) {
	somePrice := decimal.NewFromFloat(98.765)
	someTestType := testA
	strukt := testStruct{somePrice, &somePrice, nil, someTestType, &someTestType, nil}
	for inputForm, expectedOutputI := range map[string]interface{}{
		`(get obj "Price")`:        somePrice,
		`(get obj "PricePtr1")`:    somePrice,
		`(get obj "PricePtr2")`:    nil,
		`(get obj "TestTypePtr1")`: someTestType,
		`(get obj "TestTypePtr2")`: nil,
		`(get obj "TestType")`:     someTestType,
	} {
		readSexp, idx, err := Read(inputForm, 0)
		require.Nil(t, err, inputForm)
		require.Equal(t, idx, len(inputForm), inputForm)

		printed := Print(readSexp)
		require.Equal(t, inputForm, printed)

		evalledSexp, err := Eval(StdEnv, []map[string]interface{}{{"obj": &strukt}}, readSexp)
		require.Nil(t, err, inputForm)
		switch expectedOutput := expectedOutputI.(type) {
		case decimal.Decimal:
			if expectedOutput.Cmp(evalledSexp.(decimal.Decimal)) != 0 {
				require.Equal(t, expectedOutput, evalledSexp, inputForm)
			}
		default:
			require.Equal(t, expectedOutput, evalledSexp, inputForm)
		}
	}
}

func TestGetFails(t *testing.T) {
	for _, inputForm := range []string{
		`(get obj)`,
		`(get "PricePtr1")`,
	} {
		readSexp, idx, err := Read(inputForm, 0)
		require.Nil(t, err, inputForm)
		require.Equal(t, idx, len(inputForm), inputForm)

		printed := Print(readSexp)
		require.Equal(t, inputForm, printed)

		strukt := testStruct{decimal.NewFromFloat(1234), nil, nil, testA, nil, nil}
		evalledSexp, err := Eval(StdEnv, []map[string]interface{}{{"obj": &strukt}}, readSexp)
		require.NotNil(t, err, inputForm)
		require.Nil(t, evalledSexp, inputForm)
	}
}

func TestSet(t *testing.T) {
	somePrice := decimal.NewFromFloat(98.765)
	someTestType := testA
	for inputForm, expectedOutput := range map[string]interface{}{
		`(set obj "Price" 1.23456)`:                         &testStruct{decimal.NewFromFloat(1.23456), nil, nil, testA, nil, nil},
		`(set obj "PricePtr1" nil)`:                         &testStruct{decimal.NewFromFloat(1234), nil, nil, testA, nil, nil},
		`(set obj "PricePtr1" 98.765)`:                      &testStruct{decimal.NewFromFloat(1234), &somePrice, nil, testA, nil, nil},
		`(set obj "Price" 1.23456 "TestType" "b")`:          &testStruct{decimal.NewFromFloat(1.23456), nil, nil, testB, nil, nil},
		`(get (set obj "Price" 1.23456) "Price")`:           decimal.NewFromFloat(1.23456),
		`(set obj "TestType" "b")`:                          &testStruct{decimal.NewFromFloat(1234), nil, nil, testB, nil, nil},
		`(get (set obj "TestType" "b") "TestType")`:         testB,
		`(set obj "TestTypePtr1" "a")`:                      &testStruct{decimal.NewFromFloat(1234), nil, nil, testA, &someTestType, nil},
		`(get (set obj "TestTypePtr1" "a") "TestTypePtr1")`: testA,
	} {
		readSexp, idx, err := Read(inputForm, 0)
		require.Nil(t, err, inputForm)
		require.Equal(t, idx, len(inputForm), inputForm)

		printed := Print(readSexp)
		require.Equal(t, inputForm, printed)

		strukt := testStruct{decimal.NewFromFloat(1234), nil, nil, testA, nil, nil}
		evalledSexp, err := Eval(StdEnv, []map[string]interface{}{{"obj": &strukt}}, readSexp)
		require.Nil(t, err, inputForm)
		require.Equal(t, expectedOutput, evalledSexp, inputForm)
	}
}

func TestSetFails(t *testing.T) {
	for _, inputForm := range []string{
		`(set obj "Price")`,
		`(set obj "PricePtr1")`,
		`(set obj "Price" 1.23456 "TestType")`,
	} {
		readSexp, idx, err := Read(inputForm, 0)
		require.Nil(t, err, inputForm)
		require.Equal(t, idx, len(inputForm), inputForm)

		printed := Print(readSexp)
		require.Equal(t, inputForm, printed)

		strukt := testStruct{decimal.NewFromFloat(1234), nil, nil, testA, nil, nil}
		evalledSexp, err := Eval(StdEnv, []map[string]interface{}{{"obj": &strukt}}, readSexp)
		require.NotNil(t, err, inputForm)
		require.Nil(t, evalledSexp, inputForm)
	}
}
