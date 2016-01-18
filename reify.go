package ucfg

import "reflect"

func (c *Config) Unpack(to interface{}) error {
	return reifyInto(reflect.ValueOf(to), c.fields)
}

func reifyInto(to reflect.Value, from map[string]value) error {
	to = chaseValuePointers(to)

	if to.Type() == tConfig {
		return mergeConfig(to.Addr().Interface().(*Config).fields, from)
	}

	switch to.Kind() {
	case reflect.Map:
		return reifyMap(to, from)
	case reflect.Struct:
		return reifyStruct(to, from)
	}

	return ErrTypeMismatch
}

func reifyMap(to reflect.Value, from map[string]value) error {
	if to.Type().Key().Kind() != reflect.String {
		return ErrTypeMismatch
	}

	for k, value := range from {
		key := reflect.ValueOf(k)

		old := to.MapIndex(key)
		var v reflect.Value
		var err error

		if !old.IsValid() {
			v, err = reifyValue(to.Type().Elem(), value)
		} else {
			v, err = reifyMergeValue(old, value)
		}

		if err != nil {
			return err
		}
		to.SetMapIndex(key, v)
	}

	return nil
}

func reifyStruct(to reflect.Value, from map[string]value) error {
	to = chaseValuePointers(to)
	numField := to.NumField()

	for i := 0; i < numField; i++ {
		stField := to.Type().Field(i)
		name, _ := parseTags(stField.Tag.Get("config"))
		name = fieldName(name, stField.Name)

		value, ok := from[name]
		if !ok {
			// TODO: handle missing config
			continue
		}

		vField := to.Field(i)
		v, err := reifyMergeValue(vField, value)
		if err != nil {
			return err
		}
		vField.Set(v)
	}

	return nil
}

func reifyValue(t reflect.Type, value value) (reflect.Value, error) {
	if t == tInterface {
		return reifyTopValue(value)
	}

	baseType := chaseTypePointers(t)
	if baseType == tConfig {
		if _, ok := value.(cfgSub); !ok {
			return reflect.Value{}, ErrTypeMismatch
		}

		v := value.reflect()
		if t == baseType { // copy config
			v = v.Elem()
		} else {
			v = pointerize(t, baseType, v)
		}
		return v, nil
	}

	if baseType.Kind() == reflect.Struct {
		if _, ok := value.(cfgSub); !ok {
			return reflect.Value{}, ErrTypeMismatch
		}

		newSt := reflect.New(baseType)
		if err := reifyInto(newSt, value.(cfgSub).c.fields); err != nil {
			return reflect.Value{}, err
		}

		if t.Kind() != reflect.Ptr {
			return newSt.Elem(), nil
		}
		return pointerize(t, baseType, newSt), nil
	}

	if baseType.Kind() == reflect.Map {
		if _, ok := value.(cfgSub); !ok {
			return reflect.Value{}, ErrTypeMismatch
		}

		if baseType.Key().Kind() != reflect.String {
			return reflect.Value{}, ErrTypeMismatch
		}

		newMap := reflect.MakeMap(baseType)
		if err := reifyInto(newMap, value.(cfgSub).c.fields); err != nil {
			return reflect.Value{}, err
		}
		return newMap, nil
	}

	v := value.reflect()
	if v.Type().ConvertibleTo(baseType) {
		v = pointerize(t, baseType, v.Convert(baseType))
		return v, nil
	}

	return reflect.Value{}, ErrTypeMismatch
}

func reifyTopValue(value value) (reflect.Value, error) {
	if _, ok := value.(cfgSub); ok {
		return reifyValue(tConfigMap, value)
	}

	if _, ok := value.(*cfgArray); ok {
		return reifyValue(tInterfaceArray, value)
	}

	return reflect.ValueOf(value.reflect().Interface()), nil
}

func reifyMergeValue(old reflect.Value, value value) (reflect.Value, error) {
	t := old.Type()
	baseType := chaseTypePointers(t)
	v := value.reflect()
	if v.Type().ConvertibleTo(baseType) {
		return pointerize(t, baseType, v.Convert(baseType)), nil
	}

	return reflect.Value{}, ErrTODO
}

func pointerize(t, base reflect.Type, v reflect.Value) reflect.Value {
	if t == base {
		return v
	}

	for t != v.Type() {
		if !v.CanAddr() {
			tmp := reflect.New(v.Type())
			tmp.Elem().Set(v)
			v = tmp
		} else {
			v = v.Addr()
		}
	}
	return v
}
