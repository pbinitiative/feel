package tests

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pbinitiative/feel"
	"github.com/pbinitiative/feel/cmd/testcase-extractor/fs"
	"github.com/pbinitiative/feel/cmd/testcase-extractor/model/testconfig"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"testing"
)

func Test_FeelTests(t *testing.T) {
	testConfigFiles := fs.FindFiles("./", "*.yaml", false)
	for _, testConfigFile := range testConfigFiles {
		runFeelTests(t, testConfigFile)
	}
}

func runFeelTests(t *testing.T, testConfigFile string) {
	testCases, err := loadTestConfigs(testConfigFile)

	if err != nil {
		t.Fatalf("Failed to load test configuration for FEEL expressions tests: %v", err)
	}

	for _, testCase := range testCases {
		for _, test := range testCase.Tests {
			testName := fmt.Sprintf(
				"description: %s, expression: %s",
				testCase.Description,
				test.FeelExpression,
			)
			t.Run(testName, func(t *testing.T) {
				log.Default().Println(testName)

				result, err := SafeCall(
					func() (any, error) {
						return feel.EvalString(test.FeelExpression)
					})

				if err != nil {
					t.Fatalf(
						"Failed: %v"+
							"\ndescription: %s"+
							"\nexpression: %s"+
							"\nexpected:"+
							"\n\t%v",
						err,
						testCase.Description,
						test.FeelExpression,
						test.ExpectedResult,
					)
				}

				expected := CreateExpected(t, test.ExpectedResult)
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
						"\ndescription: %s"+
						"\nexpression: %s"+
						"\nexpected:"+
						"\n\t%v",
					testCase.Description,
					test.FeelExpression,
					test.ExpectedResult,
				)
			})
		}
	}
}

func loadTestConfigs(filepath string) ([]testconfig.TestCase, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var testConfigs []testconfig.TestCase
	err = yaml.Unmarshal(data, &testConfigs)
	if err != nil {
		return nil, err
	}

	return testConfigs, nil
}
