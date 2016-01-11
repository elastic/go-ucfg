package ucfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeMapPrimitives(t *testing.T) {
	c := New()
	err := c.Merge(map[string]interface{}{
		"b": true,
		"i": 42,
		"f": 3.14,
		"s": "string",
	})
	assert.Nil(t, err)

	verify := struct {
		B bool
		I int
		F float64
		S string
	}{}
	err = c.Materialize(&verify)
	assert.Nil(t, err)

	assert.Equal(t, true, verify.B)
	assert.Equal(t, 42, verify.I)
	assert.Equal(t, 3.14, verify.F)
	assert.Equal(t, "string", verify.S)
}

func TestMergeStructPrimitives(t *testing.T) {
	type st struct {
		B bool
		I int
		F float64
		S string
	}

	c := New()
	err := c.Merge(st{
		B: true,
		I: 42,
		F: 3.14,
		S: "string",
	})

	verify := st{}
	err = c.Materialize(&verify)
	assert.Nil(t, err)

	assert.Equal(t, true, verify.B)
	assert.Equal(t, 42, verify.I)
	assert.Equal(t, 3.14, verify.F)
	assert.Equal(t, "string", verify.S)
}

func TestMergeMixedPrimitives(t *testing.T) {
	c := New()

	err := c.Merge(map[string]interface{}{
		"b": true,
		"i": 42,
	})
	assert.Nil(t, err)

	err = c.Merge(struct {
		F float64
		S string
	}{3.14, "string"})
	assert.Nil(t, err)

	verify := struct {
		B bool
		I int
		F float64
		S string
	}{}
	err = c.Materialize(&verify)
	assert.Nil(t, err)

	assert.Equal(t, true, verify.B)
	assert.Equal(t, 42, verify.I)
	assert.Equal(t, 3.14, verify.F)
	assert.Equal(t, "string", verify.S)
}

func TestMergeConfig(t *testing.T) {
	type st struct {
		B bool
		I int
		F float64
		S string
	}

	c := New()
	tmp := New()
	err := tmp.Merge(st{
		B: true,
		I: 42,
		F: 3.14,
		S: "string",
	})
	assert.Nil(t, err)

	err = c.Merge(tmp)
	assert.Nil(t, err)

	verify := st{}
	err = c.Materialize(&verify)
	assert.Nil(t, err)

	assert.Equal(t, true, verify.B)
	assert.Equal(t, 42, verify.I)
	assert.Equal(t, 3.14, verify.F)
	assert.Equal(t, "string", verify.S)

}
