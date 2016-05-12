package ucfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVarExpParserSuccess(t *testing.T) {
	str := func(s string) stringPiece { return stringPiece(s) }
	ref := func(s string) *reference { return newReference(parsePath(s, ".")) }
	exp := func(l, r []splicePiece) *expansion {
		var sr splicePiece
		if r != nil {
			sr = &splice{r}
		}
		return &expansion{&splice{l}, sr, "."}
	}

	tests := []struct {
		title, exp string
		expected   []splicePiece
	}{
		{"plain string", "string", []splicePiece{str("string")}},
		{"reference", "${reference}", []splicePiece{ref("reference")}},
		{"exp in middle", "test ${splice} this",
			[]splicePiece{str("test "), ref("splice"), str(" this")}},
		{"exp at beginning", "${splice} test",
			[]splicePiece{ref("splice"), str(" test")}},
		{"exp at end", "test ${this}",
			[]splicePiece{str("test "), ref("this")}},
		{"exp nested", "${${nested}}",
			[]splicePiece{exp([]splicePiece{ref("nested")}, nil)}},
		{"exp nested in middle", "${test.${this}.test}",
			[]splicePiece{exp([]splicePiece{str("test."), ref("this"), str(".test")}, nil)}},
		{"exp nested at beginning", "${${test}.this}",
			[]splicePiece{exp([]splicePiece{ref("test"), str(".this")}, nil)}},
		{"exp nested at beginning", "${test.${this}}",
			[]splicePiece{exp([]splicePiece{str("test."), ref("this")}, nil)}},
		{"exp with default", "${test:default}",
			[]splicePiece{exp(
				[]splicePiece{str("test")},
				[]splicePiece{str("default")})}},
		{"exp with defautl exp", "${test:the ${default} value}",
			[]splicePiece{exp(
				[]splicePiece{str("test")},
				[]splicePiece{str("the "), ref("default"), str(" value")})},
		},
	}

	for _, test := range tests {
		t.Logf("test %v: %v", test.title, test.exp)
		actual, err := parseSplice(test.exp, ".")
		if err != nil {
			t.Errorf("failed to parse with %v", err)
			continue
		}

		t.Logf("expected: %v", test.expected)
		t.Logf("actual: %v", actual)
		if assert.Equal(t, test.expected, actual) {
			t.Logf("success")
		}
	}
}

func TestVarExpParseErrors(t *testing.T) {
	tests := []struct{ title, exp string }{
		{"empty expansion fail", "${}"},
		{"default expansion with left side", "${:abc}"},
	}

	for _, test := range tests {
		t.Logf("test %v: %v", test.title, test.exp)
		_, err := parseSplice(test.exp, ".")
		assert.True(t, err != nil)
	}
}
