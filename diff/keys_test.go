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

package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"

	ucfg "github.com/elastic/go-ucfg"
)

var opts = []ucfg.Option{ucfg.PathSep(".")}

func TestDiff(t *testing.T) {
	oneGraph := map[string]interface{}{
		"n.a.b.c": "hello",
		"n.a.d":   "world",
	}

	twoGraph := map[string]interface{}{
		"n.a.b.c": "hello",
		"o":       "new",
	}

	g1, err := ucfg.NewFrom(oneGraph, opts...)
	assert.NoError(t, err)

	g2, err := ucfg.NewFrom(twoGraph, opts...)
	assert.NoError(t, err)

	expected := Diff{
		Keep:   []string{"n.a.b.c"},
		Remove: []string{"n.a.d"},
		Add:    []string{"o"},
	}

	result := CompareConfigs(g1, g2, opts...)
	assert.Equal(t, expected, result)
}

func TestConfigurationWithAddedCompareConfigs(t *testing.T) {
	oneGraph := map[string]interface{}{
		"n.a.b.c": "hello",
	}

	twoGraph := map[string]interface{}{
		"n.a.b.c": "hello",
		"o":       "new",
	}

	g1, err := ucfg.NewFrom(oneGraph, opts...)
	assert.NoError(t, err)

	g2, err := ucfg.NewFrom(twoGraph, opts...)
	assert.NoError(t, err)

	d := CompareConfigs(g1, g2, opts...)
	assert.True(t, d.HasChanged())
	assert.True(t, d.HasKeyAdded())
}

func TestConfigurationWithRemovedKey(t *testing.T) {
	oneGraph := map[string]interface{}{
		"n.a.b.c": "hello",
		"o":       "new",
	}

	twoGraph := map[string]interface{}{
		"o": "new",
	}

	g1, err := ucfg.NewFrom(oneGraph, opts...)
	assert.NoError(t, err)

	g2, err := ucfg.NewFrom(twoGraph, opts...)
	assert.NoError(t, err)

	d := CompareConfigs(g1, g2, opts...)
	assert.True(t, d.HasChanged())
	assert.True(t, d.HasKeyRemoved())
}

func TestConfigurationWithAddedAndRemovedKey(t *testing.T) {
	oneGraph := map[string]interface{}{
		"n.a.b.c": "hello",
		"o":       "new",
	}

	twoGraph := map[string]interface{}{
		"o": "new",
		"l": "new-new",
	}

	g1, err := ucfg.NewFrom(oneGraph, opts...)
	assert.NoError(t, err)

	g2, err := ucfg.NewFrom(twoGraph, opts...)
	assert.NoError(t, err)

	d := CompareConfigs(g1, g2, opts...)
	assert.True(t, d.HasChanged())
}

func TestConfigurationHasNotChanged(t *testing.T) {
	oneGraph := map[string]interface{}{
		"n.a.b.c": "hello",
		"n.a.d":   "world",
	}

	g1, err := ucfg.NewFrom(oneGraph, opts...)
	assert.NoError(t, err)

	d := CompareConfigs(g1, g1, opts...)
	assert.False(t, d.HasChanged())
	assert.False(t, d.HasKeyRemoved())
	assert.False(t, d.HasKeyRemoved())
}
