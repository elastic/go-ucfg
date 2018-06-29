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

type node map[string]interface{}

// C rebrands Config
type C Config

func newC() *C {
	return fromConfig(New())
}

func newCFrom(from interface{}) *C {
	c, err := NewFrom(from)
	if err != nil {
		panic(err)
	}
	return fromConfig(c)
}

func fromConfig(in *Config) *C {
	return (*C)(in)
}

func (c *C) asConfig() *Config {
	return (*Config)(c)
}

func (c *C) SetBool(name string, idx int, value bool) {
	c.asConfig().SetBool(name, idx, value)
}

func (c *C) SetInt(name string, idx int, value int64) {
	c.asConfig().SetInt(name, idx, value)
}

func (c *C) SetUint(name string, idx int, value uint64) {
	c.asConfig().SetUint(name, idx, value)
}

func (c *C) SetFloat(name string, idx int, value float64) {
	c.asConfig().SetFloat(name, idx, value)
}

func (c *C) SetString(name string, idx int, value string) {
	c.asConfig().SetString(name, idx, value)
}

func (c *C) SetChild(name string, idx int, value *C) {
	c.asConfig().SetChild(name, idx, (*Config)(value))
}
