package markdown

func Markdown(input string) string {
	return PrettyPrint(Parse(input))
}
