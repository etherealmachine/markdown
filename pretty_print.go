package markdown

import (
	"bytes"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func PrettyPrint(tokens []*html.Token) string {
	var buf bytes.Buffer
	for i, tok := range tokens {
		buf.WriteString(tok.String())
		if tok.Type == html.EndTagToken &&
			blockTag[tok.DataAtom] &&
			i < len(tokens)-1 {
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

var (
	blockTag = map[atom.Atom]bool{
		atom.H1:  true,
		atom.H2:  true,
		atom.H3:  true,
		atom.H4:  true,
		atom.H5:  true,
		atom.H6:  true,
		atom.P:   true,
		atom.Div: true,
		atom.Pre: true,
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
