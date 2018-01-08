package ucfg

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlattenKeys(t *testing.T) {
	tests := []struct {
		name    string
		pathSep string
	}{
		{"withDot", "."},
		{"emptySep", ""},
	}

	sorted := func(s []string) []string {
		sort.Strings(s)
		return s
	}

	cfg := map[string]interface{}{
		"n.a.b.c": "hello",
		"n.a.d":   "world",
		"values": []interface{}{
			map[string]interface{}{
				"j": "j-value",
				"k": "k-value",
			},
			map[string]interface{}{
				"j": "r-value",
				"o": "o-value",
			},
		},
	}

	expected := sorted([]string{
		"n.a.b.c",
		"n.a.d",
		"values.0.j",
		"values.0.k",
		"values.1.j",
		"values.1.o",
	})

	for _, test := range tests {
		sep := test.pathSep
		t.Run(test.name, func(t *testing.T) {
			opts := []Option{}
			if sep != "" {
				opts = append(opts, PathSep(sep))
			}

			c, err := NewFrom(cfg, opts...)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, expected, sorted(c.FlattenedKeys(opts...)))
		})
	}
}
