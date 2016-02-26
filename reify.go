package ucfg

import (
	"reflect"
	"strings"
	"time"
)

func (c *Config) Unpack(to interface{}, options ...Option) error {
	if c == nil {
		return ErrNilConfig
	}
	if to == nil {
		return ErrNilValue
	}

	opts := makeOptions(options)
	vTo := reflect.ValueOf(to)
	if to == nil || (vTo.Kind() != reflect.Ptr && vTo.Kind() != reflect.Map) {
		return ErrPointerRequired
	}
	return reifyInto(opts, vTo, c)
}

func reifyInto(opts options, to reflect.Value, from *Config) error {
	to = chaseValuePointers(to)

	if to.Type() == tConfig {
		return mergeConfig(to.Addr().Interface().(*Config).fields, from.fields)
	}

	switch to.Kind() {
	case reflect.Map:
		return reifyMap(opts, to, from)
	case reflect.Struct:
		return reifyStruct(opts, to, from)
	}

	return ErrTypeMismatch
}

func reifyMap(opts options, to reflect.Value, from *Config) error {
	if to.Type().Key().Kind() != reflect.String {
		return ErrTypeMismatch
	}

	if len(from.fields) == 0 {
		return nil
	}

	if to.IsNil() {
		to.Set(reflect.MakeMap(to.Type()))
	}
	for k, value := range from.fields {
		key := reflect.ValueOf(k)

		old := to.MapIndex(key)
		var v reflect.Value
		var err error

		if !old.IsValid() {
			v, err = reifyValue(opts, to.Type().Elem(), value)
		} else {
			v, err = reifyMergeValue(opts, old, value)
		}

		if err != nil {
			return err
		}
		to.SetMapIndex(key, v)
	}

	return nil
}

