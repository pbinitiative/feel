package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type TckDefinitions struct {
	XMLName   xml.Name       `xml:"definitions"`
	Decisions []*TckDecision `xml:"decision"`
}

type TckDecision struct {
	XMLName           xml.Name `xml:"decision"`
	Name              *string  `xml:"name,attr"`
	LiteralExpression *string  `xml:"literalExpression>text"`
}

type TckTestCases struct {
	XMLName   xml.Name       `xml:"testCases"`
	ModelName *string        `xml:"modelName"`
	TestCases []*TckTestCase `xml:"testCase"`
}

type TckTestCase struct {
	XMLName     xml.Name             `xml:"testCase"`
	Id          *string              `xml:"id,attr"`
	Description *string              `xml:"description"`
	Results     []*TckTestCaseResult `xml:"resultNode"`
}

type TckTestCaseResult struct {
	XMLName  xml.Name      `xml:"resultNode"`
	Name     *string       `xml:"name,attr"`
	Expected NilableString `xml:"expected>value"`
}

type NilableString struct {
	Value *string
}

func (ns *NilableString) String() string {
	return *ns.Value
}

func (ns *NilableString) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type StringNode struct {
		Value string `xml:",chardata"`
		Nil   bool   `xml:"http://www.w3.org/2001/XMLSchema-instance nil,attr"`
	}
	var stringNode StringNode
	if err := d.DecodeElement(&stringNode, &start); err != nil {
		return err
	}
	if stringNode.Nil {
		ns.Value = nil
	} else {
		if stringNode.Value != "" {
			ns.Value = &stringNode.Value
		} else {
			ns.Value = nil // Treat empty string as nil
		}
	}
	return nil
}

type TestConfig struct {
	Model     TestConfigModel `yaml:"model"`
	TestCases []TestCase      `yaml:"test-cases"`
}

type TestConfigModel struct {
	Name *string `yaml:"name"`
	Dir  *string `yaml:"dir"`
}

type TestCase struct {
	Id          *string          `yaml:"id"`
	Description *string          `yaml:"description"`
	Results     []TestCaseResult `yaml:"results"`
}

type TestCaseResult struct {
	FeelExpression *string `yaml:"feel-expression"`
	Expected       *string `yaml:"expected"`
}

const (
	DmnModelFileWildcard = "*-feel-*.dmn"
	TestCaseFileSuffix   = "-test-01.xml"
)

func parseFlags() (
	dir *string,
	outputFilename *string,
	forceStats *bool,
) {
	dir = flag.String(
		"dir",
		".",
		"Directory to start search for DMN model files and Test Cases files.",
	)
	outputFilename = flag.String(
		"output-file",
		"-",
		`Name of file to save output to. Default is "-" (STDOUT).`,
	)
	forceStats = flag.Bool(
		"force-stats",
		false,
		"Force statistics report. Useful when writing to STDOUT and want statistics.",
	)
	flag.Parse()

	return
}

func resolveOutputFile(outputFilename string) *os.File {
	var outputFile *os.File
	var err error

	if outputFilename == "-" {
		outputFile = os.Stdout
	} else {
		outputFile, err = os.Create(outputFilename)
		if err != nil {
			printError("Error getting output filename absolute path: %v\n", err)
			panic(err)
		}
	}

	return outputFile
}

// findFiles traverses the directory tree starting at dir and returns files
// matching the wildcard pattern
func findFiles(dir, wildcard string) []string {
	var filePaths []string

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		printError("Error: Directory '%s' does not exist\n", dir)
		os.Exit(1)
	}

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
		printError("Error: %v\n", err)
		os.Exit(1)
	}

	return filePaths
}

func extractTckTestCases(testCasesFilePath string, tckTestCasesCh chan *TckTestCases) {
	defer close(tckTestCasesCh)

	bytes, err := os.ReadFile(testCasesFilePath)
	if err != nil {
		printError("Error: could not read Test Cases file: %v\n", err)
	}

	var testCases *TckTestCases
	err = xml.Unmarshal(bytes, &testCases)
	if err != nil {
		printError("Error: could not read Test Cases file: %v\n", err)
	}

	tckTestCasesCh <- testCases
}

func extractTckDecisions(decisionsFilePath string, tckDecisionsCh chan []*TckDecision) {
	defer close(tckDecisionsCh)

	bytes, err := os.ReadFile(decisionsFilePath)
	if err != nil {
		printError("Error: could not read Test Cases file: %v\n", err)
	}

	var definitions *TckDefinitions
	err = xml.Unmarshal(bytes, &definitions)
	if err != nil {
		printError("Error: could not read Test Cases file: %v\n", err)
	}

	tckDecisionsCh <- definitions.Decisions
}

