package markdown

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"strings"
)

type scanner interface {
	Next() *Token
}

type ErrUnexpectedToken struct {
	tok *Token
}

func (e ErrUnexpectedToken) Error() string {
	return fmt.Sprintf("%v", e.tok)
}

type Parser struct {
	pos        int
	input      []*Token
	tokens     []*html.Token
	inlineMode bool
	saved      savePoint
	tableAttrs []html.Attribute
}

type savePoint struct {
	pos, tokenCount int
	inlineMode      bool
}

func Parse(input string) []*html.Token {
	p := &Parser{}
	p.parse(NewScanner(input))
	return p.tokens
}

func (p *Parser) parse(scanner scanner) {
	for tok := scanner.Next(); tok.Type != EOF; tok = scanner.Next() {
		p.input = append(p.input, tok)
	}
	for tok := p.next(); tok.Type != EOF; tok = p.next() {
		p.consume(tok)
	}
	if p.inlineMode {
		p.append(endP)
		p.inlineMode = false
	}
}

func (p *Parser) consume(tok *Token) {
	p.save()
	var err error
	switch tok.Type {
	case H1, H2, H3, H4, H5, H6:
		p.parseHeader(tok.Type)
	case CODE_BLOCK:
		err = p.parseCodeBlock()
	case ORDERED_LIST:
		p.parseOrderedList()
	case UNORDERED_LIST:
		p.parseUnorderedList()
	case TD:
		err = p.parseTD()
	default:
		p.consumeInline(tok)
	}
	if err != nil {
		p.revert()
	}
}

func (p *Parser) consumeInline(tok *Token) {
	p.save()
	var err error
	switch tok.Type {
	case EM:
		p.parseEm(tok.Lit)
	case STRONG:
		p.parseStrong(tok.Lit)
	case NEWLINE:
		p.parseNewline()
	case TEXT:
		p.parseText(tok.Lit)
	case LINK_TEXT:
		err = p.parseLink(tok.Lit)
	case IMG_ALT:
		err = p.parseImg(tok.Lit)
	case CODE:
		p.parseCode(tok.Lit)
	case HTML_TAG:
		p.parseHTMLTag(tok.Lit)
	default:
		p.parseText(tok.Raw)
	}
	if err != nil {
		p.revert()
	}
}

func (p *Parser) expect(tok TokenType) (string, error) {
	t := p.next()
	if t.Type != tok {
		return "", ErrUnexpectedToken{t}
	}
	return t.Lit, nil
}

func (p *Parser) save() {
	p.saved.pos = p.pos - 1
	p.saved.tokenCount = len(p.tokens)
	p.saved.inlineMode = p.inlineMode
}

func (p *Parser) revert() {
	var buf bytes.Buffer
	for i := p.saved.pos; i < p.pos && i < len(p.input); i++ {
		buf.WriteString(p.input[i].Raw)
	}
	p.tokens = p.tokens[:p.saved.tokenCount]
	p.inlineMode = p.saved.inlineMode
	p.parseText(buf.String())
}

func (p *Parser) next() *Token {
	if p.pos >= len(p.input) {
		return &Token{EOF, "EOF", ""}
	}
	p.pos++
	return p.input[p.pos-1]
}

func (p *Parser) prev() *html.Token {
	if len(p.tokens) == 0 {
		return &html.Token{Type: html.CommentToken}
	}
	return p.tokens[len(p.tokens)-1]
}

func (p *Parser) peek() *Token {
	if p.pos >= len(p.input) {
		return &Token{EOF, "EOF", ""}
	}
	return p.input[p.pos]
}

func (p *Parser) append(tok *html.Token) {
	p.tokens = append(p.tokens, tok)
}

func (p *Parser) inline() {
	if !p.inlineMode {
		p.append(startP)
		p.inlineMode = true
	}
}

func (p *Parser) block() {
	if p.inlineMode {
		p.append(endP)
		p.inlineMode = false
	}
}

