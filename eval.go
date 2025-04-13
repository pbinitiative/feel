package feel

import (
	"encoding/json"
	"errors"
)

// values
var null = &NullValue{}

type NullValue struct {
}

func (self NullValue) Equal(other NullValue) bool {
	return true
}

func (self NullValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(nil)
}

func boolValue(condVal any) bool {
	switch v := condVal.(type) {
	case int64:
		return v != 0
	case float64:
		return v != 0.0
	case *Number:
		return !v.Equal(*Zero)
	case bool:
		return v
	case string:
		return v != ""
	case []any:
		return v != nil && len(v) > 0
	case map[string]any:
		return v != nil && len(v) > 0
	default:
		return v != nil
	}
}

func typeName(a any) string {
	switch a.(type) {
	case int64:
		return "number"
	case float64:
		return "number"
	case *Number:
		return "number"
	case bool:
		return "bool"
	case string:
		return "string"
	case []any:
		return "list"
	case map[string]any:
		return "context"
	case *NullValue:
		return "null"
	case *FEELDate:
		return "date"
	case *FEELTime:
		return "time"
	case *FEELDatetime:
		return "datetime"
	case *FEELDuration:
		return "duration"
	case *RangeValue:
		return "range"
	case *NativeFun:
		return "function"
	case *FunDef:
		return "function"
	case *Macro:
		return "function"
	default:
		return "unknown"
	}
}

func normalizeValue(v any) any {
	if v == nil {
		return null
	}
	switch vv := v.(type) {
	case int:
		return NewNumberFromInt64(int64(vv))
	case int64:
		return NewNumberFromInt64(vv)
	case float64:
		return NewNumberFromFloat(vv)
	default:
		return vv
	}
}

func normalizeScope(scope Scope) Scope {
	newScope := make(Scope)
	for key, value := range scope {
		if nestedScope, ok := value.(map[string]interface{}); ok {
			newScope[key] = normalizeScope(nestedScope)
		} else {
			newScope[key] = normalizeValue(value)
		}
	}
	return newScope
}

func NewIntepreter() *Interpreter {
	intp := &Interpreter{}
	intp.PushEmpty()
	return intp
}

func (i Interpreter) String() string {
	return "interpreter"
}

func (i Interpreter) Len() int {
	return len(i.ScopeStack)
}

func (i *Interpreter) Push(scp Scope) {
	i.ScopeStack = append(i.ScopeStack, normalizeScope(scp))
}

func (i *Interpreter) PushEmpty() {
	vars := make(Scope)
	i.Push(vars)
}

func (i *Interpreter) Pop() Scope {
	if i.Len() > 0 {
		top := i.ScopeStack[len(i.ScopeStack)-1]
		i.ScopeStack = i.ScopeStack[:len(i.ScopeStack)-1]
		return top
	}
	return nil
}

// Resolve a name from the top of scopestack to bottom
func (i Interpreter) Resolve(name string) (any, bool) {
	for at := len(i.ScopeStack) - 1; at >= 0; at-- {
		if v, ok := i.ScopeStack[at][name]; ok {
			return v, true
		}
	}
	if prelude, ok := GetPrelude().Resolve(name); ok {
		return prelude, ok
	}
	return nil, false
}

// Set the name and set to new value
func (i Interpreter) Set(name string, value any) bool {
	for at := len(i.ScopeStack) - 1; at >= 0; at-- {
		if _, ok := i.ScopeStack[at][name]; ok {
			i.ScopeStack[at][name] = value
			return true
		}
	}
	return false
}

// Bind the value to the name of current scope
func (i *Interpreter) Bind(name string, value any) {
	if i.Len() > 0 {
		i.ScopeStack[i.Len()-1][name] = normalizeValue(value)
	} else {
		panic("empty stack")
	}
}

// Node's eval functions

// Eval Evaluate Number node
func (numberNode NumberNode) Eval(intp *Interpreter) (any, error) {
	return NewNumber(numberNode.Value), nil
}

// Eval Evaluate bool node
func (boolNode BoolNode) Eval(intp *Interpreter) (any, error) {
	return boolNode.Value, nil
}

func (nullNode NullNode) Eval(intp *Interpreter) (any, error) {
	return null, nil
}

func (stringNode StringNode) Eval(intp *Interpreter) (any, error) {
	return stringNode.Content(), nil
}

