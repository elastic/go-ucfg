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
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type myNonzeroInt int
type myCustomList []int
type myCustomMap map[string]int
type myNonzeroList []myNonzeroInt
type myNonzeroMap map[string]myNonzeroInt

type mapValidator myNonzeroMap

type structValidator struct{ I int }
type ptrStructValidator struct{ I int }
type structMapValidator struct {
	I int
	M mapValidator
}
type structNestedValidator struct {
	I int
	N structMapValidator
}

type structWithValidationTags struct {
	I int `validate:"positive"`
}

type nestedPtrStructValidator struct {
	A *ptrStructValidator `validate:"required"`
	B *ptrStructValidator `validate:"required"`
}

type nestedNestedPtrStructValidator struct {
	A *nestedPtrStructValidator
	B *nestedPtrStructValidator
}

var errZeroTest = errors.New("value must not be 0")
var errEmptyTest = errors.New("value must not be empty")
var errMoreTest = errors.New("value must have more than 1 element")

func (m myNonzeroInt) Validate() error {
	return testZeroErr(int(m))
}

func (l myCustomList) Validate() error {
	if len(l) == 0 {
		return errEmptyTest
	}
	return nil
}

func (p myCustomMap) Validate() error {
	if len(p) == 0 {
		return errEmptyTest
	}
	return nil
}

func (p mapValidator) Validate() error {
	if len(p) <= 1 {
		return errMoreTest
	}
	return nil
}

func (s structValidator) Validate() error {
	return testZeroErr(s.I)
}

func (p *ptrStructValidator) Validate() error {
	return testZeroErr(p.I)
}

func (p structMapValidator) Validate() error {
	return testZeroErr(p.I)
}

func (p structNestedValidator) Validate() error {
	return testZeroErr(p.I)
}

func testZeroErr(i int) error {
	if i == 0 {
		return errZeroTest
	}
	return nil
}

