package feel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_builtin_boolean_functions_not(t *testing.T) {
	actual, err := EvalString(`not( true )`)
	assert.NoError(t, err)
	assert.Equal(t, false, actual)

	actual, err = EvalString(`not( 1 > 5 )`)
	assert.NoError(t, err)
	assert.Equal(t, true, actual)
}