func (p *Parser) handleDirective(s string) bool {
	if strings.HasPrefix(s, "table ") {
		tt := html.NewTokenizer(strings.NewReader("<" + s + ">"))
		tt.Next()
		tok := tt.Token()
		p.tableAttrs = tok.Attr
		return true
	}
	return false
}

func (p *Parser) parseHeader(headerToken TokenType) {
	p.append(hStartTag[headerToken])
	p.inlineMode = true
	for {
		next := p.peek()
		if next.Type == EOF || next.Type == NEWLINE {
			p.next()
			break
		}
		if next.Type == ORDERED_LIST || next.Type == UNORDERED_LIST {
			break
		}
		p.next()
		p.consumeInline(next)
	}
	p.append(hEndTag[headerToken])
	p.inlineMode = false
}

func (p *Parser) parseEm(lit string) {
	p.inline()
	p.append(startEm)
	p.append(text(lit))
	p.append(endEm)
}

func (p *Parser) parseStrong(lit string) {
	p.inline()
	p.append(startStrong)
	p.append(text(lit))
	p.append(endStrong)
}

func (p *Parser) parseNewline() {
	next := p.next()
	if next.Type == NEWLINE {
		p.block()
	} else {
		if next.Type == EOF || (next.Type == TEXT && strings.TrimSpace(next.Lit) == "") {
			return
		}
		p.tokens = append(p.tokens, text("\n"))
		p.consume(next)
	}
}

func (p *Parser) parseText(s string) {
	if p.prev().Type != html.TextToken && strings.TrimLeft(s, " ") == "" {
		return
	}
	if !p.inlineMode {
		p.append(startP)
		p.inlineMode = true
		s = strings.TrimLeft(s, " ")
	}
	p.append(text(s))
}

func (p *Parser) parseLink(s string) error {
	p.inline()
	href, err := p.expect(HREF)
	if err != nil {
		return err
	}
	p.append(&html.Token{
		Type:     html.StartTagToken,
		DataAtom: atom.A,
		Data:     "a",
		Attr: []html.Attribute{{
			Key: "href",
			Val: href,
		}},
	})
	p.append(text(s))
	p.append(endA)
	return nil
}

func (p *Parser) parseImg(alt string) error {
	p.inline()
	src, err := p.expect(HREF)
	if err != nil {
		return err
	}
	p.tokens = append(p.tokens, &html.Token{
		Type:     html.SelfClosingTagToken,
		DataAtom: atom.Img,
		Data:     "img",
		Attr: []html.Attribute{
			{Key: "alt", Val: alt},
			{Key: "src", Val: src},
		},
	})
	return nil
}

func (p *Parser) parseCode(code string) {
	p.inline()
	p.append(startCode)
	p.append(text(code))
	p.append(endCode)
}

func (p *Parser) parseCodeBlock() error {
	p.append(startPre)
	var buf bytes.Buffer
	tok := p.next()
	if tok.Type == TEXT {
		p.tokens = append(p.tokens, &html.Token{
			Type:     html.StartTagToken,
			DataAtom: atom.Code,
			Data:     "code",
			Attr: []html.Attribute{
				{Key: "class", Val: tok.Lit},
			},
		})
	} else if tok.Type == NEWLINE {
		p.append(startCode)
	} else {
		return ErrUnexpectedToken{tok}
	}
	for tok = p.next(); tok.Type != CODE_BLOCK; tok = p.next() {
		if tok.Type == EOF {
			return ErrUnexpectedToken{tok}
		}
		buf.WriteString(tok.Raw)
	}
	code := strings.Trim(buf.String(), "\n")
	p.append(text(code))
	p.append(endCode)
	p.append(endPre)
	return nil
}

func (p *Parser) parseHTMLTag(tag string) {
	tt := html.NewTokenizer(strings.NewReader(tag))
	tt.Next()
	tok := tt.Token()
	if tok.Type == html.CommentToken {
		if p.handleDirective(tok.Data) {
			return
		}
	} else {
		if !p.inlineMode && inline(&tok) &&
			(tok.Type == html.StartTagToken && inlineBlock[tok.DataAtom]) {
			p.append(startP)
			p.inlineMode = true
		} else if p.inlineMode && !inline(&tok) &&
			!(tok.Type == html.EndTagToken && inlineBlock[tok.DataAtom]) {
			p.append(endP)
			p.inlineMode = false
		}
		if blockTag[tok.DataAtom] {
			p.inlineMode = false
		} else {
			p.inlineMode = true
		}
	}
	p.append(&tok)
}

