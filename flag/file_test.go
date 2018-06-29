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

package flag

import (
	"errors"
	goflag "flag"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/go-ucfg"
)

func TestFlagFileParsePrimitives(t *testing.T) {
	config, err := parseTestFileFlags("-c test.a -c test.b",
		map[string]FileLoader{
			".a": makeLoadTestConfig(map[string]interface{}{
				"b": true,
				"i": 42,
			}),
			".b": makeLoadTestConfig(map[string]interface{}{
				"b2": true,
				"f":  3.14,
				"s":  "string",
			}),
		},
	)

	assert.NoError(t, err)
	checkFields(t, config)
}

func TestFlagFileParseOverwrites(t *testing.T) {
	config, err := parseTestFileFlags("-c test.a -c test.b",
		map[string]FileLoader{
			".a": makeLoadTestConfig(map[string]interface{}{
				"b":  true,
				"b2": false,
				"i":  23,
				"s":  "test",
			}),
			".b": makeLoadTestConfig(map[string]interface{}{
				"b2": true,
				"f":  3.14,
				"i":  42,
				"s":  "string",
			}),
		},
	)

	assert.NoError(t, err)
	checkFields(t, config)
}

func TestFlagFileParseFail(t *testing.T) {
	var expectedErr = errors.New("test fail")
	_, err := parseTestFileFlags("-c test.a -c test.b",
		map[string]FileLoader{
			".a": makeLoadTestConfig(map[string]interface{}{
				"b":  true,
				"b2": false,
				"i":  23,
				"s":  "test",
			}),
			".b": makeLoadTestFail(expectedErr),
		},
	)
	assert.EqualError(t, err, expectedErr.Error())
}

func makeLoadTestConfig(
	c map[string]interface{},
) func(string, ...ucfg.Option) (*ucfg.Config, error) {
	return func(path string, opts ...ucfg.Option) (*ucfg.Config, error) {
		return ucfg.NewFrom(c, opts...)
	}
}

func makeLoadTestFail(
	err error,
) func(string, ...ucfg.Option) (*ucfg.Config, error) {
	return func(path string, opts ...ucfg.Option) (*ucfg.Config, error) {
		return nil, err
	}
}

func parseTestFileFlags(args string, exts map[string]FileLoader) (*ucfg.Config, error) {
	fs := goflag.NewFlagSet("test", goflag.ContinueOnError)
	v := ConfigFiles(fs, "c", "config file", exts, ucfg.PathSep("."))
	err := fs.Parse(strings.Split(args, " "))
	if err != nil {
		return nil, err
	}
	return v.Config(), v.Error()
}
