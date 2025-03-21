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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stUnpackable struct {
	value int
}

type primUnpackable int

type (
	unpackBool    struct{ b bool }
	unpackInt     struct{ i int }
	unpackUint    struct{ u int }
	unpackFloat   struct{ f float64 }
	unpackString  struct{ s string }
	unpackConfig  struct{ c *Config }
	unpackRebrand struct{ c *Config }
)

func (u *unpackBool) Unpack(b bool) error      { u.b = b; return nil }
func (u *unpackInt) Unpack(i int64) error      { u.i = int(i); return nil }
func (u *unpackUint) Unpack(v uint64) error    { u.u = int(v); return nil }
func (u *unpackFloat) Unpack(f float64) error  { u.f = f; return nil }
func (u *unpackString) Unpack(s string) error  { u.s = s; return nil }
func (u *unpackConfig) Unpack(c *Config) error { u.c = c; return nil }
func (u *unpackRebrand) Unpack(c *C) error     { u.c = c.asConfig(); return nil }

func (s *stUnpackable) Unpack(v interface{}) error {
	i, err := unpackI(v)
	s.value = i
	return err
}

func (s *stUnpackable) Value() int {
	return s.value
}

func (p *primUnpackable) Unpack(v interface{}) error {
	i, err := unpackI(v)
	*p = primUnpackable(i)
	return err
}

func (p primUnpackable) Value() int {
	return int(p)
}

func unpackI(v interface{}) (int, error) {
	switch n := v.(type) {
	case int64:
		return int(n), nil
	case uint64:
		return int(n), nil
	case float64:
		return int(n), nil
	}

	m, ok := v.(map[string]interface{})
	if !ok {
		return 0, errors.New("expected dictionary")
	}

	val, ok := m["i"]
	if !ok {
		return 0, errors.New("missing field i")
	}

	switch n := val.(type) {
	case int64:
		return int(n), nil
	case uint64:
		return int(n), nil
	case float64:
		return int(n), nil
	default:
		return 0, errors.New("not a number")
	}
}

func TestReifyUnpackerInterface(t *testing.T) {
	cfg, _ := NewFrom(map[string]interface{}{
		"i": 10,
	})

	st := stUnpackable{}
	err := cfg.Unpack(&st)
	assert.NoError(t, err)
	assert.Equal(t, 10, st.Value())

	p := struct {
		I primUnpackable
	}{}
	err = cfg.Unpack(&p)
	assert.NoError(t, err)
	assert.Equal(t, 10, p.I.Value())
}

func TestReifyUnpackers(t *testing.T) {
	to := &struct {
		B unpackBool
		I unpackInt
		U unpackUint
		F unpackFloat
		S unpackString
		C unpackConfig
		R unpackRebrand
	}{}

	sub, _ := NewFrom(map[string]interface{}{"v": 1})
	expectedSub := map[string]interface{}{}
	if err := sub.Unpack(&expectedSub); err != nil {
		t.Fatal(err)
	}

	configs := []map[string]interface{}{
		{"b": true},
		{"i": -42},
		{"u": 23},
		{"f": 3.14},
		{"s": "string"},
		{"c": sub},
		{"r": sub},
	}

	// apply configurations
	for i, c := range configs {
		t.Logf("Unpacking config (%v): %#v", i, c)

		cfg, err := NewFrom(c)
		if err != nil {
			t.Fatal(err)
		}

		if err := cfg.Unpack(to); err != nil {
			t.Fatal(err)
		}
	}

	// validate unpackers
	assert.Equal(t, true, to.B.b)
	assert.Equal(t, -42, to.I.i)
	assert.Equal(t, 23, to.U.u)
	assert.Equal(t, 3.14, to.F.f)
	assert.Equal(t, "string", to.S.s)

	assertSubConfig := func(c *Config) {
		actual := map[string]interface{}{}
		if err := sub.Unpack(&actual); err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, expectedSub, actual)
	}
	assertSubConfig(to.C.c)
	assertSubConfig(to.R.c)
}

func TestReifyUnpackersPtr(t *testing.T) {
	to := &struct {
		B *unpackBool
		I *unpackInt
		U *unpackUint
		F *unpackFloat
		S *unpackString
		C *unpackConfig
		R *unpackRebrand
	}{}

	sub, _ := NewFrom(map[string]interface{}{"v": 1})
	expectedSub := map[string]interface{}{}
	if err := sub.Unpack(&expectedSub); err != nil {
		t.Fatal(err)
	}

	configs := []map[string]interface{}{
		{"b": true},
		{"i": -42},
		{"u": 23},
		{"f": 3.14},
		{"s": "string"},
		{"c": sub},
		{"r": sub},
	}

	// apply configurations
	for i, c := range configs {
		t.Logf("Unpacking config (%v): %#v", i, c)

		cfg, err := NewFrom(c)
		if err != nil {
			t.Fatal(err)
		}

		if err := cfg.Unpack(to); err != nil {
			t.Fatal(err)
		}
	}

	// validate unpackers
	assert.Equal(t, true, to.B.b)
	assert.Equal(t, -42, to.I.i)
	assert.Equal(t, 23, to.U.u)
	assert.Equal(t, 3.14, to.F.f)
	assert.Equal(t, "string", to.S.s)

	assertSubConfig := func(c *Config) {
		actual := map[string]interface{}{}
		if err := sub.Unpack(&actual); err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, expectedSub, actual)
	}
	assertSubConfig(to.C.c)
	assertSubConfig(to.R.c)
}

func TestUnpack(t *testing.T) {
	type CustomString string
	type StructWithCustomString struct {
		Foo CustomString `config:"foo"`
	}
	type StructWithCustomStringPtr struct {
		Foo *CustomString `config:"foo"`
	}

	type StructWithString struct {
		Foo string `config:"foo"`
	}
	type StructWithStringPtr struct {
		Foo *string `config:"foo"`
	}

	strPtr := func(s string) *string { return &s }
	customStrPtr := func(s CustomString) *CustomString { return &s }

	tests := map[string]struct {
		config   func() *Config
		unpackTo interface{}
		expect   interface{}
	}{
		"string to string": {
			config: func() *Config {
				return MustNewFrom(map[string]interface{}{
					"foo": "bar",
				})
			},
			unpackTo: &StructWithString{},
			expect:   &StructWithString{Foo: "bar"},
		},
		"string to *string": {
			config: func() *Config {
				return MustNewFrom(map[string]interface{}{
					"foo": "bar",
				})
			},
			unpackTo: &StructWithStringPtr{},
			expect:   &StructWithStringPtr{Foo: strPtr("bar")},
		},
		"string to CustomString": {
			config: func() *Config {
				return MustNewFrom(map[string]interface{}{
					"foo": "bar",
				})
			},
			unpackTo: &StructWithCustomString{},
			expect:   &StructWithCustomString{Foo: "bar"},
		},
		"string to *CustomString": {
			config: func() *Config {
				return MustNewFrom(map[string]interface{}{
					"foo": "bar",
				})
			},
			unpackTo: &StructWithCustomStringPtr{},
			expect:   &StructWithCustomStringPtr{Foo: customStrPtr("bar")},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			config := tc.config()
			err := config.Unpack(tc.unpackTo)
			require.NoError(t, err)
			require.Equal(t, tc.expect, tc.unpackTo)
		})
	}
}
