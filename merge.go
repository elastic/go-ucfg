package ucfg

import "reflect"

func (c *Config) Merge(from interface{}) error {
	return mergeInto(c.fields, from)
}

func mergeInto(to map[string]interface{}, from interface{}) error {
	from = chasePointers(reflect.ValueOf(from)).Interface()
	switch from.(type) {
	case Config:
		return mergeMap(to, from.(Config).fields)
	case map[string]interface{}:
		return mergeMap(to, from.(map[string]interface{}))
	default:
		if !isStruct(from) {
			return ErrTypeMismatch
		}
		return mergeStruct(to, from)
	}
}

func mergeMap(to, from map[string]interface{}) error {
	for k, v := range from {
		var new interface{}
		var err error

		if old, ok := to[k]; ok {
			new, err = mergeValue(old, v)
		} else {
			new, err = normalizeValue(v)
		}

		if err != nil {
			return err
		}
		to[k] = new
	}
	return nil
}

func mergeStruct(to map[string]interface{}, from interface{}) error {
	v := reflect.ValueOf(from)
	numField := v.NumField()

	for i := 0; i < numField; i++ {
		stField := v.Type().Field(i)
		name, opts := parseTags(stField.Tag.Get("config"))

		if opts.squash {
			var err error

			vField := chasePointers(v.Field(i))
			if vField.Kind() == reflect.Struct {
				err = mergeStruct(to, vField.Interface())
			} else if mapField, ok := vField.Interface().(map[string]interface{}); ok {
				err = mergeMap(to, mapField)
			} else {
				err = ErrTypeMismatch
			}

			if err != nil {
				return err
			}
		} else {
			name = fieldName(name, stField.Name)
			var new interface{}
			var err error

			if old, ok := to[name]; ok {
				new, err = mergeValue(old, v.Field(i).Interface())
			} else {
				new, err = normalizeValue(v.Field(i).Interface())
			}

			if err != nil {
				return err
			}

			to[name] = new
		}
	}

	return nil
}

func mergeValue(old, new interface{}) (interface{}, error) {
	vNew := chasePointers(reflect.ValueOf(new))
	if isPrimitive(vNew.Kind()) {
		return vNew.Interface(), nil
	}

	if vNew.IsNil() {
		return nil, nil
	}

	if vNew.Kind() == reflect.Array || vNew.Kind() == reflect.Slice {
		return normalizeArray(vNew)
	}

	if vNew.Kind() == reflect.Map {
		if newMap, ok := vNew.Interface().(map[string]interface{}); ok {
			oldMap, ok := old.(map[string]interface{})
			if ok {
				err := mergeMap(oldMap, newMap)
				return oldMap, err
			}
			return normalizeMap(newMap)
		}
		return nil, ErrTypeMismatch
	}

	if vNew.Kind() == reflect.Struct {
		oldMap, ok := old.(map[string]interface{})
		if !ok {
			oldMap = make(map[string]interface{})
		}
		err := mergeStruct(oldMap, vNew.Interface())
		return oldMap, err
	}

	return nil, ErrTypeMismatch
}

func normalizeValue(ifc interface{}) (interface{}, error) {
	v := chasePointers(reflect.ValueOf(ifc))
	if isPrimitive(v.Kind()) {
		return v.Interface(), nil
	}

	if v.IsNil() {
		return nil, nil
	}

	if v.Kind() == reflect.Array || v.Kind() == reflect.Slice {
		return normalizeArray(v)
	}

	if v.Kind() == reflect.Map {
		if inMap, ok := v.Interface().(map[string]interface{}); ok {
			return normalizeMap(inMap)
		}
		return nil, ErrTypeMismatch
	}

	if v.Kind() == reflect.Struct {
		tmp := make(map[string]interface{})
		if err := mergeStruct(tmp, v.Interface()); err != nil {
			return nil, err
		}
		return tmp, nil
	}

	return nil, ErrTypeMismatch
}

func normalizeArray(v reflect.Value) (interface{}, error) {
	l := v.Len()
	out := make([]interface{}, 0, l)
	for i := 0; i < l; i++ {
		tmp, err := normalizeValue(v.Field(i).Interface())
		if err != nil {
			return nil, err
		}
		out = append(out, tmp)
	}
	return out, nil
}

func normalizeMap(v map[string]interface{}) (interface{}, error) {
	tmp := make(map[string]interface{})
	if err := mergeMap(tmp, v); err != nil {
		return nil, err
	}
	return tmp, nil
}

func chasePointers(v reflect.Value) reflect.Value {
	for {
		k := v.Kind()
		if k != reflect.Ptr && k != reflect.Interface {
			return v
		}
		v = v.Elem()
	}
}

func isStruct(v interface{}) bool {
	return isKind(v, reflect.Struct)
}

func isPtr(v interface{}) bool {
	return isKind(v, reflect.Ptr)
}

func isKind(v interface{}, k reflect.Kind) bool {
	return reflect.TypeOf(v).Kind() == k
}

func isPrimitive(k reflect.Kind) bool {
	switch k {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	default:
		return false
	}
}
