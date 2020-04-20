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
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"

	ucfg "github.com/elastic/go-ucfg"
)

var opts = []ucfg.Option{ucfg.PathSep(".")}

func mustNewFrom(t *testing.T, in map[string]interface{}, opts []ucfg.Option) *ucfg.Config {
	c, err := ucfg.NewFrom(in, opts...)
	if err != nil {
		t.Fatalf("failed to create new config for:%v reason:%v", spew.Sdump(in), err)
	}
	return c
}

func TestDiff(t *testing.T) {
	oneGraph := map[string]interface{}{
		"n.a.b.c": "hello",
		"n.a.d":   "world",
	}

	twoGraph := map[string]interface{}{
		"n.a.b.c": "hello",
		"o":       "new",
	}

	g1 := mustNewFrom(t, oneGraph, opts)
	g2 := mustNewFrom(t, twoGraph, opts)

	expected := Diff{
		Keep:   []string{"n.a.b.c"},
		Remove: []string{"n.a.d"},
		Add:    []string{"o"},
	}

	result := CompareConfigs(g1, g2, opts...)
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("expected:%v got:%v", spew.Sdump(expected), spew.Sdump(result))
	}
}

func TestConfigurationWithAddedCompareConfigs(t *testing.T) {
	oneGraph := map[string]interface{}{
		"n.a.b.c": "hello",
	}

	twoGraph := map[string]interface{}{
		"n.a.b.c": "hello",
		"o":       "new",
	}

	g1 := mustNewFrom(t, oneGraph, opts)
	g2 := mustNewFrom(t, twoGraph, opts)

	d := CompareConfigs(g1, g2, opts...)
	if !d.HasChanged() {
		t.Fatal("expected Diff.HasChanged() to be true")
	}
	if !d.HasKeyAdded() {
		t.Fatal("expected Diff.HasKeyAdded() to be true")
	}
}

func TestConfigurationWithRemovedKey(t *testing.T) {
	oneGraph := map[string]interface{}{
		"n.a.b.c": "hello",
		"o":       "new",
	}

	twoGraph := map[string]interface{}{
		"o": "new",
	}

	g1 := mustNewFrom(t, oneGraph, opts)
	g2 := mustNewFrom(t, twoGraph, opts)

	d := CompareConfigs(g1, g2, opts...)
	if !d.HasChanged() {
		t.Fatal("expected Diff.HasChanged() to be true")
	}
	if !d.HasKeyRemoved() {
		t.Fatal("expected Diff.HasKeyRemoved() to be true")
	}
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

	g1 := mustNewFrom(t, oneGraph, opts)
	g2 := mustNewFrom(t, twoGraph, opts)

	d := CompareConfigs(g1, g2, opts...)
	if !d.HasChanged() {
		t.Fatal("expected Diff.HasChanged() to be true")
	}
}

func TestConfigurationHasNotChanged(t *testing.T) {
	oneGraph := map[string]interface{}{
		"n.a.b.c": "hello",
		"n.a.d":   "world",
	}

	g1 := mustNewFrom(t, oneGraph, opts)

	d := CompareConfigs(g1, g1, opts...)
	if d.HasChanged() {
		t.Fatal("expected Diff.HasChanged() to be false")
	}
	if d.HasKeyRemoved() {
		t.Fatal("expected Diff.HasKeyRemoved() to be false")
	}
}
