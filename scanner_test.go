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
	{"# Non-unix line endings\r\nMore stuff\n\r", []Token{
		H1, TEXT, NEWLINE, TEXT, NEWLINE,
	}},
	{"[a link](www.example.com)", []Token{
		LINK_TEXT, HREF,
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
		HTML_START, TEXT, HTML_END, TEXT, HTML_END_TAG,
	}},
	{"`some code`", []Token{
		CODE, TEXT, CODE,
	}},
	{"# A more complicated\nExample with elements [mixed](www.example.com)\nin **together**.\n\nAnd multiple paragraphs.", []Token{
		H1, TEXT, NEWLINE, TEXT, LINK_TEXT, HREF, NEWLINE, TEXT, STRONG, TEXT, STRONG, TEXT, NEWLINE, NEWLINE, TEXT,
	}},
	{"```javascript\nA block of code\n```", []Token{
		CODE_BLOCK, TEXT, NEWLINE, TEXT, NEWLINE, CODE_BLOCK,
	}},
	{"* One\n* Two\n* Three\n", []Token{
		UNORDERED_LIST, TEXT, NEWLINE,
		UNORDERED_LIST, TEXT, NEWLINE,
		UNORDERED_LIST, TEXT, NEWLINE,
	}},
	{"1. One\n2. Two\n3. Three\n", []Token{
		ORDERED_LIST, TEXT, NEWLINE,
		ORDERED_LIST, TEXT, NEWLINE,
		ORDERED_LIST, TEXT, NEWLINE,
	}},
	/*
		{"[[define panel]]\n[[title]]\n[[body]]\n[[end]]\n", []Token{
			TMPL_DEF, NEWLINE, TMPL_VAL, NEWLINE, TMPL_VAL, TMPL_END,
		}},
	*/
}

func TestScanner(t *testing.T) {
	for _, c := range scannerCases {
		scanner := NewScanner(c.input)
		var got []Token
		for tok, _ := scanner.Next(); tok != EOF; tok, _ = scanner.Next() {
			got = append(got, tok)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("got:\n%v\nwant:\n%v", got, c.want)
		}
	}
}
