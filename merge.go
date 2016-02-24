package ucfg

import "reflect"

type MergeOption func(mergeOpts) mergeOpts

type mergeOpts struct {
	tag string
}

func StructTag(tag string) MergeOption {
	return func(opts mergeOpts) mergeOpts {
		opts.tag = tag
		return opts
	}
}

func makeMergeOpts(options []MergeOption) mergeOpts {
	opts := mergeOpts{
		tag: "config",
	}
	for _, opt := range options {
		opts = opt(opts)
	}
	return opts
}

func (c *Config) Merge(from interface{}, options ...MergeOption) error {
	opts := makeMergeOpts(options)
	other, err := normalize(opts, from)
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
func normalize(opts mergeOpts, from interface{}) (*Config, error) {
	vFrom := chaseValue(reflect.ValueOf(from))

	switch vFrom.Type() {
	case tConfig:
		return vFrom.Addr().Interface().(*Config), nil
	case tConfigMap:
		return normalizeMap(opts, vFrom)
	default:
		switch vFrom.Kind() {
		case reflect.Struct:
			return normalizeStruct(opts, vFrom)
		case reflect.Map:
			return normalizeMap(opts, vFrom)
		}
	}

	return nil, ErrTypeMismatch
}

func normalizeMap(opts mergeOpts, from reflect.Value) (*Config, error) {
	cfg := New()
	if err := normalizeMapInto(cfg, opts, from); err != nil {
		return nil, err
	}
	return cfg, nil
}

func normalizeMapInto(cfg *Config, opts mergeOpts, from reflect.Value) error {
	k := from.Type().Key().Kind()
	if k != reflect.String && k != reflect.Interface {
		return ErrTypeMismatch
	}

	for _, k := range from.MapKeys() {
		k = chaseValueInterfaces(k)
		if k.Kind() != reflect.String {
			return ErrKeyTypeNotString
		}

		name := k.String()
		if cfg.HasField(name) {
			return errDuplicateKey(name)
		}

		v, err := normalizeValue(opts, from.MapIndex(k))
		if err != nil {
			return err
		}
		cfg.fields[name] = v
	}
	return nil
}

func normalizeStruct(opts mergeOpts, from reflect.Value) (*Config, error) {
	cfg := New()
	if err := normalizeStructInto(cfg, opts, from); err != nil {
		return nil, err
	}
	return cfg, nil
}

func normalizeStructInto(cfg *Config, opts mergeOpts, from reflect.Value) error {
	v := chaseValue(from)
	numField := v.NumField()

	for i := 0; i < numField; i++ {
		stField := v.Type().Field(i)
		name, tagOpts := parseTags(stField.Tag.Get(opts.tag))
		if tagOpts.squash {
			var err error

			vField := chaseValue(v.Field(i))
			switch vField.Kind() {
			case reflect.Struct:
				err = normalizeStructInto(cfg, opts, vField)
			case reflect.Map:
				err = normalizeMapInto(cfg, opts, vField)
			default:
				err = ErrTypeMismatch
			}

			if err != nil {
				return err
			}
		} else {
			v, err := normalizeValue(opts, v.Field(i))
			if err != nil {
				return err
			}

			name = fieldName(name, stField.Name)
			if cfg.HasField(name) {
				return errDuplicateKey(name)
			}

			cfg.fields[name] = v
		}
	}
	return nil
}

func normalizeStructValue(opts mergeOpts, from reflect.Value) (value, error) {
	sub, err := normalizeStruct(opts, from)
	if err != nil {
		return nil, err
	}
	return cfgSub{sub}, nil
}

func normalizeMapValue(opts mergeOpts, from reflect.Value) (value, error) {
	sub, err := normalizeMap(opts, from)
	if err != nil {
		return nil, err
	}
	return cfgSub{sub}, nil
}

func normalizeArray(opts mergeOpts, v reflect.Value) (value, error) {
	l := v.Len()
	out := make([]value, 0, l)
	for i := 0; i < l; i++ {
		tmp, err := normalizeValue(opts, v.Index(i))
		if err != nil {
			return nil, err
		}
		out = append(out, tmp)
	}
	return &cfgArray{arr: out}, nil
}

func normalizeValue(opts mergeOpts, v reflect.Value) (value, error) {
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
		return normalizeArray(opts, v)
	case reflect.Map:
		return normalizeMapValue(opts, v)
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
		return normalizeStructValue(opts, v)
	default:
		return nil, ErrTypeMismatch
	}
}
