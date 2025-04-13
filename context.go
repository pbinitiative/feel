package feel

func contextGetByKeys(ctx map[string]any, keys []string) (any, bool) {
	for i, key := range keys {
		if i == len(keys)-1 {
			v, ok := ctx[key]
			return v, ok
		} else {
			v, ok := ctx[key]
			if !ok {
				return nil, false
			}
			if subctx, ok := v.(map[string]any); ok {
				ctx = subctx
			} else {
				return nil, false
			}
		}
	}
	return nil, false
}

func contextProbePut(ctx map[string]any, keys []string) bool {
	for i, key := range keys {
		if i == len(keys)-1 {
			return true
		} else {
			v, ok := ctx[key]
			if !ok {
				// empty cell can be put
				return true
			}
			if subctx, ok := v.(map[string]any); ok {
				ctx = subctx
			} else {
				// sub ctx is not map
				return false
			}
		}
	}
	return false
}

func contextPutKeys(ctx map[string]any, keys []string, value any) (map[string]any, bool) {
	if !contextProbePut(ctx, keys) {
		// cannot put keys
		return ctx, false
	}

	rootCtx := ctx
	for i, key := range keys {
		if i == len(keys)-1 {
			// the last key
			ctx[key] = value
			return rootCtx, true
		} else {
			if v, ok := ctx[key]; ok {
				if subctx, ok := v.(map[string]any); ok {
					ctx = subctx
				} else {
					return rootCtx, false
				}
			} else {
				subctx := make(map[string]any)
				ctx[key] = subctx
				ctx = subctx
			}
		}
	}
	return rootCtx, false
}

func installContextFunctions(prelude *Prelude) {
	// context/map functions
	prelude.Bind("get value", NewNativeFunc(func(kwargs map[string]any) (any, error) {
		type getvalueByKey struct {
			Context map[string]any `json:"context"`
			Key     string         `json:"key"`
		}

		type getvalueByKeys struct {
			Context map[string]any `json:"context"`
			Keys    []string       `json:"key"`
		}

		argsByKey := getvalueByKey{}

		if err := decodeKWArgs(kwargs, &argsByKey); err != nil {
			argsByKeys := getvalueByKeys{}
			if err := decodeKWArgs(kwargs, &argsByKeys); err != nil {
				return nil, err
			}

			if v, ok := contextGetByKeys(argsByKeys.Context, argsByKeys.Keys); ok {
				return v, nil
			} else {
				return null, nil
			}
		} else {
			if v, ok := argsByKey.Context[argsByKey.Key]; ok {
				return v, nil
			} else {
				return null, nil
			}
		}
	}).Required("context", "key"))

	prelude.Bind("get entries", wrapTyped(func(ctx map[string]any) ([](map[string]any), error) {
		entries := make([](map[string]any), 0)
		for k, v := range ctx {
			ent := map[string]any{
				"key":   k,
				"value": v,
			}
			entries = append(entries, ent)
		}
		return entries, nil
	}).Required("context"))

	prelude.Bind("context put", NewNativeFunc(func(kwargs map[string]any) (any, error) {
		type putByKey struct {
			Context map[string]any `json:"context"`
			Key     string         `json:"key"`
			Value   any            `json:"value"`
		}

		type putByKeys struct {
			Context map[string]any `json:"context"`
			Keys    []string       `json:"key"`
			Value   any            `json:"value"`
		}

		argsByKey := putByKey{}

		if err := decodeKWArgs(kwargs, &argsByKey); err != nil {
			argsByKeys := putByKeys{}
			if err := decodeKWArgs(kwargs, &argsByKeys); err != nil {
				return nil, err
			}
			ctx, _ := contextPutKeys(argsByKeys.Context, argsByKeys.Keys, argsByKeys.Value)
			return ctx, nil
		} else {
			ctx, _ := contextPutKeys(argsByKey.Context, []string{argsByKey.Key}, argsByKey.Value)
			return ctx, nil
		}
	}).Required("context", "key", "value"))

	prelude.Bind("context merge", wrapTyped(func(contexts []map[string]any) (map[string]any, error) {
		merged := make(map[string]any)
		for _, ctx := range contexts {
			for k, v := range ctx {
				merged[k] = v
			}
		}
		return merged, nil
	}).Required("contextx"))

}