func (p *Parser) parseOrderedList() {
	p.block()
	depth := 0
	p.inlineMode = true
	p.append(startOl)
	p.append(startLi)
	for tok := p.next(); tok.Type != EOF && tok.Type != NEWLINE; tok = p.next() {
		if tok.Type == ORDERED_LIST {
			d := len(tok.Lit)
			if d > depth {
				p.append(startOl)
				p.append(startLi)
			} else if d < depth {
				p.append(endLi)
				p.append(endOl)
				p.append(endLi)
				p.append(startLi)
			} else {
				p.append(endLi)
				p.append(startLi)
			}
			depth = d
		} else {
			p.consumeInline(tok)
		}
	}
	for ; depth >= 0; depth-- {
		p.append(endLi)
		p.append(endOl)
	}
	p.inlineMode = false
}

func (p *Parser) parseUnorderedList() {
	p.block()
	depth := 0
	p.inlineMode = true
	p.append(startUl)
	p.append(startLi)
	for tok := p.next(); tok.Type != EOF && tok.Type != NEWLINE; tok = p.next() {
		if tok.Type == UNORDERED_LIST {
			d := len(tok.Lit)
			if d > depth {
				p.append(startUl)
				p.append(startLi)
			} else if d < depth {
				for ; depth > d; depth-- {
					p.append(endLi)
					p.append(endUl)
				}
				p.append(endLi)
				p.append(startLi)
			} else {
				p.append(endLi)
				p.append(startLi)
			}
			depth = d
		} else {
			p.consumeInline(tok)
		}
	}
	for ; depth >= 0; depth-- {
		p.append(endLi)
		p.append(endUl)
	}
	p.inlineMode = false
}

func (p *Parser) parseTD() error {
	p.block()
	p.inlineMode = false
	if p.tableAttrs != nil {
		p.append(&html.Token{
			Type: html.StartTagToken, DataAtom: atom.Table, Data: "table", Attr: p.tableAttrs})
	} else {
		p.append(startTable)
	}
	p.append(startTr)
	row, col := 0, 0
	nlCount := 0
	var styles []*html.Token
	for tok := p.input[p.pos-1]; tok.Type != EOF; tok = p.next() {
		if tok.Type != TD && tok.Type != EOF && tok.Type != NEWLINE {
			return ErrUnexpectedToken{tok}
		}
		if tok.Type == TD && row == 1 {
			if strings.HasPrefix(tok.Lit, ":") && strings.HasSuffix(tok.Lit, ":") {
				styles = append(styles, startTdC)
			} else if strings.HasPrefix(tok.Lit, ":") {
				styles = append(styles, startTdL)
			} else if strings.HasSuffix(tok.Lit, ":") {
				styles = append(styles, startTdR)
			} else {
				styles = append(styles, startTd)
			}
		} else if tok.Type == TD {
			if row == 0 {
				p.append(startTh)
			} else {
				p.append(styles[col])
			}
			scanner := NewScanner(tok.Lit)
			for tok := scanner.Next(); tok.Type != EOF; tok = scanner.Next() {
				p.inlineMode = true
				p.consumeInline(tok)
				p.inlineMode = false
			}
			if row == 0 {
				p.append(endTh)
			} else {
				p.append(endTd)
			}
			col++
		}
		if tok.Type == NEWLINE {
			if row != 1 && nlCount == 0 && p.peek().Type != NEWLINE {
				p.append(endTr)
				p.append(startTr)
			}
			nlCount++
			row++
			col = 0
		} else {
			nlCount = 0
		}
		if nlCount == 2 {
			break
		}
	}
	p.append(endTr)
	p.append(endTable)
	p.inlineMode = false
	return nil
}
