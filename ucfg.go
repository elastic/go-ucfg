package ucfg

import "errors"

type Config struct {
	fields map[string]interface{}
}

var (
	ErrMissing = errors.New("field name missing")

	ErrTypeNoArray = errors.New("field is no array")

	ErrTypeMismatch = errors.New("...")

	ErrIndexOutOfRange = errors.New("index out of range")
)

func New() *Config {
	return &Config{
		fields: make(map[string]interface{}),
	}
}

func (c *Config) GetFields() []string {
	var names []string
	for k := range c.fields {
		names = append(names, k)
	}
	return names
}

func (c *Config) HasField(name string) bool {
	_, ok := c.fields[name]
	return ok
}
