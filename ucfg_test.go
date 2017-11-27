package ucfg

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

var oneGraph = map[string]interface{}{
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

func TestFlattenKeys(t *testing.T) {
	opts := []Option{PathSep(".")}

	c, err := NewFrom(oneGraph, opts...)
	assert.NoError(t, err)

	expected := []string{
		"n.a.b.c",
		"n.a.d",
		"values.0.j",
		"values.0.k",
		"values.1.j",
		"values.1.o",
	}
	sort.Strings(expected)

	results := c.FlattenedKeys(opts...)
	sort.Strings(results)

	assert.Equal(t, expected, results)
}

func TestFlattenKeysWithEmptyPathSep(t *testing.T) {
	opts := []Option{}

	c, err := NewFrom(oneGraph, opts...)
	assert.NoError(t, err)

	expected := []string{
		"n.a.b.c",
		"n.a.d",
		"values.0.j",
		"values.0.k",
		"values.1.j",
		"values.1.o",
	}
	sort.Strings(expected)

	results := c.FlattenedKeys(opts...)
	sort.Strings(results)

	assert.Equal(t, expected, results)
}
