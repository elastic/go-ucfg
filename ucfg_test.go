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
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

var opts = []Option{
	PathSep("."),
	ResolveEnv,
	VarExp,
}

func TestFlattenKeys(t *testing.T) {
	tests := []struct {
		name    string
		pathSep string
	}{
		{"withDot", "."},
		{"emptySep", ""},
	}

	sorted := func(s []string) []string {
		sort.Strings(s)
		return s
	}

	cfg := map[string]interface{}{
		"n.a.b.c": "hello",
		"n.a.d":   "world",
		"values": []interface{}{
			map[string]interface{}{
				"j": "j-value",
				"k": "k-value",
			},
			map[string]interface{}{
				"j": "r-value",
				"o": "o-value",
			},
		},
	}

	expected := sorted([]string{
		"n.a.b.c",
		"n.a.d",
		"values.0.j",
		"values.0.k",
		"values.1.j",
		"values.1.o",
	})

	for _, test := range tests {
		sep := test.pathSep
		t.Run(test.name, func(t *testing.T) {
			opts := []Option{}
			if sep != "" {
				opts = append(opts, PathSep(sep))
			}

			c, err := NewFrom(cfg, opts...)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, expected, sorted(c.FlattenedKeys(opts...)))
		})
	}
}

func TestDetectCyclicReference(t *testing.T) {
	tests := []struct {
		title  string
		cfg    map[string]interface{}
		config interface{}
	}{
		{
			title: "direct reference on a struct",
			cfg: map[string]interface{}{
				"top.reference": "${top.reference}",
			},
			config: &struct {
				TopReference string `config:"top.reference"`
			}{},
		},
		{
			title: "direct compound reference on a struct",
			cfg: map[string]interface{}{
				"top.reference": "hello ${top.reference}",
			},
			config: &struct {
				TopReference string `config:"top.reference"`
			}{},
		},
		{
			title: "direct template reference on an empty map",
			cfg: map[string]interface{}{
				"top.reference": "hello ${top.reference}",
			},
			config: &map[string]interface{}{},
		},
		{
			title: "indirect template reference on an empty map",
			cfg: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": "hello ${a}",
					},
				},
			},
			config: &map[string]interface{}{},
		},
		{
			title: "indirect reference on an empty map",
			cfg: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": "${a}",
					},
				},
			},
			config: &map[string]interface{}{},
		},
		{
			title: "direct array reference into an empty map",
			cfg: map[string]interface{}{
				"c": []string{
					"a",
					"${c.1}",
				},
			},
			config: &map[string]interface{}{},
		},
		{
			title: "direct array reference into an empty map",
			cfg: map[string]interface{}{
				"c": []string{
					"a",
					"${c.1}",
				},
			},
			config: &map[string]interface{}{},
		},
		{
			title: "direct array reference into a struct",
			cfg: map[string]interface{}{
				"c": []string{
					"a",
					"${c.1}",
				},
			},
			config: &struct {
				C []string `config:"c"`
			}{},
		},
	}

	for _, test := range tests {

		t.Run(test.title, func(t *testing.T) {
			c, err := NewFrom(test.cfg, opts...)
			assert.NoError(t, err)

			err = c.Unpack(test.config, opts...)
			assert.Error(t, err)
		})
	}
}

func TestCyclicReferenceShouldFallbackToOtherResolvers(t *testing.T) {
	cfg := map[string]interface{}{
		"top.reference": "${top.reference}",
	}

	resolveFn := func(key string) (string, error) {
		if key == "top.reference" {
			return "reference-found", nil
		}
		return "", ErrMissing
	}

	opts := []Option{
		PathSep("."),
		Resolve(resolveFn),
		ResolveEnv,
		VarExp,
	}

	c, err := NewFrom(cfg, opts...)
	v, err := c.String("top.reference", -1, opts...)
	if assert.NoError(t, err) {
		assert.Equal(t, "reference-found", v)
	}
}

