package ucfg

import (
	"fmt"
	"reflect"
)

func (c *Config) Unpack(to interface{}) error {
	vTo := reflect.ValueOf(to)
	if to == nil || (vTo.Kind() != reflect.Ptr && vTo.Kind() != reflect.Map) {
		return ErrPointerRequired
	}
	return reifyInto(vTo, c.fields)
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
	fmt.Printf("reifyMap, to=%v, from=%v\n", to, from)

	if to.Type().Key().Kind() != reflect.String {
		return ErrTypeMismatch
	}

	if to.IsNil() {
		fmt.Println("to is nil")
		to.Set(reflect.MakeMap(to.Type()))
	}

	for k, value := range from {
		key := reflect.ValueOf(k)

		old := to.MapIndex(key)
		var v reflect.Value
		var err error

		fmt.Println("key: ", k)
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
	fmt.Printf("reifyStruct, to=%v, from=%v\n", to, from)

	to = chaseValuePointers(to)
	numField := to.NumField()

	for i := 0; i < numField; i++ {
		stField := to.Type().Field(i)
		name, _ := parseTags(stField.Tag.Get("config"))
		name = fieldName(name, stField.Name)

		value, ok := from[name]
		if !ok {
			fmt.Println("config missing")
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
	fmt.Printf("reifyValue, t=%v, value=%v\n", t, value)

	if t.Kind() == reflect.Interface && t.NumMethod() == 0 {
		fmt.Println("return interface")
		return reflect.ValueOf(value.reify()), nil
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

func reifyMergeValue(oldValue reflect.Value, value value) (reflect.Value, error) {
	fmt.Println("reifyMergeValue")

	old := chaseValueInterfaces(oldValue)
	t := old.Type()
	old = chaseValuePointers(old)
	if (old.Kind() == reflect.Ptr || old.Kind() == reflect.Interface) && old.IsNil() {
		return reifyValue(t, value)
	}

	baseType := chaseTypePointers(old.Type())
	fmt.Println("old:", old)
	fmt.Println("baseType: ", baseType)

	fmt.Println("test config")
	if baseType == tConfig {
		fmt.Println("is config")

		sub, ok := value.(cfgSub)
		fmt.Println("is cfgSub:", ok)
		if !ok {
			return reflect.Value{}, ErrTypeMismatch
		}

		if t == baseType {
			// no pointer -> return type mismatch
			return reflect.Value{}, ErrTypeMismatch
		}

		// check if old is nil -> copy reference only
		if old.Kind() == reflect.Ptr && old.IsNil() {
			fmt.Println("is nil")
			return pointerize(t, baseType, value.reflect()), nil
		}

		// check if old == value
		subOld := chaseValuePointers(old).Addr().Interface().(*Config)
		fmt.Println("old config")
		if sub.c == subOld {
			fmt.Println("same pointer")
			return oldValue, nil
		}

		// old != value -> merge value into old
		fmt.Println("merge configs")
		err := mergeConfig(subOld.fields, sub.c.fields)
		return oldValue, err
	}

	switch baseType.Kind() {
	case reflect.Map:
		fmt.Println("is map")
		sub, ok := value.(cfgSub)
		if !ok {
			return reflect.Value{}, ErrTypeMismatch
		}
		err := reifyMap(old, sub.c.fields)
		return old, err
	case reflect.Struct:
		fmt.Println("is struct")
		sub, ok := value.(cfgSub)
		if !ok {
			return reflect.Value{}, ErrTypeMismatch
		}
		err := reifyStruct(old, sub.c.fields)
		return oldValue, err
	case reflect.Array, reflect.Slice:
		fmt.Println("is array")
		_, ok := value.(*cfgArray)
		if !ok {
			return reflect.Value{}, ErrTODO
		}
		return reflect.Value{}, ErrTODO
	}

	// try primitive conversion
	v := value.reflect()
	fmt.Println("test convertible")
	if v.Type().ConvertibleTo(baseType) {
		return pointerize(t, baseType, v.Convert(baseType)), nil
	}

	return reflect.Value{}, ErrTODO
}

func pointerize(t, base reflect.Type, v reflect.Value) reflect.Value {
	if t == base {
		return v
	}

	if t.Kind() == reflect.Interface {
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
