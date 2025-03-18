package feel

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_compareInterfaces(t *testing.T) {
	scope := map[string]any{
		"one": 1,
		"nested": map[string]any{
			"tenK":        10000,
			"fiveHundred": 500.0,
		},
	}

	tests := []struct {
		expr     string
		expected any
	}{
		{expr: `one <= 1000`, expected: true},
		{expr: `1000 >= one`, expected: true},
		{expr: `nested.tenK <= 20000`, expected: true},
		{expr: `20000 >= nested.tenK`, expected: true},
		{expr: `nested.fiveHundred <= 20000`, expected: true},
		{expr: `20000 >= nested.fiveHundred`, expected: true},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("%s is %v", tt.expr, tt.expected)
		t.Run(name, func(t *testing.T) {
			actual, err := EvalStringWithScope(tt.expr, scope)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