func TestTopYamlKeyInEnvResolvers(t *testing.T) {
	resolveFn := func(key string) (string, error) {
		if key == "a.key" {
			return "key-found", nil
		}
		return "", fmt.Errorf("could not find the key: %s", key)
	}

	opts := []Option{
		PathSep("."),
		Resolve(resolveFn),
		ResolveEnv,
		VarExp,
	}

	tests := []struct {
		name     string
		cfg      interface{}
		expected string
	}{
		{
			name: "top level key reference exists",
			cfg: map[string]interface{}{
				"a.top":         "top-level",
				"f.l.reference": "${a.key}",
			},
		},
		{
			name: "top level key reference doesn't exist",
			cfg: map[string]interface{}{
				"f.l.reference": "${a.key}",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, err := NewFrom(test.cfg, opts...)
			assert.NoError(t, err)

			v, err := c.String("f.l.reference", -1, opts...)
			if assert.NoError(t, err) {
				assert.Equal(t, "key-found", v)
			}
		})
	}
}

func TestMultipleDirectReference(t *testing.T) {
	cfg := map[string]interface{}{
		"path.home": "hello",
		"path.logs": "${path.home}",
		"output": map[string]interface{}{
			"path": "${path.logs}",
		},
	}

	c := New()
	err := c.Merge(cfg, opts...)
	assert.NoError(t, err)

	t.Run("into a struct", func(t *testing.T) {
		config := struct {
			Path struct {
				Home string `config:"home"`
				Logs string `config:"logs"`
			} `config:"path"`
			Output struct {
				Path string `config:"path"`
			} `config:"output"`
		}{}

		err = c.Unpack(&config, opts...)
		if assert.NoError(t, err) {
			assert.Equal(t, "hello", config.Output.Path)
		}
	})

	t.Run("into a map", func(t *testing.T) {
		m := map[string]interface{}{}
		err = c.Unpack(&m, opts...)
		if assert.NoError(t, err) {
			v := m["output"]
			output := v.(map[string]interface{})
			path := output["path"]
			assert.Equal(t, "hello", path)
		}
	})
}

func TestResolveNOOP(t *testing.T) {
	opts := []Option{
		PathSep("."),
		ResolveNOOP,
	}

	cfg := map[string]interface{}{
		"a.top":         "top-level",
		"f.l.reference": "${a.key}",
	}

	c, err := NewFrom(cfg, opts...)
	assert.NoError(t, err)

	v, err := c.String("f.l.reference", -1, opts...)
	if assert.NoError(t, err) {
		assert.Equal(t, "${a.key}", v)
	}
}

func TestHas(t *testing.T) {
	cases := map[string]struct {
		cfg  map[string]interface{}
		has  bool
		fail bool
		path string
		idx  int
	}{
		"exists": {
			cfg:  map[string]interface{}{"a": 1},
			has:  true,
			path: "a", idx: -1,
		},
		"unkown": {
			cfg:  map[string]interface{}{"a": 1},
			has:  false,
			path: "b", idx: -1,
		},
		"nested find primitive": {
			cfg:  map[string]interface{}{"a.b": 1},
			has:  true,
			path: "a.b", idx: -1,
		},
		"nested find obj": {
			cfg:  map[string]interface{}{"a.b.c": 1},
			has:  true,
			path: "a.b", idx: -1,
		},
		"unknown nested": {
			cfg:  map[string]interface{}{"a.b.c": 1},
			has:  false,
			path: "a.d", idx: -1,
		},
		"array in len": {
			cfg:  map[string]interface{}{"a": []interface{}{1, 2}},
			has:  true,
			path: "a", idx: 1,
		},
		"array out of index": {
			cfg:  map[string]interface{}{"a": []interface{}{1, 2}},
			has:  false,
			path: "a", idx: 10,
		},
		"intermediate is primitive": {
			cfg:  map[string]interface{}{"a.b": 1},
			fail: true,
			path: "a.b.c", idx: -1,
		},
	}

	for name, test := range cases {
		test := test
		t.Run(name, func(t *testing.T) {
			opts := []Option{PathSep("."), VarExp}

			cfg := MustNewFrom(test.cfg, opts...)
			b, err := cfg.Has(test.path, test.idx, opts...)
			assert.Equal(t, test.has, b)
			assert.Equal(t, test.fail, err != nil)
		})
	}
}

