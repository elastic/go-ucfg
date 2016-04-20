package ucfg

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type myNonzeroInt int

func (m myNonzeroInt) Validate() error {
	if m == 0 {
		return errors.New("myNonzeroInt must not be 0")
	}
	return nil
}

func TestValidationPass(t *testing.T) {
	c, _ := NewFrom(map[string]interface{}{
		"a": 0,
		"b": 10,
		"d": -10,
		"f": 3.14,
	})

	tests := []interface{}{
		&struct {
			A int `validate:"positive"`
		}{},
		&struct {
			A int `validate:"positive,min=0"`
		}{},
		&struct {
			X int `config:"a" validate:"min=0"`
		}{},

		&struct {
			C int `validate:"nonzero"`
		}{},
		&struct {
			C myNonzeroInt
		}{},
		&struct {
			C int `validate:"positive"`
		}{},
		&struct {
			X int `config:"c" validate:"nonzero,min=-1"`
		}{},
		&struct {
			X int `config:"c" validate:"min=10, max=20"`
		}{},

		&struct {
			D int `validate:"nonzero"`
		}{},
		&struct {
			X int `config:"d" validate:"nonzero,min=-10"`
		}{},
		&struct {
			X int `config:"d" validate:"min=-10, max=0"`
		}{},

		&struct {
			F float64 `validate:"nonzero"`
		}{},
		&struct {
			F float64 `validate:"positive"`
		}{},
		&struct {
			X int `config:"f" validate:"nonzero,min=-1"`
		}{},
		&struct {
			X int `config:"f" validate:"min=3, max=20"`
		}{},
	}

	for i, test := range tests {
		t.Logf("Test config (%v): %#v", i, test)

		err := c.Unpack(test)
		assert.NoError(t, err)
	}
}

func TestValidationFail(t *testing.T) {
	c, _ := NewFrom(map[string]interface{}{
		"a": 0,
		"b": 10,
		"d": -10,
		"f": 3.14,
	})

	tests := []interface{}{
		&struct {
			X int `config:"a" validate:"nonzero"`
		}{},
		&struct {
			X myNonzeroInt `config:"a"`
		}{},
		&struct {
			X int `config:"a" validate:"min=10"`
		}{},

		&struct {
			X int `config:"b" validate:"max=8"`
		}{},
		&struct {
			X int `config:"b" validate:"min=20"`
		}{},

		&struct {
			X int `config:"d" validate:"positive"`
		}{},
		&struct {
			X int `config:"d" validate:"max=-11"`
		}{},
		&struct {
			X int `config:"d" validate:"min=20"`
		}{},

		&struct {
			X float64 `config:"f" validate:"max=1"`
		}{},
		&struct {
			X float64 `config:"f" validate:"min=20"`
		}{},
	}

	for i, test := range tests {
		t.Logf("Test config (%v): %#v", i, test)

		err := c.Unpack(test)
		assert.True(t, err != nil)
	}
}
