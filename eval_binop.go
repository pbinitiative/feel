package feel

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
)

func (self Binop) Eval(intp *Interpreter) (any, error) {
	switch self.Op {
	case "and":
		return self.andOp(intp)
	case "or":
		return self.orOp(intp)
	case "+":
		return self.addOp(intp)
	case "-":
		return self.subOp(intp)
	case "*":
		return self.mulOp(intp)
	case "/":
		return self.divOp(intp)
	case "%":
		return self.modOp(intp)
	case ">":
		return self.compareGTOp(intp)
	case ">=":
		return self.compareGEOp(intp)
	case "<":
		return self.compareLTOp(intp)
	case "<=":
		return self.compareLEOp(intp)
	case "!=":
		return self.notEqalOp(intp)
	case "[]":
		return self.indexAtOp(intp)
	case "=":
		return self.equalOp(intp)
	case "in":
		return self.inOp(intp)
	default:
		return nil, NewEvalError(-3000, "no such binary op", fmt.Sprintf("Binary op %s not exist or not supported", self.Op))
	}
}

type evalNumbers func(a, b *Number) any
type evalStrings func(a, b string) any

func (self Binop) numberOp(intp *Interpreter, en evalNumbers, op string) (any, error) {
	leftVal, err := self.Left.Eval(intp)
	if err != nil {
		return nil, err
	}
	rightVal, err := self.Right.Eval(intp)
	if err != nil {
		return nil, err
	}
	if leftNumber, ok := leftVal.(*Number); ok {
		if rightNumber, ok := rightVal.(*Number); ok {
			return en(leftNumber, rightNumber), nil
		}
	}
	return nil, NewEvalError(-3101, "invalid types", fmt.Sprintf("bad type in op, %T %s %T", leftVal, op, rightVal))
}

func CompareValues(leftVal, rightVal any) int {
	r, err := compareInterfaces(leftVal, rightVal)
	if err != nil {
		panic(err)
	} else {
		return r
	}
}

func compareInterfaces(leftVal, rightVal any) (int, error) {
	switch v := leftVal.(type) {
	case string:
		if rightString, ok := rightVal.(string); ok {
			return strings.Compare(v, rightString), nil
		}
	case *Number:
		if rightNumber, ok := rightVal.(*Number); ok {
			return v.Compare(*rightNumber), nil
		}
	case *NullValue:
		if _, ok := rightVal.(*NullValue); ok {
			return 0, nil
		}
	case bool:
		if rightBool, ok := rightVal.(bool); ok {
			if v == rightBool {
				return 0, nil
			} else if v == false {
				return -1, nil
			} else {
				return 1, nil
			}
		}
	case HasTime:
		if rightHasTime, ok := rightVal.(HasTime); ok {
			if v.Time().Equal(rightHasTime.Time()) {
				return 0, nil
			} else if v.Time().Before(rightHasTime.Time()) {
				return -1, nil
			} else {
				return 1, nil
			}
		}
	case []any:
		if rightArr, ok := rightVal.([]any); ok {
			return compareArrays(v, rightArr)
		}
	case map[string]any:
		if rightMap, ok := rightVal.(map[string]any); ok {
			return compareMaps(v, rightMap)
		}
	}
	return 0, NewEvalError(-3106, "invalid types", fmt.Sprintf("bad type in comparation, %T vs. %T", leftVal, rightVal))
}

func compareArrays(a, b []any) (int, error) {
	minSize := len(a)
	if minSize > len(b) {
		minSize = len(b)
	}
	for i := 0; i < minSize; i++ {
		leftVal := a[i]
		rightVal := b[i]
		r, err := compareInterfaces(leftVal, rightVal)
		if err != nil {
			return 0, err
		}
		if r != 0 {
			return r, nil
		}
	}
	if len(b) > minSize {
		return -1, nil
	} else {
		return 1, nil
	}
}

func compareMaps(a, b map[string]any) (int, error) {
	if len(a) > len(b) {
		return 1, nil
	} else if len(a) < len(b) {
		return -1, nil
	}
	for k, leftVal := range a {
		if rightVal, ok := b[k]; ok {
			r, err := compareInterfaces(leftVal, rightVal)
			if err != nil {
				return 0, err
			}
			if r != 0 {
				return r, nil
			}
		} else {
			return 1, nil
		}
	}
	return 0, nil
}

func (self Binop) compareValues(intp *Interpreter) (int, error) {
	leftVal, err := self.Left.Eval(intp)
	if err != nil {
		return 0, err
	}
	rightVal, err := self.Right.Eval(intp)
	if err != nil {
		return 0, err
	}
	return compareInterfaces(leftVal, rightVal)
}

func (self Binop) typedOp(intp *Interpreter, es evalStrings, en evalNumbers, op string) (any, error) {
	leftVal, err := self.Left.Eval(intp)
	if err != nil {
		return nil, err
	}
	rightVal, err := self.Right.Eval(intp)
	if err != nil {
		return nil, err
	}

	switch v := leftVal.(type) {
	case string:
		if es != nil {
			if rightString, ok := rightVal.(string); ok {
				return es(v, rightString), nil
			}
		}
	case *Number:
		if en != nil {
			if rightNumber, ok := rightVal.(*Number); ok {
				return en(v, rightNumber), nil
			}
		}
	case *FEELDatetime:
		if op == "+" {
			if rightDur, ok := rightVal.(*FEELDuration); ok {
				return v.Add(rightDur), nil
			}
		} else if op == "-" {
			if rightDur, ok := rightVal.(*FEELDuration); ok {
				return v.Add(rightDur.Negative()), nil
			} else if rightTime, ok := rightVal.(HasTime); ok {
				return v.Sub(rightTime), nil
			}
		}
	}
	return nil, NewEvalError(-3101, "invalid types", fmt.Sprintf("bad type in op, %T %s %T", leftVal, op, rightVal))
}

