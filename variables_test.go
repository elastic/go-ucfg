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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVarExpParserSuccess(t *testing.T) {
	str := func(s string) varEvaler { return constExp(s) }
	ref := func(s string) *reference { return newReference(parsePath(s, ".")) }
	cat := func(e ...varEvaler) *splice { return &splice{e} }
	nested := func(n ...varEvaler) varEvaler {
		return &expansionSingle{&splice{n}, "."}
	}
	exp := func(op string, l, r varEvaler) varEvaler {
		return makeOpExpansion(l, r, op, ".")
	}

	tests := []struct {
		title, exp string
		expected   varEvaler
	}{
		{"plain string", "string", str("string")},
		{"string containing :", "just:a:string", str("just:a:string")},
		{"string containing }", "abc } def", str("abc } def")},
		{"string containging regex with $", "log$|leg$", str("log$|leg$")},
		{"string with escaped var", "escaped $${var}", str("escaped ${var}")},
		{"reference", "${reference}", ref("reference")},
		{"exp in middle", "test ${splice} this",
			cat(str("test "), ref("splice"), str(" this"))},
		{"exp at beginning", "${splice} test",
			cat(ref("splice"), str(" test"))},
		{"exp at end", "test ${this}",
			cat(str("test "), ref("this"))},
		{"exp nested", "${${nested}}",
			nested(ref("nested"))},
		{"exp nested in middle", "${test.${this}.test}",
			nested(str("test."), ref("this"), str(".test"))},
		{"exp nested at beginning", "${${test}.this}",
			nested(ref("test"), str(".this"))},
		{"exp nested at end", "${test.${this}}",
			nested(str("test."), ref("this"))},
		{"exp with default", "${test:default}",
			exp(opDefault, str("test"), str("default"))},
		{"exp with defautl exp", "${test:the ${default} value}",
			exp(opDefault,
				str("test"),
				cat(str("the "), ref("default"), str(" value")))},
		{"exp with default containing }", "${test:abc$}def}",
			exp(opDefault, str("test"), str("abc}def"))},
		{"exp with default containing :", "${test:http://default:1234}",
			exp(opDefault, str("test"), str("http://default:1234"))},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s %s", test.title, test.exp), func(t *testing.T) {
			actual, err := parseSplice(test.exp, ".")
			if assert.NoError(t, err) {
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

func TestVarExpParseErrors(t *testing.T) {
	tests := []struct{ title, exp string }{
		{"empty expansion fail", "${}"},
		{"default expansion with left side", "${:abc}"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("test %v: %v", test.title, test.exp), func(t *testing.T) {
			res, err := parseSplice(test.exp, ".")
			assert.True(t, err != nil)
			assert.Error(t, err, fmt.Sprintf("result: %v, error: %v", res, err))
		})
	}
}