func TestValidationPass(t *testing.T) {
	c, _ := NewFrom(map[string]interface{}{
		"a": 0,
		"b": 10,
		"i": 5,
		"d": -10,
		"f": 3.14,
		"l": []int{0, 1},
		"m": myCustomMap{
			"key": 1,
		},
		"n": myNonzeroList{myNonzeroInt(1)},
		"o": myNonzeroMap{
			"key": myNonzeroInt(1),
		},
		"p": mapValidator{
			"one": 1,
			"two": 2,
		},
		"q": structMapValidator{
			I: 1,
			M: mapValidator{
				"one": 1,
				"two": 2,
			},
		},
		"r": structNestedValidator{
			I: 1,
			N: structMapValidator{
				I: 1,
				M: mapValidator{
					"one": 1,
					"two": 2,
				},
			},
		},
	})

	tests := []interface{}{
		// validate field 'a'
		&struct {
			A int `validate:"positive"`
		}{},
		&struct {
			A int `validate:"positive,min=0"`
		}{},
		&struct {
			X int `config:"a" validate:"min=0"`
		}{},
		&struct {
			A time.Duration `validate:"positive"`
		}{},
		&struct {
			A time.Duration `validate:"positive,min=0"`
		}{},
		&struct {
			X time.Duration `config:"a" validate:"min=0"`
		}{},

		// validate field 'b'
		&struct {
			B int `validate:"nonzero"`
		}{},
		&struct {
			B myNonzeroInt
		}{},
		&struct {
			B int `validate:"positive"`
		}{},
		&struct {
			Tmp structValidator `config:",inline"`
		}{},
		&struct {
			Tmp ptrStructValidator `config:",inline"`
		}{},
		&struct {
			X int `config:"b" validate:"nonzero,min=-1"`
		}{},
		&struct {
			X int `config:"b" validate:"min=10, max=20"`
		}{},
		&struct {
			B time.Duration `validate:"nonzero"`
		}{},
		&struct {
			B time.Duration `validate:"positive"`
		}{},
		&struct {
			X time.Duration `config:"b" validate:"min=10, max=20"`
		}{},
		&struct {
			X time.Duration `config:"b" validate:"min=10s, max=20s"`
		}{},

		// validate field 'd'
		&struct {
			D int `validate:"nonzero"`
		}{},
		&struct {
			X int `config:"d" validate:"nonzero,min=-10"`
		}{},
		&struct {
			X int `config:"d" validate:"min=-10, max=0"`
		}{},
		&struct {
			D time.Duration `validate:"nonzero"`
		}{},
		&struct {
			X time.Duration `config:"d" validate:"nonzero,min=-10"`
		}{},
		&struct {
			X time.Duration `config:"d" validate:"min=-10, max=0"`
		}{},

		// validate field 'f'
		&struct {
			F float64 `validate:"nonzero"`
		}{},
		&struct {
			F float64 `validate:"positive"`
		}{},
		&struct {
			X int `config:"f" validate:"nonzero,min=-1"`
		}{},
		&struct {
			X int `config:"f" validate:"min=3, max=20"`
		}{},
		&struct {
			F time.Duration `validate:"nonzero"`
		}{},
		&struct {
			F time.Duration `validate:"positive"`
		}{},
		&struct {
			X time.Duration `config:"f" validate:"nonzero,min=-1"`
		}{},
		&struct {
			X time.Duration `config:"f" validate:"min=3, max=20"`
		}{},

		// validate field 'l'
		&struct {
			L myCustomList
		}{},

		// validation field 'm'
		&struct {
			M myCustomMap
		}{},

		// validation field 'n'
		&struct {
			N myNonzeroList
		}{},

		// validation field 'o'
		&struct {
			O myNonzeroMap
		}{},

		// validation field 'p'
		&struct {
			P mapValidator
		}{},

		// validation field 'q'
		&struct {
			Q structMapValidator
		}{},

		// validation field 'r'
		&struct {
			R structNestedValidator
		}{},

		// other
		&struct {
			X int // field not present in config, but not required
		}{},
		&struct {
			X *ptrStructValidator // Validator not called as its nil value
		}{},
		&struct {
			X *nestedNestedPtrStructValidator
		}{
			X: &nestedNestedPtrStructValidator{},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test config (%v): %#v", i, test), func(t *testing.T) {
			err := c.Unpack(test)
			assert.NoError(t, err)
		})
	}
}

