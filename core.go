package minsexp

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
)

type Symbol string

func Read(sexpStr string, startIdx int) (sexp interface{}, idx int, err error) {
	defer func() {
		if r := recover(); r != nil {
			sexp = nil
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("minsexp: %v", r)
			}
		}
	}()
	return parseSexp(sexpStr, startIdx)
}

func ReadFully(sexpStr string) (sexp interface{}, err error) {
	expr, idx, err := Read(sexpStr, 0)
	if idx != len(sexpStr) {
		return nil, errors.New("expected a string containing a single sexp, but got: " + sexpStr)
	}
	return expr, err
}

// functions must have this interface: func(fnName string, args []interface{}) (interface{}, error)
// special forms must have this interface: func(env map[string]interface{}, lexicalScope []map[string]interface{}, fnName string, args []interface{}) (interface{}, error)
func Eval(env map[string]interface{}, lexicalScope []map[string]interface{}, sexp interface{}) (result interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = nil
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("minsexp: %v", r)
			}
		}
	}()
	switch sexp := sexp.(type) {
	case []interface{}:
		if len(sexp) == 0 {
			return nil, nil
		}

		if sexp[0] == Symbol("let") {
			if len(sexp)%2 != 0 {
				return nil, errors.New("let needs an uneven number of arguments: name/sexp pairs and one sexp")
			}

			//nameValuePairCount := (len(sexp) - 2) / 2
			//newLexicalScope := make([]map[string]interface{}, len(lexicalScope), len(lexicalScope) + nameValuePairCount)
			//copy(newLexicalScope, lexicalScope)
			newLexicalScope := append(lexicalScope[:0:0], lexicalScope...)
			newLexicalScopeMap := make(map[string]interface{})
			newLexicalScope = append(newLexicalScope, newLexicalScopeMap)
			for i := 1; i+1 < len(sexp); i += 2 {
				if nameSymbol, ok := sexp[i].(Symbol); ok {
					value, err := Eval(env, newLexicalScope, sexp[i+1])
					if err != nil {
						return nil, err
					}
					newLexicalScopeMap[string(nameSymbol)] = value
				} else {
					return nil, errors.New("let needs an uneven number of arguments: name-symbol/sexp pairs and one sexp")
				}
			}
			return Eval(env, newLexicalScope, sexp[len(sexp)-1])
		}

		fnOrSpecialForm, fnErr := Eval(env, lexicalScope, sexp[0])
		if fnErr != nil {
			return nil, fnErr
		}
		if funFn, ok := fnOrSpecialForm.(func([]interface{}) (interface{}, error)); ok {
			args := make([]interface{}, len(sexp)-1)
			for i, v := range sexp[1:] {
				out, fnErr := Eval(env, lexicalScope, v)
				if fnErr != nil {
					return nil, fnErr
				}
				args[i] = out
			}
			return funFn(args)
		} else if specialForm, ok := fnOrSpecialForm.(func(map[string]interface{}, []map[string]interface{}, []interface{}) (interface{}, error)); ok {
			return specialForm(env, lexicalScope, sexp[1:])
		} else {
			return nil, errors.New(fmt.Sprintf("Not a special form and not a function: %v", sexp[0]))
		}
	case Symbol:
		for i := len(lexicalScope) - 1; i >= 0; i-- {
			m := lexicalScope[i]
			v, ok := m[string(sexp)]
			if ok {
				return v, nil
			}
		}
		v, ok := env[string(sexp)]
		if ok {
			return v, nil
		} else {
			return nil, errors.New("Unbound name " + string(sexp))
		}
	default:
		return sexp, nil
	}
}

func Print(sexpI interface{}) (s string) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("minsexp: %v", r)
			}
			s = err.Error()
		}
	}()
	switch sexp := sexpI.(type) {
	case []interface{}:
		s := "("
		for _, v := range sexp {
			if s != "(" {
				s += " "
			}
			s += Print(v)
		}
		return s + ")"
	case string:
		return "\"" + sexp + "\""
	default:
		return fmt.Sprintf("%v", sexp)
	}
}

func getNextNonWSP(s string, startIdx int) int {
	for idx := startIdx; idx < len(s); idx++ {
		switch s[idx] {
		case ' ':
		case '\t':
		case '\r':
		case '\n':
		default:
			return idx
		}
	}
	return len(s)
}

