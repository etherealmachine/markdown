package markdown

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"reflect"
	"testing"
)

type fakeScanner struct {
	pos  int
	toks []Token
}

func (s *fakeScanner) Next() *Token {
	if s.pos >= len(s.toks) {
		return &Token{EOF, "EOF", ""}
	}
	s.pos++
	return &s.toks[s.pos-1]
}

type parserCase struct {
	input []Token
	want  []*html.Token
}

var parserCases = []*parserCase{
	{
		[]Token{
			{H1, "# ", "# "}, {TEXT, "Some header", "Some header"}, {NEWLINE, "\n", "\n"},
		},
		[]*html.Token{
			startH1, text("Some header"), endH1,
		},
	},
	{
		[]Token{
			{H2, "## ", "## "}, {TEXT, "Some header", "Some header"}, {NEWLINE, "\n", "\n"},
		},
		[]*html.Token{
			startH2, text("Some header"), endH2,
		},
	},
	{
		[]Token{
			{LINK_TEXT, "A link", "[A link]"}, {HREF, "www.example.com", "(www.example.com)"},
		},
		[]*html.Token{
			startP,
			{
				Type:     html.StartTagToken,
				DataAtom: atom.A,
				Data:     "a",
				Attr: []html.Attribute{{
					Key: "href",
					Val: "www.example.com",
				}},
			},
			text("A link"),
			endA,
			endP,
		},
	},
	{
		[]Token{
			{IMG_ALT, "An image", "![An image]"}, {HREF, "www.example.com", "(www.example.com)"},
		},
		[]*html.Token{
			startP,
			{
				Type:     html.SelfClosingTagToken,
				DataAtom: atom.Img,
				Data:     "img",
				Attr: []html.Attribute{
					{
						Key: "alt",
						Val: "An image",
					},
					{
						Key: "src",
						Val: "www.example.com",
					},
				},
			},
			endP,
		},
	},
	{
		[]Token{
			{CODE, "Some code", "`Some code`"},
		},
		[]*html.Token{
			startP, startCode, text("Some code"), endCode, endP,
		},
	},
	{
		[]Token{
			{CODE_BLOCK, "```", "```"}, {NEWLINE, "\n", "\n"}, {TEXT, "Some code ", "Some code "}, {EM, "*", "*"}, {CODE_BLOCK, "```", "```"},
		},
		[]*html.Token{
			startPre, startCode, text("\nSome code *\n"), endCode, endPre,
		},
	},
	{
		[]Token{
			{TEXT, "foo", "foo"}, {NEWLINE, "\n", "\n"}, {TEXT, "bar", "bar"},
		},
		[]*html.Token{
			startP, text("foo"), text("\n"), text("bar"), endP,
		},
	},
	{
		[]Token{
			{TEXT, "foo", "foo"}, {NEWLINE, "\n", "\n"}, {TEXT, "bar", "bar"}, {NEWLINE, "\n", "\n"}, {NEWLINE, "\n", "\n"}, {TEXT, "baz", "bar"}, {CODE, "bang", "`bang`"},
		},
		[]*html.Token{
			startP, text("foo"), text("\n"), text("bar"), endP,
			startP, text("baz"), startCode, text("bang"), endCode, endP,
		},
	},
	{
		[]Token{
			{STRONG, "foo", "**foo**"},
		},
		[]*html.Token{
			startP, startStrong, text("foo"), endStrong, endP,
		},
	},
	{
		[]Token{
			{EM, "foo", "*foo*"},
		},
		[]*html.Token{
			startP, startEm, text("foo"), endEm, endP,
		},
	},
	{
		[]Token{
			{H2, "##", "## "}, {TEXT, "header", "header"},
			{UNORDERED_LIST, "", "* "}, {TEXT, "foo", "foo"},
			{UNORDERED_LIST, "", "* "}, {TEXT, "bar", "bar"},
			{UNORDERED_LIST, "", "* "}, {TEXT, "baz", "baz"}, {NEWLINE, "\n", "\n"},
		},
		[]*html.Token{
			startH2, text("header"), endH2,
			startUl,
			startLi, text("foo"), endLi,
			startLi, text("bar"), endLi,
			startLi, text("baz"), endLi,
			endUl,
		},
	},
	{
		[]Token{
			{ORDERED_LIST, "", "1. "}, {TEXT, "foo", "foo"},
			{ORDERED_LIST, "", "2. "}, {TEXT, "bar", "bar"},
			{ORDERED_LIST, "", "2. "}, {TEXT, "baz", "baz"}, {NEWLINE, "\n", "\n"},
		},
		[]*html.Token{
			startOl,
			startLi, text("foo"), endLi,
			startLi, text("bar"), endLi,
			startLi, text("baz"), endLi,
			endOl,
		},
	},
	{[]Token{
		{TD, "Col1", "Col1 |"}, {TD, "Col2", "Col2 |"}, {TD, "Col3", "Col3"}, {NEWLINE, "\n", "\n"},
		{TD, ":-", "-|"}, {TD, ":-:", "-"}, {TD, "-:", "-"}, {NEWLINE, "\n", "\n"},
		{TD, "A", "A |"}, {TD, "B", "B"}, {TD, "F", "F"}, {NEWLINE, "\n", "\n"},
		{TD, "C", "C |"}, {TD, "D *E*", "D"}, {TD, "G", "G"},
	},
		[]*html.Token{
			startTable,
			startTr,
			startTh,
			text("Col1"),
			endTh,
			startTh,
			text("Col2"),
			endTh,
			startTh,
			text("Col3"),
			endTh,
			endTr,
			startTr,
			startTdL,
			text("A"),
			endTd,
			startTdC,
			text("B"),
			endTd,
			startTdR,
			text("F"),
			endTd,
			endTr,
			startTr,
			startTdL,
			text("C"),
			endTd,
			startTdC,
			text("D "),
			startEm,
			text("E"),
			endEm,
			endTd,
			startTdR,
			text("G"),
			endTd,
			endTr,
			endTable,
		},
	},
	{
		[]Token{
			{HTML_TAG, "<!--comment-->", "<!--comment-->"},
		},
		[]*html.Token{
			{Type: html.CommentToken, Data: "comment"},
		},
	},
	{
		[]Token{
			{HREF, "4.1", "(4.1)"},
			{TEXT, " ", " "},
			{MATHML, "$F = ma$", "$F = ma$"},
		},
		[]*html.Token{
			startP,
			{Type: html.TextToken, Data: "(4.1)"},
			{Type: html.TextToken, Data: " "},
			{Type: html.TextToken, Data: "$F = ma$"},
			endP,
		},
	},
	{
		[]Token{
			{HTML_TAG, "<div>", "<div"},
			{NEWLINE, "\n", "\n"},
			{TEXT, "  ", "  "},
			{HTML_TAG, "</div>", "</div>"},
		},
		[]*html.Token{
			{Type: html.StartTagToken, DataAtom: atom.Div, Data: "div"},
			{Type: html.EndTagToken, DataAtom: atom.Div, Data: "div"},
		},
	},
}

func TestParser(t *testing.T) {
	for _, c := range parserCases {
		p := &Parser{}
		p.parse(&fakeScanner{toks: c.input})
		got := p.tokens
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("got:\n%v\nwant:\n%v", got, c.want)
		}
	}
}
