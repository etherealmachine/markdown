package markdown

type Token int

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
	HTML_START
	HTML_END
	HTML_END_TAG
	CODE_BLOCK
	ORDERED_LIST
	UNORDERED_LIST
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
	HTML_START:     "HTML_START",
	HTML_END:       "HTML_END",
	HTML_END_TAG:   "HTML_END_TAG",
	CODE_BLOCK:     "CODE_BLOCK",
	ORDERED_LIST:   "ORDERED_LIST",
	UNORDERED_LIST: "UNORDERED_LIST",
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