func (tempNode TemporalNode) Eval(intp *Interpreter) (any, error) {
	return ParseTemporalValue(tempNode.Content())
}

func (v Var) Eval(intp *Interpreter) (any, error) {
	if v, ok := intp.Resolve(v.Name); ok {
		return v, nil
	} else {
		//return nil, NewErrKeyNotFound(v.Name)
		return null, nil
	}
}

func (rangeNode RangeNode) Eval(intp *Interpreter) (any, error) {
	startVal, err := rangeNode.Start.Eval(intp)
	if err != nil {
		return nil, err
	}
	endVal, err := rangeNode.End.Eval(intp)
	if err != nil {
		return nil, err
	}
	return &RangeValue{
		Start:     startVal,
		StartOpen: rangeNode.StartOpen,
		End:       endVal,
		EndOpen:   rangeNode.EndOpen,
	}, nil
}

func (arrNode ArrayNode) Eval(intp *Interpreter) (any, error) {
	var arr []any
	for _, elem := range arrNode.Elements {
		v, err := elem.Eval(intp)
		if err != nil {
			return nil, err
		}
		arr = append(arr, v)
	}
	return arr, nil
}

func (exprList ExprList) Eval(intp *Interpreter) (any, error) {
	var finalRet any = nil
	for _, elem := range exprList.Elements {
		v, err := elem.Eval(intp)
		if err != nil {
			return nil, err
		}
		finalRet = v
	}
	return finalRet, nil
}

func (mt MultiTests) Eval(intp *Interpreter) (any, error) {
	for _, elem := range mt.Elements {
		v, err := elem.Eval(intp)
		if err != nil {
			return nil, err
		}
		if boolValue(v) {
			return true, nil
		}
	}
	return false, nil
}

func (mapNode MapNode) Eval(intp *Interpreter) (any, error) {
	mapVal := make(map[string]any)
	for _, item := range mapNode.Values {

		v, err := item.Value.Eval(intp)
		if err != nil {
			return nil, err
		}
		mapVal[item.Name] = v
	}
	return mapVal, nil
}

func (dotop DotOp) Eval(intp *Interpreter) (any, error) {
	leftVal, err := dotop.Left.Eval(intp)
	if err != nil {
		return nil, err
	}
	if mapVal, ok := leftVal.(Scope); ok {
		if val, found := mapVal[dotop.Attr]; found {
			return val, nil
		} else {
			return nil, NewErrKeyNotFound(dotop.Attr)
		}
	} else if mapVal, ok = leftVal.(map[string]interface{}); ok {
		// todo: consolidate duplicate code as above, same as for 'Scope'
		if val, found := mapVal[dotop.Attr]; found {
			return val, nil
		} else {
			return nil, NewErrKeyNotFound(dotop.Attr)
		}
	} else if obj, ok := leftVal.(HasAttrs); ok {
		if v, found := obj.GetAttr(dotop.Attr); found {
			return normalizeValue(v), nil
		} else {
			//return nil, NewEvalError(-4001, "attr error", fmt.Sprintf("cannot get attr %s", dotop.Attr))
			return nil, NewErrKeyNotFound(dotop.Attr)

		}
	} else {
		return nil, NewErrTypeMismatch("map")
	}
}

func (ifExpr IfExpr) Eval(intp *Interpreter) (any, error) {
	condVal, err := ifExpr.Cond.Eval(intp)
	if err != nil {
		return nil, err
	}

	if boolValue(condVal) {
		brVal, err := ifExpr.ThenBranch.Eval(intp)
		if err != nil {
			return nil, err
		}
		return brVal, nil
	} else {
		brVal, err := ifExpr.ElseBranch.Eval(intp)
		if err != nil {
			return nil, err
		}
		return brVal, nil
	}
}

func (fe ForExpr) Eval(intp *Interpreter) (any, error) {
	listLike, err := fe.ListExpr.Eval(intp)
	if err != nil {
		return nil, err
	}

	if aList, ok := listLike.([]any); ok {
		intp.PushEmpty()
		results := make([]any, 0)
		for _, val := range aList {
			intp.Bind(fe.Varname, val)

			res, err := fe.ReturnExpr.Eval(intp)
			if err != nil {
				intp.Pop()
				return nil, err
			}
			results = append(results, res)
		}
		intp.Pop()
		return results, nil
	} else {
		return nil, NewErrTypeMismatch("list")
	}
}

