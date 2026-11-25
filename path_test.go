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
	"reflect"
	"testing"
)

func Test_parsePathWithOpts(t *testing.T) {
	type args struct {
		in   string
		opts *options
	}
	tests := []struct {
		name string
		args args
		want cfgPath
	}{
		{
			name: "happy path",
			args: args{
				in:   "a.b",
				opts: &options{pathSep: "."},
			},
			want: cfgPath{
				fields: []field{
					namedField{"a"},
					namedField{"b"},
				},
				sep: ".",
			},
		},
		{
			name: "escape path",
			args: args{
				in:   "[a.b]",
				opts: &options{pathSep: ".", escapePath: true},
			},
			want: cfgPath{
				fields: []field{
					namedField{"[a.b]"}},
				sep: ".",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parsePathWithOpts(tt.args.in, tt.args.opts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePathWithOpts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parsePathWithOpts_negativeIndex(t *testing.T) {
	opts := &options{pathSep: ".", maxIdx: defaultMaxIdx}

	tests := []struct {
		name string
		in   string
		want cfgPath
	}{
		{
			name: "negative segment in the middle is a named field",
			in:   "a.-1.b",
			want: cfgPath{fields: []field{namedField{"a"}, namedField{"-1"}, namedField{"b"}}, sep: "."},
		},
		{
			name: "negative last segment is a named field",
			in:   "a.-1",
			want: cfgPath{fields: []field{namedField{"a"}, namedField{"-1"}}, sep: "."},
		},
		{
			name: "single negative segment is a named field",
			in:   "-1",
			want: cfgPath{fields: []field{namedField{"-1"}}, sep: "."},
		},
		{
			name: "negative segment in another base is a named field",
			in:   "a.-0x1.b",
			want: cfgPath{fields: []field{namedField{"a"}, namedField{"-0x1"}, namedField{"b"}}, sep: "."},
		},
		{
			name: "non-negative segment is still an index",
			in:   "a.1.b",
			want: cfgPath{fields: []field{namedField{"a"}, idxField{1}, namedField{"b"}}, sep: "."},
		},
		{
			name: "zero is still an index",
			in:   "a.0",
			want: cfgPath{fields: []field{namedField{"a"}, idxField{0}}, sep: "."},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parsePathWithOpts(tt.in, opts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePathWithOpts(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