func TestRemove(t *testing.T) {
	type spec struct {
		has   bool
		fails bool
		path  string
		idx   int
	}

	cases := map[string]struct {
		cfg   map[string]interface{}
		wants map[string]interface{}
		spec  spec
	}{
		"exist": {
			cfg:   map[string]interface{}{"field": "test"},
			wants: nil,
			spec:  spec{has: true, path: "field", idx: -1},
		},
		"unknown field": {
			cfg:   map[string]interface{}{"field": "test"},
			wants: map[string]interface{}{"field": "test"},
			spec:  spec{has: false, path: "unknown", idx: -1},
		},
		"nested": {
			cfg:   map[string]interface{}{"a.b": "keep", "a.c": "remove"},
			wants: map[string]interface{}{"a": map[string]interface{}{"b": "keep"}},
			spec:  spec{has: true, path: "a.c", idx: -1},
		},
		"unknown nested": {
			cfg:   map[string]interface{}{"a.b": "keep", "a.c": "keep"},
			wants: map[string]interface{}{"a": map[string]interface{}{"b": "keep", "c": "keep"}},
			spec:  spec{has: false, path: "a.d", idx: -1},
		},
		"array start": {
			cfg:   map[string]interface{}{"arr": []interface{}{"a", "b", "c"}},
			wants: map[string]interface{}{"arr": []interface{}{"b", "c"}},
			spec:  spec{has: true, path: "arr", idx: 0},
		},
		"array end": {
			cfg:   map[string]interface{}{"arr": []interface{}{"a", "b", "c"}},
			wants: map[string]interface{}{"arr": []interface{}{"a", "b"}},
			spec:  spec{has: true, path: "arr", idx: 2},
		},
		"array middle": {
			cfg:   map[string]interface{}{"arr": []interface{}{"a", "b", "c"}},
			wants: map[string]interface{}{"arr": []interface{}{"a", "c"}},
			spec:  spec{has: true, path: "arr", idx: 1},
		},
		"array out of index": {
			cfg:   map[string]interface{}{"arr": []interface{}{"a", "b", "c"}},
			wants: map[string]interface{}{"arr": []interface{}{"a", "b", "c"}},
			spec:  spec{has: false, path: "arr", idx: 5},
		},
		"full namespace": {
			cfg:   map[string]interface{}{"a.b": 1, "a.c": 2, "x": "keep"},
			wants: map[string]interface{}{"x": "keep"},
			spec:  spec{has: true, path: "a", idx: -1},
		},
		"shared via reference": {
			cfg: map[string]interface{}{"a": "${b}", "b.c": "remove", "b.d": "keep"},
			wants: map[string]interface{}{
				"a": map[string]interface{}{
					"d": "keep",
				},
				"b": map[string]interface{}{
					"d": "keep",
				},
			},
			spec: spec{has: true, path: "a.c", idx: -1},
		},
		"fail if no object": {
			cfg:  map[string]interface{}{"a": "test"},
			spec: spec{fails: true, path: "a.b", idx: -1},
		},
	}

	for name, test := range cases {
		test := test
		t.Run(name, func(t *testing.T) {
			opts := []Option{PathSep("."), VarExp}

			cfg := MustNewFrom(test.cfg, opts...)
			b, err := cfg.Remove(test.spec.path, test.spec.idx, opts...)
			if err != nil && !test.spec.fails {
				t.Fatal("unexpected error:", err)
			}
			assert.Equal(t, test.spec.has, b, "unexpected remove operation result")

			if test.spec.fails {
				return
			}

			var actual map[string]interface{}
			if err := cfg.Unpack(&actual, opts...); err != nil {
				t.Fatal("unpacking config after remove failed:", err)
			}

			assert.Equal(t, test.wants, actual)
		})
	}
}
