package markdown

import (
	"regexp"
)

type matcher func(s string) *Tok

type Scanner struct {
	pos        int
	src        string
	next       *Tok
	matchers   []matcher
	inOl, inUl bool
}

func NewScanner(src string) *Scanner {
	s := &Scanner{
		src: src,
	}
	s.matchers = []matcher{
		matchHeader,
		s.matchOrderedList,
		s.matchUnorderedList,
		groupMatcher(regexp.MustCompile("^\r?(\n)"), NEWLINE),
		groupMatcher(regexp.MustCompile(`^\[(.*?)\]`), LINK_TEXT),
		groupMatcher(regexp.MustCompile(`^!\[(.*?)\]`), IMG_ALT),
		groupMatcher(regexp.MustCompile(`^\((.*?)\)`), HREF),
		groupMatcher(regexp.MustCompile(`^\*\*(.+?)\*\*`), STRONG),
		groupMatcher(regexp.MustCompile(`^\*(.+?)\*`), EM),
		groupMatcher(regexp.MustCompile(`^__(.+?)__`), STRONG),
		groupMatcher(regexp.MustCompile(`^_(.+?)_`), EM),
		groupMatcher(regexp.MustCompile("^(```)"), CODE_BLOCK),
		groupMatcher(regexp.MustCompile("^`(.*?)`"), CODE),
		groupMatcher(regexp.MustCompile("^(<.*?>)"), HTML_TAG),
	}
	return s
}

func (s *Scanner) Next() *Tok {
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
			s.inOl, s.inUl = false, false
		}
		for _, match := range s.matchers {
			if tok := match(s.src[s.pos:]); tok != nil {
				if last != s.pos {
					text := &Tok{TEXT, s.src[last:s.pos], s.src[last:s.pos]}
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
		return &Tok{TEXT, s.src[last:], s.src[last:]}
	}
	return &Tok{EOF, "EOF", ""}
}
func groupMatcher(re *regexp.Regexp, tok Token) matcher {
	return func(s string) *Tok {
		groups := re.FindStringSubmatch(s)
		if len(groups) == 0 {
			return nil
		}
		return &Tok{tok, groups[1], groups[0]}
	}
}

var headerRe = regexp.MustCompile(`^([#]+)\s*`)

func matchHeader(s string) *Tok {
	groups := headerRe.FindStringSubmatch(s)
	if len(groups) == 0 {
		return nil
	}
	return &Tok{headers[len(groups[1])], groups[1], groups[0]}
}

var orderedListMatcher = groupMatcher(regexp.MustCompile(`^\n*([\t ]*)\d+\. `), ORDERED_LIST)

func (s *Scanner) matchOrderedList(str string) *Tok {
	if !(s.pos == 0 || s.inOl || (len(str) >= 2 && str[0] == '\n' && str[1] == '\n')) {
		return nil
	}
	if tok := orderedListMatcher(str); tok != nil {
		s.inOl = true
		return tok
	}
	return nil
}

var unorderedListMatcher = groupMatcher(regexp.MustCompile(`^\n*([\t ]*)[*-] `), UNORDERED_LIST)

func (s *Scanner) matchUnorderedList(str string) *Tok {
	if !(s.pos == 0 || s.inUl || (len(str) >= 2 && str[0] == '\n' && str[1] == '\n')) {
		return nil
	}
	if tok := unorderedListMatcher(str); tok != nil {
		s.inUl = true
		return tok
	}
	return nil
}
