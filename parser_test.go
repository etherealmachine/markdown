package markdown

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"reflect"
	"testing"
)

type fakeScanner struct {
	pos  int
	toks []Tok
}

func (s *fakeScanner) Next() *Tok {
	if s.pos >= len(s.toks) {
		return &Tok{EOF, "EOF", ""}
	}
	s.pos++
	return &s.toks[s.pos-1]
}

type parserCase struct {
	input []Tok
	want  []*html.Token
}

var parserCases = []*parserCase{
	{
		[]Tok{
			{H1, "# ", "# "}, {TEXT, "Some header", "Some header"}, {NEWLINE, "\n", "\n"},
		},
		[]*html.Token{
			startH1, text("Some header"), endH1,
		},
	},
	{
		[]Tok{
			{H2, "## ", "## "}, {TEXT, "Some header", "Some header"}, {NEWLINE, "\n", "\n"},
		},
		[]*html.Token{
			startH2, text("Some header"), endH2,
		},
	},
	{
		[]Tok{
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
		[]Tok{
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
		[]Tok{
			{CODE, "`", "`"}, {TEXT, "Some code", "Some code"}, {CODE, "`", "`"},
		},
		[]*html.Token{
			startP, startCode, text("Some code"), endCode, endP,
		},
	},
	{
		[]Tok{
			{CODE_BLOCK, "```", "```"}, {NEWLINE, "\n", "\n"}, {TEXT, "Some code ", "Some code "}, {EM, "*", "*"}, {CODE_BLOCK, "```", "```"},
		},
		[]*html.Token{
			startPre, startCode, text("\nSome code *\n"), endCode, endPre,
		},
	},
	{
		[]Tok{
			{TEXT, "foo", "foo"}, {NEWLINE, "\n", "\n"}, {TEXT, "bar", "bar"},
		},
		[]*html.Token{
			startP, text("foo"), text("\n"), text("bar"), endP,
		},
	},
	{
		[]Tok{
			{TEXT, "foo", "foo"}, {NEWLINE, "\n", "\n"}, {TEXT, "bar", "bar"}, {NEWLINE, "\n", "\n"}, {NEWLINE, "\n", "\n"}, {TEXT, "baz", "bar"}, {CODE, "`", "`"}, {TEXT, "bang", "bang"}, {CODE, "`", "`"},
		},
		[]*html.Token{
			startP, text("foo"), text("\n"), text("bar"), endP,
			startP, text("baz"), startCode, text("bang"), endCode, endP,
		},
	},
	{
		[]Tok{
			{STRONG, "**", "**"}, {TEXT, "foo", "foo"}, {STRONG, "**", "**"},
		},
		[]*html.Token{
			startP, startStrong, text("foo"), endStrong, endP,
		},
	},
	{
		[]Tok{
			{EM, "*", "*"}, {TEXT, "foo", "foo"}, {EM, "*", "*"},
		},
		[]*html.Token{
			startP, startEm, text("foo"), endEm, endP,
		},
	},
	{
		[]Tok{
			{UNORDERED_LIST, "* ", "* "}, {TEXT, "foo", "foo"}, {NEWLINE, "\n", "\n"},
			{UNORDERED_LIST, "* ", "* "}, {TEXT, "bar", "bar"}, {NEWLINE, "\n", "\n"},
			{UNORDERED_LIST, "* ", "* "}, {TEXT, "baz", "baz"}, {NEWLINE, "\n", "\n"},
		},
		[]*html.Token{
			startUl,
			startLi, text("foo"), endLi,
			startLi, text("bar"), endLi,
			startLi, text("baz"), endLi,
			endUl,
		},
	},
	{
		[]Tok{
			{ORDERED_LIST, "1. ", "1. "}, {TEXT, "foo", "foo"}, {NEWLINE, "\n", "\n"},
			{ORDERED_LIST, "2. ", "2. "}, {TEXT, "bar", "bar"}, {NEWLINE, "\n", "\n"},
			{ORDERED_LIST, "3. ", "2. "}, {TEXT, "baz", "baz"}, {NEWLINE, "\n", "\n"},
		},
		[]*html.Token{
			startOl,
			startLi, text("foo"), endLi,
			startLi, text("bar"), endLi,
			startLi, text("baz"), endLi,
			endOl,
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
