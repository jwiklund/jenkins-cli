package jenkins

import (
	"net/http"
)

type Build struct {
	node string
	// node_url string always /computer/$name
	// state can be offline
	build string
}

func (b Build) String() string {
	if b.build == "" {
		return b.node
	}
	return b.node + " building " + b.build
}

type Jenkins interface {
	Builds() ([]Build, error)
}

func New(url string) Jenkins {
	return jenkins(url)
}

type jenkins string

func (j jenkins) Builds() ([]Build, error) {
	resp, err := http.Get(string(j) + "/ajaxExecutors")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return parseExecutors(resp.Body)
}
