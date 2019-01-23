// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package ucfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldsetAddAField(t *testing.T) {
	fs := newFieldSet(nil)
	fs.Add("hello")
	assert.True(t, fs.Has("hello"))
}

func TestFieldsetReturnsTheListOfFields(t *testing.T) {
	fs1 := newFieldSet(nil)
	fs1.Add("hello")
	fs1.Add("bye")
	fs2 := newFieldSet(fs1)
	fs2.Add("adios")
	assert.ElementsMatch(t, []string{"hello", "bye", "adios"}, fs2.Names())
}

func TestFieldSetHas(t *testing.T) {
	fs1 := newFieldSet(nil)
	fs1.Add("parent")
	fs2 := newFieldSet(fs1)
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
	fs1 := newFieldSet(nil)
	fs1.Add("parent")
	fs2 := newFieldSet(fs1)
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
