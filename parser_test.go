package feel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimpleParser(t *testing.T) {
	input := `
	abc + 3 * sum(2, 7) - @"2023-06-01 05:01:00@Asia/Shanghai"
	`
	ast, err := ParseString(input)
	assert.NoError(t, err)
	assert.Equal(t, `(- (+ abc (* 3 (call sum [2, 7]))) @"2023-06-01 05:01:00@Asia/Shanghai")`, ast.Repr())

}

func TestStringVal(t *testing.T) {
	input := `
	"a string\n value"
	`
	ast, err := ParseString(input)
	assert.NoError(t, err)
	strVal, ok := ast.(*StringNode)
	assert.True(t, ok)

	assert.Equal(t, "a string\u000A value", strVal.Content())
}

func TestContBinop(t *testing.T) {
	input := `
	abc + 3 * u - eight.value
	`
	ast, err := ParseString(input)
	assert.NoError(t, err)
	assert.Equal(t, "(- (+ abc (* 3 u)) (. eight value))", ast.Repr())
}

func TestCompare(t *testing.T) {
	input := `
	a = 3
	`
	ast, err := ParseString(input)
	assert.NoError(t, err)
	assert.Equal(t, "(= a 3)", ast.Repr())
}
func TestIfExpression(t *testing.T) {
	input := `
	if a > 3 
	then "yes" else "no"
	`
	ast, err := ParseString(input)
	assert.NoError(t, err)
	assert.Equal(t, "(if (> a 3) \"yes\" \"no\")", ast.Repr())
}

func TestForExpression(t *testing.T) {
	input := `
	for x in [3, 4], y in [5, 9] return x * y
	`
	ast, err := ParseString(input)
	assert.NoError(t, err)
	assert.Equal(t, "(for x [3, 4] (for y [5, 9] (* x y)))", ast.Repr())
}

func TestSomeExpression(t *testing.T) {
	input := `
	some x in [3, 4, 5, 6, 9] satisfies x >= 5
	`
	ast, err := ParseString(input)
	assert.NoError(t, err)
	assert.Equal(t, "(some \"x\" [3, 4, 5, 6, 9] (>= x 5))", ast.Repr())
}

func TestRange(t *testing.T) {
	input := `
	(2..5]
	`
	ast, err := ParseString(input)
	assert.NoError(t, err)
	assert.Equal(t, "(2..5]", ast.Repr())

	input1 := `
	[2..5 * 9)
	`
	ast1, err := ParseString(input1)
	assert.NoError(t, err)
	assert.Equal(t, "[2..(* 5 9))", ast1.Repr())
}

func TestUnaryTests(t *testing.T) {
	input := `
	(2..5], >= 6, 100, !=888
	`
	ast, err := ParseString(input)
	assert.NoError(t, err)
	assert.Equal(t, "(multitests (2..5] (>= ? 6) 100 (!= ? 888))", ast.Repr())
}

func TestFunCallAndIndex(t *testing.T) {
	input := `
	a.b.c(3, 4).d.e[4].f + 1001
	`
	ast, err := ParseString(input)
	assert.NoError(t, err)
	assert.Equal(t, `(+ (. ([] (. (. (call (. (. a b) c) [3, 4]) d) e) 4) f) 1001)`, ast.Repr())

	input1 := `
	string contains("abc def" , "e")
	`
	ast1, err := ParseString(input1)
	assert.NoError(t, err)
	assert.Equal(t, "(call `string contains` [\"abc def\", \"e\"])", ast1.Repr())
}

func TestFunDef(t *testing.T) {
	input := `
	function(a, b) a + b * 2
	`
	ast, err := ParseString(input)
	assert.NoError(t, err)
	assert.Equal(t, `(function [a, b] (+ a (* b 2)))`, ast.Repr())

	_, err1 := ParseString(`function(a, b`)
	assert.NotNil(t, err1)
	une, ok := err1.(*UnexpectedToken)
	assert.True(t, ok)
	assert.Equal(t, TokenEOF, une.token.Kind)
	assert.Equal(t, []string{")", ","}, une.expects)
}

func TestMapValue(t *testing.T) {
	input := `
	{ a: 1, b: @"2023-06-01", c: [1, 2, 3]}
	`
	ast, err := ParseString(input)
	assert.NoError(t, err)
	assert.Equal(t, `(map ("a" 1) ("b" @"2023-06-01") ("c" [1, 2, 3]))`, ast.Repr())

	// parse incomplete inputs
	input1 := `
	{ a: 1, b: @"2023-06-01",
	`
	_, err1 := ParseString(input1)
	assert.NotNil(t, err1)
	une, ok := err1.(*UnexpectedToken)
	assert.True(t, ok)
	assert.Equal(t, TokenEOF, une.token.Kind)
	assert.Equal(t, []string{"name", "string"}, une.expects)
}

func TestTemporal(t *testing.T) {
	input := `@"2023-06-07"`
	ast, err := ParseString(input)
	assert.NoError(t, err)
	assert.Equal(t, `@"2023-06-07"`, ast.Repr())
	assert.Equal(t, 0, ast.TextRange().Start.Column)
	assert.Equal(t, 13, ast.TextRange().End.Column)

	node, ok := ast.(*TemporalNode)
	assert.True(t, ok)
	assert.Equal(t, "2023-06-07", node.Content())
}