func (self Binop) addOp(intp *Interpreter) (any, error) {
	return self.typedOp(
		intp,
		func(a, b string) any { return a + b },
		func(a, b *Number) any { return a.Add(b) },
		"+",
	)
}

func (self Binop) subOp(intp *Interpreter) (any, error) {
	return self.typedOp(
		intp,
		nil,
		func(a, b *Number) any { return a.Sub(b) },
		"-")
}

func (self Binop) mulOp(intp *Interpreter) (any, error) {
	return self.numberOp(
		intp,
		func(a, b *Number) any { return a.Mul(b) },
		"*")
}

func (self Binop) divOp(intp *Interpreter) (any, error) {
	return self.numberOp(
		intp,
		func(a, b *Number) any { return a.IntDiv(b) },
		"/")
}

func (self Binop) compareGTOp(intp *Interpreter) (any, error) {
	r, err := self.compareValues(intp)
	if err != nil {
		return false, err
	} else {
		return r > 0, nil
	}
}

func (self Binop) compareGEOp(intp *Interpreter) (any, error) {
	r, err := self.compareValues(intp)
	if err != nil {
		return false, err
	} else {
		return r >= 0, nil
	}
}

func (self Binop) compareLTOp(intp *Interpreter) (any, error) {
	r, err := self.compareValues(intp)
	if err != nil {
		return false, err
	} else {
		return r < 0, nil
	}
}

func (self Binop) compareLEOp(intp *Interpreter) (any, error) {
	r, err := self.compareValues(intp)
	if err != nil {
		return false, err
	} else {
		return r <= 0, nil
	}
}

func (self Binop) equalOp(intp *Interpreter) (any, error) {
	r, err := self.compareValues(intp)
	if err != nil {
		return false, err
	} else {
		return r == 0, nil
	}
}

func (self Binop) notEqalOp(intp *Interpreter) (any, error) {
	r, err := self.compareValues(intp)
	if err != nil {
		var evalError *EvalError
		if errors.As(err, &evalError) && evalError.Code == -3106 {
			// type mismatch
			return false, nil
		}
		return false, err
	} else {
		return r != 0, nil
	}
}

func (self Binop) modOp(intp *Interpreter) (any, error) {
	return self.numberOp(
		intp,
		func(a, b *Number) any { return a.IntMod(b) },
		"%")
}

// circuit break operators
func (self Binop) andOp(intp *Interpreter) (any, error) {
	leftVal, err := self.Left.Eval(intp)
	if err != nil {
		return nil, err
	}
	leftBool := boolValue(leftVal)
	if !leftBool {
		return false, nil
	}
	rightVal, err := self.Right.Eval(intp)
	if err != nil {
		return nil, err
	}
	rightBool := boolValue(rightVal)
	return rightBool, nil
}

func (self Binop) orOp(intp *Interpreter) (any, error) {
	leftVal, err := self.Left.Eval(intp)
	if err != nil {
		return nil, err
	}
	leftBool := boolValue(leftVal)
	if leftBool {
		return true, nil
	}
	rightVal, err := self.Right.Eval(intp)
	if err != nil {
		return nil, err
	}
	rightBool := boolValue(rightVal)
	return rightBool, nil
}

func (self Binop) indexAtOp(intp *Interpreter) (any, error) {
	leftVal, err := self.Left.Eval(intp)
	if err != nil {
		return nil, err
	}
	rightVal, err := self.Right.Eval(intp)
	if err != nil {
		return nil, err
	}
	switch v := leftVal.(type) {
	case []any:
		if nRight, ok := rightVal.(*Number); ok {
			return v[nRight.Int64()], nil
		} else {
			return nil, NewEvalError(-3200, "non int index")
		}
	case map[string]any:
		if strRight, ok := rightVal.(string); ok {
			if elem, ok := v[strRight]; ok {
				return elem, nil
			} else {
				return nil, NewEvalError(-3201, "key not found")
			}
		} else {
			return nil, NewEvalError(-3200, "non string index")
		}
	default:
		return nil, NewEvalError(-3202, "non indexable value")
	}
}

func (self Binop) inOp(intp *Interpreter) (any, error) {
	leftVal, err := self.Left.Eval(intp)
	if err != nil {
		return nil, err
	}
	rightVal, err := self.Right.Eval(intp)
	if err != nil {
		return nil, err
	}
	switch rv := rightVal.(type) {
	case *RangeValue:
		return rv.Contains(leftVal), nil
	case []any:
		for _, kv := range rv {
			if cmp.Equal(leftVal, kv) {
				return true, nil
			}
		}
		return false, nil
	default:
		return nil, NewEvalError(-3202, "non in value")
	}
}
