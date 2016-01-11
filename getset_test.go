package ucfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetGetPrimitives(t *testing.T) {
	c := New()

	c.SetBool("bool", 0, true)
	c.SetInt("int", 0, 42)
	c.SetFloat("float", 0, 2.3)
	c.SetString("str", 0, "abc")

	assert.True(t, c.HasField("bool"))
	assert.True(t, c.HasField("int"))
	assert.True(t, c.HasField("float"))
	assert.True(t, c.HasField("str"))
	assert.Len(t, c.GetFields(), 4)

	cnt, err := c.CountField("bool")
	assert.Nil(t, err)
	assert.Equal(t, 1, cnt)

	cnt, err = c.CountField("int")
	assert.Nil(t, err)
	assert.Equal(t, 1, cnt)

	cnt, err = c.CountField("float")
	assert.Nil(t, err)
	assert.Equal(t, 1, cnt)

	cnt, err = c.CountField("str")
	assert.Nil(t, err)
	assert.Equal(t, 1, cnt)

	b, err := c.Bool("bool", 0)
	assert.Nil(t, err)
	assert.Equal(t, true, b)

	i, err := c.Int("int", 0)
	assert.Nil(t, err)
	assert.Equal(t, 42, i)

	f, err := c.Float("float", 0)
	assert.Nil(t, err)
	assert.Equal(t, 2.3, f)

	s, err := c.String("str", 0)
	assert.Nil(t, err)
	assert.Equal(t, "abc", s)
}

func TestSetGetChild(t *testing.T) {
	var err error
	c := New()
	child := New()

	child.SetInt("test", 0, 42)
	c.SetChild("child", 0, child)

	child, err = c.Child("child", 0)
	assert.Nil(t, err)

	i, err := child.Int("test", 0)
	assert.Nil(t, err)
	assert.Equal(t, 42, i)
}
