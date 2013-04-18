package jenkins

import (
	"code.google.com/p/go-html-transform/h5"
	"code.google.com/p/go.net/html"
	"code.google.com/p/go.net/html/atom"
	"errors"
	"fmt"
	"io"
)

func parseGetChild(node *html.Node, childType atom.Atom, count int) (*html.Node, error) {
	if node.FirstChild == nil {
		return nil, errors.New("No children")
	}
	if node.LastChild == node.FirstChild {
		if node.FirstChild.DataAtom != childType {
			return parseGetChild(node.FirstChild, childType, count)
		}
	}
	child := node.FirstChild
	for {
		if child.DataAtom == childType {
			count = count - 1
			if count == 0 {
				return child, nil
			}
		}
		if child == node.LastChild {
			return nil, errors.New("Atom not found " + childType.String() + " in " + h5.NewTree(node).String())
		}
		child = child.NextSibling
	}
	return nil, errors.New("Atom not found " + childType.String() + " in " + h5.NewTree(node).String())
}

func parsePrint(n *html.Node) {
	fmt.Println(h5.NewTree(n).String())
}
func parseExecutors(rdr io.Reader) ([]Build, error) {
	tree, err := h5.New(rdr)
	if err != nil {
		return nil, err
	}
	body, err := parseGetChild(tree.Top(), atom.Body, 1)
	if err != nil {
		return nil, err
	}
	table, err := parseGetChild(body, atom.Table, 1)
	if err != nil {
		return nil, err
	}
	tbody, err := parseGetChild(table, atom.Tbody, 2)
	if err != nil {
		return nil, err
	}
	tr := tbody.FirstChild
	var builds []Build
	for {
		th, err := parseGetChild(tr, atom.Th, 1)
		if err == nil {
			nameLink, err := parseGetChild(th, atom.A, 1)
			if err != nil {
				fmt.Println("link not found")
				return nil, err
			}
			if tr.NextSibling == nil {
				builds = append(builds, Build{nameLink.FirstChild.Data, ""})
				break
			}
			tr = tr.NextSibling
			_, err = parseGetChild(tr, atom.Th, 1)
			if err == nil {
				// no data row
				builds = append(builds, Build{nameLink.FirstChild.Data, ""})
				continue
			}
			if tr.FirstChild == nil || tr.FirstChild.NextSibling == nil {
				return nil, errors.New("Build without div")
			}
			buildTd := tr.FirstChild.NextSibling
			if buildTd.DataAtom != atom.Td {
				return nil, errors.New("Expected td but got " + h5.NewTree(buildTd).String())
			}
			buildDiv, err := parseGetChild(buildTd, atom.Div, 1)
			if err != nil {
				// empty data row
				builds = append(builds, Build{nameLink.FirstChild.Data, ""})
			} else {
				build, err := parseGetChild(buildDiv, atom.A, 1)
				if err != nil {
					return nil, err
				}
				builds = append(builds, Build{nameLink.FirstChild.Data, build.FirstChild.Data})
			}
		}
		if tr == tbody.LastChild {
			break
		}
		tr = tr.NextSibling
	}

	return builds, nil
}
