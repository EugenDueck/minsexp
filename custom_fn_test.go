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
		readSexp, idx, err := Read(inputForm, 0)
		require.Nil(t, err, inputForm)
		require.Equal(t, idx, len(inputForm), inputForm)

		printed := Print(readSexp)
		require.Equal(t, inputForm, printed)

		evalledSexp, err := Eval(env, nil, readSexp)
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
