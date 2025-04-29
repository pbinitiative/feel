package testconfig

import (
	"fmt"
	"strings"
)

type TestConfig struct {
	Model     Model      `yaml:"model"`
	TestCases []TestCase `yaml:"test-cases"`
}

type Model struct {
	Dir  string `yaml:"dir"`
	Name string `yaml:"name"`
}

type TestCase struct {
	Id          string `yaml:"id"`
	Description string `yaml:"description"`
	Tests       []Test `yaml:"tests"`
}

type Test struct {
	FeelExpression string `yaml:"feel-expression"`
	ExpectedResult `yaml:"expected"`
}

type ExpectedResult struct {
	Components *[]Component      `yaml:"components,omitempty"`
	Value      *ExpectedValue    `yaml:"result,omitempty"`
	Values     *[]ExpectedResult `yaml:"results,omitempty"`
}

func joinStrings(vals []string) string {
	return strings.Join(vals, ", ")
}

func (er ExpectedResult) String() string {
	switch {
	case er.Components != nil:
		componentsStrings := make([]string, 0)
		for _, component := range *er.Components {
			componentsStrings = append(componentsStrings, component.String())
		}

		return fmt.Sprintf("components: {%s}", joinStrings(componentsStrings))
	case er.Value != nil:
		if er.Value.Value == nil {
			return fmt.Sprintf("value: nil")
		}

		var typeString string
		if er.Value.Type == nil {
			typeString = "nil"
		} else {
			typeString = *er.Value.Type
		}

		return fmt.Sprintf("%s(%s)", typeString, *er.Value.Value)
	case er.Values != nil:
		valuesStrings := make([]string, 0)
		for _, value := range *er.Values {
			valuesStrings = append(valuesStrings, value.String())
		}
		return fmt.Sprintf("values: {%s}", joinStrings(valuesStrings))
	default:
		panic("Unsupported 'ExpectedResult' type. Supported types are 'Components', 'Value' and 'Values'")
	}
}

type Component struct {
	Name           string `yaml:"name"`
	ExpectedResult `yaml:"expected"`
}

func (c Component) String() string {
	return fmt.Sprintf("%s: %s", c.Name, c.ExpectedResult.String())
}

type ExpectedValue struct {
	Value *string `yaml:"value"`
	Type  *string `yaml:"type,omitempty"`
}
