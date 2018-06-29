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

package json

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/go-ucfg"
)

func TestPrimitives(t *testing.T) {
	input := []byte(`
  {
    "b": true,
    "i": 42,
    "u": 23,
    "f": 3.14,
    "s": "string"
  }`)

	c, err := NewConfig(input)
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	verify := struct {
		B bool
		I int
		U uint
		F float64
		S string
	}{}
	err = c.Unpack(&verify)
	assert.Nil(t, err)

	assert.Equal(t, true, verify.B)
	assert.Equal(t, 42, verify.I)
	assert.Equal(t, uint(23), verify.U)
	assert.Equal(t, 3.14, verify.F)
	assert.Equal(t, "string", verify.S)
}

func TestNested(t *testing.T) {
	input := []byte(`
  {
    "c": {
      "b": true
    }
  }`)

	c, err := NewConfig(input)
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	var verify struct {
		C struct{ B bool }
	}
	err = c.Unpack(&verify)
	assert.NoError(t, err)
	assert.True(t, verify.C.B)
}

func TestNestedPath(t *testing.T) {
	input := []byte(`
  {
    "c.b": true
  }`)

	c, err := NewConfig(input, ucfg.PathSep("."))
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	var verify struct {
		C struct{ B bool }
	}
	err = c.Unpack(&verify)
	assert.NoError(t, err)
	assert.True(t, verify.C.B)
}

func TestArray(t *testing.T) {
	input := []byte(`
[
  {
    "b": 2,
    "c": 3
  },
  {
    "c": 4
  }
]
`)

	c, err := NewConfig(input)
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	verify := []map[string]int{}
	err = c.Unpack(&verify)
	assert.Nil(t, err)

	assert.Equal(t, verify[0]["b"], 2)
	assert.Equal(t, verify[0]["c"], 3)
	assert.Equal(t, verify[1]["c"], 4)
}
