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
		err = p.parseEm()
	case STRONG:
		err = p.parseStrong()
	case NEWLINE:
		p.parseNewline()
	case TEXT:
		p.parseText(tok.Lit)
	case LINK_TEXT:
		err = p.parseLink(tok.Lit)
	case IMG_ALT:
		err = p.parseImg(tok.Lit)
	case CODE:
		err = p.parseCode()
	case CODE_BLOCK:
		err = p.parseCodeBlock()
	case HTML_START:
		err = p.parseHTMLStart(tok.Lit)
	case HTML_END_TAG:
		p.parseHTMLEnd(tok.Lit)
	case ORDERED_LIST:
		p.parseOrderedList()
	case UNORDERED_LIST:
		p.parseUnorderedList()
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
	for i := p.saved.pos; i <= p.pos && i < len(p.input); i++ {
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

func (p *Parser) append(tok *html.Token) {
	p.tokens = append(p.tokens, tok)
}

func (p *Parser) prev() *html.Token {
	if len(p.tokens) == 0 {
		return nil
	}
	return p.tokens[len(p.tokens)-1]
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
	for tok := p.next(); tok.Tok != NEWLINE && tok.Tok != EOF; tok = p.next() {
		p.consume(tok)
	}
	p.append(hEndTag[headerToken])
	p.inlineMode = false
}

func (p *Parser) parseEm() error {
	p.inline()
	p.append(startEm)
	var buf bytes.Buffer
	for tok := p.next(); tok.Tok != EM; tok = p.next() {
		if tok.Tok == NEWLINE || tok.Tok == EOF {
			return ErrUnexpectedToken{tok}
		}
		buf.WriteString(tok.Lit)
	}
	p.append(text(buf.String()))
	p.append(endEm)
	return nil
}

func (p *Parser) parseStrong() error {
	p.inline()
	p.append(startStrong)
	var buf bytes.Buffer
	for tok := p.next(); tok.Tok != STRONG; tok = p.next() {
		if tok.Tok == NEWLINE || tok.Tok == EOF {
			return ErrUnexpectedToken{tok}
		}
		buf.WriteString(tok.Lit)
	}
	p.append(text(buf.String()))
	p.append(endStrong)
	return nil
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

func (p *Parser) parseCode() error {
	p.inline()
	s, err := p.expect(TEXT)
	if err != nil {
		return err
	}
	if _, err := p.expect(CODE); err != nil {
		return err
	}
	p.append(startCode)
	p.append(text(s))
	p.append(endCode)
	return nil
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

func (p *Parser) parseHTMLStart(start string) error {
	var buf bytes.Buffer
	buf.WriteRune('<')
	for tok := p.next(); tok.Tok != HTML_END; tok = p.next() {
		buf.WriteString(tok.Raw)
	}
	buf.WriteRune('>')
	tt := html.NewTokenizer(&buf)
	tt.Next()
	tok := tt.Token()
	if !p.inlineMode && !blockTag[tok.DataAtom] {
		p.append(startP)
		p.inlineMode = true
	}
	p.append(&tok)
	return nil
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
	p.append(startOl)
	p.append(startLi)
	p.inlineMode = true
	for {
		tok := p.next()
		switch tok.Tok {
		case EOF:
			break
		case NEWLINE:
			tok = p.next()
			if tok.Tok == ORDERED_LIST {
				p.append(endLi)
				p.append(startLi)
			} else if tok.Tok == EOF || tok.Tok == NEWLINE {
				goto EmitEnd
			} else {
				p.consume(tok)
			}
		default:
			p.consume(tok)
		}
	}
EmitEnd:
	p.append(endLi)
	p.append(endOl)
	p.inlineMode = false
}

func (p *Parser) parseUnorderedList() {
	p.append(startUl)
	p.append(startLi)
	p.inlineMode = true
	for {
		tok := p.next()
		switch tok.Tok {
		case EOF:
			break
		case NEWLINE:
			tok = p.next()
			if tok.Tok == UNORDERED_LIST {
				p.append(endLi)
				p.append(startLi)
			} else if tok.Tok == EOF || tok.Tok == NEWLINE {
				goto EmitEnd
			} else {
				p.consume(tok)
			}
		default:
			p.consume(tok)
		}
	}
EmitEnd:
	p.append(endLi)
	p.append(endUl)
	p.inlineMode = false
}
