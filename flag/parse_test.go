package flag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlagValueParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// null
		{"", nil},
		{"null", nil},

		// booleans
		{`true`, true},
		{`false`, false},
		{`on`, true},
		{`off`, false},

		// unsigned numbers
		{`23`, uint64(23)},

		// negative number
		{`-42`, int64(-42)},

		// floating point
		{`3.14`, float64(3.14)},

		// strings
		{`'single quoted'`, `single quoted`},
		{`'single quoted \"'`, `single quoted \"`},
		{`"double quoted"`, `double quoted`},
		{`"double quoted \""`, `double quoted "`},
		{`plain string`, `plain string`},

		// test arrays
		{`[]`, nil},
		{
			`a,b,c`,
			[]interface{}{"a", "b", "c"},
		},
		{
			`[array, 1, true, "abc"]`,
			[]interface{}{"array", uint64(1), true, "abc"},
		},
		{
			`[test, [1,2,3], on]`,
			[]interface{}{
				"test",
				[]interface{}{uint64(1), uint64(2), uint64(3)},
				true,
			},
		},

		// test dictionaries:
		{`{}`, nil},
		{`{'key1': true,
       "key2": 1,
       key 3: ['test', "test2", off],
       nested key: {"a" : 2}}`,
			map[string]interface{}{
				"key1":  true,
				"key2":  uint64(1),
				"key 3": []interface{}{"test", "test2", false},
				"nested key": map[string]interface{}{
					"a": uint64(2),
				},
			},
		},

		// array of top-level dictionaries
		{
			`{key: 1},{key: 2}`,
			[]interface{}{
				map[string]interface{}{
					"key": uint64(1),
				},
				map[string]interface{}{
					"key": uint64(2),
				},
			},
		},
	}

	for i, test := range tests {
		t.Logf("run test (%v): %v", i, test.input)

		v, err := parseValue(test.input)
		if err != nil {
			t.Error(err)
			continue
		}

		assert.Equal(t, test.expected, v)
	}

}

func TestFlagValueParsingFails(t *testing.T) {
	tests := []string{
		// strings:
		`'abc`,
		`"abc`,

		// arrays
		`[1,2,3`,         // missing ']'
		`['abc' 'def']`,  // missing comma
		`['abc', 'def,]`, // nested

		// objects
		`{a: 1, b:2`,      // missing '}'
		`{'a' 1, b: 2}`,   // missing ':'
		`{'a': '1' b: 2}`, // missing ','
		`{'abc: 2}`,       // key parsing error
		`{key: 'fail}`,    // value parsing error
		`{:'abc'}`,        // object with missing key
		`{nested: {a: 2}`, // nested object with missing '}'
	}
	for i, test := range tests {
		t.Logf("run test(%v): %v", i, test)

		_, err := parseValue(test)
		if err == nil {
			t.Errorf("parsing '%v' did not fail", test)
			continue
		}

		t.Log("  Failed with: ", err.Error())
	}
}
