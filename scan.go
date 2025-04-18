package feel

import (
	"errors"
	"regexp"
	"strings"
)

const (
	TokenEOF               = "eof"
	TokenSpace             = "space"
	TokenCommentSingleLine = "comment/singleline"
	TokenCommentMultiline  = "comment/multiline"

	TokenName = "name"
	//TokenFuncall  = "funcall"
	TokenTemporal = "temporal"
	TokenString   = "string"

	TokenKeyword = "keyword"
	TokenNumber  = "number"
)

type tokenMatcher struct {
	token string
	reg   *regexp.Regexp
}

func match(token string, regstr string) tokenMatcher {
	if regstr == "" {
		return tokenMatcher{token: token}
	}
	return tokenMatcher{
		token: token,
		reg:   regexp.MustCompile("^" + regstr),
	}
}

var tokenMatchers = []tokenMatcher{
	match(TokenSpace, `\s+`),
	match(TokenCommentSingleLine, `//.*\n`),
	match(TokenCommentMultiline, `\/\*(.|\n)*\*\/`),

	match(TokenKeyword, `\b(true|false|and|or|null|function|if|then|else|loop|for|some|every|in|return|satisfies)\b`),

	match(TokenTemporal, `@"(\\.|[^"])*"`),
	match(TokenString, `"(\\.|[^"])*"`),
	match(TokenNumber, `\-?[0-9]+(\.[0-9]+)?\b`),

	match("?", ""),
	match("..", ""),
	match(".", ""),
	match(",", ""),
	match(";", ""),

	match(">=", ""),
	match(">", ""),

	match("<=", ""),
	match("<", ""),

	match("!=", ""),
	match("!", ""),

	match("=", ""),

	match("(", ""),
	match(")", ""),
	match("[", ""),
	match("]", ""),
	match("{", ""),
	match("}", ""),

	match(":=", ""),
	match(":", ""),

	match("+", ""),
	match("-", ""),
	match("*", ""),
	match("/", ""),
	match("%", ""),

	// variable name support unicode chars, currently Han and Greek is in the list
	// refer to https://github.com/google/re2/wiki/Syntax
	match(TokenName, `[a-zA-Z_\$\p{Han}\p{Greek}\p{Bopomofo}\p{Hangul}][a-zA-Z_\$0-9\p{Han}\p{Greek}}\p{Bopomofo}\p{Hangul}]*`),
}

type ScanPosition struct {
	Row    int
	Column int
}

type ScannerToken struct {
	Kind  string
	Value string
	Pos   ScanPosition
}

var opTokens = map[string]bool{
	"+": true,
	"-": true,
	"*": true,
	"/": true,
	"%": true,
}

func (token ScannerToken) IsOp() bool {
	if isOp, ok := opTokens[token.Kind]; ok {
		return isOp
	}
	return false
}

func (token ScannerToken) Expect(tokenKinds ...string) bool {
	for _, kind := range tokenKinds {
		if token.Kind == kind {
			return true
		}
	}
	return false
}

func (token ScannerToken) ExpectKeywords(words ...string) bool {
	if token.Kind != TokenKeyword {
		return false
	}
	for _, kw := range words {
		if token.Value == kw {
			return true
		}
	}
	return false
}

type Scanner struct {
	input string
	rest  string

	currentToken ScannerToken

	Pos   ScanPosition
	Eaten int
}

func NewScanner(input string) *Scanner {
	return &Scanner{
		input: input,
		rest:  input,
	}
}

// Tokens Find all tokens
func (scanner *Scanner) Tokens() ([]ScannerToken, error) {
	err := scanner.Next()
	if err != nil {
		return nil, err
	}
	var arr []ScannerToken
	for scanner.currentToken.Kind != TokenEOF {
		// fmt.Printf("scanner current token %q, rest %q\n", scanner.currentToken.Kind, scanner.rest)
		arr = append(arr, scanner.currentToken)
		err := scanner.Next()
		if err != nil {
			return nil, err
		}
	}
	return arr, nil
}

func (scanner Scanner) Current() ScannerToken {
	return scanner.currentToken
}

func (scanner *Scanner) goAhead(matched string) {
	// find \n
	lastIndexOfLn := strings.LastIndex(matched, "\n")
	if lastIndexOfLn >= 0 {
		// found \n
		scanner.Pos.Column = len(matched) - lastIndexOfLn - 1
		scanner.Pos.Row += strings.Count(matched, "\n")
	} else {
		scanner.Pos.Column += len(matched)
	}
	scanner.rest = scanner.rest[len(matched):]
	scanner.Eaten += len(matched)
}

func (scanner *Scanner) Next() error {
	if scanner.currentToken.Kind == TokenEOF {
		return errors.New("EOF met")
	} else if scanner.rest == "" {
		scanner.currentToken = ScannerToken{Kind: TokenEOF, Pos: scanner.Pos}
		return nil
	}

	for _, matcher := range tokenMatchers {
		matched := ""
		if matcher.reg == nil {
			if strings.HasPrefix(scanner.rest, matcher.token) {
				matched = matcher.token
			}
		} else {
			matched = matcher.reg.FindString(scanner.rest)
		}
		if matched != "" {
			// find a matching
			if matcher.token == TokenSpace || matcher.token == TokenCommentMultiline || matcher.token == TokenCommentSingleLine {
				scanner.goAhead(matched)
				return scanner.Next()
			} else {
				scanner.currentToken = ScannerToken{
					Kind:  matcher.token,
					Value: matched,
					Pos:   scanner.Pos,
				}
				scanner.goAhead(matched)
				return nil
			}
		}
	}

	if scanner.rest == "" {
		scanner.currentToken = ScannerToken{Kind: TokenEOF, Pos: scanner.Pos}
		return nil
	} else {
		return NewErrBadInput(scanner.Pos.Row, scanner.Pos.Column, scanner.rest)
	}
}
