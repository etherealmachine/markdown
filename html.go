package markdown

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	startH1   = &html.Token{Type: html.StartTagToken, DataAtom: atom.H1, Data: "h1"}
	endH1     = &html.Token{Type: html.EndTagToken, DataAtom: atom.H1, Data: "h1"}
	startH2   = &html.Token{Type: html.StartTagToken, DataAtom: atom.H2, Data: "h2"}
	endH2     = &html.Token{Type: html.EndTagToken, DataAtom: atom.H2, Data: "h2"}
	startH3   = &html.Token{Type: html.StartTagToken, DataAtom: atom.H3, Data: "h3"}
	endH3     = &html.Token{Type: html.EndTagToken, DataAtom: atom.H3, Data: "h3"}
	startH4   = &html.Token{Type: html.StartTagToken, DataAtom: atom.H4, Data: "h4"}
	endH4     = &html.Token{Type: html.EndTagToken, DataAtom: atom.H4, Data: "h4"}
	startH5   = &html.Token{Type: html.StartTagToken, DataAtom: atom.H5, Data: "h5"}
	endH5     = &html.Token{Type: html.EndTagToken, DataAtom: atom.H5, Data: "h5"}
	startH6   = &html.Token{Type: html.StartTagToken, DataAtom: atom.H6, Data: "h6"}
	endH6     = &html.Token{Type: html.EndTagToken, DataAtom: atom.H6, Data: "h6"}
	hStartTag = map[TokenType]*html.Token{
		H1: startH1,
		H2: startH2,
		H3: startH3,
		H4: startH4,
		H5: startH5,
		H6: startH6,
	}
	hEndTag = map[TokenType]*html.Token{
		H1: endH1,
		H2: endH2,
		H3: endH3,
		H4: endH4,
		H5: endH5,
		H6: endH6,
	}
	endA        = &html.Token{Type: html.EndTagToken, DataAtom: atom.A, Data: "a"}
	startCode   = &html.Token{Type: html.StartTagToken, DataAtom: atom.Code, Data: "code"}
	endCode     = &html.Token{Type: html.EndTagToken, DataAtom: atom.Code, Data: "code"}
	startPre    = &html.Token{Type: html.StartTagToken, DataAtom: atom.Pre, Data: "pre"}
	endPre      = &html.Token{Type: html.EndTagToken, DataAtom: atom.Pre, Data: "pre"}
	startP      = &html.Token{Type: html.StartTagToken, DataAtom: atom.P, Data: "p"}
	endP        = &html.Token{Type: html.EndTagToken, DataAtom: atom.P, Data: "p"}
	startOl     = &html.Token{Type: html.StartTagToken, DataAtom: atom.Ol, Data: "ol"}
	endOl       = &html.Token{Type: html.EndTagToken, DataAtom: atom.Ol, Data: "ol"}
	startUl     = &html.Token{Type: html.StartTagToken, DataAtom: atom.Ul, Data: "ul"}
	endUl       = &html.Token{Type: html.EndTagToken, DataAtom: atom.Ul, Data: "ul"}
	startLi     = &html.Token{Type: html.StartTagToken, DataAtom: atom.Li, Data: "li"}
	endLi       = &html.Token{Type: html.EndTagToken, DataAtom: atom.Li, Data: "li"}
	startEm     = &html.Token{Type: html.StartTagToken, DataAtom: atom.Em, Data: "em"}
	endEm       = &html.Token{Type: html.EndTagToken, DataAtom: atom.Em, Data: "em"}
	startStrong = &html.Token{Type: html.StartTagToken, DataAtom: atom.Strong, Data: "strong"}
	endStrong   = &html.Token{Type: html.EndTagToken, DataAtom: atom.Strong, Data: "strong"}
	startTable  = &html.Token{Type: html.StartTagToken, DataAtom: atom.Table, Data: "table"}
	endTable    = &html.Token{Type: html.EndTagToken, DataAtom: atom.Table, Data: "table"}
	startTr     = &html.Token{Type: html.StartTagToken, DataAtom: atom.Tr, Data: "tr"}
	endTr       = &html.Token{Type: html.EndTagToken, DataAtom: atom.Tr, Data: "tr"}
	startTh     = &html.Token{Type: html.StartTagToken, DataAtom: atom.Th, Data: "th"}
	endTh       = &html.Token{Type: html.EndTagToken, DataAtom: atom.Th, Data: "th"}
	startTd     = &html.Token{Type: html.StartTagToken, DataAtom: atom.Td, Data: "td"}
	endTd       = &html.Token{Type: html.EndTagToken, DataAtom: atom.Td, Data: "td"}
)

func text(s string) *html.Token {
	return &html.Token{Type: html.TextToken, Data: s}
}

var (
	blockTag = map[atom.Atom]bool{
		atom.H1:    true,
		atom.H2:    true,
		atom.H3:    true,
		atom.H4:    true,
		atom.H5:    true,
		atom.H6:    true,
		atom.P:     true,
		atom.Div:   true,
		atom.Pre:   true,
		atom.Ol:    true,
		atom.Ul:    true,
		atom.Li:    true,
		atom.Table: true,
		atom.Tr:    true,
	}
	inlineTag = map[atom.Atom]bool{
		atom.B:        true,
		atom.Big:      true,
		atom.I:        true,
		atom.Small:    true,
		atom.Tt:       true,
		atom.Abbr:     true,
		atom.Cite:     true,
		atom.Code:     true,
		atom.Dfn:      true,
		atom.Em:       true,
		atom.Kbd:      true,
		atom.Strong:   true,
		atom.Samp:     true,
		atom.Var:      true,
		atom.A:        true,
		atom.Bdo:      true,
		atom.Br:       true,
		atom.Img:      true,
		atom.Map:      true,
		atom.Object:   true,
		atom.Q:        true,
		atom.Script:   true,
		atom.Span:     true,
		atom.Sub:      true,
		atom.Sup:      true,
		atom.Button:   true,
		atom.Input:    true,
		atom.Label:    true,
		atom.Select:   true,
		atom.Textarea: true,
	}
)

func inline(token *html.Token) bool {
	if token == nil {
		return true
	}
	if token.Type == html.TextToken {
		return true
	}
	return inlineTag[token.DataAtom]
}
