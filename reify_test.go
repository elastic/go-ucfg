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
			U uint
			F float64
			S string
		}{},
		&struct {
			B interface{}
			I interface{}
			U interface{}
			F interface{}
			S interface{}
		}{},
		&struct {
			B *bool
			I *int
			U *uint
			F *float64
			S *string
		}{},
	}

	c, _ := NewFrom(node{
		"b": true,
		"i": 42,
		"u": 23,
		"f": 3.14,
		"s": "string",
	})

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

		c, err := NewFrom(in)
		if err != nil {
			t.Errorf("failed")
			continue
		}

		b, err := c.Bool("b", -1)
		assert.NoError(t, err)

		i, err := c.Int("i", -1)
		assert.NoError(t, err)

		u, err := c.Uint("u", -1)
		assert.NoError(t, err)

		f, err := c.Float("f", -1)
		assert.NoError(t, err)

		s, err := c.String("s", -1)
		assert.NoError(t, err)

		assert.Equal(t, true, b)
		assert.Equal(t, 42, int(i))
		assert.Equal(t, 23, int(u))
		assert.Equal(t, 3.14, f)
		assert.Equal(t, "string", s)
	}
}

func TestUnpackPrimitivesValuesResolve(t *testing.T) {
	tests := []interface{}{
		New(),
		&map[string]interface{}{},
		map[string]interface{}{},
		node{},
		&node{},
		&struct {
			B bool
			I int
			U uint
			F float64
			S string
		}{},
		&struct {
			B interface{}
			I interface{}
			U interface{}
			F interface{}
			S interface{}
		}{},
		&struct {
			B *bool
			I *int
			U *uint
			F *float64
			S *string
		}{},
	}

	cfgOpts := []Option{
		VarExp,
		Resolve(func(name string) (string, error) {
			return map[string]string{
				"v_b": "true",
				"v_i": "42",
				"v_u": "23",
				"v_f": "3.14",
				"v_s": "string",
			}[name], nil
		}),
	}

	c, _ := NewFrom(node{
		"b": "${v_b}",
		"i": "${v_i}",
		"u": "${v_u}",
		"f": "${v_f}",
		"s": "${v_s}",
	}, cfgOpts...)

	for i, out := range tests {
		t.Logf("test unpack primitives(%v) into: %v", i, out)
		err := c.Unpack(out, cfgOpts...)
		if err != nil {
			t.Fatalf("failed to unpack: %v", err)
		}
	}

	// validate content by merging struct
	for i, in := range tests {
		t.Logf("test unpack primitives(%v) check: %v", i, in)

		c, err := NewFrom(in, cfgOpts...)
		if err != nil {
			t.Errorf("failed")
			continue
		}

		b, err := c.Bool("b", -1, cfgOpts...)
		assert.NoError(t, err)

		i, err := c.Int("i", -1, cfgOpts...)
		assert.NoError(t, err)

		u, err := c.Uint("u", -1, cfgOpts...)
		assert.NoError(t, err)

		f, err := c.Float("f", -1, cfgOpts...)
		assert.NoError(t, err)

		s, err := c.String("s", -1, cfgOpts...)
		assert.NoError(t, err)

		assert.Equal(t, true, b)
		assert.Equal(t, 42, int(i))
		assert.Equal(t, 23, int(u))
		assert.Equal(t, 3.14, f)
		assert.Equal(t, "string", s)
	}
}

