package tck

import (
	"encoding/xml"
)

// TestCaseType ...
type TestCaseType string

// Labels ...
type Labels struct {
	XMLName xml.Name `xml:"labels"`
	Labels  []string `xml:"label"`
}

// InputNode ...
type InputNode struct {
	XMLName       xml.Name `xml:"inputNode"`
	NameAttr      string   `xml:"name,attr"`
	NamespaceAttr string   `xml:"namespace,attr,omitempty"`
	*ValueType
}

// ResultNode ...
type ResultNode struct {
	XMLName         xml.Name   `xml:"resultNode"`
	ErrorResultAttr bool       `xml:"errorResult,attr,omitempty"`
	NameAttr        string     `xml:"name,attr"`
	NamespaceAttr   string     `xml:"namespace,attr,omitempty"`
	TypeAttr        string     `xml:"type,attr,omitempty"`
	CastAttr        string     `xml:"cast,attr,omitempty"`
	Computed        *ValueType `xml:"computed"`
	Expected        *ValueType `xml:"expected"`
}

// ExtensionElements ...
type ExtensionElements struct {
	XMLName xml.Name `xml:"extensionElements"`
}

// TestCase ...
type TestCase struct {
	XMLName           xml.Name           `xml:"testCase"`
	IdAttr            string             `xml:"id,attr,omitempty"`
	TypeAttr          string             `xml:"type,attr,omitempty"`
	InvocableNameAttr string             `xml:"invocableName,attr,omitempty"`
	NameAttr          string             `xml:"name,attr,omitempty"`
	NamespaceAttr     string             `xml:"namespace,attr,omitempty"`
	Description       string             `xml:"description"`
	InputNodes        []*InputNode       `xml:"inputNode"`
	ResultNodes       []*ResultNode      `xml:"resultNode"`
	ExtensionElements *ExtensionElements `xml:"extensionElements"`
}

// TestCases ...
type TestCases struct {
	XMLName       xml.Name    `xml:"testCases"`
	NamespaceAttr string      `xml:"namespace,attr,omitempty"`
	ModelName     string      `xml:"modelName"`
	Labels        *Labels     `xml:"labels"`
	TestCases     []*TestCase `xml:"testCase"`
}

// Component ...
type Component struct {
	XMLName  xml.Name `xml:"component"`
	NameAttr string   `xml:"name,attr,omitempty"`
	*ValueType
}

// List ...
type List struct {
	XMLName xml.Name     `xml:"list"`
	Items   []*ValueType `xml:"item"`
}

// ValueType ...
type ValueType struct {
	//XMLName           xml.Name           `xml:"valueType"`
	Value             *AnySimpleType     `xml:"value"`
	Components        *[]Component       `xml:"component"`
	List              *List              `xml:"list"`
	ExtensionElements *ExtensionElements `xml:"extensionElements"`
}

type AnySimpleType struct {
	Content *string
	Type    *string
}

func (simpleType *AnySimpleType) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type SimpleTypeNode struct {
		Value string  `xml:",chardata"`
		Type  *string `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
		Nil   bool    `xml:"http://www.w3.org/2001/XMLSchema-instance nil,attr"`
	}
	var simpleTypeNode SimpleTypeNode
	if err := d.DecodeElement(&simpleTypeNode, &start); err != nil {
		return err
	}
	if simpleTypeNode.Nil {
		simpleType.Content = nil // Treat empty string as nil
	} else {
		simpleType.Content = &simpleTypeNode.Value
	}
	simpleType.Type = simpleTypeNode.Type
	return nil
}
