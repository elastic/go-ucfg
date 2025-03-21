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
