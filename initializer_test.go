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

type structInitializer struct {
	I int
	J int
}

func (s *structInitializer) InitDefaults() {
	s.J = 10
}

type nestedStructInitializer struct {
	N structInitializer
	O int
}

func (n *nestedStructInitializer) InitDefaults() {
	n.O = 20

	// overridden by InitDefaults from structInitializer
	n.N.J = 15
}

func TestInitDefaultsSingle(t *testing.T) {
	c, _ := NewFrom(map[string]interface{}{
		"s": map[string]interface{}{
			"i": 5,
		},
	})

	// unpack S
	r := &struct {
		S structInitializer
	}{}

	err := c.Unpack(r)
	assert.NoError(t, err)
	assert.Equal(t, 5, r.S.I)
	assert.Equal(t, 10, r.S.J)
}

func TestInitDefaultsNested(t *testing.T) {
	c, _ := NewFrom(map[string]interface{}{
		"s": map[string]interface{}{
			"n": map[string]interface{}{
				"i": 5,
			},
		},
	})

	// unpack S
	r := &struct {
		S nestedStructInitializer
	}{}

	err := c.Unpack(r)
	assert.NoError(t, err)
	assert.Equal(t, 5, r.S.N.I)
	assert.Equal(t, 10, r.S.N.J)
	assert.Equal(t, 20, r.S.O)
}
