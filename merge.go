package ucfg

import "reflect"

func (c *Config) Merge(from interface{}) error {
	other, err := normalize(from)
	if err != nil {
		return err
	}
	return mergeConfig(c.fields, other.fields)
}

func mergeConfig(to, from map[string]value) error {
	for k, v := range from {
		old, ok := to[k]
		if !ok {
			to[k] = v
			continue
		}

		subOld, ok := old.(cfgSub)
		if !ok {
			to[k] = v
			continue
		}

		subFrom, ok := v.(cfgSub)
		if !ok {
			to[k] = v
			continue
		}

		err := mergeConfig(subOld.c.fields, subFrom.c.fields)
		if err != nil {
			return err
		}
	}
	return nil
}

// convert from into normalized *Config checking for errors
// before merging generated(normalized) config with current config
func normalize(from interface{}) (*Config, error) {
	vFrom := chaseValue(reflect.ValueOf(from))

	switch vFrom.Type() {
	case tConfig:
		return vFrom.Addr().Interface().(*Config), nil
	case tConfigMap:
		return normalizeMap(vFrom)
	default:
		switch vFrom.Kind() {
		case reflect.Struct:
			return normalizeStruct(vFrom)
		case reflect.Map:
			return normalizeMap(vFrom)
		}
	}

	return nil, ErrTypeMismatch
}

func normalizeMap(from reflect.Value) (*Config, error) {
	cfg := New()
	if err := normalizeMapInto(cfg, from); err != nil {
		return nil, err
	}
	return cfg, nil
}

func normalizeMapInto(cfg *Config, from reflect.Value) error {
	k := from.Type().Key().Kind()
	if k != reflect.String && k != reflect.Interface {
		return ErrTypeMismatch
	}

	for _, k := range from.MapKeys() {
		k = chaseValueInterfaces(k)
		if k.Kind() != reflect.String {
			return ErrKeyTypeNotString
		}

		v, err := normalizeValue(from.MapIndex(k))
		if err != nil {
			return err
		}
		cfg.fields[k.String()] = v
	}
	return nil
}

func normalizeStruct(from reflect.Value) (*Config, error) {
	cfg := New()
	if err := normalizeStructInto(cfg, from); err != nil {
		return nil, err
	}
	return cfg, nil
}

func normalizeStructInto(cfg *Config, from reflect.Value) error {
	v := chaseValue(from)
	numField := v.NumField()

	for i := 0; i < numField; i++ {
		stField := v.Type().Field(i)
		name, opts := parseTags(stField.Tag.Get("config"))

		if opts.squash {
			var err error

			vField := chaseValue(v.Field(i))
			switch vField.Kind() {
			case reflect.Struct:
				err = normalizeStructInto(cfg, vField)
			case reflect.Map:
				err = normalizeMapInto(cfg, vField)
			default:
				err = ErrTypeMismatch
			}

			if err != nil {
				return err
			}
		} else {
			v, err := normalizeValue(v.Field(i))
			if err != nil {
				return err
			}

			name = fieldName(name, stField.Name)
			cfg.fields[name] = v
		}
	}
	return nil
}

func normalizeStructValue(from reflect.Value) (value, error) {
	sub, err := normalizeStruct(from)
	if err != nil {
		return nil, err
	}
	return cfgSub{sub}, nil
}

func normalizeMapValue(from reflect.Value) (value, error) {
	sub, err := normalizeMap(from)
	if err != nil {
		return nil, err
	}
	return cfgSub{sub}, nil
}

func normalizeArray(v reflect.Value) (value, error) {
	l := v.Len()
	out := make([]value, 0, l)
	for i := 0; i < l; i++ {
		tmp, err := normalizeValue(v.Index(i))
		if err != nil {
			return nil, err
		}
		out = append(out, tmp)
	}
	return &cfgArray{arr: out}, nil
}

func normalizeValue(v reflect.Value) (value, error) {
	v = chaseValue(v)

	// handle primitives
	switch v.Kind() {
	case reflect.Bool:
		return &cfgBool{b: v.Bool()}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &cfgInt{i: v.Int()}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &cfgInt{i: int64(v.Uint())}, nil
	case reflect.Float32, reflect.Float64:
		return &cfgFloat{f: v.Float()}, nil
	case reflect.String:
		return &cfgString{s: v.String()}, nil
	case reflect.Array, reflect.Slice:
		return normalizeArray(v)
	case reflect.Map:
		return normalizeMapValue(v)
	case reflect.Struct:
		if v.Type().ConvertibleTo(tConfig) {
			var c *Config
			if !v.CanAddr() {
				vTmp := reflect.New(tConfig)
				vTmp.Elem().Set(v)
				c = vTmp.Interface().(*Config)
			} else {
				c = v.Addr().Interface().(*Config)
			}
			return cfgSub{c}, nil
		}
		return normalizeStructValue(v)
	default:
		return nil, ErrTypeMismatch
	}
}
