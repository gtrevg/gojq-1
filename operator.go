package gojq

import (
	"reflect"
	"strings"
)

// Operator ...
type Operator int

// Operators ...
const (
	OpAdd Operator = iota
	OpSub
	OpMul
	OpDiv
)

var operatorMap = map[string]Operator{
	"+": OpAdd,
	"-": OpSub,
	"*": OpMul,
	"/": OpDiv,
}

// Capture implements  participle.Capture.
func (op *Operator) Capture(s []string) error {
	*op = operatorMap[s[0]]
	return nil
}

// String implements Stringer.
func (op Operator) String() string {
	switch op {
	case OpAdd:
		return "+"
	case OpSub:
		return "-"
	case OpMul:
		return "*"
	case OpDiv:
		return "/"
	}
	panic(op)
}

// Eval the expression.
func (op Operator) Eval(l, r interface{}) interface{} {
	switch op {
	case OpAdd:
		return funcOpAdd(l, r)
	case OpSub:
		return funcOpSub(l, r)
	case OpMul:
		return funcOpMul(l, r)
	case OpDiv:
		return funcOpDiv(l, r)
	}
	panic("unsupported operator")
}

func binopTypeSwitch(
	l, r interface{},
	callbackInts func(int, int) interface{},
	callbackFloats func(float64, float64) interface{},
	callbackStrings func(string, string) interface{},
	callbackArrays func(l, r []interface{}) interface{},
	callbackMaps func(l, r map[string]interface{}) interface{},
	fallback func(interface{}, interface{}) interface{}) interface{} {
	switch l := l.(type) {
	case int:
		switch r := r.(type) {
		case int:
			return callbackInts(l, r)
		case float64:
			return callbackFloats(float64(l), r)
		default:
			return fallback(l, r)
		}
	case float64:
		switch r := r.(type) {
		case int:
			return callbackFloats(l, float64(r))
		case float64:
			return callbackFloats(l, r)
		default:
			return fallback(l, r)
		}
	case string:
		switch r := r.(type) {
		case string:
			return callbackStrings(l, r)
		default:
			return fallback(l, r)
		}
	case []interface{}:
		switch r := r.(type) {
		case []interface{}:
			return callbackArrays(l, r)
		default:
			return fallback(l, r)
		}
	case map[string]interface{}:
		switch r := r.(type) {
		case map[string]interface{}:
			return callbackMaps(l, r)
		default:
			return fallback(l, r)
		}
	default:
		return fallback(l, r)
	}
}

func funcOpAdd(l, r interface{}) interface{} {
	if l == nil {
		return r
	} else if r == nil {
		return l
	}
	return binopTypeSwitch(l, r,
		func(l, r int) interface{} { return l + r },
		func(l, r float64) interface{} { return l + r },
		func(l, r string) interface{} { return l + r },
		func(l, r []interface{}) interface{} { return append(l, r...) },
		func(l, r map[string]interface{}) interface{} {
			m := make(map[string]interface{})
			for k, v := range l {
				m[k] = v
			}
			for k, v := range r {
				m[k] = v
			}
			return m
		},
		func(l, r interface{}) interface{} { return &binopTypeError{"add", l, r} },
	)
}

func funcOpSub(l, r interface{}) interface{} {
	return binopTypeSwitch(l, r,
		func(l, r int) interface{} { return l - r },
		func(l, r float64) interface{} { return l - r },
		func(l, r string) interface{} { return &binopTypeError{"subtract", l, r} },
		func(l, r []interface{}) interface{} {
			a := make([]interface{}, 0, len(l))
			for _, v := range l {
				var found bool
				for _, w := range r {
					if reflect.DeepEqual(v, w) {
						found = true
						break
					}
				}
				if !found {
					a = append(a, v)
				}
			}
			return a
		},
		func(l, r map[string]interface{}) interface{} { return &binopTypeError{"subtract", l, r} },
		func(l, r interface{}) interface{} { return &binopTypeError{"subtract", l, r} },
	)
}

func funcOpMul(l, r interface{}) interface{} {
	return binopTypeSwitch(l, r,
		func(l, r int) interface{} { return l * r },
		func(l, r float64) interface{} { return l * r },
		func(l, r string) interface{} { return &binopTypeError{"multiply", l, r} },
		func(l, r []interface{}) interface{} { return &binopTypeError{"multiply", l, r} },
		func(l, r map[string]interface{}) interface{} {
			m := make(map[string]interface{})
			for k, v := range l {
				m[k] = v
			}
			for k, v := range r {
				m[k] = v
			}
			return m
		},
		func(l, r interface{}) interface{} {
			multiplyString := func(s string, cnt float64) interface{} {
				if cnt < 0.0 {
					return nil
				}
				if cnt < 1.0 {
					return l
				}
				return strings.Repeat(s, int(cnt))
			}
			if l, ok := l.(string); ok {
				switch r := r.(type) {
				case int:
					return multiplyString(l, float64(r))
				case float64:
					return multiplyString(l, r)
				}
			}
			if r, ok := r.(string); ok {
				switch l := l.(type) {
				case int:
					return multiplyString(r, float64(l))
				case float64:
					return multiplyString(r, l)
				}
			}
			return &binopTypeError{"multiply", l, r}
		},
	)
}

func funcOpDiv(l, r interface{}) interface{} {
	return binopTypeSwitch(l, r,
		func(l, r int) interface{} { return l / r },
		func(l, r float64) interface{} { return l / r },
		func(l, r string) interface{} { return strings.Split(l, r) },
		func(l, r []interface{}) interface{} { return &binopTypeError{"divide", l, r} },
		func(l, r map[string]interface{}) interface{} { return &binopTypeError{"divide", l, r} },
		func(l, r interface{}) interface{} { return &binopTypeError{"divide", l, r} },
	)
}