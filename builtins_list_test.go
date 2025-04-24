package feel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_builtin_list_functions_list_contains(t *testing.T) {
	actual, err := EvalString(`list contains( [1,2,3], 2 )`)
	assert.NoError(t, err)
	assert.Equal(t, true, actual)
}

func Test_builtin_list_functions_list_replace(t *testing.T) {
	t.Skip("method not yet supported")

	actual, err := EvalString(`list replace( [2, 4, 7, 8], 3, 6)`)
	assert.NoError(t, err)
	assert.Equal(t, []int{2, 4, 6, 8}, actual)
}

func Test_builtin_list_functions_list_replace_function(t *testing.T) {
	t.Skip("method not yet supported")

	actual, err := EvalString(`list replace ( [2, 4, 7, 8], function(item, newItem) item < newItem, 5)`)
	assert.NoError(t, err)
	assert.Equal(t, []int{5, 5, 7, 8}, actual)
}

func Test_builtin_list_functions_count(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`count( [1,2,3] )`)
	assert.NoError(t, err)
	assert.Equal(t, 3, actual)
}

func Test_builtin_list_functions_min(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`min( [1,2,3] )`)
	assert.NoError(t, err)
	assert.Equal(t, 1, actual)

	actual, err = EvalString(`min( ["a","b","c"] )`)
	assert.NoError(t, err)
	assert.Equal(t, "a", actual)

	actual, err = EvalString(`min( [] )`)
	assert.NoError(t, err)
	assert.Equal(t, nil, actual)
}

func Test_builtin_list_functions_max(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`max( [1,2,3] )`)
	assert.NoError(t, err)
	assert.Equal(t, 3, actual)

	actual, err = EvalString(`max( ["a","b","c"] )`)
	assert.NoError(t, err)
	assert.Equal(t, "c", actual)

	actual, err = EvalString(`max( [] )`)
	assert.NoError(t, err)
	assert.Equal(t, nil, actual)
}

func Test_builtin_list_functions_sum(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`sum( [1,2,3] )`)
	assert.NoError(t, err)
	assert.Equal(t, 6, actual)
}

func Test_builtin_list_functions_mean(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`mean( [1,2,3] )`)
	assert.NoError(t, err)
	assert.Equal(t, 2, actual)
}

func Test_builtin_list_functions_all(t *testing.T) {
	actual, err := EvalString(`all( [false,null,true] )`)
	assert.NoError(t, err)
	assert.Equal(t, false, actual)
}

func Test_builtin_list_functions_any(t *testing.T) {
	actual, err := EvalString(`any( [false,null,true] )`)
	assert.NoError(t, err)
	assert.Equal(t, true, actual)
}

func Test_builtin_list_functions_sublist(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`sublist( [4,5,6], 1, 2 )`)
	assert.NoError(t, err)
	assert.Equal(t, []int{4, 5}, actual)
}

func Test_builtin_list_functions_append(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`append( [1], 2, 3 )`)
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, actual)
}

func Test_builtin_list_functions_concatenate(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`concatenate( ["a","b"],["c"] )`)
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, actual)
}

func Test_builtin_list_functions_insert_before(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`insert before( ["a","c"],1,"b")`)
	assert.NoError(t, err)
	assert.Equal(t, []string{"b", "a", "c"}, actual)
}

func Test_builtin_list_functions_remove(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`remove( ["a","b", "c"], 2 )`)
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "c"}, actual)
}

func Test_builtin_list_functions_reverse(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`reverse( ["a", "b", "c"])`)
	assert.NoError(t, err)
	assert.Equal(t, []string{"c", "b", "a"}, actual)
}

func Test_builtin_list_functions_index_of(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`index of( [1,2,3,2],2 )`)
	assert.NoError(t, err)
	assert.Equal(t, []int{2, 4}, actual)
}

func Test_builtin_list_functions_union(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`union( [1,2],[2,3] )`)
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, actual)
}

func Test_builtin_list_functions_distinct_values(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`distinct values( [1,2,3,2,1] )`)
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, actual)
}

func Test_builtin_list_functions_flatten(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`flatten( [[1,2],[[3]], 4] )`)
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4}, actual)
}

func Test_builtin_list_functions_product(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`product( [2, 3, 4] )`)
	assert.NoError(t, err)
	assert.Equal(t, []int{24}, actual)
}

func Test_builtin_list_functions_median(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`median( 8, 2, 5, 3, 4 )`)
	assert.NoError(t, err)
	assert.Equal(t, 4, actual)
}

func Test_builtin_list_functions_stddev(t *testing.T) {
	t.Skip("method implemented, but return type should be native Go type")

	actual, err := EvalString(`stddev( 2, 4, 7, 5 )`)
	assert.NoError(t, err)
	assert.Equal(t, 2.0816659994, actual)
}

func Test_builtin_list_functions_mode(t *testing.T) {
	t.Skip("method not yet supported")

	actual, err := EvalString(`mode( 6, 3, 9, 6, 6 )`)
	assert.NoError(t, err)
	assert.Equal(t, []int{6}, actual)
}
