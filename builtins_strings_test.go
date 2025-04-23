package feel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_builtin_string_functions(t *testing.T) {
	tests := []struct {
		expr   string
		result any
	}{
		{
			expr:   `substring ( "foobar" , 2 , 2)`,
			result: "oo",
		},
		{
			expr:   `upper case ("foobar")`,
			result: "FOOBAR",
		},
		{
			expr:   `lower case ("FOOBAR")`,
			result: "foobar",
		},
		{
			expr:   `contains ("foobar", "oo")`,
			result: true,
		},
		{
			expr:   `starts with ("foobar", "foo")`,
			result: true,
		},
		{
			expr:   `ends with ("foobar", "bar")`,
			result: true,
		},
		{
			expr:   `string join(["foo","bar"], "-")`,
			result: "foo-bar",
		},
		{
			expr:   `string join(["foo","bar"])`,
			result: "foobar",
		},
		{
			expr:   `"foo" + "bar"`,
			result: "foobar",
		},
		{
			expr:   `string(123)`, // method implemented, but return type should be native Go type
			result: "123",
		},

		//{
		//	expr:   `string length("foobar")`,    // method implemented, but return type should be native Go type
		//	result: 6,
		//},
		//{
		//	expr:   `substring before("foobar", "b")`, // method not yet implemented
		//	result: "foo",
		//},
		//{
		//	expr:   `substring after("foobar", "b")`, // method not yet implemented
		//	result: "ar",
		//},
		//{
		//	expr:   `replace ( "fooXXbar" , "XX" , "" )`, // method not yet implemented
		//	result: "foobar",
		//},
		//{
		//	expr:   `matches("foobar", "^foo")`, // method not yet implemented
		//	result: true,
		//},
		//{
		//	expr:   `split("foo,bar", ",")`, 	// method not yet implemented
		//	result: []string{"foo", "bar"},
		//},
	}
	for _, test := range tests {
		t.Run(test.expr, func(t *testing.T) {
			actual, err := EvalString(test.expr)
			assert.NoError(t, err)
			assert.Equal(t, test.result, actual)
		})
	}
}