func TestValidationFail(t *testing.T) {
	c, _ := NewFrom(map[string]interface{}{
		"a": 0,
		"b": 10,
		"i": 0,
		"d": -10,
		"f": 3.14,
		"l": []int{},
		"m": myCustomMap{},
		"n": myNonzeroList{myNonzeroInt(0)},
		"o": myNonzeroMap{
			"key": myNonzeroInt(0),
		},
		"p0": mapValidator{},
		"p1": mapValidator{
			"one": 0,
		},
		"p2": mapValidator{
			"one": 1,
		},
		"q0": structMapValidator{
			I: 0,
			M: mapValidator{
				"one": 1,
				"two": 2,
			},
		},
		"q1": structMapValidator{
			I: 1,
			M: mapValidator{},
		},
		"q2": structMapValidator{
			I: 1,
			M: mapValidator{
				"one": 1,
			},
		},
		"r0": structNestedValidator{
			I: 0,
			N: structMapValidator{
				I: 1,
				M: mapValidator{
					"one": 1,
					"two": 2,
				},
			},
		},
		"r1": structNestedValidator{
			I: 1,
			N: structMapValidator{
				I: 0,
				M: mapValidator{
					"one": 1,
					"two": 2,
				},
			},
		},
		"r2": structNestedValidator{
			I: 1,
			N: structMapValidator{
				I: 1,
				M: mapValidator{},
			},
		},
		"r3": structNestedValidator{
			I: 1,
			N: structMapValidator{
				I: 1,
				M: mapValidator{
					"one": 1,
				},
			},
		},
	})

	tests := []interface{}{
		// test field 'a'
		&struct {
			X int `config:"a" validate:"nonzero"`
		}{},
		&struct {
			X myNonzeroInt `config:"a"`
		}{},
		&struct {
			Tmp structValidator `config:",inline"`
		}{},
		&struct {
			Tmp ptrStructValidator `config:",inline"`
		}{},
		&struct {
			X int `config:"a" validate:"min=10"`
		}{},
		&struct {
			X time.Duration `config:"a" validate:"nonzero"`
		}{},
		&struct {
			X time.Duration `config:"a" validate:"min=10"`
		}{},
		&struct {
			X time.Duration `config:"a" validate:"min=10ns"`
		}{},

		// test field 'b'
		&struct {
			X int `config:"b" validate:"max=8"`
		}{},
		&struct {
			X int `config:"b" validate:"min=20"`
		}{},
		&struct {
			X time.Duration `config:"b" validate:"max=8ms"`
		}{},
		&struct {
			X time.Duration `config:"b" validate:"min=20h"`
		}{},

		// test field 'd'
		&struct {
			X int `config:"d" validate:"positive"`
		}{},
		&struct {
			X int `config:"d" validate:"max=-11"`
		}{},
		&struct {
			X int `config:"d" validate:"min=20"`
		}{},
		&struct {
			X time.Duration `config:"d" validate:"positive"`
		}{},
		&struct {
			X time.Duration `config:"d" validate:"max=-11s"`
		}{},
		&struct {
			X time.Duration `config:"d" validate:"min=20h"`
		}{},

		// test field 'f'
		&struct {
			X float64 `config:"f" validate:"max=1"`
		}{},
		&struct {
			X float64 `config:"f" validate:"min=20"`
		}{},
		&struct {
			X time.Duration `config:"f" validate:"max=1s"`
		}{},
		&struct {
			X time.Duration `config:"f" validate:"min=20s"`
		}{},

		// test field 'l'
		&struct {
			X myCustomList `config:"l"`
		}{},

		// validation field 'm'
		&struct {
			M myCustomMap
		}{},

		// validation field 'n'
		&struct {
			N myNonzeroList
		}{},

		// validation field 'o'
		&struct {
			O myNonzeroMap
		}{},

		// validation 'p' fields
		&struct {
			P mapValidator `config:"p0"`
		}{},
		&struct {
			P mapValidator `config:"p1"`
		}{},
		&struct {
			P mapValidator `config:"p2"`
		}{},

		// validation 'q' fields
		&struct {
			Q structMapValidator `config:"q0"`
		}{},
		&struct {
			Q structMapValidator `config:"q1"`
		}{},
		&struct {
			Q structMapValidator `config:"q2"`
		}{},

		// validation 'r' fields
		&struct {
			R structNestedValidator `config:"r0"`
		}{},
		&struct {
			R structNestedValidator `config:"r1"`
		}{},
		&struct {
			R structNestedValidator `config:"r2"`
		}{},
		&struct {
			R structNestedValidator `config:"r3"`
		}{},

		// other
		&struct {
			X int `validate:"required"`
		}{},
		&struct {
			X *nestedPtrStructValidator
		}{
			X: &nestedPtrStructValidator{},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test config (%v): %#v", i, test), func(t *testing.T) {
			err := c.Unpack(test)
			assert.True(t, err != nil)
		})
	}
}

