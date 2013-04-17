package jenkins

import (
	"code.google.com/p/go-html-transform/h5"
	"code.google.com/p/go.net/html"
	"code.google.com/p/go.net/html/atom"
	"io"
	"strings"
)

func attrs(attributes []html.Attribute) string {
	res := ""
	for _, a := range attributes {
		res = res + a.Key + "=" + a.Val + ";"
	}
	return res
}

func hrefToName(attributes []html.Attribute, delimiter string) string {
	for _, a := range attributes {
		if a.Key == "href" {
			ind := strings.Index(a.Val, delimiter)
			if ind > 0 {
				res := a.Val[ind+len(delimiter):]
				if res[len(res)-1] == '/' {
					res = res[0 : len(res)-1]
				}
				return res
			}
		}
	}
	return ""
}

func parseExecutors(rdr io.Reader) ([]Build, error) {
	h5, err := h5.New(rdr)
	if err != nil {
		return nil, err
	}
	state := "none"
	node := ""
	build := ""
	var builds []Build
	save := func() {
		builds = append(builds, Build{node, build})
	}
	h5.Walk(func(n *html.Node) {
		if n.DataAtom == atom.Th {
			if node != "" {
				save()
			}
			state = "node"
		}
		if n.DataAtom == atom.Div {
			state = "build"
		}
		if n.DataAtom == atom.A {
			if state == "node" {
				node = hrefToName(n.Attr, "/computer/")
				state = "none"
			} else if state == "build" {
				build = hrefToName(n.Attr, "/job/")
			}
		}
	})
	save()
	return builds, nil
}
