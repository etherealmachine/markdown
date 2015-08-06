package markdown

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"strings"
)

type scanner interface {
	Next() (Token, string)
}

type Parser struct {
	scanner    scanner
	tokens     []*html.Token
	inlineMode bool
}

func Parse(input string) []*html.Token {
	p := &Parser{
		scanner: NewScanner(input),
	}
	p.parse()
	return p.tokens
}

func (p *Parser) parse() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()
	_ = "breakpoint"
	for tok, lit := p.scanner.Next(); tok != EOF; tok, lit = p.scanner.Next() {
		p.consume(tok, lit)
	}
	if p.inlineMode {
		p.tokens = append(p.tokens, endP)
		p.inlineMode = false
	}
}

func (p *Parser) consume(tok Token, lit string) {
	switch tok {
	case H1, H2, H3, H4, H5, H6:
		p.parseHeader(tok)
	case EM:
		p.parseEm()
	case STRONG:
		p.parseStrong()
	case NEWLINE:
		p.parseNewline()
	case TEXT:
		p.parseText(lit)
	case LINK_TEXT:
		p.parseLink(lit)
	case IMG_ALT:
		p.parseImg(lit)
	case CODE:
		p.parseCode()
	case CODE_BLOCK:
		p.parseCodeBlock()
	case HTML_START:
		p.parseHTMLStart(lit)
	case HTML_END_TAG:
		p.parseHTMLEnd(lit)
	case ORDERED_LIST:
		p.parseOrderedList()
	case UNORDERED_LIST:
		p.parseUnorderedList()
	default:
		panic(fmt.Sprintf("unexpected token %s", tok))
	}
}

func (p *Parser) expect(tok Token) string {
	t, lit := p.scanner.Next()
	if t != tok {
		panic(fmt.Sprintf("unexpected token %s", t))
	}
	return lit
}

func (p *Parser) prev() *html.Token {
	if len(p.tokens) == 0 {
		return nil
	}
	return p.tokens[len(p.tokens)-1]
}

func (p *Parser) parseHeader(headerToken Token) {
	p.tokens = append(p.tokens, hStartTag[headerToken])
	p.inlineMode = true
	for {
		tok, lit := p.scanner.Next()
		if tok == NEWLINE || tok == EOF {
			break
		}
		p.consume(tok, lit)
	}
	p.tokens = append(p.tokens, hEndTag[headerToken])
	p.inlineMode = false
}

func (p *Parser) parseEm() {
	if !p.inlineMode {
		p.tokens = append(p.tokens, startP)
		p.inlineMode = true
	}
	p.tokens = append(p.tokens, startEm)
	for {
		tok, lit := p.scanner.Next()
		if tok == NEWLINE || tok == EOF || tok == EM {
			break
		}
		p.consume(tok, lit)
	}
	p.tokens = append(p.tokens, endEm)
}

func (p *Parser) parseStrong() {
	if !p.inlineMode {
		p.tokens = append(p.tokens, startP)
		p.inlineMode = true
	}
	p.tokens = append(p.tokens, startStrong)
	for {
		tok, lit := p.scanner.Next()
		if tok == NEWLINE || tok == EOF || tok == STRONG {
			break
		}
		p.consume(tok, lit)
	}
	p.tokens = append(p.tokens, endStrong)
}

func (p *Parser) parseNewline() {
	tok, lit := p.scanner.Next()
	if tok == NEWLINE {
		if p.inlineMode {
			p.tokens = append(p.tokens, endP)
			p.inlineMode = false
		}
	} else {
		p.tokens = append(p.tokens, text("\n"))
		p.consume(tok, lit)
	}
}

func (p *Parser) parseText(s string) {
	if !p.inlineMode {
		p.tokens = append(p.tokens, startP)
		p.inlineMode = true
	}
	p.tokens = append(p.tokens, text(s))
}