func getNextNonSymbolChar(s string, startIdx int) int {
	for idx := startIdx; idx < len(s); idx++ {
		switch s[idx] {
		case ' ':
			fallthrough
		case '(':
			fallthrough
		case ')':
			fallthrough
		case '[':
			fallthrough
		case ']':
			fallthrough
		case '{':
			fallthrough
		case '}':
			fallthrough
		case ',':
			fallthrough
		case '\t':
			fallthrough
		case '\r':
			fallthrough
		case '\n':
			return idx
		}
	}
	return len(s)
}

func parseList(s string, startIdx int) ([]interface{}, int, error) {
	//fmt.Println("parseList", startIdx)
	if s[startIdx] != '(' {
		return nil, startIdx, errors.New("expecting '(' at start of list")
	}
	startIdx++
	var list []interface{}
	for {
		i := getNextNonWSP(s, startIdx)
		if i >= len(s) {
			return nil, i, errors.New("reached end of input parsing list")
		}
		if s[i] == ')' {
			return list, i + 1, nil
		}
		var value interface{}
		var err error
		value, startIdx, err = parseSexp(s, i)
		if err != nil {
			return nil, startIdx, err
		}
		list = append(list, value)
	}
}

func parseSymbol(s string, startIdx int) (Symbol, int, error) {
	//fmt.Println("parseSymbol", startIdx)
	i := getNextNonSymbolChar(s, startIdx)
	return Symbol(s[startIdx:i]), i, nil
}

func parseString(s string, startIdx int) (string, int, error) {
	//fmt.Println("parseString", startIdx)
	if s[startIdx] != '"' {
		return "", startIdx, errors.New("expecting '\"' at start of string")
	}

	j := startIdx + 1
	for ; j < len(s); j++ {
		// todo: handle '\' escapes
		if s[j] == '"' {
			return s[startIdx+1 : j], j + 1, nil
		}
	}
	return "", j, errors.New(fmt.Sprintf("string starting at %v not terminated by double quote", startIdx))
}

func isAfterNumber(c uint8) bool {
	switch c {
	case ' ':
		fallthrough
	case '\t':
		fallthrough
	case '\r':
		fallthrough
	case '\n':
		fallthrough
	case ')':
		return true
	default:
		return false
	}

}
func parseNumber(s string, startIdx int) (interface{}, int, error) {
	// todo: allow scientific notation
	afterSignIdx := startIdx
	firstChar := s[startIdx]
	if firstChar == '-' || firstChar == '+' {
		afterSignIdx++
	}
	dotIdx := -1
	for i := startIdx + 1; i <= len(s); i++ {
		if i == len(s) || isAfterNumber(s[i]) {
			if i-1 == dotIdx {
				return nil, i - 1, errors.New(fmt.Sprintf("not a valid number at %v: %v", i-1, s[startIdx:i]))
			}
			f, e := decimal.NewFromString(s[startIdx:i])
			if e != nil {
				return nil, i, e
			} else {
				return f, i, nil
			}
		}

		c := s[i]
		if (c < '0' || c > '9') && (c != '.' || dotIdx > -1 || i == afterSignIdx) {
			return nil, i, errors.New(fmt.Sprintf("not a valid number at %v: %v", i, s[startIdx:i+1]))
		}
		if c == '.' {
			dotIdx = i
		}
	}
	return nil, len(s), errors.New("impossible error")
}

func parseSexp(s string, startIdx int) (value interface{}, nextIndex int, err error) {
	i := getNextNonWSP(s, startIdx)
	if i >= len(s) {
		return nil, i, errors.New("reached end of input parsing sexp")
	}
	b := s[i]

	switch b {
	case '(':
		return parseList(s, i)
	case '"':
		return parseString(s, i)

	case '+':
		fallthrough
	case '-':
		if i+1 == len(s) || s[i+1] < '0' || s[i+1] > '9' {
			return parseSymbol(s, i)
		}
		return parseNumber(s, i)
	case '0':
		fallthrough
	case '1':
		fallthrough
	case '2':
		fallthrough
	case '3':
		fallthrough
	case '4':
		fallthrough
	case '5':
		fallthrough
	case '6':
		fallthrough
	case '7':
		fallthrough
	case '8':
		fallthrough
	case '9':
		return parseNumber(s, i)

	case ')':
		fallthrough
	case '[':
		fallthrough
	case ']':
		fallthrough
	case '{':
		fallthrough
	case '}':
		fallthrough
	case ',':
		fallthrough
	case '.':
		return nil, i, errors.New(fmt.Sprintf("Syntax error. Unexpected character '%c'", b))

	default:
		return parseSymbol(s, i)
	}
}
