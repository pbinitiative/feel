package feel

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_time_is_supported_in_evaluation_scope(t *testing.T) {
	t.Skip("Go's time, and date, and duration values should be supported in scope")
	tests := []struct {
		a      any
		expr   string
		result any
	}{
		{
			a:      time.Date(2025, 04, 13, 17, 27, 58, 123456, time.UTC),
			expr:   `a != null`,
			result: true,
		},
		{
			a:      FEELTime{t: time.Date(2025, 04, 13, 17, 27, 58, 123456, time.UTC)},
			expr:   `a != null`,
			result: true,
		},
	}

	for _, test := range tests {
		t.Run(test.expr, func(t *testing.T) {
			scope := map[string]any{
				"a": test.a,
			}
			actual, err := EvalStringWithScope(test.expr, scope)
			assert.NoError(t, err)
			assert.Equal(t, test.result, actual)
		})
	}
}
