package feel

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_builtin_misc_today_just_returns_a_date(t *testing.T) {
	actual, err := EvalString(`today()`)
	assert.NoError(t, err)

	now := time.Now()
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	expected := FEELDate{t: date}

	assert.Equal(t, &expected, actual)
}
