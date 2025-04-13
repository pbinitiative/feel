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

func Test_implicit_type_conversion_in_math_operation(t *testing.T) {
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

func Test_compare_with_null_values_is_always_true(t *testing.T) {
	tests := []struct {
		a      any
		expr   string
		result any
	}{
		{
			a:      "aString",
			expr:   `a != null`,
			result: true,
		},
		{
			a:      42,
			expr:   `a != null`,
			result: true,
		},
		{
			a:      true,
			expr:   `a != null`,
			result: true,
		},
		{
			a:      false,
			expr:   `a != null`,
			result: true,
		},
		//{
		//	a:      time.Now(), // FIXME: this is not detected in compareInterfaces()
		//	expr:   `a != null`,
		//	result: true,
		//},
		//{
		//	a:      []string{"x", "y"}, // FIXME: this is not detected in compareInterfaces()
		//	expr:   `a != null`,
		//	result: true,
		//},
		//{
		//	a:      map[string]any{"x": 1, "y": true}, // FIXME: this is not detected in compareInterfaces()
		//	expr:   `a != null`,
		//	result: true,
		//},
	}

	for _, test := range tests {
		t.Run(test.expr, func(t *testing.T) {
			scope := map[string]any{
				"a": test.a,
			}
			actual, err := EvalStringWithScope(test.expr, scope)
			assert.Nil(t, err)
			assert.Equal(t, test.result, actual)
		})
	}
}

func Test_eval_multiple_operators_with_logical_precedence(t *testing.T) {
	// FIXME: implement operator precedence
	t.Skip("implement operator precedence")
	tests := []struct {
		a      any
		b      any
		x      any
		y      any
		expr   string
		result any
	}{
		{
			a: true,
			b: 1,
			// hint: no-equal must be before and
			expr:   `a and b != null`,
			result: true,
		},
		{
			a:      true,
			b:      false,
			x:      true,
			y:      true,
			expr:   `a and b or x and y`,
			result: true,
		},
	}

	for _, test := range tests {
		t.Run(test.expr, func(t *testing.T) {
			scope := map[string]any{
				"a": test.a,
				"b": test.b,
				"x": test.x,
				"y": test.y,
			}
			actual, err := EvalStringWithScope(test.expr, scope)
			assert.Nil(t, err)
			assert.Equal(t, test.result, actual)
		})
	}
}