func (sexpr SomeExpr) Eval(intp *Interpreter) (any, error) {
	listLike, err := sexpr.ListExpr.Eval(intp)
	if err != nil {
		return nil, err
	}

	if aList, ok := listLike.([]any); ok {
		intp.PushEmpty()
		for _, val := range aList {
			intp.Bind(sexpr.Varname, val)

			res, err := sexpr.FilterExpr.Eval(intp)
			if err != nil {
				intp.Pop()
				return nil, err
			}
			if boolValue(res) {
				return val, nil
			}
		}
		intp.Pop()
		return nil, nil
	} else {
		return nil, NewErrTypeMismatch("list")
	}
}

func (ee EveryExpr) Eval(intp *Interpreter) (any, error) {
	listLike, err := ee.ListExpr.Eval(intp)
	if err != nil {
		return nil, err
	}

	if aList, ok := listLike.([]any); ok {
		intp.PushEmpty()
		chooses := make([]any, 0)
		for _, val := range aList {
			intp.Bind(ee.Varname, val)

			res, err := ee.FilterExpr.Eval(intp)
			if err != nil {
				intp.Pop()
				return nil, err
			}

			if boolValue(res) {
				chooses = append(chooses, val)
			}
		}
		intp.Pop()
		return chooses, nil
	} else {
		return nil, NewErrTypeMismatch("list")
	}
}

func (fd FunDef) Eval(intp *Interpreter) (any, error) {
	return &FunDef{
		Args: fd.Args,
		Body: fd.Body,
	}, nil
}

func (fd FunDef) EvalCall(intp *Interpreter, args []any) (any, error) {
	if len(args) != len(fd.Args) {
		return nil, errors.New("eval call argument size mismatch")
	}
	intp.PushEmpty()
	defer intp.Pop()
	for i, argName := range fd.Args {
		intp.Bind(argName, args[i])
	}
	return fd.Body.Eval(intp)
}

func (fc FunCall) Eval(intp *Interpreter) (any, error) {
	v, err := fc.FunRef.Eval(intp)
	if err != nil {
		return nil, err
	}
	switch r := v.(type) {
	case *FunDef:
		return fc.EvalFunDef(intp, r)
	case *NativeFun:
		return fc.EvalNativeFun(intp, r)
	case *Macro:
		return fc.EvalMacro(intp, r)
	default:
		return nil, NewErrTypeMismatch("function")
	}
}

func (fc FunCall) EvalNativeFun(intp *Interpreter, funDef *NativeFun) (any, error) {
	argVals := make(map[string]any)
	if fc.keywordArgs {
		kwArgMap, err := fc.evalArgsToMap(intp)
		if err != nil {
			return nil, err
		}

		for _, argName := range funDef.requiredArgNames {
			if v, ok := kwArgMap[argName]; ok {
				argVals[argName] = v
			} else {
				//return nil, NewEvalError(-5001, "no keyword argument", fmt.Sprintf("no keyword argument %s", argName))
				return nil, NewErrKeywordArgument(argName)
			}
		}

		for _, argName := range funDef.optionalArgNames {
			if v, ok := kwArgMap[argName]; ok {
				argVals[argName] = v
			}
		}
	} else {
		if len(fc.Args) < len(funDef.requiredArgNames) {
			required := funDef.requiredArgNames[len(fc.Args):len(funDef.requiredArgNames)]
			return nil, NewErrTooFewArguments(required)
		}
		for i, argNode := range fc.Args {
			a, err := argNode.arg.Eval(intp)
			if err != nil {
				return nil, err
			}
			if i < len(funDef.requiredArgNames) {
				argVals[funDef.requiredArgNames[i]] = a
			} else if i < len(funDef.requiredArgNames)+len(funDef.optionalArgNames) {
				argVals[funDef.optionalArgNames[i-len(funDef.requiredArgNames)]] = a
			} else if funDef.varArgName != "" {
				if vars, ok := argVals[funDef.varArgName]; ok {
					varargs := vars.([]any)
					varargs = append(varargs, a)
					argVals[funDef.varArgName] = varargs
				} else {
					argVals[funDef.varArgName] = []any{a}
				}
			} else {
				//return nil, NewEvalError(-5002, "too many arguments")
				return nil, NewErrTooManyArguments()
			}
		}
	}
	return funDef.Call(intp, argVals)
}