func TestUnpackNested(t *testing.T) {
	var genSub = func(name string) *Config {
		s := New()
		s.SetBool(name, 0, false)
		return s
	}

	sub, _ := NewFrom(node{"b": true})
	c, _ := NewFrom(node{"c": sub})

	t.Logf("sub: %v", sub)
	t.Logf("c: %v", c)

	tests := []interface{}{
		New(),

		newC(),

		map[string]interface{}{},
		map[string]*Config{},
		map[string]*C{},
		map[string]map[string]bool{},
		map[string]map[string]interface{}{},
		map[string]interface{}{
			"c": map[string]interface{}{
				"b": false,
			},
		},
		map[string]interface{}{
			"c": nil,
		},
		map[string]*Config{
			"c": nil,
		},
		map[string]*C{
			"c": nil,
		},
		map[string]interface{}{
			"c": New(),
		},
		map[string]interface{}{
			"c": newC(),
		},
		map[string]interface{}{
			"c": genSub("b"),
		},
		map[string]interface{}{
			"c": genSub("d"),
		},
		map[string]interface{}{
			"c": fromConfig(genSub("b")),
		},
		map[string]interface{}{
			"c": fromConfig(genSub("d")),
		},
		map[string]*struct{ B bool }{},
		map[string]*struct{ B bool }{"c": nil},
		map[string]struct{ B bool }{},

		node{},
		node{"c": node{}},
		node{"c": node{"b": false}},
		node{"c": genSub("d")},
		node{"c": fromConfig(genSub("d"))},

		&struct{ C *Config }{},
		&struct{ C *Config }{sub},
		&struct{ C *Config }{genSub("d")},
		&struct{ C map[string]interface{} }{},
		&struct{ C node }{},
		&struct{ C struct{ B bool } }{},
		&struct{ C *struct{ B bool } }{&struct{ B bool }{}},
		&struct{ C *struct{ B bool } }{},

		&struct{ C *C }{},
		&struct{ C *C }{fromConfig(sub)},
		&struct{ C *C }{fromConfig(genSub("d"))},
	}

	for i, out := range tests {
		t.Logf("test unpack nested(%v) into: %v", i, out)
		err := c.Unpack(out)
		if err != nil {
			t.Fatalf("failed to unpack: %v", err)
		}
	}

	// validate content by merging struct
	for i, in := range tests {
		t.Logf("test unpack nested(%v) check: %v", i, in)

		c, err := NewFrom(in)
		if err != nil {
			t.Errorf("failed")
			continue
		}

		sub, err := c.Child("c", -1)
		assert.NoError(t, err)

		b, err := sub.Bool("b", -1)
		assert.NoError(t, err)
		assert.True(t, b)
	}
}

func TestUnpackNestedPath(t *testing.T) {
	tests := []interface{}{
		&struct {
			B bool `config:"c.b"`
		}{},

		&struct {
			B interface{} `config:"c.b"`
		}{},
	}

	sub, _ := NewFrom(node{"b": true})
	c, _ := NewFrom(node{"c": sub})

	for i, out := range tests {
		t.Logf("test unpack nested path(%v) into: %v", i, out)
		err := c.Unpack(out, PathSep("."))
		if err != nil {
			t.Fatalf("failed to unpack: %v", err)
		}
	}

	// validate content by merging struct (unnested)
	for i, in := range tests {
		t.Logf("test unpack nested(%v) check: %v", i, in)

		c, err := NewFrom(in)
		if err != nil {
			t.Errorf("failed")
			continue
		}

		b, err := c.Bool("c.b", 0)
		assert.NoError(t, err)
		assert.True(t, b)
	}
}

func TestUnpackArray(t *testing.T) {
	c, _ := NewFrom(node{"a": []int{1, 2, 3}})

	tests := []interface{}{
		map[string]interface{}{},
		map[string]interface{}{
			"a": []int{},
		},
		map[string][]int{"a": {}},
		map[string]interface{}{
			"a": []interface{}{},
		},
		map[string][]int{},

		node{},
		node{
			"a": []int{},
		},
		node{
			"a": []interface{}{},
		},

		&struct{ A []int }{},
		&struct{ A []uint }{},
		&struct{ A []interface{} }{},
		&struct{ A interface{} }{},
		&struct{ A [3]int }{},
		&struct{ A [3]uint }{},
		&struct{ A [3]interface{} }{},
	}

	for i, out := range tests {
		t.Logf("test unpack array(%v) into: %v", i, out)
		err := c.Unpack(out)
		if err != nil {
			t.Fatalf("failed to unpack: %v", err)
		}
	}

	// validate content by merging struct
	for i, in := range tests {
		t.Logf("test unpack nested(%v) check: %v", i, in)

		c, err := NewFrom(in)
		if err != nil {
			t.Errorf("failed")
			continue
		}

		for i := 0; i < 3; i++ {
			v, err := c.Int("a", i)
			assert.NoError(t, err)
			assert.Equal(t, i+1, int(v))
		}
	}
}