func TestValidateRequiredFailing(t *testing.T) {
	c, _ := NewFrom(node{
		"b": "",
		"c": nil,
		"d": []string{},
	})

	tests := []struct {
		err    error
		config interface{}
	}{
		// Access missing field 'a'
		{ErrRequired, &struct {
			A *int `validate:"required"`
		}{}},
		{ErrRequired, &struct {
			A int `validate:"required"`
		}{}},
		{ErrRequired, &struct {
			A string `validate:"required"`
		}{}},
		{ErrRequired, &struct {
			A []string `validate:"required"`
		}{}},
		{ErrRequired, &struct {
			A time.Duration `validate:"required"`
		}{}},

		// Access empty string field "b"
		{ErrRequired, &struct {
			B string `validate:"required"`
		}{}},
		{ErrRequired, &struct {
			B *string `validate:"required"`
		}{}},
		{ErrRequired, &struct {
			B *regexp.Regexp `validate:"required"`
		}{}},

		// Access nil value "c"
		{ErrRequired, &struct {
			C *int `validate:"required"`
		}{}},
		{ErrRequired, &struct {
			C int `validate:"required"`
		}{}},
		{ErrRequired, &struct {
			C string `validate:"required"`
		}{}},
		{ErrRequired, &struct {
			C []string `validate:"required"`
		}{}},
		{ErrRequired, &struct {
			C time.Duration `validate:"required"`
		}{}},

		// Check empty []string field 'd'
		{ErrRequired, &struct {
			D []string `validate:"required"`
		}{}},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test config (%v): %#v => %v", i, test.config, test.err), func(t *testing.T) {
			err := c.Unpack(test.config)
			if err == nil {
				t.Error("Expected error")
				return
			}

			t.Logf("Unpack returned error: %v", err)
			err = err.(Error).Reason()
			assert.Equal(t, test.err, err)
		})
	}
}

func TestValidateNonzeroFailing(t *testing.T) {
	c, _ := NewFrom(node{
		"i": 0,
		"s": "",
		"a": []int{},
	})

	tests := []struct {
		err    error
		config interface{}
	}{
		// test integer types accessing 'i'
		{ErrZeroValue, &struct {
			I int `validate:"nonzero"`
		}{}},
		{ErrZeroValue, &struct {
			I int8 `validate:"nonzero"`
		}{}},
		{ErrZeroValue, &struct {
			I int16 `validate:"nonzero"`
		}{}},
		{ErrZeroValue, &struct {
			I int32 `validate:"nonzero"`
		}{}},
		{ErrZeroValue, &struct {
			I int64 `validate:"nonzero"`
		}{}},
		{ErrZeroValue, &struct {
			I uint `validate:"nonzero"`
		}{}},
		{ErrZeroValue, &struct {
			I uint8 `validate:"nonzero"`
		}{}},
		{ErrZeroValue, &struct {
			I uint16 `validate:"nonzero"`
		}{}},
		{ErrZeroValue, &struct {
			I uint32 `validate:"nonzero"`
		}{}},
		{ErrZeroValue, &struct {
			I uint64 `validate:"nonzero"`
		}{}},

		// test float types accessing 'i'
		{ErrZeroValue, &struct {
			I float32 `validate:"nonzero"`
		}{}},
		{ErrZeroValue, &struct {
			I float64 `validate:"nonzero"`
		}{}},

		// test string types accessing 's'
		{ErrEmpty, &struct {
			S string `validate:"nonzero"`
		}{}},
		{ErrEmpty, &struct {
			S *string `validate:"nonzero"`
		}{}},
		{ErrEmpty, &struct {
			S *regexp.Regexp `validate:"nonzero"`
		}{}},

		// test array type accessing 'a'
		{ErrEmpty, &struct {
			A []int `validate:"nonzero"`
		}{}},
		{ErrEmpty, &struct {
			A []uint8 `validate:"nonzero"`
		}{}},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test config (%v): %#v => %v", i, test.config, test.err), func(t *testing.T) {
			err := c.Unpack(test.config)
			if err == nil {
				t.Error("Expected error")
				return
			}

			t.Logf("Unpack returned error: %v", err)
			err = err.(Error).Reason()
			assert.Equal(t, test.err, err)
		})
	}
}

