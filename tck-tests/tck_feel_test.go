//go:build tck_feel_test
// +build tck_feel_test

package tck_tests

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pbinitiative/feel"
	"github.com/pbinitiative/feel/cmd/testcase-extractor/model/testconfig"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
)

const FeelTestCasesConfig = "./test-cases_feel.yaml"

func Test_TckFeelTests(t *testing.T) {
	testConfigs, err := loadTestConfigs(FeelTestCasesConfig)

	if err != nil {
		t.Fatalf("Failed to load test configuration for FEEL expressions tests: %v", err)
	}

	for _, testConfig := range testConfigs {
		for _, testCase := range testConfig.TestCases {
			for _, test := range testCase.Tests {
				testName := fmt.Sprintf(
					"model: '%s', testCase: '%s:%s', expression: '%s'",
					testConfig.Model.Name,
					testCase.Id, testCase.Description,
					test.FeelExpression,
				)
				t.Run(testName, func(t *testing.T) {
					log.Default().Println(testName)

					result, err := safeCall(
						func() (any, error) {
							return feel.EvalString(test.FeelExpression)
						})

					if err != nil {
						t.Fatalf(
							"Failed: %v"+
								"\ndir: %s"+
								"\nmodel: %s"+
								"\ntestcase:"+
								"\n\tid: %s"+
								"\n\tdescription: %s"+
								"\n\texpression: %s"+
								"\n\texpected:"+
								" \n\t\t%v",
							err,
							testConfig.Model.Dir,
							testConfig.Model.Name,
							testCase.Id,
							testCase.Description,
							test.FeelExpression,
							test.ExpectedResult,
						)
					}

					expected := createExpected(t, test.ExpectedResult)
					assert.NoError(t, err)

					diff := cmp.Diff(
						expected,
						result,
						// FEELDate, FEELTime don't not export `t`
						cmpopts.IgnoreUnexported(
							feel.FEELDate{},
							feel.FEELTime{},
						),
					)
					assert.Empty(t, diff,
						"\ndir: %s"+
							"\nmodel: %s"+
							"\ntestcase:"+
							"\n\tid: %s"+
							"\n\tdescription: %s"+
							"\n\texpression: %s"+
							"\n\texpected:"+
							" \n\t\t%v",
						testConfig.Model.Dir,
						testConfig.Model.Name,
						testCase.Id,
						testCase.Description,
						test.FeelExpression,
						test.ExpectedResult,
					)
				})
			}
		}
	}
}
func loadTestConfigs(filepath string) ([]testconfig.TestConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var testConfigs []testconfig.TestConfig
	err = yaml.Unmarshal(data, &testConfigs)
	if err != nil {
		return nil, err
	}

	return testConfigs, nil
}

// SafeCall wraps a function to recover from panics and return errors
func safeCall[T any](fn func() (T, error)) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred: %v", r)
		}
	}()

	return fn()
}

func createExpected(t *testing.T, result testconfig.ExpectedResult) interface{} {
	switch {
	case result.Components != nil:
		comps := make(map[string]any)
		for _, comp := range *result.Components {
			comps[comp.Name] = createExpected(t, comp.ExpectedResult)
		}
		return comps

	case result.Values != nil:
		vals := make([]any, 0)
		for _, val := range *result.Values {
			vals = append(vals, createExpected(t, val))
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
			datetime, err := feel.ParseDatetime(value)
			if err != nil {
				t.Fatalf("Cannot parse expected value as FEEL datetime: %v", err)
			}
			return datetime
		case "date":
			date, err := feel.ParseDate(value)
			if err != nil {
				t.Fatalf("Cannot parse expected value as FEEL date: %v", err)
			}
			return date
		case "decimal", "double":
			return feel.N(value)
		case "duration":
			dur, err := feel.ParseDuration(value)
			if err != nil {
				t.Fatalf("Cannot parse expected value as FEEL duration: %v", err)
			}
			return dur
		case "string":
			return value
		case "time":
			time, err := feel.ParseTime(value)
			if err != nil {
				t.Fatalf("Cannot parse expected value as FEEL time: %v", err)
			}
			return time

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
