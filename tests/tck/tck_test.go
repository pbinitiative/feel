package tck

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pbinitiative/feel"
	"github.com/pbinitiative/feel/cmd/testcase-extractor/fs"
	"github.com/pbinitiative/feel/cmd/testcase-extractor/model/tcktestconfig"
	"github.com/pbinitiative/feel/tests"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"testing"
)

func Test_TckFeelTests(t *testing.T) {
	testConfigFiles := fs.FindFiles("./", "*.yaml", true)
	for _, testConfigFile := range testConfigFiles {
		runFeelTests(t, testConfigFile)
	}
}

func runFeelTests(t *testing.T, testConfigFile string) {
	testConfigs, err := loadTckTestConfigs(testConfigFile)

	if err != nil {
		t.Fatalf("Failed to load test configuration for FEEL expressions tests: %v", err)
	}

	for _, testConfig := range testConfigs {
		for _, testCase := range testConfig.TestCases {
			for _, test := range testCase.Tests {
				testName := fmt.Sprintf(
					"model: '%s', testCase: '%s, %s', expression: '%s'",
					testConfig.Model.Name,
					testCase.Id, testCase.Description,
					test.FeelExpression,
				)
				t.Run(testName, func(t *testing.T) {
					log.Default().Println(testName)

					result, err := tests.SafeCall(
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

					expected := tests.CreateExpected(t, test.ExpectedResult)
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

func loadTckTestConfigs(filepath string) ([]tcktestconfig.TestConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var testConfigs []tcktestconfig.TestConfig
	err = yaml.Unmarshal(data, &testConfigs)
	if err != nil {
		return nil, err
	}

	return testConfigs, nil
}
