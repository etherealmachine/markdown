package markdown

import (
	"regexp"
	"strings"
)

type matcher func(s string) *Token

type Scanner struct {
	pos              int
	src              string
	next             *Token
	matchers         []matcher
	inOl, inUl, inTd bool
}

func NewScanner(src string) *Scanner {
	s := &Scanner{
		src: src,
	}
	s.matchers = []matcher{
		matchHeader,
		s.matchOrderedList,
		s.matchUnorderedList,
		s.matchTD,
		groupMatcher(regexp.MustCompile("^\r?(\n)"), NEWLINE),
		groupMatcher(regexp.MustCompile(`^\[(.*?)\]`), LINK_TEXT),
		groupMatcher(regexp.MustCompile(`^!\[(.*?)\]`), IMG_ALT),
		groupMatcher(regexp.MustCompile(`^\((.*?)\)`), HREF),
		groupMatcher(regexp.MustCompile(`^\*\*(.+?)\*\*`), STRONG),
		groupMatcher(regexp.MustCompile(`^\*(.+?)\*`), EM),
		groupMatcher(regexp.MustCompile(`^__(.+?)__`), STRONG),
		groupMatcher(regexp.MustCompile(`^\s_(.+?)_`), EM),
		groupMatcher(regexp.MustCompile("^(```)"), CODE_BLOCK),
		groupMatcher(regexp.MustCompile("^`(.*?)`"), CODE),
		groupMatcher(regexp.MustCompile("(?s)^(<.*?>)"), HTML_TAG),
		groupMatcher(regexp.MustCompile("^([$].*?[$])"), MATHML),
	}
	return s
}

func (s *Scanner) Next() *Token {
	if s.next != nil {
		tok := s.next
		s.next = nil
		return tok
	}
	last := s.pos
	for {
		if s.pos >= len(s.src) {
			break
		}
		if s.pos+1 < len(s.src) && s.src[s.pos] == '\n' && s.src[s.pos+1] == '\n' {
			s.inOl, s.inUl, s.inTd = false, false, false
		}
		if s.src[s.pos] == '\\' {
			s.pos += 2
			continue
		}
		for _, match := range s.matchers {
			if tok := match(s.src[s.pos:]); tok != nil {
				if last != s.pos {
					text := &Token{TEXT, s.src[last:s.pos], s.src[last:s.pos]}
					s.pos += len(tok.Raw)
					s.next = tok
					return text
				}
				s.pos += len(tok.Raw)
				return tok
			}
		}
		s.pos++
	}
	if last != s.pos {
		return &Token{TEXT, s.src[last:], s.src[last:]}
	}
	return &Token{EOF, "EOF", ""}
}
func groupMatcher(re *regexp.Regexp, tok TokenType) matcher {
	return func(s string) *Token {
		groups := re.FindStringSubmatch(s)
		if len(groups) == 0 {
			return nil
		}
		return &Token{tok, groups[1], groups[0]}
	}
}

var headerRe = regexp.MustCompile(`^[\t ]*([#]+)\s*`)

func matchHeader(s string) *Token {
	groups := headerRe.FindStringSubmatch(s)
	if len(groups) == 0 {
		return nil
	}
	return &Token{headers[len(groups[1])], groups[1], groups[0]}
}

var orderedListMatcher = groupMatcher(
	regexp.MustCompile(`^\n*([\t ]*)\d+\.[\t ]+`),
	ORDERED_LIST)

func (s *Scanner) matchOrderedList(str string) *Token {
	if !(s.pos == 0 || s.inOl || (len(str) >= 2 && str[0] == '\n' && str[1] == '\n')) {
		return nil
	}
	if tok := orderedListMatcher(str); tok != nil {
		s.inOl = true
		return tok
	}
	return nil
}

var unorderedListMatcher = groupMatcher(
	regexp.MustCompile(`^\n*([\t ]*)[*-][\t ]+`),
	UNORDERED_LIST)

func (s *Scanner) matchUnorderedList(str string) *Token {
	if !(s.pos == 0 || s.inUl || (len(str) >= 2 && str[0] == '\n' && str[1] == '\n')) {
		return nil
	}
	if tok := unorderedListMatcher(str); tok != nil {
		s.inUl = true
		return tok
	}
	return nil
}

var tdMatcher = groupMatcher(regexp.MustCompile(`^\s*(.*?)\s*[|]`), TD)

func (s *Scanner) matchTD(str string) *Token {
	for i, r := range str {
		if r == '\n' {
			if i > 0 && s.inTd {
				return &Token{TD, strings.TrimSpace(str[:i]), str[:i]}
			}
			return nil
		}
		if r == '|' {
			if tok := tdMatcher(str[:i+1]); tok != nil {
				s.inTd = true
				return tok
			}
		}
	}
	if s.inTd {
		s.inTd = false
		return &Token{TD, strings.TrimSpace(str), str}
	}
	return nil
}