func (fc FunCall) evalArgsToMap(intp *Interpreter) (map[string]any, error) {
	if !fc.keywordArgs {
		return nil, errors.New("funcall has no keyword args")
	}
	kwArgMap := make(map[string]any)
	for _, argNode := range fc.Args {
		a, err := argNode.arg.Eval(intp)
		if err != nil {
			return nil, err
		}
		kwArgMap[argNode.argName] = a
	}
	return kwArgMap, nil
}

func (fc FunCall) EvalMacro(intp *Interpreter, macro *Macro) (any, error) {
	if len(macro.requiredArgNames) > len(fc.Args) {
		return nil, NewErrTooFewArguments(macro.requiredArgNames[len(fc.Args):])
		//return nil, NewEvalError(-1005, "number of args of macro mismatch")
	}

	argNodes := make(map[string]Node)
	var varArgs []Node
	if fc.keywordArgs {
		kwArgMap := make(map[string]Node)
		for _, argNode := range fc.Args {
			kwArgMap[argNode.argName] = argNode.arg
		}

		for _, argName := range macro.requiredArgNames {
			if ast, ok := kwArgMap[argName]; ok {
				argNodes[argName] = ast
			} else {
				//return nil, NewEvalError(-5001, "no keyword argument", fmt.Sprintf("no keyword argument %s", argName))
				return nil, NewErrKeywordArgument(argName)
			}
		}

		for _, argName := range macro.optionalArgNames {
			if ast, ok := kwArgMap[argName]; ok {
				argNodes[argName] = ast
			}
		}
	} else {
		if len(fc.Args) < len(macro.requiredArgNames) {
			//reqArgs := strings.Join(macro.requiredArgNames[len(fc.Args):len(macro.requiredArgNames)], ", ")
			//return nil, NewEvalError(-5003, "too few arguments", fmt.Sprintf("more arguments required: %s", reqArgs))
			return nil, NewErrTooFewArguments(macro.requiredArgNames[len(fc.Args):])
		}
		for i, argNode := range fc.Args {
			if i < len(macro.requiredArgNames) {
				argNodes[macro.requiredArgNames[i]] = argNode.arg
			} else if i < len(macro.requiredArgNames)+len(macro.optionalArgNames) {
				argNodes[macro.optionalArgNames[i-len(macro.requiredArgNames)]] = argNode.arg
			} else if macro.varArgName != "" {
				varArgs = append(varArgs, argNode.arg)
			} else {
				//return nil, NewEvalError(-5002, "too many arguments")
				return nil, NewErrTooManyArguments()
			}
		}
	}
	return macro.fn(intp, argNodes, varArgs)
}

func (fc FunCall) EvalFunDef(intp *Interpreter, funDef *FunDef) (any, error) {
	if len(funDef.Args) > len(fc.Args) {
		//return nil, NewEvalError(-1004, "number of args mismatch")
		return nil, NewErrTooFewArguments(funDef.Args[len(fc.Args):])
	} else if len(funDef.Args) < len(fc.Args) {
		return nil, NewErrTooManyArguments()
	}
	//var args []any
	intp.PushEmpty()
	defer intp.Pop()

	if fc.keywordArgs {
		kwArgMap, err := fc.evalArgsToMap(intp)
		if err != nil {
			return nil, err
		}

		for _, argName := range funDef.Args {
			if v, ok := kwArgMap[argName]; ok {
				//argVals = append(argVals, v)
				intp.Bind(argName, v)
			} else {
				//return nil, NewEvalError(-5001, "no keyword argument", fmt.Sprintf("no keyword argument %s", argName))
				intp.Bind(argName, null)
			}
		}
	} else {
		for i, argNode := range fc.Args {
			a, err := argNode.arg.Eval(intp)
			if err != nil {
				return nil, err
			}
			intp.Bind(funDef.Args[i], a)
		}
	}
	ret, err := funDef.Body.Eval(intp)
	return ret, err
}

func EvalString(input string) (any, error) {
	return EvalStringWithScope(input, nil)
}

func EvalStringWithScope(input string, scope Scope) (any, error) {
	ast, err := ParseString(input)
	if err != nil {
		return nil, err
	}
	intp := NewIntepreter()
	if scope != nil {
		intp.Push(scope)
	}
	r, err := ast.Eval(intp)
	return r, err
}
