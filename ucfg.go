package ucfg

import (
	"fmt"
	"reflect"
	"time"
)

type Config struct {
	ctx      context
	metadata *Meta
	fields   *fields
}

type fields struct {
	fields map[string]value
}

// Meta holds additional meta data per config value
type Meta struct {
	source string
}

var (
	tConfig         = reflect.TypeOf(Config{})
	tConfigPtr      = reflect.PtrTo(tConfig)
	tConfigMap      = reflect.TypeOf((map[string]interface{})(nil))
	tInterfaceArray = reflect.TypeOf([]interface{}(nil))
	tDuration       = reflect.TypeOf(time.Duration(0))

	tBool    = reflect.TypeOf(true)
	tInt64   = reflect.TypeOf(int64(0))
	tFloat64 = reflect.TypeOf(float64(0))
	tString  = reflect.TypeOf("")
)

func New() *Config {
	return &Config{
		fields: &fields{map[string]value{}},
	}
}

func NewFrom(from interface{}, opts ...Option) (*Config, error) {
	c := New()
	if err := c.Merge(from, opts...); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) GetFields() []string {
	var names []string
	for k := range c.fields.fields {
		names = append(names, k)
	}
	return names
}

func (c *Config) HasField(name string) bool {
	_, ok := c.fields.fields[name]
	return ok
}

func (c *Config) Path(sep string) string {
	return c.ctx.path(sep)
}

func (c *Config) Parent() *Config {
	ctx := c.ctx
	for {
		if ctx.parent == nil {
			return nil
		}

		switch p := ctx.parent.(type) {
		case *cfgArray:
			ctx = p.Context()
		case cfgSub:
			return p.c
		default:
			return nil
		}
	}
}

func (c *Config) PathOf(field, sep string) string {
	if p := c.Path(sep); p != "" {
		return fmt.Sprintf("%v%v%v", p, sep, field)
	}
	return field
}
