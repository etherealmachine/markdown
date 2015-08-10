package markdown

import (
	"reflect"
	"testing"
)

type scannerCase struct {
	input string
	want  []Token
}

var scannerCases = []*scannerCase{
	{"# Some header", []Token{
		H1, TEXT,
	}},
	{"## Some header", []Token{
		H2, TEXT,
	}},
	{"# Non-unix line endings\r\nMore stuff\n", []Token{
		H1, TEXT, NEWLINE, TEXT, NEWLINE,
	}},
	{"[a link](www.example.com)", []Token{
		LINK_TEXT, HREF,
	}},
	{"Text with *emphasis*", []Token{
		TEXT, EM,
	}},
	{"# [A link in a header](www.example.com)", []Token{
		H1, LINK_TEXT, HREF,
	}},
	{"![An img](www.example.com)", []Token{
		IMG_ALT, HREF,
	}},
	{"A paragraph of text\npossibly on multiple\nlines.", []Token{
		TEXT, NEWLINE, TEXT, NEWLINE, TEXT,
	}},
	{`<a href="www.example.com" rel="nofollow">Some HTML</a>`, []Token{
		HTML_TAG, TEXT, HTML_TAG,
	}},
	{"`some code`", []Token{
		CODE,
	}},
	{"# A more complicated\nExample with elements [mixed](www.example.com)\nin **together**.\n\nAnd multiple paragraphs.", []Token{
		H1, TEXT, NEWLINE, TEXT, LINK_TEXT, HREF, NEWLINE, TEXT, STRONG, TEXT, NEWLINE, NEWLINE, TEXT,
	}},
	{"```javascript\nA block of code\n```", []Token{
		CODE_BLOCK, TEXT, NEWLINE, TEXT, NEWLINE, CODE_BLOCK,
	}},
	{"* One\n* Two\n* *Three* items\n", []Token{
		UNORDERED_LIST, TEXT,
		UNORDERED_LIST, TEXT,
		UNORDERED_LIST, EM, TEXT, NEWLINE,
	}},
	{"1. One\n2. Two\n3. Three\n", []Token{
		ORDERED_LIST, TEXT,
		ORDERED_LIST, TEXT,
		ORDERED_LIST, TEXT, NEWLINE,
	}},
	{"\n\n1. One\n2. Two\n3. Three\n", []Token{
		ORDERED_LIST, TEXT,
		ORDERED_LIST, TEXT,
		ORDERED_LIST, TEXT, NEWLINE,
	}},
}

func TestScanner(t *testing.T) {
	for _, c := range scannerCases {
		scanner := NewScanner(c.input)
		var got []Token
		for tok, _, _ := scanner.Next().Tuple(); tok != EOF; tok, _, _ = scanner.Next().Tuple() {
			got = append(got, tok)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("\n%s\ngot:\n%v\nwant:\n%v", c.input, got, c.want)
		}
	}
}
