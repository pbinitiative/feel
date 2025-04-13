package feel

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"math"
	"sort"
	"strings"
)

func toFEELIndex(idx int) int {
	return idx + 1
}

func fromFEELIndex(idx int) int {
	return idx - 1
}

func decodeKWArgs(input map[string]any, output any) error {
	config := &mapstructure.DecoderConfig{
		Metadata: nil,
		TagName:  "json",
		Result:   &output,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}

func extractList(args map[string]any, argName string) ([]any, error) {
	if v, ok := args[argName]; ok {
		if listV, ok := v.([]any); ok {
			if len(listV) == 1 {
				if aList, ok := listV[0].([]any); ok {
					return aList, nil
				}
			}
			return listV, nil
		} else {
			return nil, NewErrTypeMismatch("list")
		}
	} else {
		return nil, NewErrKeyNotFound(argName)
	}
}

func installBuiltinFunctions(prelude *Prelude) {
	// conversion functions
	prelude.Bind("string", wrapTyped(func(v any) (string, error) {
		return fmt.Sprintf("%s", v), nil
	}).Required("from"))

	prelude.Bind("number", wrapTyped(func(v any) (*Number, error) {
		return ParseNumberWithErr(v)
	}).Required("from"))

	// boolean functions
	prelude.Bind("not", wrapTyped(func(v any) (bool, error) {
		return !boolValue(v), nil
	}).Required("from"))

	prelude.Bind("is defined", NewMacro(func(intp *Interpreter, args map[string]Node, varArgs []Node) (any, error) {
		if varNode, ok := args["value"].(*Var); ok {
			if _, ok := intp.Resolve(varNode.Name); !ok {
				return false, nil
			}
		} else {
			_, err := args["value"].Eval(intp)
			if err != nil {
				var evalErr *EvalError
				if errors.As(err, &evalErr) {
					if evalErr.Short == "key not found" || evalErr.Short == "index error" {
						return false, nil
					}
				}
				return false, err
			}
		}
		// TODO: more condition tests
		return true, nil
	}).Required("value"))

	// string functions
	prelude.Bind("string length", wrapTyped(func(s string) (int, error) {
		return len(s), nil
	}).Required("string"))

	prelude.Bind("substring", NewNativeFunc(func(kwargs map[string]any) (any, error) {
		type substringArgs struct {
			Str      string  `json:"string"`
			StartPos *Number `json:"start position"`
			Length   *Number `json:"length,omitempty"`
		}

		args := substringArgs{}
		if err := decodeKWArgs(kwargs, &args); err != nil {
			return nil, err
		}
		startPos := fromFEELIndex(args.StartPos.Int())
		if startPos >= len(args.Str) {
			return "", nil
		}
		endPos := len(args.Str)
		if args.Length != nil {
			endPos = startPos + int(args.Length.Int64())
			if endPos >= len(args.Str) {
				endPos = len(args.Str)
			}
		}
		subs := args.Str[startPos:endPos]
		return subs, nil
	}).Required("string", "start position").Optional("length"))

	prelude.Bind("upper case", wrapTyped(func(s string) (string, error) {
		return strings.ToUpper(s), nil
	}).Required("string"))

	prelude.Bind("lower case", wrapTyped(func(s string) (string, error) {
		return strings.ToLower(s), nil
	}).Required("string"))

	prelude.Bind("contains", wrapTyped(func(s string, match string) (bool, error) {
		return strings.Contains(s, match), nil
	}).Required("string", "match"))

	prelude.Bind("starts with", wrapTyped(func(s string, match string) (bool, error) {
		return strings.HasPrefix(s, match), nil
	}).Required("string", "match"))

	prelude.Bind("ends with", wrapTyped(func(s string, match string) (bool, error) {
		return strings.HasSuffix(s, match), nil
	}).Required("string", "match"))

	// list functions
	prelude.Bind("list contains", wrapTyped(func(list []any, elem any) (bool, error) {
		for _, entry := range list {
			if cmp, err := compareInterfaces(entry, elem); err == nil && cmp == 0 {
				return true, err
			}
		}
		return false, nil
	}).Required("list", "element"))

	prelude.Bind("count", NewNativeFunc(func(args map[string]any) (any, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		return len(list), nil
	}).Vararg("list"))

	prelude.Bind("min", NewNativeFunc(func(args map[string]any) (any, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		var minValue any = nil
		for _, entry := range list {
			if minValue == nil {
				minValue = entry
			} else if cmp, err := compareInterfaces(minValue, entry); err == nil && cmp == 1 {
				// minValue > entry
				minValue = entry
			}
		}
		return minValue, nil
	}).Vararg("list"))

	prelude.Bind("max", NewNativeFunc(func(args map[string]any) (any, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		var maxValue any = nil
		for _, entry := range list {
			if maxValue == nil {
				maxValue = entry
			} else if cmp, err := compareInterfaces(maxValue, entry); err == nil && cmp == -1 {
				// maxValue < entry
				maxValue = entry
			}
		}
		return maxValue, nil
	}).Vararg("list"))

	prelude.Bind("sum", NewNativeFunc(func(args map[string]any) (any, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		sum := Zero
		for _, entry := range list {
			if numEntry, ok := entry.(*Number); ok {
				sum = sum.Add(numEntry)
			}
		}
		return sum, nil
	}).Vararg("list"))

	prelude.Bind("product", NewNativeFunc(func(args map[string]any) (any, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		sum := N(1)
		for _, entry := range list {
			if numEntry, ok := entry.(*Number); ok {
				sum = sum.Mul(numEntry)
			}
		}
		return sum, nil
	}).Vararg("list"))

	prelude.Bind("mean", NewNativeFunc(func(args map[string]any) (any, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		sum := Zero
		cnt := 0
		for _, entry := range list {
			if numEntry, ok := entry.(*Number); ok {
				sum = sum.Add(numEntry)
				cnt++
			}
		}
		r := sum.FloatDiv(N(cnt))
		return r, nil
	}).Vararg("list"))

	prelude.Bind("stddev", NewNativeFunc(func(args map[string]any) (any, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		sum := 0.0
		cnt := 0
		for _, entry := range list {
			if numEntry, ok := entry.(*Number); ok {
				//sum = sum.Add(numEntry)
				sum += numEntry.Float64()
				cnt++
			}
		}
		if cnt == 0 {
			return 0, nil
		}

		avg := sum / float64(cnt)

		dev := 0.0
		for _, entry := range list {
			if numEntry, ok := entry.(*Number); ok {
				n := numEntry.Float64()
				dev += (n - avg) * (n - avg)
			}
		}

		return math.Sqrt(dev / float64(cnt)), nil
	}).Vararg("list"))

	prelude.Bind("median", NewNativeFunc(func(args map[string]any) (any, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		var numberList []*Number

		for _, entry := range list {
			if numEntry, ok := entry.(*Number); ok {
				numberList = append(numberList, numEntry)
			}
		}
		if len(numberList) == 0 {
			return nil, nil
		} else if len(numberList) == 1 {
			return numberList[0], nil
		}

		sort.Slice(numberList, func(i, j int) bool {
			return numberList[i].Compare(*numberList[j]) == -1
		})

		if len(numberList)%2 == 1 {
			return numberList[len(numberList)/2], nil
		} else {
			medPos := (len(numberList) / 2)
			return numberList[medPos+1].Add(numberList[medPos]).Mul(N("0.5")), nil
		}
	}).Vararg("list"))

	prelude.Bind("all", NewNativeFunc(func(args map[string]any) (any, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		for _, v := range list {
			if !boolValue(v) {
				return false, nil
			}
		}
		return true, nil
	}).Vararg("list"))

	prelude.Bind("and", NewNativeFunc(func(args map[string]any) (any, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		for _, v := range list {
			if !boolValue(v) {
				return false, nil
			}
		}
		return true, nil
	}).Vararg("list"))

	prelude.Bind("any", NewNativeFunc(func(args map[string]any) (any, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		for _, v := range list {
			if boolValue(v) {
				return true, nil
			}
		}
		return false, nil
	}).Vararg("list"))

	prelude.Bind("or", NewNativeFunc(func(args map[string]any) (any, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		for _, v := range list {
			if boolValue(v) {
				return true, nil
			}
		}
		return false, nil
	}).Vararg("list"))

	prelude.Bind("sublist", NewNativeFunc(func(kwargs map[string]any) (any, error) {
		type sublistArgs struct {
			List     []any   `json:"list"`
			StartPos *Number `json:"start position"`
			Length   *Number `json:"length,omitempty"`
		}

		args := sublistArgs{}
		if err := decodeKWArgs(kwargs, &args); err != nil {
			return nil, err
		}
		startPos := fromFEELIndex(args.StartPos.Int())
		if startPos >= len(args.List) {
			return "", nil
		}
		endPos := len(args.List)
		if args.Length != nil {
			endPos = startPos + int(args.Length.Int64())
			if endPos >= len(args.List) {
				endPos = len(args.List)
			}
		}
		subs := args.List[startPos:endPos]
		return subs, nil
	}).Required("list", "start position").Optional("length"))

	prelude.Bind("append", NewNativeFunc(func(kwargs map[string]any) (any, error) {
		type appendArgs struct {
			List  []any `json:"list"`
			Items []any `json:"items"`
		}
		args := appendArgs{}
		if err := decodeKWArgs(kwargs, &args); err != nil {
			return nil, err
		}
		return append(args.List, args.Items...), nil
	}).Required("list").Vararg("items"))

	prelude.Bind("concatenate", NewNativeFunc(func(kwargs map[string]any) (any, error) {
		type concatArgs struct {
			Lists [][]any `json:"lists"`
		}
		args := concatArgs{}
		if err := decodeKWArgs(kwargs, &args); err != nil {
			return nil, err
		}
		results := make([]any, 0)
		for _, list := range args.Lists {
			results = append(results, list...)
		}
		return results, nil
	}).Vararg("lists"))

	prelude.Bind("insert before", wrapTyped(func(list []any, pos *Number, newItem any) ([]any, error) {
		// The position starts at the index 1. The last position is -1
		position := fromFEELIndex(pos.Int())
		if position > len(list) {
			position = len(list)
		}
		// make a copy of the original list
		var tmpList []any
		tmpList = append(tmpList, list[:position]...)

		//
		newList := append(tmpList, newItem)
		newList = append(newList[:], list[position:]...)
		return newList, nil
	}).Required("list", "position", "newItem"))

	prelude.Bind("remove", wrapTyped(func(list []any, pos *Number) ([]any, error) {
		// The position starts at the index 1. The last position is -1
		position := fromFEELIndex(pos.Int())
		if position > len(list) {
			position = len(list)
		}
		// make a copy of the original list
		var tmpList []any
		tmpList = append(tmpList, list[:position]...)
		//
		newList := append(tmpList, list[(position+1):]...)
		return newList, nil
	}).Required("list", "position"))

	prelude.Bind("reverse", wrapTyped(func(list []any) ([]any, error) {
		var reversed []any
		for i := len(list) - 1; i >= 0; i-- {
			reversed = append(reversed, list[i])
		}
		return reversed, nil
	}).Required("list"))

	prelude.Bind("index of", wrapTyped(func(list []any, match any) ([]any, error) {
		matched := make([]any, 0)
		for i, elem := range list {
			if cmp, err := compareInterfaces(elem, match); err == nil && cmp == 0 {
				matched = append(matched, N(toFEELIndex(i)))
			}
		}
		return matched, nil
	}).Required("list", "match"))

	prelude.Bind("union", NewNativeFunc(func(kwargs map[string]any) (any, error) {
		type unionArgs struct {
			Lists [][]any `json:"lists"`
		}
		args := unionArgs{}
		if err := decodeKWArgs(kwargs, &args); err != nil {
			return nil, err
		}
		elemSet := make([]any, 0)
		for _, list := range args.Lists {
			for _, elem := range list {
				found := false
				for _, setElem := range elemSet {
					if cmp, err := compareInterfaces(elem, setElem); err == nil && cmp == 0 {
						found = true
						break
					}
				}
				if !found {
					elemSet = append(elemSet, elem)
				}
			}
		}
		return elemSet, nil
	}).Vararg("lists"))

	prelude.Bind("distinct values", wrapTyped(func(list []any) ([]any, error) {
		elemSet := make([]any, 0)
		for _, elem := range list {
			found := false
			for _, setElem := range elemSet {
				if cmp, err := compareInterfaces(elem, setElem); err == nil && cmp == 0 {
					found = true
					break
				}
			}
			if !found {
				elemSet = append(elemSet, elem)
			}
		}
		return elemSet, nil
	}).Required("list"))

	prelude.Bind("flatten", wrapTyped(func(list []any) ([]any, error) {
		flattened := make([]any, 0)
		var flattenInterface func(v any)
		flattenInterface = func(v any) {
			if arr, ok := v.([]any); ok {
				for _, a := range arr {
					flattenInterface(a)
				}
			} else {
				flattened = append(flattened, v)
			}
		}
		for _, elem := range list {
			flattenInterface(elem)
		}
		return flattened, nil
	}).Required("list"))

	prelude.Bind("sort", NewMacro(func(intp *Interpreter, args map[string]Node, varargs []Node) (any, error) {
		vlist, err := args["list"].Eval(intp)
		if err != nil {
			return nil, err
		}
		list, ok := vlist.([]any)
		if !ok {
			//return nil, NewEvalError(-4080, "the first argument is not list")
			return nil, NewErrTypeMismatch("list")
		}

		vpred, err := args["predicates"].Eval(intp)
		if err != nil {
			return nil, err
		}
		predicates, ok := vpred.(*FunDef)
		if !ok {
			//return nil, NewEvalError(-4080, "the second argument is not function")
			return nil, NewErrTypeMismatch("function")
		}

		newList := append([]any{}, list...)
		var sortErr error
		sort.Slice(newList, func(i, j int) bool {
			r, err := predicates.EvalCall(intp, []any{newList[i], newList[j]})
			if err != nil {
				//panic(err)
				sortErr = err
				return false
			}
			return boolValue(r)
		})
		if sortErr != nil {
			return nil, sortErr
		}
		return newList, nil
	}).Required("list", "predicates"))

	prelude.Bind("string join", NewNativeFunc(func(kwargs map[string]any) (any, error) {
		type joinArgs struct {
			List      []any  `json:"list"`
			Delimiter string `json:"delimiter,omitempty"`
			Prefix    string `json:"prefix,omitempty"`
			Suffix    string `json:"suffix,omitempty"`
		}
		args := joinArgs{}
		if err := decodeKWArgs(kwargs, &args); err != nil {
			return nil, err
		}
		strArray := make([]string, 0)
		for _, v := range args.List {
			if s, ok := v.(string); ok {
				strArray = append(strArray, s)
			} else if _, ok := v.(*NullValue); !ok {
				return null, nil
			}
		}
		joined := fmt.Sprintf("%s%s%s", args.Prefix, strings.Join(strArray, args.Delimiter), args.Suffix)
		return joined, nil
	}).Required("list").Optional("delimiter", "prefix", "suffix"))

}
