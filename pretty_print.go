package markdown

import (
	"bytes"
	"golang.org/x/net/html"
)

func PrettyPrint(tokens []*html.Token) string {
	var pretty bytes.Buffer
	depth := 0
	var prev *html.Token
	for i, token := range tokens {
		tokenString := token.String()
		if !inline(prev) && inline(token) && tokenString == "\n" {
			continue
		}

		if token.Type == html.EndTagToken && blockTag[token.DataAtom] {
			depth--
		}

		if (!inline(token) && !inline(prev)) ||
			(i > 0 && token.Type == html.StartTagToken && blockTag[token.DataAtom]) {
			pretty.WriteByte('\n')
			for i := 0; i < depth; i++ {
				pretty.WriteByte('\t')
			}
		}

		pretty.WriteString(tokenString)

		if token.Type == html.StartTagToken && blockTag[token.DataAtom] {
			depth++
		}
		prev = token
	}
	return pretty.String()
}
