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

type expectedCountAndResult struct {
	count  int
	result interface{}
}

func TestDo(t *testing.T) {
	for inputForm, expectedOutput := range map[string]expectedCountAndResult{
		"(do)":                               {0, nil},
		"(do true nil)":                      {0, nil},
		"(do nil)":                           {0, nil},
		"(do true)":                          {0, true},
		"(do (count) true)":                  {1, true},
		"(do (count) (count) true)":          {2, true},
		"(do (count) (count) (count))":       {3, nil},
		"(do (count) (count) false (count))": {3, nil},
		"(do (count) (count) nil (count))":   {3, nil},
		"(do false)":                         {0, false},
		"(do 1)":                             {0, decimal.NewFromFloat(1)},
		"(do false true 3 4 false)":          {0, false},
	} {

		counter := 0
		countFn := func(args []interface{}) (interface{}, error) {
			counter++
			if len(args) == 1 {
				return args[0], nil
			} else {
				return nil, nil
			}
		}
		lexicalScopes := []map[string]interface{}{{"count": countFn}}

		readSexp, idx, err := Read(inputForm, 0)
		require.Nil(t, err, inputForm)
		require.Equal(t, idx, len(inputForm), inputForm)

		printed := Print(readSexp)
		require.Equal(t, inputForm, printed)

		evalledSexp, err := Eval(StdEnv, lexicalScopes, readSexp)
		require.Nil(t, err, inputForm)
		require.Equal(t, expectedOutput.count, counter, inputForm)
		if decV, ok := expectedOutput.result.(decimal.Decimal); ok {
			if ok {
				require.Zero(t, decV.Cmp(evalledSexp.(decimal.Decimal)), inputForm)
			}
		} else {
			require.Equal(t, expectedOutput.result, evalledSexp, inputForm)
		}
	}
}

func TestAnd(t *testing.T) {
	for inputForm, expectedOutput := range map[string]expectedCountAndResult{
		"(and)":                                            {0, true},
		"(and true nil)":                                   {0, nil},
		"(and true nil true)":                              {0, nil},
		"(and true nil (count))":                           {0, nil},
		"(and (count) true nil (count))":                   {1, nil},
		"(and (count) true false (count))":                 {1, nil},
		"(and (count true) true false (count))":            {1, false},
		"(and (count true) (count 3) false (count))":       {2, false},
		"(and (count true) (count 3) (count) false)":       {3, nil},
		"(and (count true) (count 3) (count false) false)": {3, false},
	} {

		counter := 0
		countFn := func(args []interface{}) (interface{}, error) {
			counter++
			if len(args) == 1 {
				return args[0], nil
			} else {
				return nil, nil
			}
		}
		lexicalScopes := []map[string]interface{}{{"count": countFn}}

		readSexp, idx, err := Read(inputForm, 0)
		require.Nil(t, err, inputForm)
		require.Equal(t, idx, len(inputForm), inputForm)

		printed := Print(readSexp)
		require.Equal(t, inputForm, printed)

		evalledSexp, err := Eval(StdEnv, lexicalScopes, readSexp)
		require.Nil(t, err, inputForm)
		require.Equal(t, expectedOutput.count, counter, inputForm)
		if decV, ok := expectedOutput.result.(decimal.Decimal); ok {
			if ok {
				require.Zero(t, decV.Cmp(evalledSexp.(decimal.Decimal)), inputForm)
			}
		} else {
			require.Equal(t, expectedOutput.result, evalledSexp, inputForm)
		}
	}
}

func TestOr(t *testing.T) {
	for inputForm, expectedOutput := range map[string]expectedCountAndResult{
		"(or)":                                {0, nil},
		"(or true nil)":                       {0, true},
		"(or 1 nil)":                          {0, decimal.NewFromFloat(1)},
		"(or false nil (count) (count true))": {2, true},
		"(or false nil (count) (count true) (count false))": {2, true},
		"(or (count) true (count))":                         {1, true},
		"(or true (count))":                                 {0, true},
		"(or true nil (count))":                             {0, true},
	} {

		counter := 0
		countFn := func(args []interface{}) (interface{}, error) {
			counter++
			if len(args) == 1 {
				return args[0], nil
			} else {
				return nil, nil
			}
		}
		lexicalScopes := []map[string]interface{}{{"count": countFn}}

		readSexp, idx, err := Read(inputForm, 0)
		require.Nil(t, err, inputForm)
		require.Equal(t, idx, len(inputForm), inputForm)

		printed := Print(readSexp)
		require.Equal(t, inputForm, printed)

		evalledSexp, err := Eval(StdEnv, lexicalScopes, readSexp)
		require.Nil(t, err, inputForm)
		require.Equal(t, expectedOutput.count, counter, inputForm)
		if decV, ok := expectedOutput.result.(decimal.Decimal); ok {
			if ok {
				require.Zero(t, decV.Cmp(evalledSexp.(decimal.Decimal)), inputForm)
			}
		} else {
			require.Equal(t, expectedOutput.result, evalledSexp, inputForm)
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