func TestUnpackArrayDirect(t *testing.T) {
	c, _ := NewFrom(node{"a": []int{1, 2, 3}})
	a, _ := c.Child("a", -1)

	var table []int
	err := a.Unpack(&table)
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, table)

	tmp := struct {
		Table []int `config:",inline"`
	}{}
	err = a.Unpack(&tmp)
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, tmp.Table)
}

func TestUnpackInline(t *testing.T) {
	type SubType struct{ B bool }
	type SubInterface struct{ B interface{} }

	tests := []interface{}{
		&struct {
			C SubType `config:",inline"`
		}{SubType{true}},
		&struct {
			SubType `config:",inline"`
		}{SubType{true}},

		&struct {
			C SubInterface `config:",inline"`
		}{SubInterface{true}},
		&struct {
			SubInterface `config:",inline"`
		}{SubInterface{true}},

		&struct {
			C map[string]bool `config:",inline"`
		}{map[string]bool{"b": true}},

		&struct {
			C map[string]interface{} `config:",inline"`
		}{map[string]interface{}{"b": true}},

		&struct {
			C node `config:",inline"`
		}{node{"b": true}},
	}

	c, _ := NewFrom(map[string]bool{"b": true})

	for i, out := range tests {
		t.Logf("test unpack with inline(%v) into: %v", i, out)
		err := c.Unpack(out)
		if err != nil {
			t.Fatalf("failed to unpack: %v", err)
		}
	}

	// validate content by merging struct
	for i, in := range tests {
		t.Logf("test unpack inlined(%v) check: %v", i, in)

		c, err := NewFrom(in)
		if err != nil {
			t.Fatalf("failed with: %v", err)
		}

		b, err := c.Bool("b", -1)
		assert.NoError(t, err)
		assert.Equal(t, true, b)
	}
}

func TestUnpackUnknown(t *testing.T) {
	c := New()

	tests := []interface{}{
		&struct {
			B bool   `config:"b"`
			I int    `config:"i"`
			U uint   `config:"u"`
			S string `config:"s"`
		}{true, 23, 42, "test"},

		map[string]interface{}{
			"b": true,
			"i": 23,
			"u": 42,
			"s": "test",
		},

		node{
			"b": true,
			"i": 23,
			"u": 42,
			"s": "test",
		},
	}

	for i, test := range tests {
		t.Logf("test (%v): %v", i, test)

		err := c.Unpack(test)
		if err != nil {
			assert.NoError(t, err)
			continue
		}

		t.Logf("unpacked empty (%v): %v", i, test)

		tmp, err := NewFrom(test, PathSep("."))
		if err != nil {
			assert.NoError(t, err)
			continue
		}

		b, err := tmp.Bool("b", -1, PathSep("."))
		assert.NoError(t, err)
		assert.Equal(t, true, b)

		i, err := tmp.Int("i", -1, PathSep("."))
		assert.NoError(t, err)
		assert.Equal(t, 23, int(i))

		u, err := tmp.Uint("u", -1, PathSep("."))
		assert.NoError(t, err)
		assert.Equal(t, 42, int(u))

		s, err := tmp.String("s", -1, PathSep("."))
		assert.NoError(t, err)
		assert.Equal(t, "test", s)
	}
}

