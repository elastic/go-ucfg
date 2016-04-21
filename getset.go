package ucfg

import "fmt"

// ******************************************************************************
// Low level getters and setters (do we actually need this?)
// ******************************************************************************

func convertErr(v value, err error, to string) Error {
	if err == nil {
		return nil
	}
	return raiseConversion(v, err, to)
}

// number of elements for this field. If config value is a list, returns number
// of elements in list
func (c *Config) CountField(name string) (int, error) {
	if v, ok := c.fields.fields[name]; ok {
		return v.Len(), nil
	}
	return -1, raiseMissing(c, name)
}

func (c *Config) Bool(name string, idx int, opts ...Option) (bool, error) {
	v, err := c.getField(name, idx, opts)
	if err != nil {
		return false, err
	}
	b, fail := v.toBool()
	return b, convertErr(v, fail, "bool")
}

func (c *Config) String(name string, idx int, opts ...Option) (string, error) {
	v, err := c.getField(name, idx, opts)
	if err != nil {
		return "", err
	}
	s, fail := v.toString()
	return s, convertErr(v, fail, "string")
}

func (c *Config) Int(name string, idx int, opts ...Option) (int64, error) {
	v, err := c.getField(name, idx, opts)
	if err != nil {
		return 0, err
	}
	i, fail := v.toInt()
	return i, convertErr(v, fail, "int")
}

func (c *Config) Uint(name string, idx int, opts ...Option) (uint64, error) {
	v, err := c.getField(name, idx, opts)
	if err != nil {
		return 0, err
	}
	u, fail := v.toUint()
	return u, convertErr(v, fail, "uint")
}

func (c *Config) Float(name string, idx int, opts ...Option) (float64, error) {
	v, err := c.getField(name, idx, opts)
	if err != nil {
		return 0, err
	}
	f, fail := v.toFloat()
	return f, convertErr(v, fail, "float")
}

func (c *Config) Child(name string, idx int, opts ...Option) (*Config, error) {
	v, err := c.getField(name, idx, opts)
	if err != nil {
		return nil, err
	}
	c, fail := v.toConfig()
	return c, convertErr(v, fail, "object")
}

func (c *Config) SetBool(name string, idx int, value bool, opts ...Option) error {
	return c.setField(name, idx, &cfgBool{b: value}, opts)
}

func (c *Config) SetInt(name string, idx int, value int64, opts ...Option) error {
	return c.setField(name, idx, &cfgInt{i: value}, opts)
}

func (c *Config) SetUint(name string, idx int, value uint64, opts ...Option) error {
	return c.setField(name, idx, &cfgUint{u: value}, opts)
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

func (c *Config) getField(name string, idx int, options []Option) (value, Error) {
	opts := makeOptions(options)

	cfg, field, err := reifyCfgPath(c, opts, name)
	if err != nil {
		return nil, err
	}

	v, ok := cfg.fields.fields[field]
	if !ok {
		return nil, raiseMissing(cfg, field)
	}

	if s, ok := v.(cfgSub); ok && len(s.c.fields.arr) > 0 {
		arr := s.c.fields.arr
		if idx >= len(arr) {
			return nil, raiseIndexOutOfBounds(v, idx)
		}

		v = arr[idx]
		if v == nil {
			return nil, raiseMissingArr(s.c.ctx, s.c.metadata, idx)
		}
		return arr[idx], nil
	}

	if idx > 0 {
		return nil, raiseIndexOutOfBounds(v, idx)
	}

	return v, nil
}

func (c *Config) setField(name string, idx int, v value, options []Option) Error {
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
			slice := New()
			slice.ctx = ctx
			slice.fields.arr = make([]value, idx+1)
			slice.fields.arr[idx] = v
			v = cfgSub{slice}
		} else {
			idx = -1
		}
	} else if sub, ok := old.(cfgSub); ok {
		fields := sub.c.fields
		for idx >= len(fields.arr) {
			tmp := make([]value, idx+1)
			copy(tmp, fields.arr)
			fields.arr = tmp
		}
		fields.arr[idx] = v
		v = sub
	} else if idx > 0 {
		slice := New()
		slice.ctx = ctx
		slice.fields.arr = make([]value, idx+1)
		slice.fields.arr[0] = old
		slice.fields.arr[idx] = v
		v = cfgSub{slice}
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
