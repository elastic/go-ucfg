package ucfg

import "fmt"

// ******************************************************************************
// Low level getters and setters (do we actually need this?)
// ******************************************************************************

// number of elements for this field. If config value is a list, returns number
// of elements in list
func (c *Config) CountField(name string) (int, error) {
	if v, ok := c.fields.fields[name]; ok {
		return v.Len(), nil
	}
	return -1, raise(ErrMissing)
}

func (c *Config) Bool(name string, idx int, opts ...Option) (bool, error) {
	v, err := c.getField(name, idx, opts)
	if err != nil {
		return false, err
	}
	return v.toBool()
}

func (c *Config) String(name string, idx int, opts ...Option) (string, error) {
	v, err := c.getField(name, idx, opts)
	if err != nil {
		return "", err
	}
	return v.toString()
}

func (c *Config) Int(name string, idx int, opts ...Option) (int64, error) {
	v, err := c.getField(name, idx, opts)
	if err != nil {
		return 0, err
	}
	return v.toInt()
}

func (c *Config) Float(name string, idx int, opts ...Option) (float64, error) {
	v, err := c.getField(name, idx, opts)
	if err != nil {
		return 0, err
	}
	return v.toFloat()
}

func (c *Config) Child(name string, idx int, opts ...Option) (*Config, error) {
	v, err := c.getField(name, idx, opts)
	if err != nil {
		return nil, err
	}
	return v.toConfig()
}

func (c *Config) SetBool(name string, idx int, value bool, opts ...Option) error {
	return c.setField(name, idx, &cfgBool{b: value}, opts)
}

func (c *Config) SetInt(name string, idx int, value int64, opts ...Option) error {
	return c.setField(name, idx, &cfgInt{i: value}, opts)
}

func (c *Config) SetFloat(name string, idx int, value float64, opts ...Option) error {
	return c.setField(name, idx, &cfgFloat{f: value}, opts)
}

func (c *Config) SetString(name string, idx int, value string, opts ...Option) error {
	return c.setField(name, idx, &cfgString{s: value}, opts)
}

func (c *Config) SetChild(name string, idx int, value *Config, opts ...Option) error {
	return c.setField(name, idx, cfgSub{c: value}, opts)
}

func (c *Config) getField(name string, idx int, options []Option) (value, error) {
	opts := makeOptions(options)

	cfg, field, err := reifyCfgPath(c, opts, name)
	if err != nil {
		return nil, err
	}

	v, ok := cfg.fields.fields[field]
	if !ok {
		return nil, raise(ErrMissing)
	}

	if idx >= v.Len() {
		return nil, raise(ErrIndexOutOfRange)
	}

	if arr, ok := v.(*cfgArray); ok {
		v = arr.arr[idx]
		if v == nil {
			return nil, raise(ErrMissing)
		}
		return arr.arr[idx], nil
	}
	return v, nil
}

func (c *Config) setField(name string, idx int, v value, options []Option) error {
	opts := makeOptions(options)
	ctx := context{
		parent: cfgSub{c},
		field:  name,
	}
	orig := v

	cfg, field, err := normalizeCfgPath(c, opts, name)
	if err != nil {
		return err
	}

	old, ok := cfg.fields.fields[field]
	if !ok {
		if idx > 0 {
			slice := &cfgArray{
				cfgPrimitive: cfgPrimitive{ctx: ctx},
				arr:          make([]value, idx+1),
			}
			slice.arr[idx] = v
			v = slice
		} else {
			idx = -1
		}
	} else if slice, ok := old.(*cfgArray); ok {
		for idx >= len(slice.arr) {
			slice.arr = append(slice.arr, nil)
		}
		slice.arr[idx] = v
		v = slice
	} else if idx > 0 {
		slice := &cfgArray{
			cfgPrimitive: cfgPrimitive{ctx: ctx},
			arr:          make([]value, idx+1),
		}
		slice.arr[0] = old
		slice.arr[idx] = v
		v = slice
	} else {
		idx = -1
	}

	if idx >= 0 {
		ctx.parent = v
		ctx.field = fmt.Sprintf("%v", idx)
	}
	orig.SetContext(ctx)

	if opts.meta != nil {
		v.setMeta(opts.meta)
	}

	cfg.fields.fields[field] = v

	return nil
}
