package minsexp

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"reflect"
	"strings"
)

var (
	StdEnv = map[string]interface{}{
		// symbols
		"nil":   nil,
		"true":  true,
		"false": false,

		// special forms
		"and": andForm,
		"or":  orForm,
		"if":  ifForm,

		// functions
		"not":     notFn,
		"=":       equalsFn,
		"not=":    notEqualsFn,
		"compare": compareFn,
		"<=":      lessThanOrEqualFn,
		"<":       lessThanFn,
		">=":      greaterThanOrEqualFn,
		">":       greaterThanFn,
		"+":       plusFn,
		"-":       minusFn,
		"*":       multiplyFn,
		"/":       divideFn,
		"get":     getFn,
		"set":     setFn,
	}
)

func ifForm(env map[string]interface{}, lexicalScope []map[string]interface{}, args []interface{}) (interface{}, error) {
	if len(args) != 2 && len(args) != 3 {
		return nil, errors.New("(if condition then [else]) expects a 'condition', a 'then' expression, and may have an 'else' expression")
	}

	conditionResult, err := Eval(env, lexicalScope, args[0])
	if err != nil {
		return nil, err
	}
	if trueish(conditionResult) {
		return Eval(env, lexicalScope, args[1])
	} else if len(args) == 3 {
		return Eval(env, lexicalScope, args[2])
	} else {
		return nil, nil
	}
}

func trueish(value interface{}) bool {
	return value != nil && value != false
}

func andForm(env map[string]interface{}, lexicalScope []map[string]interface{}, args []interface{}) (interface{}, error) {
	var lastTrueish interface{} = true
	for _, arg := range args {
		result, err := Eval(env, lexicalScope, arg)
		if err != nil {
			return nil, err
		}
		if trueish(result) {
			lastTrueish = result
		} else {
			return nil, nil
		}
	}
	return lastTrueish, nil
}

func orForm(env map[string]interface{}, lexicalScope []map[string]interface{}, args []interface{}) (interface{}, error) {
	var lastFalseish interface{} = true
	for _, arg := range args {
		result, err := Eval(env, lexicalScope, arg)
		if err != nil {
			return nil, err
		}
		if trueish(result) {
			return result, nil
		} else {
			lastFalseish = result
		}
	}
	return lastFalseish, nil
}

func notFn(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.New("not expects one argument")
	}
	if trueish(args[0]) {
		return false, nil
	} else {
		return true, nil
	}
}

func equalsFn(args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, errors.New("= expects at least one argument")
	}
	cmp := args[0]
	for _, v := range args[1:] {
		if d, ok := v.(decimal.Decimal); ok {
			dCmp, bothDecimals := cmp.(decimal.Decimal)
			if bothDecimals {
				if dCmp.Cmp(d) != 0 {
					return false, nil
				}
			} else {
				return false, nil
			}
		} else if cmp != v {
			return false, nil
		}
	}
	return true, nil
}

func notEqualsFn(args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, errors.New("not= expects at least one argument")
	}
	cmp := args[0]
	for _, v := range args[1:] {
		if d, ok := v.(decimal.Decimal); ok {
			dCmp, bothDecimals := cmp.(decimal.Decimal)
			if bothDecimals {
				if dCmp.Cmp(d) != 0 {
					return true, nil
				}
			} else {
				return true, nil
			}
		} else if cmp != v {
			return true, nil
		}
	}
	return false, nil
}

func compareFn(args []interface{}) (interface{}, error) {
	if len(args) != 2 || reflect.TypeOf(args[0]) != reflect.TypeOf(args[1]) {
		return nil, errors.New("compare expects two arguments of the same type")
	}

	arg0I := args[0]
	switch arg0 := arg0I.(type) {
	case decimal.Decimal:
		return arg0.Cmp(args[1].(decimal.Decimal)), nil
	case string:
		return strings.Compare(arg0, args[1].(string)), nil
	default:
		return nil, errors.New(fmt.Sprintf("cannot compare items of type %T", arg0I))
	}
}

func lessThanOrEqualFn(args []interface{}) (interface{}, error) {
	cmp, e := compareFn(args)
	if e != nil {
		return nil, e
	} else {
		return cmp.(int32) <= 0, nil
	}
}

func lessThanFn(args []interface{}) (interface{}, error) {
	cmp, e := compareFn(args)
	if e != nil {
		return nil, e
	} else {
		return cmp.(int32) < 0, nil
	}
}

func greaterThanFn(args []interface{}) (interface{}, error) {
	cmp, e := compareFn(args)
	if e != nil {
		return nil, e
	} else {
		return cmp.(int32) > 0, nil
	}
}

func greaterThanOrEqualFn(args []interface{}) (interface{}, error) {
	cmp, e := compareFn(args)
	if e != nil {
		return nil, e
	} else {
		return cmp.(int32) >= 0, nil
	}
}

func plusFn(args []interface{}) (interface{}, error) {
	result := decimal.Zero
	for _, v := range args {
		if d, ok := v.(decimal.Decimal); ok {
			result = result.Add(d)
		} else {
			return nil, errors.New("+ works on numbers of type decimal.Decimal")
		}
	}
	return result, nil
}

func minusFn(args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, errors.New("- expects at least one argument of type decimal.Decimal")
	}
	var result decimal.Decimal
	for idx, v := range args {
		if d, ok := v.(decimal.Decimal); ok {
			if idx == 0 {
				if len(args) == 1 {
					return d.Neg(), nil
				} else {
					result = d
				}
			} else {
				result = result.Sub(d)
			}
		} else {
			return nil, errors.New("- works on numbers of type decimal.Decimal")
		}
	}
	return result, nil
}

func multiplyFn(args []interface{}) (interface{}, error) {
	result := decimal.NewFromFloat(1)
	for _, v := range args {
		if d, ok := v.(decimal.Decimal); ok {
			result = result.Mul(d)
		} else {
			return nil, errors.New("* works on numbers of type decimal.Decimal")
		}
	}
	return result, nil
}

func divideFn(args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, errors.New("/ expects at least one argument of type decimal.Decimal")
	}
	result := decimal.NewFromFloat(1)
	for _, v := range args {
		if d, ok := v.(decimal.Decimal); ok {
			result = result.Div(d)
		} else {
			return nil, errors.New("/ works on numbers of type decimal.Decimal")
		}
	}
	return result, nil
}

func getFn(args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, errors.New("get expects a struct argument and a string argument containing the field name")
	}
	fieldName, ok := args[1].(string)
	if !ok {
		return nil, errors.New("get expects a struct argument and a string argument containing the field name")
	}
	structValue := reflect.ValueOf(args[0]).Elem()
	field := structValue.FieldByName(fieldName)
	return field.Interface(), nil
}

func setFn(args []interface{}) (interface{}, error) {
	if len(args) != 3 {
		return nil, errors.New("set expects a struct argument, a string argument containing the field name, and a value argument")
	}
	fieldName, ok := args[1].(string)
	if !ok {
		return nil, errors.New("set expects a struct argument, a string argument containing the field name, and a value argument")
	}
	obj := args[0]
	objValueElem := reflect.ValueOf(obj).Elem()
	field := objValueElem.FieldByName(fieldName)
	field.Set(reflect.ValueOf(args[2]).Convert(field.Type())) // we need to use Convert to allow setting aliased types using instances of the underlying type
	return obj, nil
}
