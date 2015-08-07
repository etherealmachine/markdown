package markdown

import (
	"bytes"
	"io"
	"strings"
)

type Scanner struct {
	src  io.RuneScanner
	prev Token
	buf  bytes.Buffer
}

func NewScanner(src string) *Scanner {
	return &Scanner{
		src:  strings.NewReader(src),
		prev: EOF,
	}
}

func (s *Scanner) Next() *Tok {
	tok := s.next()
	s.prev = tok.Tok
	return tok
}

var specialRunes = map[rune]bool{
	'#':  true,
	'*':  true,
	'_':  true,
	'!':  true,
	'[':  true,
	'`':  true,
	'<':  true,
	'>':  true,
	'1':  true,
	'2':  true,
	'3':  true,
	'4':  true,
	'5':  true,
	'6':  true,
	'7':  true,
	'8':  true,
	'9':  true,
	'\r': true,
	'\n': true,
}

func (s *Scanner) next() *Tok {
	for {
		r, _, err := s.src.ReadRune()
		if err == io.EOF {
			if tok := s.drainText(); tok != nil {
				return tok
			}
			return &Tok{EOF, "EOF", ""}
		}
		if specialRunes[r] {
			if tok := s.drainText(); tok != nil {
				return tok
			}
		}
		switch r {
		case '#':
			return s.scanHeader()
		case '*':
			if next, _, _ := s.src.ReadRune(); next == '*' {
				return &Tok{STRONG, "**", "**"}
			} else if (s.prev == EOF || s.prev == NEWLINE) && next == ' ' {
				return &Tok{UNORDERED_LIST, "* ", "* "}
			}
			s.src.UnreadRune()
			return &Tok{EM, "*", "*"}
		case '_':
			if next, _, _ := s.src.ReadRune(); next == '_' {
				return &Tok{STRONG, "__", "__"}
			}
			s.src.UnreadRune()
			return &Tok{EM, "_", "_"}
		case '!':
			if next, _, _ := s.src.ReadRune(); next == '[' {
				return s.scanImgAlt()
			}
			s.src.UnreadRune()
			s.buf.WriteRune(r)
		case '[':
			return s.scanLinkText()
		case '(':
			if s.prev == LINK_TEXT || s.prev == IMG_ALT {
				return s.scanHref()
			}
			s.buf.WriteRune(r)
		case '`':
			if next, _, _ := s.src.ReadRune(); next == '`' {
				if next, _, _ = s.src.ReadRune(); next == '`' {
					return &Tok{CODE_BLOCK, "```", "```"}
				}
				s.src.UnreadRune()
			}
			s.src.UnreadRune()
			return &Tok{CODE, "`", "`"}
		case '<':
			if next, _, _ := s.src.ReadRune(); next == '/' {
				return s.scanHtmlEndTag()
			}
			s.src.UnreadRune()
			return &Tok{HTML_START, "<", "<"}
		case '>':
			return &Tok{HTML_END, ">", ">"}
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if s.prev == EOF || s.prev == NEWLINE {
				if next, _, _ := s.src.ReadRune(); next == '.' {
					if next, _, _ = s.src.ReadRune(); next == ' ' {
						numeral := string([]rune{r, '.', ' '})
						return &Tok{ORDERED_LIST, numeral, numeral}
					}
					s.src.UnreadRune()
				}
				s.src.UnreadRune()
			}
			s.buf.WriteRune(r)
		case '\r':
		case '\n':
			return &Tok{NEWLINE, "\n", "\n"}
		default:
			s.buf.WriteRune(r)
		}
	}
	return nil
}

func (s *Scanner) drainText() *Tok {
	if s.buf.Len() != 0 {
		s.src.UnreadRune()
		text := s.buf.String()
		s.buf.Reset()
		return &Tok{TEXT, text, text}
	}
	return nil
}

func (s *Scanner) peek() rune {
	next, _, err := s.src.ReadRune()
	s.src.UnreadRune()
	if err != nil {
		return ' '
	}
	return next
}

func (s *Scanner) scanHeader() *Tok {
	lit := bytes.NewBufferString("#")
	count := 1
	for {
		r, _, err := s.src.ReadRune()
		if err != nil {
			return &Tok{EOF, "EOF", ""}
		}
		if r == '#' {
			lit.WriteByte('#')
			count++
		} else {
			s.src.UnreadRune()
			break
		}
	}
	header := lit.String()
	raw := header
	for s.peek() == ' ' {
		raw += " "
		s.src.ReadRune()
	}
	return &Tok{headers[count], header, raw}
}

func (s *Scanner) scanLinkText() *Tok {
	var lit bytes.Buffer
	for {
		r, _, err := s.src.ReadRune()
		if err != nil {
			return &Tok{EOF, "EOF", ""}
		}
		if r != ']' {
			lit.WriteRune(r)
		} else {
			break
		}
	}
	text := lit.String()
	raw := "[" + text + "]"
	return &Tok{LINK_TEXT, text, raw}
}

func (s *Scanner) scanImgAlt() *Tok {
	var lit bytes.Buffer
	for {
		r, _, err := s.src.ReadRune()
		if err != nil {
			return &Tok{EOF, "EOF", ""}
		}
		if r != ']' {
			lit.WriteRune(r)
		} else {
			break
		}
	}
	alt := lit.String()
	raw := "![" + alt + "]"
	return &Tok{IMG_ALT, alt, raw}
}

func (s *Scanner) scanHref() *Tok {
	var lit bytes.Buffer
	for {
		r, _, err := s.src.ReadRune()
		if err != nil {
			return &Tok{EOF, "EOF", ""}
		}
		if r != ')' {
			lit.WriteRune(r)
		} else {
			break
		}
	}
	href := lit.String()
	raw := "(" + href + ")"
	return &Tok{HREF, href, raw}
}

func (s *Scanner) scanHtmlEndTag() *Tok {
	var lit bytes.Buffer
	for {
		r, _, err := s.src.ReadRune()
		if err != nil {
			return &Tok{EOF, "EOF", ""}
		}
		if r != '>' {
			lit.WriteRune(r)
		} else {
			break
		}
	}
	inner := lit.String()
	raw := "<" + inner + ">"
	return &Tok{HTML_END_TAG, inner, raw}
}
