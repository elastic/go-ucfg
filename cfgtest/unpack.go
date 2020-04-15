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

package cfgtest

import (
	"github.com/davecgh/go-spew/spew"
)

type testingT interface {
	Fatalf(format string, args ...interface{})
}

type config interface {
	UnpackWithoutOptions(to interface{}) error
}

// MustFailUnpack method fails the testing if unpacking passed.
func MustFailUnpack(t testingT, cfg config, test interface{}) {
	if err := cfg.UnpackWithoutOptions(test); err == nil {
		t.Fatalf("expected failure, config:%s test:%s", spew.Sdump(cfg), spew.Sdump(test))
	}
}

// MustUnpack method fails the testing if unpacking failed too.
func MustUnpack(t testingT, cfg config, test interface{}) {
	if err := cfg.UnpackWithoutOptions(test); err != nil {
		t.Fatalf("config:%s test:%s error:%v", spew.Sdump(cfg), spew.Sdump(test), err)
	}
}
