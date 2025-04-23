package feel

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleScannerFinding(t *testing.T) {
	input := `
	< alice + bob 
	// single line comment
	/** multi
	line
	comment
	 */

	+ "op\nq" * true ( haha )

+ "multiline
	string中文"

	中文变量andγβ
	`

	scanner := NewScanner(input)
	tokens, err := scanner.Tokens()
	assert.NoError(t, err)
	// tokens is ["<", "name", "+", "name"]
	assert.Equal(t, "<", tokens[0].Kind)

	assert.Equal(t, "name", tokens[1].Kind)
	assert.Equal(t, "alice", tokens[1].Value)

	assert.Equal(t, "+", tokens[2].Kind)

	assert.Equal(t, "name", tokens[3].Kind)
	assert.Equal(t, "bob", tokens[3].Value)

	assert.Equal(t, "+", tokens[4].Kind)

	assert.Equal(t, "string", tokens[5].Kind)
	assert.Equal(t, `"op\nq"`, tokens[5].Value)
	assert.Equal(t, 7, len(tokens[5].Value))

	assert.Equal(t, "*", tokens[6].Kind)

	assert.Equal(t, "keyword", tokens[7].Kind)
	assert.Equal(t, "true", tokens[7].Value)

	assert.Equal(t, "(", tokens[8].Kind)

	assert.Equal(t, "name", tokens[9].Kind)
	assert.Equal(t, "haha", tokens[9].Value)

	assert.Equal(t, ")", tokens[10].Kind)

	assert.Equal(t, "string", tokens[12].Kind)
	assert.Equal(t, "\"multiline\n\tstring中文\"", tokens[12].Value)

	assert.Equal(t, "name", tokens[13].Kind)
	assert.Equal(t, "中文变量andγβ", tokens[13].Value)

	// test against position
	assert.Equal(t, 14, scanner.Pos.Row)
	assert.Equal(t, 1, scanner.Pos.Column)
}

func TestUnicodeRegexp(t *testing.T) {
	reName := regexp.MustCompile(`\p{Han}+`)

	found := reName.FindString("abc(汉字)")
	assert.Equal(t, "汉字", found)
}
