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

func Test_implicit_type_conversation_in_math_operation(t *testing.T) {
	tests := []struct {
		a      any
		b      any
		expr   string
		result any
	}{
		{
			a:      1,
			b:      "foo",
			expr:   `a + b`,
			result: "1foo",
		},
		{
			a:      2,
			b:      "bar",
			expr:   `b + a`,
			result: "bar2",
		},
		{
			a:      37,
			expr:   `a != null`,
			result: true,
		},
		{
			b:      "foobar",
			expr:   `b != null`,
			result: true,
		},
	}

	for _, test := range tests {
		t.Run(test.expr, func(t *testing.T) {
			scope := map[string]any{
				"a": test.a,
				"b": test.b,
			}
			actual, err := EvalStringWithScope(test.expr, scope)
			assert.Nil(t, err)
			assert.Equal(t, test.result, actual)
		})
	}
}
