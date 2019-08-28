package minsexp

import (
	"errors"
	"github.com/shopspring/decimal"
	"testing"
)
import "github.com/stretchr/testify/require"

func TestCustomFnInEnv(t *testing.T) {
	for inputForm, expectedOutput := range map[string]interface{}{
		"(concat \"foo-\" \"bar\")": "foo-bar",
	} {
		env := make(map[string]interface{}, len(StdEnv)+1)
		for k, v := range StdEnv {
			env[k] = v
		}
		env["concat"] = concatFn
		readExpr, idx, err := Read(inputForm, 0)
		require.Nil(t, err, inputForm)
		require.Equal(t, idx, len(inputForm), inputForm)

		printed := Print(readExpr)
		require.Equal(t, inputForm, printed)

		evalledExpr, err := Eval(env, nil, readExpr)
		require.Nil(t, err, inputForm)
		if decV, ok := expectedOutput.(decimal.Decimal); ok {
			if ok {
				require.Zero(t, decV.Cmp(evalledExpr.(decimal.Decimal)), inputForm)
			}
		} else {
			require.Equal(t, expectedOutput, evalledExpr, inputForm)
		}
	}
}

func concatFn(args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, errors.New("concat expects 2 string arguments")
	}
	arg0, ok := args[0].(string)
	if !ok {
		return nil, errors.New("concat expects 2 string arguments")
	}
	arg1, ok := args[1].(string)
	if !ok {
		return nil, errors.New("concat expects 2 string arguments")
	}
	return arg0 + arg1, nil
}
