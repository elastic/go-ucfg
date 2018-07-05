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
	goflag "flag"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/go-ucfg"
)

func TestFlagValuePrimitives(t *testing.T) {
	config, err := parseTestFlags("-D b=true -D b2 -D i=42 -D f=3.14 -D s=string")
	assert.NoError(t, err)

	//validate
	checkFields(t, config)
}

func TestFlagValueLast(t *testing.T) {
	config, err := parseTestFlags("-D b2=false -D i=23 -D s=test -D b=true -D b2 -D i=42 -D f=3.14 -D s=string")
	assert.NoError(t, err)

	//validate
	checkFields(t, config)

}

func TestFlagValueMissing(t *testing.T) {
	config, err := parseTestFlags("-D b=true -D -D s=")
	assert.NoError(t, err)

	b, err := config.Bool("b", -1, ucfg.PathSep("."))
	assert.NoError(t, err)
	assert.Equal(t, true, b)

	s, err := config.String("s", -1, ucfg.PathSep("."))
	assert.Error(t, err)
	assert.Equal(t, "missing field accessing 's'", err.Error())
	assert.Equal(t, "", s)
}

func TestFlagValueNested(t *testing.T) {
	config, err := parseTestFlags("-D c.b=true -D c.b2 -D c.i=42 -D c.f=3.14 -D c.s=string")
	assert.NoError(t, err)

	// validate
	sub, err := config.Child("c", -1)
	assert.NoError(t, err)
	assert.NotNil(t, sub)

	if sub != nil {
		checkFields(t, sub)
	}
}

func TestFlagValueList(t *testing.T) {
	config, err := parseTestFlags("-D c.0.b=true -D c.0.b2 -D c.0.i=42 -D c.0.f=3.14 -D c.0.s=string")
	assert.NoError(t, err)

	// validate
	sub, err := config.Child("c", 0)
	assert.NoError(t, err)
	assert.NotNil(t, sub)

	if sub != nil {
		checkFields(t, sub)
	}
}

func TestMergeFlagValueNewList(t *testing.T) {
	config, _ := ucfg.NewFrom(map[string]interface{}{
		"c.0.b":  true,
		"c.0.b2": true,
		"c.0.i":  42,
		"c.0.c":  3.14,
		"c.0.s":  "wrong",
	}, ucfg.PathSep("."))

	cliConfig, err := parseTestFlags("-D c.0.s=string -D c.0.f=3.14 -D c.1.b=true")
	assert.NoError(t, err)

	err = config.Merge(cliConfig, ucfg.PathSep("."))
	assert.NoError(t, err)

	// validate
	sub, err := config.Child("c", 0)
	assert.NoError(t, err)
	assert.NotNil(t, sub)
	if sub != nil {
		checkFields(t, sub)
	}

	sub, err = config.Child("c", 1)
	assert.NoError(t, err)
	assert.NotNil(t, sub)
	if sub == nil {
		return
	}

	b, err := sub.Bool("b", -1)
	assert.NoError(t, err)
	assert.True(t, b)
}

func parseTestFlags(args string) (*ucfg.Config, error) {
	fs := goflag.NewFlagSet("test", goflag.ContinueOnError)
	config := Config(fs, "D", "overwrite", ucfg.PathSep("."))
	err := fs.Parse(strings.Split(args, " "))
	return config, err
}

func checkFields(t *testing.T, config *ucfg.Config) {
	b, err := config.Bool("b", -1, ucfg.PathSep("."))
	assert.NoError(t, err)
	assert.Equal(t, true, b)

	b, err = config.Bool("b2", -1, ucfg.PathSep("."))
	assert.NoError(t, err)
	assert.Equal(t, true, b)

	i, err := config.Int("i", -1, ucfg.PathSep("."))
	assert.NoError(t, err)
	assert.Equal(t, 42, int(i))

	f, err := config.Float("f", -1, ucfg.PathSep("."))
	assert.NoError(t, err)
	assert.Equal(t, 3.14, f)

	s, err := config.String("s", -1, ucfg.PathSep("."))
	assert.NoError(t, err)
	assert.Equal(t, "string", s)
}
