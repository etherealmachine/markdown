package markdown

import (
	"reflect"
	"testing"
)

type scannerCase struct {
	input string
	want  []TokenType
}

var scannerCases = []*scannerCase{
	{"# Some header", []TokenType{
		H1, TEXT,
	}},
	{"## Some header", []TokenType{
		H2, TEXT,
	}},
	{"# Non-unix line endings\r\nMore stuff\n", []TokenType{
		H1, TEXT, NEWLINE, TEXT, NEWLINE,
	}},
	{"[a link](www.example.com)", []TokenType{
		LINK_TEXT, HREF,
	}},
	{"Text with *emphasis*", []TokenType{
		TEXT, EM,
	}},
	{"# [A link in a header](www.example.com)", []TokenType{
		H1, LINK_TEXT, HREF,
	}},
	{"![An img](www.example.com)", []TokenType{
		IMG_ALT, HREF,
	}},
	{"A paragraph of text\npossibly on multiple\nlines.", []TokenType{
		TEXT, NEWLINE, TEXT, NEWLINE, TEXT,
	}},
	{`<a href="www.example.com" rel="nofollow">Some HTML</a>`, []TokenType{
		HTML_TAG, TEXT, HTML_TAG,
	}},
	{"`some code`", []TokenType{
		CODE,
	}},
	{"# A more complicated\nExample with elements [mixed](www.example.com)\nin **together**.\n\nAnd multiple paragraphs.", []TokenType{
		H1, TEXT, NEWLINE, TEXT, LINK_TEXT, HREF, NEWLINE, TEXT, STRONG, TEXT, NEWLINE, NEWLINE, TEXT,
	}},
	{"```javascript\nA block of code\n```", []TokenType{
		CODE_BLOCK, TEXT, NEWLINE, TEXT, NEWLINE, CODE_BLOCK,
	}},
	{"* One\n* Two\n* *Three* items\n", []TokenType{
		UNORDERED_LIST, TEXT,
		UNORDERED_LIST, TEXT,
		UNORDERED_LIST, EM, TEXT, NEWLINE,
	}},
	{"1. One\n2. Two\n3. Three\n", []TokenType{
		ORDERED_LIST, TEXT,
		ORDERED_LIST, TEXT,
		ORDERED_LIST, TEXT, NEWLINE,
	}},
	{"\n\n1. One\n2. Two\n3. Three\n", []TokenType{
		ORDERED_LIST, TEXT,
		ORDERED_LIST, TEXT,
		ORDERED_LIST, TEXT, NEWLINE,
	}},
	{"Col1 | Col2 | Col3\n-|-|-\nA | B | C\nD | *E* | F\n", []TokenType{
		TEXT, TD, TEXT, TD, TEXT, NEWLINE,
		TEXT, TD, TEXT, TD, TEXT, NEWLINE,
		TEXT, TD, TEXT, TD, TEXT, NEWLINE,
		TEXT, TD, EM, TD, TEXT, NEWLINE,
	}},
}

func TestScanner(t *testing.T) {
	for _, c := range scannerCases {
		scanner := NewScanner(c.input)
		var got []TokenType
		for tok := scanner.Next().Type; tok != EOF; tok = scanner.Next().Type {
			got = append(got, tok)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("\n%s\ngot:\n%v\nwant:\n%v", c.input, got, c.want)
		}
	}
}
