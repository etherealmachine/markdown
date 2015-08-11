package main

import (
	"flag"
	"fmt"
	"github.com/etherealmachine/markdown"
	"io/ioutil"
	"os"
)

var scan = flag.Bool("scan", false, "Print the lexical analysis.")

func main() {
	flag.Parse()
	buf, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	if *scan {
		scanner := markdown.NewScanner(string(buf))
		for tok := scanner.Next(); tok.Type != markdown.EOF; tok = scanner.Next() {
			fmt.Println(tok)
		}
	} else {
		fmt.Println(markdown.Markdown(string(buf)))
	}
}
