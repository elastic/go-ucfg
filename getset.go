package ucfg

import "reflect"

// ******************************************************************************
// Low level getters and setters (do we actually need this?)
// ******************************************************************************

// number of elements for this field. If config value is a list, returns number
// of elements in list
func (c *Config) CountField(name string) (int, error) {
	ifc, ok := c.fields[name]
	if !ok {
		return -1, ErrMissing
	}

	value := reflect.ValueOf(ifc)
	if value.Kind() != reflect.Array {
		return 1, nil
	}
	return value.Len(), nil
}

// Config getters. If config value is actually a list, idx can be used to access
// values at said index

var (
	tBool   = reflect.TypeOf(false)
	tInt    = reflect.TypeOf(int64(0))
	tFloat  = reflect.TypeOf(float64(0.0))
	tString = reflect.TypeOf(string("string"))

	tConfigMap = reflect.TypeOf(map[string]interface{}(nil))
)

func (c *Config) Bool(name string, idx int) (bool, error) {
	value, err := c.getFieldValue(name, idx, tBool)
	if err != nil {
		return false, err
	}
	return value.Bool(), nil
}

func (c *Config) String(name string, idx int) (string, error) {
	value, err := c.getFieldValue(name, idx, tString)
	if err != nil {
		return "", err
	}
	return value.String(), nil
}

func (c *Config) Int(name string, idx int) (int, error) {
	value, err := c.getFieldValue(name, idx, tInt)
	if err != nil {
		return 0, err
	}
	return int(value.Int()), nil
}

func (c *Config) Float(name string, idx int) (float64, error) {
	value, err := c.getFieldValue(name, idx, tFloat)
	if err != nil {
		return 0, err
	}
	return float64(value.Float()), nil
}

func (c *Config) Child(name string, idx int) (*Config, error) {
	value, err := c.getFieldValue(name, idx, tConfigMap)
	if err != nil {
		if err == ErrTypeMismatch {
			value, err = c.getFieldValue(name, idx, reflect.TypeOf(c))
			if err != nil {
				return nil, err
			}
			return value.Interface().(*Config), nil
		}
		return nil, err
	}

	return &Config{value.Interface().(map[string]interface{})}, nil
}

func (c *Config) SetBool(name string, idx int, value bool) {
	c.setValue(name, idx, reflect.ValueOf(value))
}

func (c *Config) SetInt(name string, idx int, value int) {
	c.setValue(name, idx, reflect.ValueOf(value))
}

func (c *Config) SetFloat(name string, idx int, value float64) {
	c.setValue(name, idx, reflect.ValueOf(value))
}

func (c *Config) SetString(name string, idx int, value string) {
	c.setValue(name, idx, reflect.ValueOf(value))
}

func (c *Config) SetChild(name string, idx int, value *Config) {
	c.setValue(name, idx, reflect.ValueOf(value.fields))
}

func (c *Config) setValue(name string, idx int, newValue reflect.Value) {
	old, ok := c.fields[name]
	if !ok {
		if idx == 0 {
			c.fields[name] = newValue.Interface()
		} else {
			slice := make([]interface{}, idx+1)
			slice[idx] = newValue.Interface()
			c.fields[name] = slice
		}
		return
	}

	vOld := reflect.ValueOf(old)
	tOld := vOld.Type()
	if tOld.Kind() != reflect.Array {
		if idx == 0 {
			c.fields[name] = newValue.Interface()
		} else {
			slice := make([]interface{}, idx+1)
			slice[idx] = newValue.Interface()
			c.fields[name] = slice
		}
		return
	}

	if idx >= vOld.Len() {
		vOld.SetLen(idx + 1)
	}
	vOld.Index(idx).Set(newValue)
}

func (c *Config) getFieldValue(
	name string,
	idx int,
	typ reflect.Type,
) (reflect.Value, error) {
	ifc, ok := c.fields[name]
	if !ok {
		return errGet(ErrMissing)
	}

	value := reflect.ValueOf(ifc)
	kind := value.Kind()
	if kind != reflect.Array && idx > 0 {
		return errGet(ErrTypeNoArray)
	} else if kind == reflect.Array {
		if idx >= value.Len() {
			return errGet(ErrIndexOutOfRange)
		}

		value = reflect.ValueOf(value.Index(idx).Interface())
		kind = value.Kind()
	}

	tValue := value.Type()
	if !tValue.ConvertibleTo(typ) {
		return errGet(ErrTypeMismatch)
	}

	return value.Convert(typ), nil
}

func errGet(err error) (reflect.Value, error) {
	return reflect.Value{}, err
}