func compileTestConfigs(dmnFiles []string, searchDir string) []TestConfig {
	testConfigsCh := make(chan TestConfig)

	for _, dmnFile := range dmnFiles {
		go compileTestConfig(dmnFile, searchDir, testConfigsCh)
	}

	testConfigs := make([]TestConfig, 0)
	for _ = range len(dmnFiles) {
		testConfigs = append(testConfigs, <-testConfigsCh)
	}

	sort.Slice(testConfigs, func(i, j int) bool {
		return *testConfigs[i].Model.Name < *testConfigs[j].Model.Name
	})

	return testConfigs
}

func compileTestConfig(
	dmnFile string,
	searchDir string,
	testConfigsCh chan TestConfig,
) {
	testCaseDir := strings.TrimPrefix(filepath.Dir(dmnFile), searchDir)
	testCaseFile := strings.TrimSuffix(
		dmnFile,
		filepath.Ext(dmnFile),
	) + TestCaseFileSuffix

	tckTestCasesCh := make(chan *TckTestCases)
	go extractTckTestCases(testCaseFile, tckTestCasesCh)

	tckDecisionsCh := make(chan []*TckDecision)
	go extractTckDecisions(dmnFile, tckDecisionsCh)

	tckTestCases, tckDecisions := <-tckTestCasesCh, <-tckDecisionsCh

	tckDecisionsMap := mapTckDecisionsByName(tckDecisions)

	testCases := make([]TestCase, 0)
	for _, tckTestCase := range tckTestCases.TestCases {
		results := make([]TestCaseResult, 0)
		for _, tckResult := range tckTestCase.Results {
			tckDecision := tckDecisionsMap[*tckResult.Name]
			result := TestCaseResult{
				FeelExpression: tckDecision.LiteralExpression,
				Expected:       tckResult.Expected.Value,
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

	testConfigsCh <- TestConfig{
		Model: TestConfigModel{
			tckTestCases.ModelName,
			&testCaseDir,
		},
		TestCases: testCases,
	}
}

func mapTckDecisionsByName(tckDecisions []*TckDecision) map[string]*TckDecision {
	resultsMap := make(map[string]*TckDecision)
	for _, tckDecision := range tckDecisions {
		resultsMap[*tckDecision.Name] = tckDecision
	}

	return resultsMap
}

func saveTestConfigs(testConfigs []TestConfig, outputFile *os.File) {
	yamlData, err := yaml.Marshal(&testConfigs)
	if err != nil {
		printError("Error marshaling to YAML: %v\n", err)
		os.Exit(1)
	}

	_, err = outputFile.Write(yamlData)
	if err != nil {
		printError("Error writing YAML file: %v\n", err)
		os.Exit(1)
	}
}

func printError(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

func isOutputPiped() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		printError("Error checking pipe: %v\n", err)
	}
	// Check if the file mode indicates a named pipe (FIFO)
	return (fileInfo.Mode() & os.ModeNamedPipe) != 0
}

func printExtractionStats(
	outputFile *os.File,
	countOfFeelExpressions int,
	countOfFiles int,
	duration time.Duration,
) {

	outputFileInfo := ""
	if outputFile != os.Stdout {
		outputFilenameAbs, _ := filepath.Abs(outputFile.Name())
		outputFileInfo = fmt.Sprintf(
			" to file "+
				"\n\t%v",
			outputFilenameAbs,
		)
	}

	fmt.Printf(
		"\n"+
			"%d test cases for feel expressions extracted"+
			"%v"+
			"\n%d TCK DMN and Test Case files processed"+
			"\nIt took %v"+
			"\n",
		countOfFeelExpressions,
		outputFileInfo,
		countOfFiles,
		duration,
	)
}

func main() {
	startTs := time.Now()

	dir, outputFilename, forceStats := parseFlags()

	searchDir, err := filepath.Abs(*dir)
	if err != nil {
		printError("Error getting search dir absolute path: %v\n", err)
		panic(err)
	}

	outputFile := resolveOutputFile(*outputFilename)

	defer func(outputFile *os.File) {
		err := outputFile.Close()
		if err != nil {
			printError("Error closing output file: %v\n", err)
		}
	}(outputFile)

	modelFiles := findFiles(searchDir, DmnModelFileWildcard)
	testConfigs := compileTestConfigs(modelFiles, searchDir)

	countOfFeelExpressions := 0
	for _, testConfig := range testConfigs {
		countOfFeelExpressions += len(testConfig.TestCases)
	}

	saveTestConfigs(testConfigs, outputFile)

	duration := time.Since(startTs)

	if !isOutputPiped() || *forceStats {
		printExtractionStats(
			outputFile,
			countOfFeelExpressions,
			len(modelFiles)*2, // *2 because both DMN file and TestCase files are processed
			duration,
		)
	}
}
