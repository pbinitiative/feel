package feel

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_N(t *testing.T) {
	input := "5 + 6"
	res, err := EvalString(input)
	if err != nil {
		fmt.Printf("bad input %s\n", input)
	}
	assert.Nil(t, err)
	assert.Empty(t, cmp.Diff(N(11), res))
}

func Test_unknown_symbols_raise_error(t *testing.T) {
	scope := map[string]any{
		"COUNT": 15,
	}
	input := `COUNT <= 20 && COUNT > 10`
	_, err := EvalStringWithScope(input, scope)
	assert.ErrorContains(t, err, "bad input")
}

func Test_EvalString(t *testing.T) {
	tests := []struct {
		input  string
		expect any
	}{
		// empty input outputs nil
		{"", nil},

		{"5 + -6", N(-1)},
		{"5 + 6", N(11)},
		{"(function(a) 2 * a)(5)", N(10)},
		{"true", true},
		{"false", false},
		{`"hello" + " world"`, "hello world"},

		{`{a if c: "hello", b: "world"}`, map[string]any{"a if c": "hello", "b": "world"}},

		// in range and array
		{`5 in (5..8]`, false},
		{`5 in [5..8)`, true},
		{`8 in [5..8)`, false},
		{`8 in [5..8]`, true},

		{`"a" in ["a".."z"]`, true},
		{`5 in [3,5, 8]`, true},
		{`5 in [3, 6, 8]`, false},
		{`5 in []`, false},
		//{`not(5 in [3, 5, 9])`, false},

		// if then else
		{`bind("a", 5); if a > 3 then "larger" else "smaller"`, "larger"},
		{`bind("a", 5); if a = 5 then "equal" else "not equal"`, "equal"},
		{`bind("a b", 5); if a b = 5 then "equal" else "not equal"`, "equal"}, // a name has multiple chunks

		{`help(bind)`, "bind value to name in current top scope"},

		// test not
		{`not( 5 >  6)`, true},

		// loop functions
		{`some x in [3, 4, 5] satisfies x >= 4`, N(4)},
		{`every y in [3, 4, 5] satisfies y >= 4`, []any{N(4), N(5)}},

		// null check
		{`a != null and a.b > 10`, false},
		{`a = null or a.b > 10`, true},

		// keyword arguments
		{`bind("sub", function(a, b) a - b); sub(a: 4, b: 2)`, N(2)},

		// temporal expressions
		{`last day of month(@"2020-02-11")`, N(29)},
		{`last day of month(@"2021-01-07")`, N(31)},
		{`last day of month(@"2023-06-11")`, N(30)},
		{`last day of month(@"2023-07-11")`, N(31)},

		{`@"2023-07-21T13:57:32@CST" - @"PT2H3M"`, MustParseDatetime("2023-07-21T11:54:32@CST")}, // test day/hour/min duration
		{`@"2023-06-01T10:33:20@CST" + @"P3Y11M"`, MustParseDatetime("2027-05-01T10:33:20@CST")}, // test year/month duration

		// builtin functions
		{`is defined(x)`, false},
		{`bind("x", [1, 2, 3]); is defined(x[5])`, false},
		{`bind("x", {a: 3, b: 5}); is defined(x.c)`, false},
		{`bind("x", {a: 3, b: 5}); is defined(x.a)`, true},

		{`bind("x", 666); is defined(x)`, true},        // `x` is bound
		{`bind("x", 888); is defined(value: x)`, true}, // macro can use keyword arguments

		{`substring(string: "abcdef", start position: 3, length: 3)`, "cde"},
		{`substring(string: "abcdef", start position: 200, length: 3)`, ""},
		{`not({})`, true},
		{`not({a: 1})`, false},

		// list functions
		{`median([3, 5, 9, 1, "hello", -2])`, N(3)},

		{`append(["hello"], " ", "world")`, []any{"hello", " ", "world"}},
		{`concatenate([2, 1], [3])`, []any{N(2), N(1), N(3)}},
		{`insert before(["hello", "world"], 2, "another")`, []any{"hello", "another", "world"}},
		{`remove(["hello", "a", "world"], 2)`, []any{"hello", "world"}},

		{`index of([1,2,3,2],2)`, []any{N(2), N(4)}},

		{`distinct values([1, 2, 1, 2, 3, 2, 1])`, []any{N(1), N(2), N(3)}},
		{`flatten([["a"], [["b", ["c"]]], ["d"]])`, []any{"a", "b", "c", "d"}},
		{`union(["a", "b"], ["b", "c"], ["d"])`, []any{"a", "b", "c", "d"}},

		{`sort(["hello", "a", "world"], function(x, y) x < y)`, []any{"a", "hello", "world"}},
		{`sort([8, -1, 3], function(x, y) x > y)`, []any{N(8), N(3), N(-1)}},

		{`string join(["hello", "world"])`, "helloworld"},
		{`string join(["hello", "world"], " ", "[", "]")`, "[hello world]"},

		{`or([false, 0, true, false, 1])`, true},
		{`and([false, 0, true, false, 1])`, false},
		{`and([true, 1, true, "ok"])`, true},

		// context/map functions
		{`get value({a: 2}, "b")`, null},
		{`get value({a: 2}, "a")`, N(2)},
		{`get value({a: {b: {c: 4}}}, ["a", "b", "c"])`, N(4)},
		{`get value({a: {b: {c: 4}}}, ["a", "b"])`, map[string]any{"c": N(4)}},
		{`get value({a: {b: {c: 4}}}, ["a", "k"])`, null},
		{`get value(context put({a: false}, ["b", "c", "d"], 4), ["b", "c"])`, map[string]any{"d": N(4)}},
		{`context merge([{x:1, y: 0}, {y:2}])`, map[string]any{"x": N(1), "y": N(2)}},

		// range functions
		{`before(1, 10)`, true},
		{`before(10, 1)`, false},
		{`before([1..5], 10)`, true},
		{`before(1, [2..5])`, true},
		{`before(3, [2..5])`, false},

		{`before([1..5),[5..10])`, true},
		{`before([1..5),(5..10])`, true},
		{`before([1..5],[5..10])`, false},
		{`before([1..5),(5..10])`, true},

		{`after([5..10], [1..5))`, true},
		{`after((5..10], [1..5))`, true},
		{`after([5..10], [1..5])`, false},
		{`after((5..10], [1..5))`, true},

		{`meets([1..5], [5..10])`, true},
		{`meets([1..3], [4..6])`, false},
		{`meets([1..3], [3..5])`, true},
		{`meets([1..5], (5..8])`, false},

		{`met by([5..10], [1..5])`, true},
		{`met by([3..4], [1..2])`, false},
		{`met by([3..5], [1..3])`, true},
		{`met by((5..8], [1..5))`, false},
		{`met by([5..10], [1..5))`, false},

		{`overlaps([5..10], [1..6])`, true},
		{`overlaps((3..7], [1..4])`, true},
		{`overlaps([1..3], (3..6])`, false},
		{`overlaps((5..8], [1..5))`, false},
		{`overlaps([4..10], [1..5))`, true},

		{`overlaps before([1..5], [4..10])`, true},
		{`overlaps before([3..4], [1..2])`, false},
		{`overlaps before([1..3], (3..5])`, false},
		{`overlaps before([1..5), (3..8])`, true},
		{`overlaps before([1..5), [5..10])`, false},

		{`overlaps after([4..10], [1..5])`, true},
		{`overlaps after([3..4], [1..2])`, false},
		{`overlaps after([3..5], [1..3))`, false},
		{`overlaps after((5..8], [1..5))`, false},
		{`overlaps after([4..10], [1..5))`, true},

		{`finishes(5, [1..5])`, true},
		{`finishes(10, [1..7])`, false},
		{`finishes([3..5], [1..5])`, true},
		{`finishes((1..5], [1..5))`, false},
		{`finishes([5..10], [1..10))`, false},

		{`finished by([5..10], 10)`, true},
		{`finished by([3..4], 2)`, false},

		{`finished by([3..5], [1..5])`, true},
		{`finished by((5..8], [1..5))`, false},
		{`finished by([5..10], (1..10))`, true},

		{`includes([5..10], 6)`, true},
		{`includes([3..4], 5)`, false},
		{`includes([1..10], [4..6])`, true},
		{`includes((5..8], [1..5))`, false},
		{`includes([1..10], [1..5))`, true},

		{`during(5, [1..10])`, true},
		{`during(12, [1..10])`, false},
		{`during(1, (1..10])`, false},
		{`during([4..6], [1..10))`, true},
		{`during((1..5], (1..10])`, true},

		{`starts(1, [1..5])`, true},
		{`starts(1, (1..8])`, false},
		{`starts((1..5], [1..5])`, false},
		{`starts([1..10], [1..10])`, true},
		{`starts((1..10), (1..10))`, true},

		{`started by([1..10], 1)`, true},
		{`started by((1..10], 1)`, false},
		{`started by([1..10], [1..5])`, true},
		{`started by((1..10], [1..5))`, false},
		{`started by([1..10], [1..10))`, true},

		{`coincides([1..5], [1..5])`, true},
		{`coincides((1..5], [1..5))`, false},
		{`coincides([1..5], [2..6])`, false},
	}

	for _, p := range tests {
		t.Run(fmt.Sprintf("eval: %s", p.input), func(t *testing.T) {
			res, err := EvalString(p.input)
			if err != nil {
				fmt.Printf("bad input %s\n", p.input)
			}
			assert.Nil(t, err)
			assert.Empty(t, cmp.Diff(p.expect, res))
		})
	}
}

