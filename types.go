package ucfg

import (
	"fmt"
	"reflect"
)

type value interface {
	cpy(c context) value

	Context() context
	SetContext(c context)

	Len() int

	reflect() reflect.Value
	typ() reflect.Type
	reify() interface{}

	toBool() (bool, error)
	toString() (string, error)
	toInt() (int64, error)
	toFloat() (float64, error)
	toConfig() (*Config, error)
}

type context struct {
	parent value
	field  string
}

type cfgArray struct {
	cfgPrimitive
	arr []value
}

type cfgBool struct {
	cfgPrimitive
	b bool
}

type cfgInt struct {
	cfgPrimitive
	i int64
}

type cfgFloat struct {
	cfgPrimitive
	f float64
}

type cfgString struct {
	cfgPrimitive
	s string
}

type cfgSub struct {
	c *Config
}

type cfgNil struct{ cfgPrimitive }

type cfgPrimitive struct {
	ctx context
}

func (c *context) empty() bool {
	return c.parent == nil
}

func (c *context) path(sep string) string {
	if c.field == "" {
		return ""
	}

	if c.parent != nil {
		p := c.parent.Context()
		if parent := p.path(sep); parent != "" {
			return fmt.Sprintf("%v%v%v", parent, sep, c.field)
		}
	}

	return c.field
}

func newBool(ctx context, b bool) *cfgBool {
	return &cfgBool{cfgPrimitive{ctx}, b}
}

func newInt(ctx context, i int64) *cfgInt {
	return &cfgInt{cfgPrimitive{ctx}, i}
}

func newFloat(ctx context, f float64) *cfgFloat {
	return &cfgFloat{cfgPrimitive{ctx}, f}
}

func newString(ctx context, s string) *cfgString {
	return &cfgString{cfgPrimitive{ctx}, s}
}

func (p cfgPrimitive) Context() context         { return p.ctx }
func (p cfgPrimitive) SetContext(c context)     { p.ctx = c }
func (cfgPrimitive) Len() int                   { return 1 }
func (cfgPrimitive) toBool() (bool, error)      { return false, raise(ErrTypeMismatch) }
func (cfgPrimitive) toString() (string, error)  { return "", raise(ErrTypeMismatch) }
func (cfgPrimitive) toInt() (int64, error)      { return 0, raise(ErrTypeMismatch) }
func (cfgPrimitive) toFloat() (float64, error)  { return 0, raise(ErrTypeMismatch) }
func (cfgPrimitive) toConfig() (*Config, error) { return nil, raise(ErrTypeMismatch) }

func (c *cfgArray) Len() int               { return len(c.arr) }
func (c *cfgArray) reflect() reflect.Value { return reflect.ValueOf(c.arr) }
func (cfgArray) typ() reflect.Type         { return tInterfaceArray }

func (c *cfgArray) cpy(ctx context) value {
	return &cfgArray{cfgPrimitive{ctx}, c.arr}
}

func (c *cfgArray) reify() interface{} {
	r := make([]interface{}, len(c.arr))
	for i, v := range c.arr {
		r[i] = v.reify()
	}
	return r
}

func (cfgNil) cpy(ctx context) value     { return cfgNil{cfgPrimitive{ctx}} }
func (cfgNil) Len() int                  { return 0 }
func (cfgNil) toString() (string, error) { return "null", nil }
func (cfgNil) toInt() (int64, error)     { return 0, raise(ErrTypeMismatch) }
func (cfgNil) toFloat() (float64, error) { return 0, raise(ErrTypeMismatch) }
func (cfgNil) reflect() reflect.Value    { return reflect.ValueOf(nil) }
func (cfgNil) reify() interface{}        { return nil }
func (cfgNil) typ() reflect.Type         { return reflect.PtrTo(tConfig) }

func (n cfgNil) toConfig() (*Config, error) {
	c := New()
	c.ctx = n.ctx
	return c, nil
}

func (c *cfgBool) cpy(ctx context) value     { return newBool(ctx, c.b) }
func (c *cfgBool) toBool() (bool, error)     { return c.b, nil }
func (c *cfgBool) reflect() reflect.Value    { return reflect.ValueOf(c.b) }
func (c *cfgBool) reify() interface{}        { return c.b }
func (c *cfgBool) toString() (string, error) { return fmt.Sprintf("%t", c.b), nil }
func (c *cfgBool) typ() reflect.Type         { return tBool }

func (c *cfgInt) cpy(ctx context) value     { return newInt(ctx, c.i) }
func (c *cfgInt) toInt() (int64, error)     { return c.i, nil }
func (c *cfgInt) reflect() reflect.Value    { return reflect.ValueOf(c.i) }
func (c *cfgInt) reify() interface{}        { return c.i }
func (c *cfgInt) toString() (string, error) { return fmt.Sprintf("%d", c.i), nil }
func (c *cfgInt) typ() reflect.Type         { return tInt64 }

func (c *cfgFloat) cpy(ctx context) value     { return newFloat(ctx, c.f) }
func (c *cfgFloat) toFloat() (float64, error) { return c.f, nil }
func (c *cfgFloat) reflect() reflect.Value    { return reflect.ValueOf(c.f) }
func (c *cfgFloat) reify() interface{}        { return c.f }
func (c *cfgFloat) toString() (string, error) { return fmt.Sprintf("%v", c.f), nil }
func (c *cfgFloat) typ() reflect.Type         { return tFloat64 }

func (c *cfgString) cpy(ctx context) value     { return newString(ctx, c.s) }
func (c *cfgString) toString() (string, error) { return c.s, nil }
func (c *cfgString) reflect() reflect.Value    { return reflect.ValueOf(c.s) }
func (c *cfgString) reify() interface{}        { return c.s }
func (c *cfgString) typ() reflect.Type         { return tString }

func (cfgSub) Len() int                     { return 1 }
func (c cfgSub) Context() context           { return c.c.ctx }
func (cfgSub) toBool() (bool, error)        { return false, raise(ErrTypeMismatch) }
func (cfgSub) toString() (string, error)    { return "", raise(ErrTypeMismatch) }
func (cfgSub) toInt() (int64, error)        { return 0, raise(ErrTypeMismatch) }
func (cfgSub) toFloat() (float64, error)    { return 0, raise(ErrTypeMismatch) }
func (c cfgSub) toConfig() (*Config, error) { return c.c, nil }
func (cfgSub) typ() reflect.Type            { return reflect.PtrTo(tConfig) }
func (c cfgSub) reflect() reflect.Value     { return reflect.ValueOf(c.c) }

func (c cfgSub) cpy(ctx context) value {
	return cfgSub{
		c: &Config{ctx: ctx, fields: c.c.fields},
	}
}

func (c cfgSub) SetContext(ctx context) {
	if c.c.ctx.empty() {
		c.c.ctx = ctx
	} else {
		c.c = &Config{
			ctx:    ctx,
			fields: c.c.fields,
		}
	}
}

func (c cfgSub) reify() interface{} {
	m := make(map[string]interface{})
	for k, v := range c.c.fields.fields {
		m[k] = v.reify()
	}
	return m
}