func TestUnpackUnknownNested(t *testing.T) {
	c, _ := NewFrom(map[string]interface{}{
		"s": nil,
	})

	tests := []interface{}{
		&struct {
			B bool   `config:"s.b"`
			I int    `config:"s.i"`
			U uint   `config:"s.u"`
			S string `config:"s.s"`
		}{true, 23, 42, "test"},

		node{
			"s": node{
				"b": true,
				"i": 23,
				"u": 42,
				"s": "test",
			},
		},

		node{
			"s": &struct {
				B bool
				I int
				U uint
				S string
			}{true, 23, 42, "test"},
		},
	}

	for i, test := range tests {
		t.Logf("test (%v): %v", i, test)

		err := c.Unpack(test)
		if err != nil {
			assert.NoError(t, err)
			continue
		}

		t.Logf("unpacked empty (%v): %v", i, test)

		tmp, err := NewFrom(test, PathSep("."))
		if err != nil {
			assert.NoError(t, err)
			continue
		}

		b, err := tmp.Bool("s.b", -1, PathSep("."))
		assert.NoError(t, err)
		assert.Equal(t, true, b)

		i, err := tmp.Int("s.i", -1, PathSep("."))
		assert.NoError(t, err)
		assert.Equal(t, 23, int(i))

		u, err := tmp.Uint("s.u", -1, PathSep("."))
		assert.NoError(t, err)
		assert.Equal(t, 42, int(u))

		s, err := tmp.String("s.s", -1, PathSep("."))
		assert.NoError(t, err)
		assert.Equal(t, "test", s)
	}
}

func TestUnpackStructWithConfig(t *testing.T) {
	type target struct {
		to     interface{}
		config **Config
	}

	type testCase struct {
		target   target
		config   interface{}
		expected interface{}
	}

	cases := map[string]testCase{
		"merge config by default": testCase{
			target: func() target {
				to := struct{ Config *Config }{}
				to.Config = MustNewFrom(map[string]interface{}{"a": 1})
				return target{&to, &to.Config}
			}(),

			config: map[string]interface{}{
				"config": map[string]interface{}{"b": 2},
			},

			expected: map[string]interface{}{
				"a": uint64(1),
				"b": uint64(2),
			},
		},

		"replace config": testCase{
			target: func() target {
				to := struct {
					Config *Config `config:",replace"`
				}{}
				to.Config = MustNewFrom(map[string]interface{}{"a": 1})
				return target{&to, &to.Config}
			}(),

			config: map[string]interface{}{
				"config": map[string]interface{}{"b": 2},
			},

			expected: map[string]interface{}{
				"b": uint64(2),
			},
		},

		"merge array by default": testCase{
			target: func() target {
				to := struct{ Config *Config }{}
				to.Config = MustNewFrom([]interface{}{
					map[string]interface{}{"a": 1},
				})
				return target{&to, &to.Config}
			}(),

			config: map[string]interface{}{
				"config": []interface{}{
					map[string]interface{}{"b": 2},
				},
			},

			expected: []map[string]interface{}{
				{"a": uint64(1), "b": uint64(2)},
			},
		},

		"replace array": testCase{
			target: func() target {
				to := struct {
					Config *Config `config:",replace"`
				}{}
				to.Config = MustNewFrom([]interface{}{
					map[string]interface{}{"a": 1},
					map[string]interface{}{"c": 1},
				})
				return target{&to, &to.Config}
			}(),

			config: map[string]interface{}{
				"config": []interface{}{
					map[string]interface{}{
						"b": 2,
					},
				},
			},

			expected: []map[string]interface{}{
				{"b": uint64(2)},
			},
		},

		"append array": testCase{
			target: func() target {
				to := struct {
					Config *Config `config:",append"`
				}{}
				to.Config = MustNewFrom([]interface{}{
					map[string]interface{}{"a": 1},
				})
				return target{&to, &to.Config}
			}(),

			config: map[string]interface{}{
				"config": []interface{}{
					map[string]interface{}{"b": 2},
				},
			},

			expected: []map[string]interface{}{
				{"a": uint64(1)},
				{"b": uint64(2)},
			},
		},

		"prepend array": testCase{
			target: func() target {
				to := struct {
					Config *Config `config:",prepend"`
				}{}
				to.Config = MustNewFrom([]interface{}{
					map[string]interface{}{"a": 1},
				})
				return target{&to, &to.Config}
			}(),

			config: map[string]interface{}{
				"config": []interface{}{
					map[string]interface{}{"b": 2},
				},
			},

			expected: []map[string]interface{}{
				{"b": uint64(2)},
				{"a": uint64(1)},
			},
		},
	}

	for name, test := range cases {
		test := test
		t.Run(name, func(t *testing.T) {
			config := MustNewFrom(test.config)
			if assert.NoError(t, config.Unpack(test.target.to)) {
				assertConfig(t, *test.target.config, test.expected)
			}
		})
	}
}