func (p *Parser) parseLink(s string) {
	if !p.inlineMode {
		p.tokens = append(p.tokens, startP)
		p.inlineMode = true
	}
	href := p.expect(HREF)
	p.tokens = append(p.tokens, &html.Token{
		Type:     html.StartTagToken,
		DataAtom: atom.A,
		Data:     "a",
		Attr: []html.Attribute{{
			Key: "href",
			Val: href,
		}},
	})
	p.tokens = append(p.tokens, text(s))
	p.tokens = append(p.tokens, endA)
}

func (p *Parser) parseImg(alt string) {
	if !p.inlineMode {
		p.tokens = append(p.tokens, startP)
		p.inlineMode = true
	}
	src := p.expect(HREF)
	p.tokens = append(p.tokens, &html.Token{
		Type:     html.SelfClosingTagToken,
		DataAtom: atom.Img,
		Data:     "img",
		Attr: []html.Attribute{
			{Key: "alt", Val: alt},
			{Key: "src", Val: src},
		},
	})
}

func (p *Parser) parseCode() {
	if !p.inlineMode {
		p.tokens = append(p.tokens, startP)
		p.inlineMode = true
	}

	s := p.expect(TEXT)
	p.expect(CODE)
	p.tokens = append(p.tokens, startCode)
	p.tokens = append(p.tokens, text(s))
	p.tokens = append(p.tokens, endCode)
}

func (p *Parser) parseCodeBlock() {
	p.tokens = append(p.tokens, startPre)
	var buf bytes.Buffer
	tok, lit := p.scanner.Next()
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
		p.tokens = append(p.tokens, startCode)
	} else {
		panic(fmt.Sprintf("unexpected token %s", tok))
	}
	for {
		tok, lit = p.scanner.Next()
		if tok == EOF {
			panic("unexpected EOF")
		}
		if tok == CODE_BLOCK {
			break
		}
		buf.WriteString(lit)
	}
	code := strings.Trim(buf.String(), "\n")
	p.tokens = append(p.tokens, text("\n"+code+"\n"))
	p.tokens = append(p.tokens, endCode)
	p.tokens = append(p.tokens, endPre)
}

func (p *Parser) parseHTMLStart(start string) {
	text := p.expect(TEXT)
	end := p.expect(HTML_END)
	tt := html.NewTokenizer(strings.NewReader(start + text + end))
	tt.Next()
	tok := tt.Token()
	if !p.inlineMode && !blockTag[tok.DataAtom] {
		p.tokens = append(p.tokens, startP)
		p.inlineMode = true
	}
	p.tokens = append(p.tokens, &tok)
}

func (p *Parser) parseHTMLEnd(tag string) {
	tok := html.Token{
		Type:     html.EndTagToken,
		DataAtom: atom.Lookup([]byte(tag)),
		Data:     tag,
	}
	if p.inlineMode && !(tok.Type == html.EndTagToken && inlineTag[tok.DataAtom]) {
		p.tokens = append(p.tokens, endP)
		p.inlineMode = false
	}
	p.tokens = append(p.tokens, &tok)
}

func (p *Parser) parseOrderedList() {
	p.tokens = append(p.tokens, startOl)
	p.tokens = append(p.tokens, startLi)
	p.inlineMode = true
	for {
		tok, lit := p.scanner.Next()
		switch tok {
		case EOF:
			break
		case NEWLINE:
			tok, lit = p.scanner.Next()
			if tok == ORDERED_LIST {
				p.tokens = append(p.tokens, endLi)
				p.tokens = append(p.tokens, startLi)
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
	p.tokens = append(p.tokens, endLi)
	p.tokens = append(p.tokens, endOl)
	p.inlineMode = false
}

func (p *Parser) parseUnorderedList() {
	p.tokens = append(p.tokens, startUl)
	p.tokens = append(p.tokens, startLi)
	p.inlineMode = true
	for {
		tok, lit := p.scanner.Next()
		switch tok {
		case EOF:
			break
		case NEWLINE:
			tok, lit = p.scanner.Next()
			if tok == UNORDERED_LIST {
				p.tokens = append(p.tokens, endLi)
				p.tokens = append(p.tokens, startLi)
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
	p.tokens = append(p.tokens, endLi)
	p.tokens = append(p.tokens, endUl)
	p.inlineMode = false
}
