package testutils

import "golang.org/x/net/html"

func AttrMap(node *html.Node) map[string]string {
	m := make(map[string]string, len(node.Attr))
	for _, attr := range node.Attr {
		m[attr.Key] = attr.Val
	}
	return m
}