func TestValidationFailOnDefaults(t *testing.T) {
	c := New()

	tests := []interface{}{
		// test field 'a'
		&struct {
			X int `config:"a" validate:"nonzero"`
		}{
			X: 0,
		},
		&struct {
			X myNonzeroInt `config:"a"`
		}{
			X: 0,
		},
		&struct {
			Tmp structValidator `config:",inline"`
		}{
			Tmp: structValidator{
				I: 0,
			},
		},
		&struct {
			Tmp ptrStructValidator `config:",inline"`
		}{
			Tmp: ptrStructValidator{
				I: 0,
			},
		},
		&struct {
			X int `config:"a" validate:"min=10"`
		}{
			X: 9,
		},
		&struct {
			X time.Duration `config:"a" validate:"nonzero"`
		}{
			X: time.Duration(0),
		},
		&struct {
			X time.Duration `config:"a" validate:"min=10"`
		}{
			X: time.Duration(9),
		},
		&struct {
			X time.Duration `config:"a" validate:"min=10ns"`
		}{
			X: time.Duration(9 * time.Nanosecond),
		},

		// test field 'b'
		&struct {
			X int `config:"b" validate:"max=8"`
		}{
			X: 9,
		},
		&struct {
			X int `config:"b" validate:"min=20"`
		}{
			X: 19,
		},
		&struct {
			X time.Duration `config:"b" validate:"max=8ms"`
		}{
			X: time.Duration(9 * time.Millisecond),
		},
		&struct {
			X time.Duration `config:"b" validate:"min=20h"`
		}{
			X: time.Duration(19 * time.Hour),
		},

		// test field 'd'
		&struct {
			X int `config:"d" validate:"positive"`
		}{
			X: -1,
		},
		&struct {
			X int `config:"d" validate:"max=-11"`
		}{
			X: -10,
		},
		&struct {
			X int `config:"d" validate:"min=20"`
		}{
			X: 19,
		},
		&struct {
			X time.Duration `config:"d" validate:"positive"`
		}{
			X: time.Duration(-1),
		},
		&struct {
			X time.Duration `config:"d" validate:"max=-11s"`
		}{
			X: time.Duration(-10 * time.Second),
		},
		&struct {
			X time.Duration `config:"d" validate:"min=20h"`
		}{
			X: time.Duration(19 * time.Hour),
		},

		// test field 'f'
		&struct {
			X float64 `config:"f" validate:"max=1"`
		}{
			X: 2,
		},
		&struct {
			X float64 `config:"f" validate:"min=20"`
		}{
			X: 19,
		},
		&struct {
			X time.Duration `config:"f" validate:"max=1s"`
		}{
			X: time.Duration(2 * time.Second),
		},
		&struct {
			X time.Duration `config:"f" validate:"min=20s"`
		}{
			X: time.Duration(19 * time.Second),
		},

		// test field 'l'
		&struct {
			X myCustomList `config:"l"`
		}{
			X: myCustomList{},
		},

		// validation field 'm'
		&struct {
			M myCustomMap
		}{
			M: myCustomMap{},
		},

		// validation field 'n'
		&struct {
			N myNonzeroList
		}{
			N: myNonzeroList{0},
		},

		// validation field 'o'
		&struct {
			O myNonzeroMap
		}{
			O: myNonzeroMap{
				"zero": 0,
			},
		},

		// validation 'p' field
		&struct {
			P mapValidator `config:"p"`
		}{
			P: mapValidator{
				"zero": 0,
			},
		},

		// validation 'q' field
		&struct {
			Q structMapValidator
		}{
			Q: structMapValidator{
				I: 0,
			},
		},

		// validation 'r' field
		&struct {
			R structNestedValidator
		}{
			R: structNestedValidator{
				I: 0,
			},
		},
		&struct {
			R *structNestedValidator
		}{
			R: &structNestedValidator{
				I: 1,
				N: structMapValidator{
					I: 1,
					M: mapValidator{
						"one": 1,
					},
				},
			},
		},

		// validation 's' field
		&struct {
			S *structWithValidationTags
		}{
			S: &structWithValidationTags{
				I: -1,
			},
		},

		// validate array
		&myNonzeroList{0},

		// validate map
		&myNonzeroMap{
			"zero": 0,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test config (%v): %#v", i, test), func(t *testing.T) {
			err := c.Unpack(test)
			assert.True(t, err != nil)
		})
	}
}
