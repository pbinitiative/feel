package tests

import (
	"fmt"
	"github.com/pbinitiative/feel"
	"github.com/pbinitiative/feel/cmd/testcase-extractor/model/testconfig"
	"strconv"
	"strings"
	"testing"
	"time"
)

// SafeCall wraps a function to recover from panics and return errors
func SafeCall[T any](fn func() (T, error)) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred: %v", r)
		}
	}()

	return fn()
}

// CreateExpected creates an expected result ready for comparison in test
// functions. Expected results are create based on expected result type
// (boolean, dateTime, ...)
func CreateExpected(t *testing.T, result testconfig.ExpectedResult) interface{} {
	switch {
	case result.Components != nil:
		comps := make(map[string]any)
		for _, comp := range *result.Components {
			comps[comp.Name] = CreateExpected(t, comp.ExpectedResult)
		}
		return comps

	case result.Values != nil:
		vals := make([]any, 0)
		for _, val := range *result.Values {
			vals = append(vals, CreateExpected(t, val))
		}
		return vals

	case result.Value != nil:
		if result.Value.Value == nil {
			return nil
		}

		value := *result.Value.Value
		theType := *result.Value.Type
		switch theType {
		case "boolean":
			b, err := strconv.ParseBool(value)
			if err != nil {
				t.Fatalf("Cannot parse expected value as boolean: %v", err)
			}
			return b
		case "dateTime":
			formats := []string{
				"2006-01-02T15:04:05",
				"2006-01-02T15:04:05-07:00",
				"2006-01-02T15:04:05@MST",
			}
			for _, format := range formats {
				if datetime, err := time.Parse(format, value); err == nil {
					return datetime
				}
			}
			t.Fatalf("Cannot parse expected value as datetime: %s", value)
		case "date":
			date, err := time.Parse(time.DateOnly, value)
			if err != nil {
				t.Fatalf("Cannot parse expected value as date: %v", err)
			}
			return date
		case "decimal", "double":
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				t.Fatalf("Cannot parse expected value as date: %v", err)
			}
			return f
		case "duration":
			dur, err := feel.ParseDuration(value)
			if err != nil {
				t.Fatalf("Cannot parse expected value as duration: %v", err)
			}
			return dur.Duration() // A bit of a cheat here because go does have a std lib that can read ISO duration
		case "string":
			return value
		case "time":
			formats := []string{
				"15:04:05Z07:00",
				time.TimeOnly,
			}
			for _, format := range formats {
				if datetime, err := time.Parse(format, value); err == nil {
					return datetime
				}
			}
			t.Fatalf("Cannot parse expected value as datetime: %s", value)

		default:
			t.Fatalf(
				"Unsupported value type: '%s'. Supported types are: %s",
				theType,
				strings.Join([]string{
					"boolean",
					"date",
					"dateTime",
					"decimal", "double",
					"duration",
					"string",
					"time",
				}, ", "),
			)
		}
	default:
		t.Fatalf(
			"Unsupported 'ExpectedResult' type: %#v."+
				" Supported types are 'Components', 'Value' and 'Values'",
			result,
		)
	}

	return nil
}
