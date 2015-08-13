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
		"<p><a href=\"www.example.com\">a link</a></p>",
	},
	{
		"# [A link in a header](www.example.com)",
		"<h1><a href=\"www.example.com\">A link in a header</a></h1>",
	},
	{
		"![An img](www.example.com)",
		"<p><img alt=\"An img\" src=\"www.example.com\"/></p>",
	},
	{
		"A paragraph of text\npossibly on multiple\nlines.",
		"<p>A paragraph of text\npossibly on multiple\nlines.</p>",
	},
	{
		"<p><a href=\"www.example.com\" rel=\"nofollow\">Some HTML</a></p>",
		"<p><a href=\"www.example.com\" rel=\"nofollow\">Some HTML</a></p>",
	},
	{
		"`some code`",
		"<p><code>some code</code></p>",
	},
	{
		"```javascript\nA block of code\n```",
		"<pre><code class=\"javascript\">\nA block of code\n</code></pre>",
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