func Test_EvalStringWithScope_unary_with_default_scope(t *testing.T) {
	input := `> 8, <= 5`
	v, err := EvalStringWithScope(input, Scope{"?": 4})
	assert.Nil(t, err)
	assert.True(t, v.(bool))
}

func Test_EvalStringWithScope(t *testing.T) {
	input := `foo + bar`
	v, err := EvalStringWithScope(input, Scope{"foo": 5, "bar": 7})
	assert.Nil(t, err)
	assert.True(t, N(12).Equal(*v.(*Number)))
}

func Test_EvalStringWithScope_contexts(t *testing.T) {
	scope := Scope{
		"data": Scope{
			"foo": "foo", "bar": "bar",
		},
	}
	v, err := EvalStringWithScope(`get value( data, "foo" ) + get value( data, "bar" )`, scope)
	assert.Nil(t, err)
	assert.Equal(t, "foobar", v)
}

func Test_EvalStringWithScope_contexts_with_struct(t *testing.T) {
	type TestItem struct {
		Key string
	}
	scope := Scope{
		"data": TestItem{Key: "foobar"},
	}
	v, err := EvalStringWithScope(`get value( data, "Key" )`, scope)
	assert.Nil(t, err)
	assert.Equal(t, "foobar", v)
}

func TestTemporalValue(t *testing.T) {
	input := `@"2023-06-07".day`
	v, err := EvalString(input)
	assert.Nil(t, err)
	assert.Empty(t, cmp.Diff(v, N(7)))

	input1 := `@"2023-06-07T15:08:39".second`
	v1, err := EvalString(input1)
	assert.Nil(t, err)
	assert.Empty(t, cmp.Diff(v1, N(39)))

	input2 := `@"P1DT3H25M60S".minutes`
	v2, err := EvalString(input2)
	assert.Nil(t, err)
	assert.Empty(t, cmp.Diff(v2, N(25)))

	dt, err := ParseDatetime(`2023-06-07T15:04:05`)
	assert.Nil(t, err)
	assert.Empty(t, cmp.Diff(dt.t.Hour(), 15))
	assert.Empty(t, cmp.Diff(dt.t.Second(), 5))

	dur, err := ParseDuration("P12Y2M")
	assert.Nil(t, err)
	assert.Equal(t, 12, dur.Years)
	assert.Equal(t, 2, dur.Months)

	dur1, err := ParseDuration("P7M")
	assert.Nil(t, err)
	assert.Equal(t, 0, dur1.Years)
	assert.Equal(t, 7, dur1.Months)

	dur2, err := ParseDuration("PT20H")
	assert.Nil(t, err)
	assert.Equal(t, 20, dur2.Hours)
	assert.Equal(t, 0, dur2.Seconds)

	td, err := time.ParseDuration("3h37m20s")
	assert.Nil(t, err)
	dur3 := NewFEELDuration(td)
	assert.Equal(t, 3, dur3.Hours)
	assert.Equal(t, 37, dur3.Minutes)
	assert.Equal(t, 20, dur3.Seconds)
}
