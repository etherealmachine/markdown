package markdown

import (
	"bytes"
	"io"
	"strings"
)

type Scanner struct {
	src  io.RuneScanner
	prev Token
}

func NewScanner(src string) *Scanner {
	return &Scanner{
		src:  strings.NewReader(src),
		prev: EOF,
	}
}

func (s *Scanner) Next() (Token, string) {
	tok, lit := s.next()
	s.prev = tok
	return tok, lit
}

func (s *Scanner) next() (Token, string) {
	var buf bytes.Buffer
	for {
		r, _, err := s.src.ReadRune()
		if err == io.EOF {
			if buf.Len() != 0 {
				s.src.UnreadRune()
				return TEXT, buf.String()
			}
			return EOF, ""
		}
		switch r {
		case '#':
			if buf.Len() != 0 {
				s.src.UnreadRune()
				return TEXT, buf.String()
			}
			return s.scanHeader()
		case '*':
			if buf.Len() != 0 {
				s.src.UnreadRune()
				return TEXT, buf.String()
			}
			if next, _, _ := s.src.ReadRune(); next == '*' {
				return STRONG, "__"
			} else if (s.prev == EOF || s.prev == NEWLINE) && next == ' ' {
				return UNORDERED_LIST, "* "
			}
			s.src.UnreadRune()
			return EM, "_"
		case '_':
			if buf.Len() != 0 {
				s.src.UnreadRune()
				return TEXT, buf.String()
			}
			if next, _, _ := s.src.ReadRune(); next == '_' {
				return STRONG, "__"
			}
			s.src.UnreadRune()
			return EM, "_"
		case '!':
			if next, _, _ := s.src.ReadRune(); next == '[' {
				return s.scanImgAlt()
			}
			s.src.UnreadRune()
			buf.WriteRune(r)
		case '[':
			if buf.Len() != 0 {
				s.src.UnreadRune()
				return TEXT, buf.String()
			}
			return s.scanLinkText()
		case '(':
			if s.prev == LINK_TEXT || s.prev == IMG_ALT {
				return s.scanHref()
			}
			buf.WriteRune(r)
		case '`':
			if buf.Len() != 0 {
				s.src.UnreadRune()
				return TEXT, buf.String()
			}
			if next, _, _ := s.src.ReadRune(); next == '`' {
				if next, _, _ = s.src.ReadRune(); next == '`' {
					return CODE_BLOCK, "```"
				}
				s.src.UnreadRune()
			}
			s.src.UnreadRune()
			return CODE, "`"
		case '<':
			if buf.Len() != 0 {
				s.src.UnreadRune()
				return TEXT, buf.String()
			}
			if next, _, _ := s.src.ReadRune(); next == '/' {
				return s.scanHtmlEndTag()
			}
			s.src.UnreadRune()
			return HTML_START, "<"
		case '>':
			if buf.Len() != 0 {
				s.src.UnreadRune()
				return TEXT, buf.String()
			}
			return HTML_END, ">"
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if s.prev == EOF || s.prev == NEWLINE {
				if next, _, _ := s.src.ReadRune(); next == '.' {
					if next, _, _ = s.src.ReadRune(); next == ' ' {
						return ORDERED_LIST, string([]rune{r, '.', ' '})
					}
					s.src.UnreadRune()
				}
				s.src.UnreadRune()
			}
			buf.WriteRune(r)
		case '\n':
			if buf.Len() != 0 {
				s.src.UnreadRune()
				return TEXT, buf.String()
			}
			return NEWLINE, "\n"
		default:
			buf.WriteRune(r)
		}
	}
	return EOF, "EOF"
}

func (s *Scanner) peek() rune {
	next, _, err := s.src.ReadRune()
	s.src.UnreadRune()
	if err == io.EOF {
		return ' '
	}
	return next
}

func (s *Scanner) scanHeader() (Token, string) {
	lit := bytes.NewBufferString("#")
	count := 1
	for {
		r, _, err := s.src.ReadRune()
		if err == io.EOF {
			panic("EOF")
		}
		if r == '#' {
			lit.WriteByte('#')
			count++
		} else {
			s.src.UnreadRune()
			break
		}
	}
	for s.peek() == ' ' {
		s.src.ReadRune()
	}
	return headers[count], lit.String()
}

func (s *Scanner) scanLinkText() (Token, string) {
	var lit bytes.Buffer
	for {
		r, _, err := s.src.ReadRune()
		if err == io.EOF {
			panic("EOF")
		}
		if r != ']' {
			lit.WriteRune(r)
		} else {
			break
		}
	}
	return LINK_TEXT, lit.String()
}

func (s *Scanner) scanImgAlt() (Token, string) {
	var lit bytes.Buffer
	for {
		r, _, err := s.src.ReadRune()
		if err == io.EOF {
			panic("EOF")
		}
		if r != ']' {
			lit.WriteRune(r)
		} else {
			break
		}
	}
	return IMG_ALT, lit.String()
}

func (s *Scanner) scanHref() (Token, string) {
	var lit bytes.Buffer
	for {
		r, _, err := s.src.ReadRune()
		if err == io.EOF {
			panic("EOF")
		}
		if r != ')' {
			lit.WriteRune(r)
		} else {
			break
		}
	}
	return HREF, lit.String()
}

func (s *Scanner) scanHtmlEndTag() (Token, string) {
	var lit bytes.Buffer
	for {
		r, _, err := s.src.ReadRune()
		if err == io.EOF {
			panic("EOF")
		}
		if r != '>' {
			lit.WriteRune(r)
		} else {
			break
		}
	}
	return HTML_END_TAG, lit.String()
}
