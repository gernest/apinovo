package apinovo

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

//TextFromHTML extracts texts from html document.
func TextFromHTML(input io.Reader) []string {
	doc, err := html.Parse(input)
	if err != nil {
		panic(err)
	}
	return getText(doc)
}

func getText(n *html.Node) (result []string) {
	if n.Type == html.TextNode {
		txt := strings.TrimSpace(n.Data)
		if txt != "" {
			result = append(result, txt)
		}
		return
	}
	for i := n.FirstChild; i != nil; i = i.NextSibling {
		result = append(result, getText(i)...)
	}
	return
}
