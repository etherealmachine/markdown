package markdown

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"reflect"
	"testing"
)

type token struct {
	tok Token
	lit string
}

type fakeScanner struct {
	pos  int
	toks []token
}

func (s *fakeScanner) Next() (Token, string) {
	if s.pos >= len(s.toks) {
		return EOF, "EOF"
	}
	s.pos++
	return s.toks[s.pos-1].tok, s.toks[s.pos-1].lit
}

type parserCase struct {
	input []token
	want  []*html.Token
}

var parserCases = []*parserCase{
	{
		[]token{
			{H1, "H1"}, {TEXT, "Some header"}, {NEWLINE, "\n"},
		},
		[]*html.Token{
			startH1, text("Some header"), endH1,
		},
	},
	{
		[]token{
			{H2, "H2"}, {TEXT, "Some header"}, {NEWLINE, "\n"},
		},
		[]*html.Token{
			startH2, text("Some header"), endH2,
		},
	},
	{
		[]token{
			{LINK_TEXT, "A link"}, {HREF, "www.example.com"},
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
		[]token{
			{IMG_ALT, "An image"}, {HREF, "www.example.com"},
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
		[]token{
			{CODE, "`"}, {TEXT, "Some code"}, {CODE, "`"},
		},
		[]*html.Token{
			startP, startCode, text("Some code"), endCode, endP,
		},
	},
	{
		[]token{
			{CODE_BLOCK, "```"}, {NEWLINE, "\n"}, {TEXT, "Some code "}, {EM, "*"}, {CODE_BLOCK, "```"},
		},
		[]*html.Token{
			startPre, startCode, text("\nSome code *\n"), endCode, endPre,
		},
	},
	{
		[]token{
			{TEXT, "foo"}, {NEWLINE, "\n"}, {TEXT, "bar"},
		},
		[]*html.Token{
			startP, text("foo"), text("\n"), text("bar"), endP,
		},
	},
	{
		[]token{
			{TEXT, "foo"}, {NEWLINE, "\n"}, {TEXT, "bar"}, {NEWLINE, "\n"}, {NEWLINE, "\n"}, {TEXT, "baz"}, {CODE, "`"}, {TEXT, "bang"}, {CODE, "`"},
		},
		[]*html.Token{
			startP, text("foo"), text("\n"), text("bar"), endP,
			startP, text("baz"), startCode, text("bang"), endCode, endP,
		},
	},
	{
		[]token{
			{STRONG, "**"}, {TEXT, "foo"}, {STRONG, "**"},
		},
		[]*html.Token{
			startP, startStrong, text("foo"), endStrong, endP,
		},
	},
	{
		[]token{
			{EM, "*"}, {TEXT, "foo"}, {EM, "*"},
		},
		[]*html.Token{
			startP, startEm, text("foo"), endEm, endP,
		},
	},
	{
		[]token{
			{UNORDERED_LIST, "* "}, {TEXT, "foo"}, {NEWLINE, "\n"},
			{UNORDERED_LIST, "* "}, {TEXT, "bar"}, {NEWLINE, "\n"},
			{UNORDERED_LIST, "* "}, {TEXT, "baz"}, {NEWLINE, "\n"},
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
		[]token{
			{ORDERED_LIST, "1. "}, {TEXT, "foo"}, {NEWLINE, "\n"},
			{ORDERED_LIST, "2. "}, {TEXT, "bar"}, {NEWLINE, "\n"},
			{ORDERED_LIST, "3. "}, {TEXT, "baz"}, {NEWLINE, "\n"},
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
		p := &Parser{scanner: &fakeScanner{toks: c.input}}
		p.parse()
		got := p.tokens
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("got:\n%v\nwant:\n%v", got, c.want)
		}
	}
}
