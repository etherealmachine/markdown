package markdown

import (
	"bytes"
	"golang.org/x/net/html"
)

func PrettyPrint(tokens []*html.Token) string {
	var pretty bytes.Buffer
	depth := 0
	prevTt := html.CommentToken
	for _, token := range tokens {
		tt := token.Type
		tokenString := token.String()
		if prevTt != html.TextToken && tt == html.TextToken && tokenString == "\n" {
			continue
		}

		if tt == html.EndTagToken {
			depth--
		}

		if prevTt != html.CommentToken && tt != html.TextToken && prevTt != html.TextToken {
			pretty.WriteByte('\n')
			for i := 0; i < depth; i++ {
				pretty.WriteByte('\t')
			}
		}

		pretty.WriteString(tokenString)

		if tt == html.StartTagToken {
			depth++
		}
		prevTt = tt
	}
	return pretty.String()
}
