package ucfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnpackPrimitiveValues(t *testing.T) {
	tests := []interface{}{
		New(),
		&map[string]interface{}{},
		map[string]interface{}{},
		node{},
		&node{},
		&struct {
			B bool
			I int
			F float64
			S string
		}{},
		&struct {
			B interface{}
			I interface{}
			F interface{}
			S interface{}
		}{},
		&struct {
			B *bool
			I *int
			F *float64
			S *string
		}{},
	}

	c := New()
	c.SetBool("b", 0, true)
	c.SetInt("i", 0, 42)
	c.SetFloat("f", 0, 3.14)
	c.SetString("s", 0, "string")

	for i, out := range tests {
		t.Logf("test unpack primitives(%v) into: %v", i, out)
		err := c.Unpack(out)
		if err != nil {
			t.Fatalf("failed to unpack: %v", err)
		}
	}

	// validate content by merging struct
	for i, in := range tests {
		t.Logf("test unpack primitives(%v) check: %v", i, in)

		c := New()
		err := c.Merge(in)
		if err != nil {
			t.Errorf("failed")
			continue
		}

		b, err := c.Bool("b", 0)
		assert.NoError(t, err)

		i, err := c.Int("i", 0)
		assert.NoError(t, err)

		f, err := c.Float("f", 0)
		assert.NoError(t, err)

		s, err := c.String("s", 0)
		assert.NoError(t, err)

		assert.Equal(t, true, b)
		assert.Equal(t, 42, int(i))
		assert.Equal(t, 3.14, f)
		assert.Equal(t, "string", s)
	}
}
