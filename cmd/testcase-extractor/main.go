package main

import (
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"github.com/pbinitiative/feel/cmd/testcase-extractor/fs"
	"github.com/pbinitiative/feel/cmd/testcase-extractor/model/dmn"
	"github.com/pbinitiative/feel/cmd/testcase-extractor/model/tck"
	"github.com/pbinitiative/feel/cmd/testcase-extractor/model/tcktestconfig"
	"github.com/pbinitiative/feel/cmd/testcase-extractor/model/testconfig"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	dmnModelFileWildcard     = "*-feel-*.dmn"
	testCaseFileSuffix       = "-test-01.xml"
	xmlSchemaNamespacePrefix = "xsd:"
)

func parseFlags() (
	dir *string,
	outputDir *string,
	forceStats *bool,
) {
	dir = flag.String(
		"dir",
		".",
		"Directory to start search for TCK DMN model files and Test Cases files.",
	)
	outputDir = flag.String(
		"output-dir",
		"./feel",
		`Dir to save output to. Default is "./feel".`,
	)
	forceStats = flag.Bool(
		"force-stats",
		false,
		"Force statistics report. Useful when writing to STDOUT and want statistics.",
	)
	flag.Parse()

	return
}

func ensureOutputFile(outputFilename string) *os.File {
	var outputFile *os.File
	var err error

	if outputFilename == "-" {
		outputFile = os.Stdout
	} else {
		fs.EnsureDir(filepath.Dir(outputFilename))
		outputFile, err = os.Create(outputFilename)
		if err != nil {
			printError("Error getting output filename absolute path: %v\n", err)
			panic(err)
		}
	}

	return outputFile
}

func extractTckTestCases(
	testCasesFilePath string,
	tckTestCasesCh chan *tck.TestCases,
) {
	defer close(tckTestCasesCh)

	bytes, err := os.ReadFile(testCasesFilePath)
	if err != nil {
		printError("Error: could not read Test Cases file %v: %v\n", testCasesFilePath, err)
	}

	var testCases *tck.TestCases
	err = xml.Unmarshal(bytes, &testCases)
	if err != nil {
		printError("Error: could not read Test Cases file %v: %v\n", testCasesFilePath, err)
	}

	tckTestCasesCh <- testCases
}

func extractDmnDecisions(decisionsFilePath string, dmnDecisionsCh chan []*dmn.Decision) {
	defer close(dmnDecisionsCh)

	bytes, err := os.ReadFile(decisionsFilePath)
	if err != nil {
		printError("Error: could not read Test Cases file: %v\n", err)
	}

	var definitions *dmn.Definitions
	err = xml.Unmarshal(bytes, &definitions)
	if err != nil {
		printError("Error: could not read Test Cases file: %v\n", err)
	}

	dmnDecisionsCh <- definitions.Decisions
}

func compileTestConfigs(dmnFiles []string, searchDir string) []tcktestconfig.TestConfig {
	testConfigsCh := make(chan tcktestconfig.TestConfig)

	for _, dmnFile := range dmnFiles {
		go compileTestConfig(dmnFile, searchDir, testConfigsCh)
	}

	testConfigs := make([]tcktestconfig.TestConfig, 0)
	for range len(dmnFiles) {
		testConfigs = append(testConfigs, <-testConfigsCh)
	}

	sort.Slice(testConfigs, func(i, j int) bool {
		return testConfigs[i].Model.Name < testConfigs[j].Model.Name
	})

	return testConfigs
}

func compileExpectedResult(tckValue *tck.ValueType) testconfig.ExpectedResult {
	switch {
	case tckValue.Components != nil:
		components := make([]testconfig.Component, 0)
		for _, tckComponent := range *tckValue.Components {
			expectedResult := compileExpectedResult(tckComponent.ValueType)
			components = append(
				components,
				testconfig.Component{
					Name:           tckComponent.NameAttr,
					ExpectedResult: expectedResult,
				},
			)
		}

		return testconfig.ExpectedResult{Components: &components}
	case tckValue.Value != nil:
		return testconfig.ExpectedResult{
			Value: &testconfig.ExpectedValue{
				Value: tckValue.Value.Content,
				Type:  stripNamespacePrefix(tckValue.Value.Type),
			},
		}
	case tckValue.List != nil:
		results := make([]testconfig.ExpectedResult, 0)
		for _, result := range tckValue.List.Items {
			expectedResult := compileExpectedResult(result)
			results = append(
				results,
				expectedResult,
			)
		}

		return testconfig.ExpectedResult{Values: &results}
	default:
		panic(errors.New("unsupported Test Case Result type. Should not happen"))
	}
}

