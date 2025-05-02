package dmn

type Definitions struct {
	Decisions []*Decision `xml:"decision"`
}

type Decision struct {
	Name string `xml:"name,attr"`
	// TODO[JSot]: Add support for "contextEntry>literalExpression>text"
	LiteralExpression string `xml:"literalExpression>text"`
}
