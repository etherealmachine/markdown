package markdown

import (
	"fmt"
)

type Token int

type Tok struct {
	Tok Token
	Lit string
	Raw string
}

func (t *Tok) Tuple() (Token, string, string) {
	return t.Tok, t.Lit, t.Raw
}

func (t *Tok) String() string {
	return fmt.Sprintf("&{%s %q %q}", t.Tok, t.Lit, t.Raw)
}

const (
	EOF = iota
	H1
	H2
	H3
	H4
	H5
	H6
	EM
	STRONG
	NEWLINE
	TEXT
	LINK_TEXT
	IMG_ALT
	HREF
	CODE
	HTML_TAG
	CODE_BLOCK
	ORDERED_LIST
	UNORDERED_LIST
	MATHML
)

var tokenNames = map[Token]string{
	EOF:            "EOF",
	H1:             "H1",
	H2:             "H2",
	H3:             "H3",
	H4:             "H4",
	H5:             "H5",
	H6:             "H6",
	EM:             "EM",
	STRONG:         "STRONG",
	NEWLINE:        "NEWLINE",
	TEXT:           "TEXT",
	LINK_TEXT:      "LINK_TEXT",
	IMG_ALT:        "IMG_ALT",
	HREF:           "HREF",
	CODE:           "CODE",
	HTML_TAG:       "HTML_TAG",
	CODE_BLOCK:     "CODE_BLOCK",
	ORDERED_LIST:   "ORDERED_LIST",
	UNORDERED_LIST: "UNORDERED_LIST",
	MATHML:         "MATHML",
}

func (t Token) String() string {
	return tokenNames[t]
}

var headers = map[int]Token{
	1: H1,
	2: H2,
	3: H3,
	4: H4,
	5: H5,
	6: H6,
}

var blockToken = map[Token]bool{
	EOF:            true,
	H1:             true,
	H2:             true,
	H3:             true,
	H4:             true,
	H5:             true,
	H6:             true,
	CODE_BLOCK:     true,
	ORDERED_LIST:   true,
	UNORDERED_LIST: true,
}
