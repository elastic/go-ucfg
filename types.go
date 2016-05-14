package ucfg

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
)

type value interface {
	typ() (typeInfo, error)

	cpy(c context) value

	Context() context
	SetContext(c context)

	meta() *Meta
	setMeta(m *Meta)

	Len() (int, error)

	reflect() (reflect.Value, error)
	reify() (interface{}, error)

	toBool() (bool, error)
	toString() (string, error)
	toInt() (int64, error)
	toUint() (uint64, error)
	toFloat() (float64, error)
	toConfig() (*Config, error)
}

type typeInfo struct {
	name   string
	gotype reflect.Type
}

type context struct {
	parent value
	field  string
}

type cfgBool struct {
	cfgPrimitive
	b bool
}

type cfgInt struct {
	cfgPrimitive
	i int64
}

type cfgUint struct {
	cfgPrimitive
	u uint64
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

type cfgRef struct {
	cfgPrimitive
	ref *reference
}

type cfgSplice struct {
	cfgPrimitive
	splice varEvaler
}

type cfgNil struct{ cfgPrimitive }

type cfgPrimitive struct {
	ctx      context
	metadata *Meta
}

func (c *context) empty() bool {
	return c.parent == nil
}

func (c *context) getParent() *Config {
	if c.parent == nil {
		return nil
	}

	if cfg, ok := c.parent.(cfgSub); ok {
		return cfg.c
	}
	return nil
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

func (c *context) pathOf(field, sep string) string {
	if p := c.path(sep); p != "" {
		return fmt.Sprintf("%v%v%v", p, sep, field)
	}
	return field
}

func newBool(ctx context, m *Meta, b bool) *cfgBool {
	return &cfgBool{cfgPrimitive{ctx, m}, b}
}

func newInt(ctx context, m *Meta, i int64) *cfgInt {
	return &cfgInt{cfgPrimitive{ctx, m}, i}
}

func newUint(ctx context, m *Meta, u uint64) *cfgUint {
	return &cfgUint{cfgPrimitive{ctx, m}, u}
}

func newFloat(ctx context, m *Meta, f float64) *cfgFloat {
	return &cfgFloat{cfgPrimitive{ctx, m}, f}
}

func newString(ctx context, m *Meta, s string) *cfgString {
	return &cfgString{cfgPrimitive{ctx, m}, s}
}

func newRef(ctx context, m *Meta, ref *reference) *cfgRef {
	return &cfgRef{cfgPrimitive{ctx, m}, ref}
}

func newSplice(ctx context, m *Meta, s varEvaler) *cfgSplice {
	return &cfgSplice{cfgPrimitive{ctx, m}, s}
}

func (p *cfgPrimitive) Context() context        { return p.ctx }
func (p *cfgPrimitive) SetContext(c context)    { p.ctx = c }
func (p *cfgPrimitive) meta() *Meta             { return p.metadata }
func (p *cfgPrimitive) setMeta(m *Meta)         { p.metadata = m }
func (cfgPrimitive) Len() (int, error)          { return 1, nil }
func (cfgPrimitive) toBool() (bool, error)      { return false, ErrTypeMismatch }
func (cfgPrimitive) toString() (string, error)  { return "", ErrTypeMismatch }
func (cfgPrimitive) toInt() (int64, error)      { return 0, ErrTypeMismatch }
func (cfgPrimitive) toUint() (uint64, error)    { return 0, ErrTypeMismatch }
func (cfgPrimitive) toFloat() (float64, error)  { return 0, ErrTypeMismatch }
func (cfgPrimitive) toConfig() (*Config, error) { return nil, ErrTypeMismatch }

func (c *cfgNil) cpy(ctx context) value     { return &cfgNil{cfgPrimitive{ctx, c.metadata}} }
func (*cfgNil) Len() (int, error)           { return 0, nil }
func (*cfgNil) toString() (string, error)   { return "null", nil }
func (*cfgNil) toInt() (int64, error)       { return 0, ErrTypeMismatch }
func (*cfgNil) toUint() (uint64, error)     { return 0, ErrTypeMismatch }
func (*cfgNil) toFloat() (float64, error)   { return 0, ErrTypeMismatch }
func (*cfgNil) reify() (interface{}, error) { return nil, nil }
func (*cfgNil) typ() (typeInfo, error)      { return typeInfo{"any", reflect.PtrTo(tConfig)}, nil }
func (c *cfgNil) meta() *Meta               { return c.metadata }
func (c *cfgNil) setMeta(m *Meta)           { c.metadata = m }

func (c *cfgNil) reflect() (reflect.Value, error) {
	cfg, _ := c.toConfig()
	return reflect.ValueOf(cfg), nil
}

func (c *cfgNil) toConfig() (*Config, error) {
	n := New()
	n.ctx = c.ctx
	return n, nil
}

func (c *cfgBool) cpy(ctx context) value           { return newBool(ctx, c.meta(), c.b) }
func (c *cfgBool) toBool() (bool, error)           { return c.b, nil }
func (c *cfgBool) reflect() (reflect.Value, error) { return reflect.ValueOf(c.b), nil }
func (c *cfgBool) reify() (interface{}, error)     { return c.b, nil }
func (c *cfgBool) toString() (string, error)       { return fmt.Sprintf("%t", c.b), nil }
func (c *cfgBool) typ() (typeInfo, error)          { return typeInfo{"bool", tBool}, nil }

func (c *cfgInt) cpy(ctx context) value           { return newInt(ctx, c.meta(), c.i) }
func (c *cfgInt) toInt() (int64, error)           { return c.i, nil }
func (c *cfgInt) reflect() (reflect.Value, error) { return reflect.ValueOf(c.i), nil }
func (c *cfgInt) reify() (interface{}, error)     { return c.i, nil }
func (c *cfgInt) toString() (string, error)       { return fmt.Sprintf("%d", c.i), nil }
func (c *cfgInt) typ() (typeInfo, error)          { return typeInfo{"int", tInt64}, nil }
func (c *cfgInt) toUint() (uint64, error) {
	if c.i < 0 {
		return 0, ErrNegative
	}
	return uint64(c.i), nil
}

func (c *cfgUint) cpy(ctx context) value           { return newUint(ctx, c.meta(), c.u) }
func (c *cfgUint) reflect() (reflect.Value, error) { return reflect.ValueOf(c.u), nil }
func (c *cfgUint) reify() (interface{}, error)     { return c.u, nil }
func (c *cfgUint) toString() (string, error)       { return fmt.Sprintf("%d", c.u), nil }
func (c *cfgUint) typ() (typeInfo, error)          { return typeInfo{"uint", tUint64}, nil }
func (c *cfgUint) toUint() (uint64, error)         { return c.u, nil }
func (c *cfgUint) toInt() (int64, error) {
	if c.u > math.MaxInt64 {
		return 0, ErrOverflow
	}
	return int64(c.u), nil
}

func (c *cfgFloat) cpy(ctx context) value           { return newFloat(ctx, c.meta(), c.f) }
func (c *cfgFloat) toFloat() (float64, error)       { return c.f, nil }
func (c *cfgFloat) reflect() (reflect.Value, error) { return reflect.ValueOf(c.f), nil }
func (c *cfgFloat) reify() (interface{}, error)     { return c.f, nil }
func (c *cfgFloat) toString() (string, error)       { return fmt.Sprintf("%v", c.f), nil }
func (c *cfgFloat) typ() (typeInfo, error)          { return typeInfo{"float", tFloat64}, nil }

func (c *cfgFloat) toUint() (uint64, error) {
	if c.f < 0 {
		return 0, ErrNegative
	}
	if c.f > math.MaxUint64 {
		return 0, ErrOverflow
	}
	return uint64(c.f), nil
}

func (c *cfgFloat) toInt() (int64, error) {
	if c.f < math.MinInt64 || math.MaxInt64 < c.f {
		return 0, ErrOverflow
	}
	return int64(c.f), nil
}

func (c *cfgString) cpy(ctx context) value           { return newString(ctx, c.meta(), c.s) }
func (c *cfgString) toString() (string, error)       { return c.s, nil }
func (c *cfgString) reflect() (reflect.Value, error) { return reflect.ValueOf(c.s), nil }
func (c *cfgString) reify() (interface{}, error)     { return c.s, nil }
func (c *cfgString) typ() (typeInfo, error)          { return typeInfo{"string", tString}, nil }

func (cfgSub) Len() (int, error)                 { return 1, nil }
func (c cfgSub) Context() context                { return c.c.ctx }
func (cfgSub) toBool() (bool, error)             { return false, ErrTypeMismatch }
func (cfgSub) toString() (string, error)         { return "", ErrTypeMismatch }
func (cfgSub) toInt() (int64, error)             { return 0, ErrTypeMismatch }
func (cfgSub) toUint() (uint64, error)           { return 0, ErrTypeMismatch }
func (cfgSub) toFloat() (float64, error)         { return 0, ErrTypeMismatch }
func (c cfgSub) toConfig() (*Config, error)      { return c.c, nil }
func (cfgSub) typ() (typeInfo, error)            { return typeInfo{"object", reflect.PtrTo(tConfig)}, nil }
func (c cfgSub) reflect() (reflect.Value, error) { return reflect.ValueOf(c.c), nil }
func (c cfgSub) meta() *Meta                     { return c.c.metadata }
func (c cfgSub) setMeta(m *Meta)                 { c.c.metadata = m }

func (c cfgSub) cpy(ctx context) value {
	return cfgSub{
		c: &Config{ctx: ctx, fields: c.c.fields, metadata: c.c.metadata},
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

func (c cfgSub) reify() (interface{}, error) {
	fields := c.c.fields.fields
	arr := c.c.fields.arr

	switch {
	case len(fields) == 0 && len(arr) == 0:
		return nil, nil
	case len(fields) > 0 && len(arr) == 0:
		m := make(map[string]interface{})
		for k, v := range c.c.fields.fields {
			var err error
			if m[k], err = v.reify(); err != nil {
				return nil, err
			}
		}
		return m, nil
	case len(fields) == 0 && len(arr) > 0:
		m := make([]interface{}, len(arr))
		for i, v := range arr {
			var err error
			if m[i], err = v.reify(); err != nil {
				return nil, err
			}
		}
		return m, nil
	default:
		m := make(map[string]interface{})
		for k, v := range c.c.fields.fields {
			var err error
			if m[k], err = v.reify(); err != nil {
				return nil, err
			}
		}
		for i, v := range arr {
			var err error
			m[fmt.Sprintf("%d", i)], err = v.reify()
			if err != nil {
				return nil, err
			}
		}
		return m, nil
	}
}

func (r *cfgRef) typ() (typeInfo, error) {
	v, err := r.resolve()
	if err != nil {
		return typeInfo{}, err
	}
	return v.typ()
}

func (r *cfgRef) cpy(ctx context) value {
	return newRef(ctx, r.meta(), r.ref)
}

func (r *cfgRef) Len() (int, error) {
	v, err := r.resolve()
	if err != nil {
		return 0, err
	}
	return v.Len()
}

func (r *cfgRef) reflect() (reflect.Value, error) {
	v, err := r.resolve()
	if err != nil {
		return reflect.Value{}, err
	}
	return v.reflect()
}

func (r *cfgRef) reify() (interface{}, error) {
	v, err := r.resolve()
	if err != nil {
		return reflect.Value{}, err
	}
	return v.reify()
}

func (r *cfgRef) toBool() (bool, error) {
	v, err := r.resolve()
	if err != nil {
		return false, err
	}
	return v.toBool()
}

func (r *cfgRef) toString() (string, error) {
	v, err := r.resolve()
	if err != nil {
		return "", err
	}
	return v.toString()
}

func (r *cfgRef) toInt() (int64, error) {
	v, err := r.resolve()
	if err != nil {
		return 0, err
	}
	return v.toInt()
}

func (r *cfgRef) toUint() (uint64, error) {
	v, err := r.resolve()
	if err != nil {
		return 0, err
	}
	return v.toUint()
}

func (r *cfgRef) toFloat() (float64, error) {
	v, err := r.resolve()
	if err != nil {
		return 0, err
	}
	return v.toFloat()
}

func (r *cfgRef) toConfig() (*Config, error) {
	v, err := r.resolve()
	if err != nil {
		return nil, err
	}
	return v.toConfig()
}

func (r *cfgRef) resolve() (value, error) {
	return r.ref.resolve(r.ctx.getParent())
}

func (*cfgSplice) typ() (typeInfo, error) {
	return typeInfo{"string", tString}, nil
}

func (s *cfgSplice) cpy(ctx context) value {
	return newSplice(ctx, s.meta(), s.splice)
}

func (s *cfgSplice) reflect() (reflect.Value, error) {
	str, err := s.toString()
	if err != nil {
		return reflect.Value{}, nil
	}
	return reflect.ValueOf(str), err
}

func (s *cfgSplice) reify() (interface{}, error) {
	return s.toString()
}

func (s *cfgSplice) toBool() (bool, error) {
	str, err := s.toString()
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(str)
}

func (s *cfgSplice) toString() (string, error) {
	return s.splice.eval(s.ctx.getParent())
}

func (s *cfgSplice) toInt() (int64, error) {
	str, err := s.toString()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(str, 0, 64)
}

func (s *cfgSplice) toUint() (uint64, error) {
	str, err := s.toString()
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(str, 0, 64)
}

func (s *cfgSplice) toFloat() (float64, error) {
	str, err := s.toString()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(str, 64)
}
