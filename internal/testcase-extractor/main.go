package main

import (
	"encoding/xml"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

type TckDefinitions struct {
	XMLName   xml.Name      `xml:"definitions"`
	Decisions []TckDecision `xml:"decision"`
}

type TckDecision struct {
	XMLName           xml.Name `xml:"decision"`
	Name              string   `xml:"name,attr"`
	LiteralExpression string   `xml:"literalExpression>text"`
}

type TckTestCases struct {
	XMLName   xml.Name      `xml:"testCases"`
	ModelName string        `xml:"modelName"`
	TestCases []TckTestCase `xml:"testCase"`
}

type TckTestCase struct {
	XMLName     xml.Name            `xml:"testCase"`
	Id          string              `xml:"id,attr"`
	Description string              `xml:"description"`
	Results     []TckTestCaseResult `xml:"resultNode"`
}

type TckTestCaseResult struct {
	XMLName  xml.Name `xml:"resultNode"`
	Name     string   `xml:"name,attr"`
	Expected string   `xml:"expected>value"`
}

type TestConfig struct {
	ModelName string     `yaml:"model-name"`
	TestCases []TestCase `yaml:"test-cases"`
}

type TestCase struct {
	Id          string           `yaml:"id"`
	Description string           `yaml:"description"`
	Results     []TestCaseResult `yaml:"results"`
}

type TestCaseResult struct {
	FeelExpression string `yaml:"feel-expression"`
	Expected       string `yaml:"expected"`
}

// findFiles traverses the directory tree starting at dir and returns files
// matching the wildcard pattern
func findFiles(dir, wildcard string) ([]string, error) {
	var filePaths []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			matched, err := filepath.Match(wildcard, info.Name())
			if err != nil {
				return err
			}
			if matched {
				filePaths = append(filePaths, path)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return filePaths, nil
}

func extractTestCases(filePath string) (*TckTestCases, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var testCases *TckTestCases
	if err := xml.Unmarshal(bytes, &testCases); err != nil {
		panic(err)
	}

	return testCases, nil
}

func extractDecisions(filePath string) ([]TckDecision, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var definitions *TckDefinitions
	if err := xml.Unmarshal(bytes, &definitions); err != nil {
		panic(err)
	}

	return definitions.Decisions, nil
}

func mapTckResultsByName(testCases []TckTestCase) map[string]TckTestCaseResult {
	resultsMap := make(map[string]TckTestCaseResult)
	for _, testCase := range testCases {
		for _, result := range testCase.Results {
			resultsMap[result.Name] = result
		}
	}

	return resultsMap
}

func printError(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

func main() {
	// TODO[JSot]: Uncomment before merge
	//dir := *flag.String("dir", ".", "Directory to start traversal")
	//flag.Parse()

	// TODO[JSot]: TESTING STUFF. Remove before merge,
	//dir := "./resources/tck/TestCases"
	dir := "./resources/tck/TestCases/compliance-level-2"
	//dir := "./resources/tck/TestCases/compliance-level-2/0101-feel-constants"

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		printError("Error: Directory '%s' does not exist\n", dir)
		os.Exit(1)
	}

	files, err := findFiles(dir, "*-feel-*.dmn")
	if err != nil {
		printError("Error: %v\n", err)
		os.Exit(1)
	}

	testConfigs := make([]TestConfig, 0)
	for _, dmnFile := range files {
		testCases := make([]TestCase, 0)
		testCaseFile := strings.TrimSuffix(dmnFile, filepath.Ext(dmnFile)) + "-test-01.xml"

		tckTestCases, err := extractTestCases(testCaseFile)
		if err != nil {
			printError("Error: could not read Test Cases: %v\n", err)
		}

		decisions, err := extractDecisions(dmnFile)
		if err != nil {
			printError("Error: could not read Decisions: %v\n", err)
		}

		tckResultsMap := mapTckResultsByName(tckTestCases.TestCases)
		for _, tckTestCase := range tckTestCases.TestCases {
			results := make([]TestCaseResult, 0)
			for _, decision := range decisions {
				tckResult := tckResultsMap[decision.Name]
				result := TestCaseResult{
					FeelExpression: decision.LiteralExpression,
					Expected:       tckResult.Expected,
				}
				results = append(results, result)
			}

			testCase := TestCase{
				Id:          tckTestCase.Id,
				Description: tckTestCase.Description,
				Results:     results,
			}
			testCases = append(testCases, testCase)
		}

		testConfig := TestConfig{
			ModelName: tckTestCases.ModelName,
			TestCases: testCases,
		}
		testConfigs = append(testConfigs, testConfig)
	}

	yamlData, err := yaml.Marshal(&testConfigs)
	if err != nil {
		printError("Error marshaling to YAML: %v\n", err)
		os.Exit(1)
	}

	// TODO[JSot]: TESTING STUFF. Remove before merge
	fmt.Println(string(yamlData))

	err = os.WriteFile("test-cases.yaml", yamlData, 0644)
	if err != nil {
		printError("Error writing YAML file: %v\n", err)
		os.Exit(1)
	}
}
