package markdown

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"strings"
)

type scanner interface {
	Next() *Tok
}

type ErrUnexpectedToken struct {
	tok *Tok
}

func (e ErrUnexpectedToken) Error() string {
	return fmt.Sprintf("%v", e.tok)
}

type Parser struct {
	pos        int
	input      []*Tok
	tokens     []*html.Token
	inlineMode bool
	saved      savePoint
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
	for tok := scanner.Next(); tok.Tok != EOF; tok = scanner.Next() {
		p.input = append(p.input, tok)
	}
	for tok := p.next(); tok.Tok != EOF; tok = p.next() {
		p.consume(tok)
	}
	if p.inlineMode {
		p.append(endP)
		p.inlineMode = false
	}
}

func (p *Parser) consume(tok *Tok) {
	p.save()
	var err error
	switch tok.Tok {
	case H1, H2, H3, H4, H5, H6:
		p.parseHeader(tok.Tok)
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
	case CODE_BLOCK:
		err = p.parseCodeBlock()
	case HTML_TAG:
		p.parseHTMLTag(tok.Lit)
	case ORDERED_LIST:
		p.parseOrderedList()
	case UNORDERED_LIST:
		p.parseUnorderedList()
	default:
		p.parseText(tok.Raw)
	}
	if err != nil {
		p.revert()
	}
}

func (p *Parser) expect(tok Token) (string, error) {
	t := p.next()
	if t.Tok != tok {
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

func (p *Parser) next() *Tok {
	if p.pos >= len(p.input) {
		return &Tok{EOF, "EOF", ""}
	}
	p.pos++
	return p.input[p.pos-1]
}

func (p *Parser) peek() *Tok {
	if p.pos >= len(p.input) {
		return &Tok{EOF, "EOF", ""}
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

func (p *Parser) parseHeader(headerToken Token) {
	p.append(hStartTag[headerToken])
	p.inlineMode = true
	for {
		next := p.peek()
		if next.Tok == EOF || next.Tok == NEWLINE {
			p.next()
			break
		}
		if next.Tok == ORDERED_LIST || next.Tok == UNORDERED_LIST {
			break
		}
		p.next()
		p.consume(next)
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
	tok := p.next()
	if tok.Tok == NEWLINE {
		p.block()
	} else {
		p.tokens = append(p.tokens, text("\n"))
		p.consume(tok)
	}
}

func (p *Parser) parseText(s string) {
	p.inline()
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
	if tok.Tok == TEXT {
		p.tokens = append(p.tokens, &html.Token{
			Type:     html.StartTagToken,
			DataAtom: atom.Code,
			Data:     "code",
			Attr: []html.Attribute{
				{Key: "class", Val: tok.Lit},
			},
		})
	} else if tok.Tok == NEWLINE {
		p.append(startCode)
	} else {
		return ErrUnexpectedToken{tok}
	}
	for tok = p.next(); tok.Tok != CODE_BLOCK; tok = p.next() {
		if tok.Tok == EOF {
			return ErrUnexpectedToken{tok}
		}
		buf.WriteString(tok.Raw)
	}
	code := strings.Trim(buf.String(), "\n")
	p.append(text("\n" + code + "\n"))
	p.append(endCode)
	p.append(endPre)
	return nil
}

func (p *Parser) parseHTMLTag(tag string) {
	tt := html.NewTokenizer(strings.NewReader(tag))
	tt.Next()
	tok := tt.Token()
	if !p.inlineMode && !blockTag[tok.DataAtom] {
		p.append(startP)
		p.inlineMode = true
	}
	p.append(&tok)
}

func (p *Parser) parseHTMLEnd(tag string) {
	tok := html.Token{
		Type:     html.EndTagToken,
		DataAtom: atom.Lookup([]byte(tag)),
		Data:     tag,
	}
	if p.inlineMode && !(tok.Type == html.EndTagToken && inlineTag[tok.DataAtom]) {
		p.append(endP)
		p.inlineMode = false
	}
	p.append(&tok)
}

func (p *Parser) parseOrderedList() {
	p.block()
	depth := 0
	p.inlineMode = true
	p.append(startOl)
	p.append(startLi)
	for tok := p.next(); tok.Tok != EOF && tok.Tok != NEWLINE; tok = p.next() {
		if tok.Tok == ORDERED_LIST {
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
			p.consume(tok)
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
	for tok := p.next(); tok.Tok != EOF && tok.Tok != NEWLINE; tok = p.next() {
		if tok.Tok == UNORDERED_LIST {
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
			p.consume(tok)
		}
	}
	for ; depth >= 0; depth-- {
		p.append(endLi)
		p.append(endUl)
	}
	p.inlineMode = false
}
