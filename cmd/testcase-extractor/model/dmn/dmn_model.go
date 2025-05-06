package dmn

type Definitions struct {
	Decisions []*Decision `xml:"decision"`
}

type Decision struct {
	Name              string    `xml:"name,attr"`
	Context           *Context  `xml:"context"`
	Variable          *Variable `xml:"variable,omitempty"`
	LiteralExpression *string   `xml:"literalExpression>text,omitempty"`
}

type Context struct {
	ContextEntries []ContextEntry `xml:"contextEntry"`
}

type ContextEntry struct {
	Variable          Variable `xml:"variable"`
	LiteralExpression string   `xml:"literalExpression>text"`
}

type Variable struct {
	Name string `xml:"name,attr"`
}