func TestUnpackStructWithArrConfig(t *testing.T) {
	type target struct {
		to     interface{}
		config *[]*Config
	}

	type testCase struct {
		target   target
		config   interface{}
		expected interface{}
	}

	cases := map[string]testCase{
		"merge entries by default": testCase{
			target: func() target {
				to := struct{ Configs []*Config }{}
				to.Configs = []*Config{
					MustNewFrom(map[string]interface{}{"a": 1}),
					MustNewFrom(map[string]interface{}{"c": 3}),
				}
				return target{&to, &to.Configs}
			}(),

			config: map[string]interface{}{
				"configs": []interface{}{
					map[string]interface{}{"b": 2},
				},
			},

			expected: []map[string]interface{}{
				{"a": uint64(1), "b": uint64(2)},
				{"c": uint64(3)},
			},
		},

		"replace array": testCase{
			target: func() target {
				to := struct {
					Configs []*Config `config:",replace"`
				}{}
				to.Configs = []*Config{
					MustNewFrom(map[string]interface{}{"a": 1}),
					MustNewFrom(map[string]interface{}{"c": 3}),
				}
				return target{&to, &to.Configs}
			}(),

			config: map[string]interface{}{
				"configs": []interface{}{
					map[string]interface{}{"b": 2},
				},
			},

			expected: []map[string]interface{}{
				{"b": uint64(2)},
			},
		},

		"append": testCase{
			target: func() target {
				to := struct {
					Configs []*Config `config:",append"`
				}{}
				to.Configs = []*Config{
					MustNewFrom(map[string]interface{}{"a": 1}),
					MustNewFrom(map[string]interface{}{"c": 3}),
				}
				return target{&to, &to.Configs}
			}(),

			config: map[string]interface{}{
				"configs": []interface{}{
					map[string]interface{}{"b": 2},
				},
			},

			expected: []map[string]interface{}{
				{"a": uint64(1)},
				{"c": uint64(3)},
				{"b": uint64(2)},
			},
		},

		"prepend": testCase{
			target: func() target {
				to := struct {
					Configs []*Config `config:",prepend"`
				}{}
				to.Configs = []*Config{
					MustNewFrom(map[string]interface{}{"a": 1}),
					MustNewFrom(map[string]interface{}{"c": 3}),
				}
				return target{&to, &to.Configs}
			}(),

			config: map[string]interface{}{
				"configs": []interface{}{
					map[string]interface{}{"b": 2},
				},
			},

			expected: []map[string]interface{}{
				{"b": uint64(2)},
				{"a": uint64(1)},
				{"c": uint64(3)},
			},
		},
	}

	for name, test := range cases {
		test := test
		t.Run(name, func(t *testing.T) {
			config := MustNewFrom(test.config)
			err := config.Unpack(test.target.to)
			if assert.NoError(t, err) {
				tmp := New()
				for i, sub := range *test.target.config {
					if sub == nil {
						t.Fatalf("array %v entry is nil", i)
					}
					tmp.SetChild("", i, sub)
				}

				assertConfig(t, tmp, test.expected)
			}
		})
	}
}

func assertConfig(t *testing.T, config *Config, expected interface{}) {
	var actual interface{}

	if config.IsArray() {
		var tmp []map[string]interface{}
		err := config.Unpack(&tmp)
		if !assert.NoError(t, err) {
			return
		}
		actual = tmp
	} else {
		var tmp map[string]interface{}
		err := config.Unpack(&tmp)
		if !assert.NoError(t, err) {
			return
		}
		actual = tmp
	}

	assert.Equal(t, expected, actual)
}
