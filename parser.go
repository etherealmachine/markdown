package markdown

import (
	"bytes"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"strings"
)

type scanner interface {
	Next() (Token, string)
}

type token struct {
	tok Token
	lit string
}

type ErrUnexpectedToken struct {
	token Token
}

func (e ErrUnexpectedToken) Error() string {
	return e.token.String()
}

type Parser struct {
	pos        int
	input      []*token
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
	for tok, lit := scanner.Next(); tok != EOF; tok, lit = scanner.Next() {
		p.input = append(p.input, &token{tok, lit})
	}
	for tok, lit := p.next(); tok != EOF; tok, lit = p.next() {
		p.consume(tok, lit)
	}
	if p.inlineMode {
		p.append(endP)
		p.inlineMode = false
	}
}

func (p *Parser) consume(tok Token, lit string) {
	p.save()
	var err error
	switch tok {
	case H1, H2, H3, H4, H5, H6:
		p.parseHeader(tok)
	case EM:
		err = p.parseEm()
	case STRONG:
		err = p.parseStrong()
	case NEWLINE:
		p.parseNewline()
	case TEXT:
		p.parseText(lit)
	case LINK_TEXT:
		err = p.parseLink(lit)
	case IMG_ALT:
		err = p.parseImg(lit)
	case CODE:
		err = p.parseCode()
	case CODE_BLOCK:
		err = p.parseCodeBlock()
	case HTML_START:
		err = p.parseHTMLStart(lit)
	case HTML_END_TAG:
		p.parseHTMLEnd(lit)
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
	t, lit := p.next()
	if t != tok {
		return "", ErrUnexpectedToken{t}
	}
	return lit, nil
}

func (p *Parser) save() {
	p.saved.pos = p.pos - 1
	p.saved.tokenCount = len(p.tokens)
	p.saved.inlineMode = p.inlineMode
}

func (p *Parser) revert() {
	var buf bytes.Buffer
	for i := p.saved.pos; i <= p.pos && i < len(p.input); i++ {
		buf.WriteString(p.input[i].lit)
	}
	p.tokens = p.tokens[:p.saved.tokenCount]
	p.inlineMode = p.saved.inlineMode
	p.parseText(buf.String())
}

func (p *Parser) next() (Token, string) {
	if p.pos >= len(p.input) {
		return EOF, "EOF"
	}
	p.pos++
	return p.input[p.pos-1].tok, p.input[p.pos-1].lit
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
	for tok, lit := p.next(); tok != NEWLINE && tok != EOF; tok, lit = p.next() {
		p.consume(tok, lit)
	}
	p.append(hEndTag[headerToken])
	p.inlineMode = false
}

func (p *Parser) parseEm() error {
	p.inline()
	p.append(startEm)
	var buf bytes.Buffer
	for tok, lit := p.next(); tok != EM; tok, lit = p.next() {
		if tok == NEWLINE && tok == EOF {
			return ErrUnexpectedToken{tok}
		}
		buf.WriteString(lit)
	}
	p.append(text(buf.String()))
	p.append(endEm)
	return nil
}

func (p *Parser) parseStrong() error {
	p.inline()
	p.append(startStrong)
	var buf bytes.Buffer
	for tok, lit := p.next(); tok != STRONG; tok, lit = p.next() {
		if tok == NEWLINE && tok == EOF {
			return ErrUnexpectedToken{tok}
		}
		buf.WriteString(lit)
	}
	p.append(text(buf.String()))
	p.append(endStrong)
	return nil
}

func (p *Parser) parseNewline() {
	tok, lit := p.next()
	if tok == NEWLINE {
		p.block()
	} else {
		p.tokens = append(p.tokens, text("\n"))
		p.consume(tok, lit)
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
	tok, lit := p.next()
	if tok == TEXT {
		p.tokens = append(p.tokens, &html.Token{
			Type:     html.StartTagToken,
			DataAtom: atom.Code,
			Data:     "code",
			Attr: []html.Attribute{
				{Key: "class", Val: lit},
			},
		})
	} else if tok == NEWLINE {
		p.append(startCode)
	} else {
		return ErrUnexpectedToken{tok}
	}
	for tok, lit = p.next(); tok != CODE_BLOCK; tok, lit = p.next() {
		if tok == EOF {
			return ErrUnexpectedToken{tok}
		}
		buf.WriteString(lit)
	}
	code := strings.Trim(buf.String(), "\n")
	p.append(text("\n" + code + "\n"))
	p.append(endCode)
	p.append(endPre)
	return nil
}

func (p *Parser) parseHTMLStart(start string) error {
	text, err := p.expect(TEXT)
	if err != nil {
		return err
	}
	end, err := p.expect(HTML_END)
	if err != nil {
		return err
	}
	tt := html.NewTokenizer(strings.NewReader(start + text + end))
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
		tok, lit := p.next()
		switch tok {
		case EOF:
			break
		case NEWLINE:
			tok, lit = p.next()
			if tok == ORDERED_LIST {
				p.append(endLi)
				p.append(startLi)
			} else if tok == EOF || tok == NEWLINE {
				goto EmitEnd
			} else {
				p.consume(tok, lit)
			}
		default:
			p.consume(tok, lit)
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
		tok, lit := p.next()
		switch tok {
		case EOF:
			break
		case NEWLINE:
			tok, lit = p.next()
			if tok == UNORDERED_LIST {
				p.append(endLi)
				p.append(startLi)
			} else if tok == EOF || tok == NEWLINE {
				goto EmitEnd
			} else {
				p.consume(tok, lit)
			}
		default:
			p.consume(tok, lit)
		}
	}
EmitEnd:
	p.append(endLi)
	p.append(endUl)
	p.inlineMode = false
}