func compileTestConfig(
	dmnFile string,
	searchDir string,
	testConfigsCh chan tcktestconfig.TestConfig,
) {
	testCaseDir := strings.TrimPrefix(filepath.Dir(dmnFile), searchDir)
	testCaseFile := strings.TrimSuffix(
		dmnFile,
		filepath.Ext(dmnFile),
	) + testCaseFileSuffix

	tckTestCasesCh := make(chan *tck.TestCases)
	go extractTckTestCases(testCaseFile, tckTestCasesCh)

	dmnDecisionsCh := make(chan []*dmn.Decision)
	go extractDmnDecisions(dmnFile, dmnDecisionsCh)

	tckTestCases, dmnDecisions := <-tckTestCasesCh, <-dmnDecisionsCh

	dmnDecisionsMap := mapDmnDecisionsByName(dmnDecisions)

	testCases := make([]testconfig.TestCase, 0)
	for _, tckTestCase := range tckTestCases.TestCases {
		tests := compileTest(tckTestCase, dmnDecisionsMap)

		testCase := testconfig.TestCase{
			Id:          tckTestCase.IdAttr,
			Description: tckTestCase.Description,
			Tests:       tests,
		}
		testCases = append(testCases, testCase)
	}

	testConfigsCh <- tcktestconfig.TestConfig{
		Model: tcktestconfig.Model{
			Dir:  testCaseDir,
			Name: tckTestCases.ModelName,
		},
		TestCases: testCases,
	}
}

func compileTest(tckTestCase *tck.TestCase, dmnDecisionsMap map[string]*dmn.Decision) []testconfig.Test {
	tests := make([]testconfig.Test, 0)
	for _, tckResult := range tckTestCase.ResultNodes {
		dmnDecision := dmnDecisionsMap[tckResult.NameAttr]

		context := compileContext(dmnDecision.Context)
		expectedResult := compileExpectedResult(tckResult.Expected)
		tests = append(
			tests,
			testconfig.Test{
				Context:        &context,
				FeelExpression: dmnDecision.LiteralExpression,
				ExpectedResult: expectedResult,
			},
		)
	}
	return tests
}

func compileContext(dmnContext *dmn.Context) []testconfig.ContextEntry {
	if dmnContext == nil {
		return nil
	}

	var contextEntries = make([]testconfig.ContextEntry, 0)
	for _, dmnCe := range dmnContext.ContextEntries {
		contextEntries = append(
			contextEntries,
			testconfig.ContextEntry{
				Variable:       dmnCe.Variable.Name,
				FeelExpression: dmnCe.LiteralExpression,
			},
		)
	}

	return contextEntries
}

// Strip namespace prefix from `valueType`, if not `nil`.
// Otherwise, return `nil`.
func stripNamespacePrefix(valueType *string) *string {
	if valueType == nil {
		return nil
	} else {
		tmp := strings.TrimPrefix(*valueType, xmlSchemaNamespacePrefix)
		return &tmp
	}
}

func mapDmnDecisionsByName(tckDecisions []*dmn.Decision) map[string]*dmn.Decision {
	resultsMap := make(map[string]*dmn.Decision)
	for _, dmnDecision := range tckDecisions {
		resultsMap[dmnDecision.Name] = dmnDecision
	}

	return resultsMap
}

func saveTestConfigs(testConfigs []tcktestconfig.TestConfig, outputDir string) {

	for _, testConfig := range testConfigs {
		testConfigFilename := strings.TrimSuffix(
			testConfig.Model.Name,
			filepath.Ext(testConfig.Model.Name),
		) + ".yaml"

		outputFilename := filepath.Join(
			outputDir,
			filepath.Dir(testConfig.Model.Dir), // No need to have dir with one file of same name
			testConfigFilename,
		)
		outputFile := ensureOutputFile(outputFilename)

		yamlData, err := yaml.Marshal([]tcktestconfig.TestConfig{testConfig})
		if err != nil {
			printError("Error marshaling to YAML: %v\n", err)
			os.Exit(1)
		}

		defer func() {
			err := outputFile.Close()
			if err != nil {
				printError("Error closing output file: %v\n", err)
			}
		}()

		_, err = outputFile.Write(yamlData)
		if err != nil {
			printError("Error writing YAML to file: %v\n", err)
			os.Exit(1)
		}
	}
}

func printError(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

func printExtractionStats(
	outputDir string,
	countOfFeelExpressions int,
	countOfFiles int,
	duration time.Duration,
) {
	fmt.Printf(
		"\n"+
			"%d test cases for feel expressions extracted to %s"+
			"\n%d TCK DMN and Test Case files processed"+
			"\nIt took %v"+
			"\n",
		countOfFeelExpressions,
		outputDir,
		countOfFiles,
		duration,
	)
}

func main() {
	startTs := time.Now()

	dir, outputDir, _ := parseFlags()

	searchDir, err := filepath.Abs(*dir)
	if err != nil {
		printError("Error getting search dir absolute path: %v\n", err)
		os.Exit(1)
	}

	modelFiles := fs.FindFiles(searchDir, dmnModelFileWildcard, true)
	testConfigs := compileTestConfigs(modelFiles, searchDir)

	countOfFeelExpressions := 0
	for _, testConfig := range testConfigs {
		countOfFeelExpressions += len(testConfig.TestCases)
	}

	saveTestConfigs(testConfigs, *outputDir)

	duration := time.Since(startTs)

	printExtractionStats(
		*outputDir,
		countOfFeelExpressions,
		len(modelFiles)*2, // *2 because both DMN file and TestCase files are processed
		duration,
	)
}
