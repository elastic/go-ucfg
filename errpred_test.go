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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCriticalResolveError(t *testing.T) {
	tests := []struct {
		title    string
		err      Error
		expected bool
	}{
		{
			title:    "error is ErrMissing",
			err:      raiseMissing(New(), "reference"),
			expected: false,
		},
		{
			title:    "error is ErrCyclicReference",
			err:      raiseCyclicErr("reference"),
			expected: false,
		},
		{
			title:    "any other error is critical",
			err:      raiseCritical(errors.New("something bad"), ""),
			expected: true,
		},
		{
			title:    "when the error is nil",
			err:      nil,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			assert.Equal(t, test.expected, criticalResolveError(test.err))
		})
	}
}
