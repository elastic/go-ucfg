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