func reifyStruct(opts options, to reflect.Value, cfg *Config) error {
	to = chaseValuePointers(to)
	numField := to.NumField()

	for i := 0; i < numField; i++ {
		var err error
		stField := to.Type().Field(i)
		vField := to.Field(i)
		name, tagOpts := parseTags(stField.Tag.Get(opts.tag))

		if tagOpts.squash {
			vField := chaseValue(vField)
			switch vField.Kind() {
			case reflect.Struct, reflect.Map:
				err = reifyInto(opts, vField, cfg)
			default:
				err = ErrTypeMismatch
			}
		} else {
			name = fieldName(name, stField.Name)
			err = reifyGetField(cfg, opts, name, vField)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func reifyGetField(cfg *Config, opts options, name string, to reflect.Value) error {
	from, field, err := reifyCfgPath(cfg, opts, name)
	if err != nil {
		return err
	}

	value, ok := from.fields[field]
	if !ok {
		// TODO: handle missing config
		return nil
	}

	v, err := reifyMergeValue(opts, to, value)
	if err != nil {
		return err
	}

	to.Set(v)
	return nil
}

func reifyCfgPath(cfg *Config, opts options, field string) (*Config, string, error) {
	if opts.pathSep == "" {
		return cfg, field, nil
	}

	path := strings.Split(field, opts.pathSep)
	for len(path) > 1 {
		field = path[0]
		path = path[1:]

		sub, exists := cfg.fields[field]
		if !exists {
			return nil, field, ErrMissing
		}

		cSub, err := sub.toConfig()
		if err != nil {
			return nil, field, ErrExpectedObject
		}
		cfg = cSub
	}
	field = path[0]

	return cfg, field, nil
}

func reifyValue(opts options, t reflect.Type, val value) (reflect.Value, error) {
	if t.Kind() == reflect.Interface && t.NumMethod() == 0 {
		return reflect.ValueOf(val.reify()), nil
	}

	baseType := chaseTypePointers(t)
	if baseType == tConfig {
		if _, err := val.toConfig(); err != nil {
			return reflect.Value{}, ErrTypeMismatch
		}

		v := val.reflect()
		if t == baseType { // copy config
			v = v.Elem()
		} else {
			v = pointerize(t, baseType, v)
		}
		return v, nil
	}

	if baseType.Kind() == reflect.Struct {
		sub, err := val.toConfig()
		if err != nil {
			return reflect.Value{}, ErrTypeMismatch
		}

		newSt := reflect.New(baseType)
		if err := reifyInto(opts, newSt, sub); err != nil {
			return reflect.Value{}, err
		}

		if t.Kind() != reflect.Ptr {
			return newSt.Elem(), nil
		}
		return pointerize(t, baseType, newSt), nil
	}

	if baseType.Kind() == reflect.Map {
		sub, err := val.toConfig()
		if err != nil {
			return reflect.Value{}, ErrTypeMismatch
		}

		if baseType.Key().Kind() != reflect.String {
			return reflect.Value{}, ErrTypeMismatch
		}

		newMap := reflect.MakeMap(baseType)
		if err := reifyInto(opts, newMap, sub); err != nil {
			return reflect.Value{}, err
		}
		return newMap, nil
	}

	if baseType.Kind() == reflect.Slice {
		arr, ok := val.(*cfgArray)
		if !ok {
			arr = &cfgArray{arr: []value{val}}
		}

		v, err := reifySlice(opts, baseType, arr)
		if err != nil {
			return reflect.Value{}, err
		}
		return pointerize(t, baseType, v), nil
	}

	v := val.reflect()
	if v.Type().ConvertibleTo(baseType) {
		v = pointerize(t, baseType, v.Convert(baseType))
		return v, nil
	}

	return reflect.Value{}, ErrTypeMismatch
}

func reifyMergeValue(
	opts options,
	oldValue reflect.Value, val value,
) (reflect.Value, error) {
	old := chaseValueInterfaces(oldValue)
	t := old.Type()
	old = chaseValuePointers(old)
	if (old.Kind() == reflect.Ptr || old.Kind() == reflect.Interface) && old.IsNil() {
		return reifyValue(opts, t, val)
	}

	baseType := chaseTypePointers(old.Type())
	if baseType == tConfig {
		sub, err := val.toConfig()
		if err != nil {
			return reflect.Value{}, ErrTypeMismatch
		}

		if t == baseType {
			// no pointer -> return type mismatch
			return reflect.Value{}, ErrTypeMismatch
		}

		// check if old is nil -> copy reference only
		if old.Kind() == reflect.Ptr && old.IsNil() {
			return pointerize(t, baseType, val.reflect()), nil
		}

		// check if old == value
		subOld := chaseValuePointers(old).Addr().Interface().(*Config)
		if sub == subOld {
			return oldValue, nil
		}

		// old != value -> merge value into old
		err = mergeConfig(subOld.fields, sub.fields)
		return oldValue, err
	}

	switch baseType.Kind() {
	case reflect.Map:
		sub, err := val.toConfig()
		if err != nil {
			return reflect.Value{}, ErrTypeMismatch
		}
		err = reifyMap(opts, old, sub)
		return old, err
	case reflect.Struct:
		sub, err := val.toConfig()
		if err != nil {
			return reflect.Value{}, ErrTypeMismatch
		}
		err = reifyStruct(opts, old, sub)
		return oldValue, err
	case reflect.Array:
		arr, ok := val.(*cfgArray)
		if !ok {
			// convert single value to array for merging
			arr = &cfgArray{
				arr: []value{val},
			}
		}
		return reifyArray(opts, old, baseType, arr)
	case reflect.Slice:
		arr, ok := val.(*cfgArray)
		if !ok {
			// convert single value to array for merging
			arr = &cfgArray{
				arr: []value{val},
			}
		}
		return reifySlice(opts, baseType, arr)
	}

	return reifyPrimitive(opts, val, t, baseType)
}

func reifyPrimitive(
	opts options,
	val value,
	t, baseType reflect.Type,
) (reflect.Value, error) {
	// try primitive conversion
	v := val.reflect()
	switch {
	case v.Type() == baseType:
		return pointerize(t, baseType, v), nil

	case v.Type().ConvertibleTo(baseType):
		return pointerize(t, baseType, v.Convert(baseType)), nil

	case baseType.Kind() == reflect.String:
		s, err := val.toString()
		if err != nil {
			return reflect.Value{}, err
		}
		return pointerize(t, baseType, reflect.ValueOf(s)), nil

	case baseType == tDuration:
		s, err := val.toString()
		if err != nil {
			return reflect.Value{}, err
		}

		d, err := time.ParseDuration(s)
		if err != nil {
			return reflect.Value{}, err
		}

		return pointerize(t, baseType, reflect.ValueOf(d)), nil
	}

	return reflect.Value{}, ErrTODO
}

func reifyArray(opts options, to reflect.Value, tTo reflect.Type, arr *cfgArray) (reflect.Value, error) {
	if arr.Len() != tTo.Len() {
		return reflect.Value{}, ErrArraySizeMistach
	}
	return reifyDoArray(opts, to, tTo.Elem(), arr)
}

func reifySlice(opts options, tTo reflect.Type, arr *cfgArray) (reflect.Value, error) {
	to := reflect.MakeSlice(tTo, arr.Len(), arr.Len())
	return reifyDoArray(opts, to, tTo.Elem(), arr)
}

func reifyDoArray(
	opts options,
	to reflect.Value, elemT reflect.Type, arr *cfgArray,
) (reflect.Value, error) {
	for i, from := range arr.arr {
		v, err := reifyValue(opts, elemT, from)
		if err != nil {
			return reflect.Value{}, ErrTODO
		}
		to.Index(i).Set(v)
	}
	return to, nil
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
