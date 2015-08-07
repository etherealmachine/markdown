package markdown

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type testCase struct {
	input string
	want  string
}

var testCases = []testCase{
	{
		"# Some header",
		"<h1>Some header</h1>",
	},
	{
		"## Some header",
		"<h2>Some header</h2>",
	},
	{
		"[a link](www.example.com)",
		"<p>\n\t<a href=\"www.example.com\">a link</a>\n</p>",
	},
	{
		"# [A link in a header](www.example.com)",
		"<h1>\n\t<a href=\"www.example.com\">A link in a header</a>\n</h1>",
	},
	{
		"![An img](www.example.com)",
		"<p>\n\t<img alt=\"An img\" src=\"www.example.com\"/>\n</p>",
	},
	{
		"A paragraph of text\npossibly on multiple\nlines.",
		"<p>A paragraph of text\npossibly on multiple\nlines.</p>",
	},
	{
		"<a href=\"www.example.com\" rel=\"nofollow\">Some HTML</a>",
		"<p>\n\t<a href=\"www.example.com\" rel=\"nofollow\">Some HTML</a>\n</p>",
	},
	{
		"`some code`",
		"<p>\n\t<code>some code</code>\n</p>",
	},
	{
		"```javascript\nA block of code\n```",
		"<pre>\n\t<code class=\"javascript\">\nA block of code\n</code>\n</pre>",
	},
}

func TestMarkdown(t *testing.T) {
	files, err := ioutil.ReadDir("examples")
	if err != nil {
		t.Fatal(err)
	}
	examples := make(map[int]*testCase)
	for _, f := range files {
		buf, err := ioutil.ReadFile(filepath.Join("examples", f.Name()))
		if err != nil {
			t.Error(err)
			continue
		}
		ext := filepath.Ext(f.Name())
		i, err := strconv.Atoi(
			strings.TrimPrefix(strings.TrimSuffix(f.Name(), ext), "example"))
		if err != nil {
			t.Error(err)
			continue
		}
		if examples[i] == nil {
			examples[i] = &testCase{}
		}
		if ext == ".md" {
			examples[i].input = string(buf)
		} else if ext == ".html" {
			examples[i].want = string(buf)
		}
	}
	for i, c := range examples {
		if i == 2 {
			_ = "breakpoint"
		}
		if got := Markdown(c.input); got != c.want {
			if err := ioutil.WriteFile(fmt.Sprintf("example%d.out", i), []byte(got), 0644); err != nil {
				t.Errorf("error writing example%d.out", i)
			} else {
				t.Errorf("see example%d.out", i)
			}
		}
	}
	for _, c := range testCases {
		if got := Markdown(c.input); got != c.want {
			t.Errorf("got\n%s\nwant\n%s", got, c.want)
		}
	}
}
