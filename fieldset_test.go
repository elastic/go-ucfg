package ucfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldsetAddAField(t *testing.T) {
	fs := NewFieldSet(nil)
	fs.Add("hello")
	assert.True(t, fs.Has("hello"))
}

func TestFieldsetReturnsTheListOfFields(t *testing.T) {
	fs1 := NewFieldSet(nil)
	fs1.Add("hello")
	fs1.Add("bye")
	fs2 := NewFieldSet(fs1)
	fs2.Add("adios")
	assert.ElementsMatch(t, []string{"hello", "bye", "adios"}, fs2.Names())
}

func TestFieldSetHas(t *testing.T) {
	fs1 := NewFieldSet(nil)
	fs1.Add("parent")
	fs2 := NewFieldSet(fs1)
	fs2.Add("child")

	t.Run("ParentHasField", func(t *testing.T) {
		assert.True(t, fs2.Has("parent"))
	})

	t.Run("ChildAndParentDontHaveTheField", func(t *testing.T) {
		assert.False(t, fs2.Has("parent-doesnt"))
	})

	t.Run("ChildHasField", func(t *testing.T) {
		assert.True(t, fs2.Has("child"))
	})
}

func TestFieldSetAddNew(t *testing.T) {
	fs1 := NewFieldSet(nil)
	fs1.Add("parent")
	fs2 := NewFieldSet(fs1)
	fs2.Add("child")

	t.Run("ParentHasField", func(t *testing.T) {
		assert.False(t, fs2.AddNew("parent"))
	})

	t.Run("ChildAndParentDontHaveTheField", func(t *testing.T) {
		assert.True(t, fs2.AddNew("none"))
	})

	t.Run("ChildHasField", func(t *testing.T) {
		assert.False(t, fs2.AddNew("child"))
	})
}
