package tcktestconfig

import (
	"github.com/pbinitiative/feel/cmd/testcase-extractor/model/testconfig"
)

type TestConfig struct {
	Model     Model                 `yaml:"model"`
	TestCases []testconfig.TestCase `yaml:"test-cases"`
}

type Model struct {
	Dir  string `yaml:"dir"`
	Name string `yaml:"name"`
}
