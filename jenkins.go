package jenkins

import (
	"bufio"
	"errors"
	"net/http"
	"os"
	"strings"
)

type Jenkins interface {
	Builds() ([]Build, error)
	NodeInfo(node string) (NodeInfo, error)
}

type Build struct {
	Node string
	// node_url string always /computer/$name
	// state can be offline
	Build string
}

func (b Build) String() string {
	if b.Build == "" {
		return b.Node
	}
	return b.Node + " building " + b.Build
}

type NodeInfo struct {
	Node string
	Ip   string
}

func New(url string) Jenkins {
	return jenkins(url)
}

func NewFromConfig() (Jenkins, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return nil, errors.New("HOME not set")
	}
	f, err := os.Open(home + "/.jenkins")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	l, err := buf.ReadString('\n')
	if err != nil {
		return nil, err
	}
	return jenkins(string(strings.TrimSpace(l))), nil
}

type jenkins string

func (j jenkins) url() string {
	str := string(j)
	ind := strings.Index(str, "@")
	if ind == -1 {
		return str
	}
	url := str[ind+1:]
	ind = strings.Index(str[0:ind], "://")
	if ind != -1 {
		url = str[0:ind+3] + url
	}
	return url
}

func (j jenkins) auth() (string, string, error) {
	str := string(j)
	ind := strings.Index(str, "@")
	if ind == -1 {
		return "", "", errors.New("no auth supplied")
	}
	auth := str[0:ind]
	ind = strings.Index(auth, "://")
	if ind != -1 {
		auth = auth[ind+3:]
	}
	ind = strings.Index(auth, ":")
	if ind == -1 {
		return "", "", errors.New("invalid auth supplied: " + auth)
	}
	return auth[0:ind], auth[ind+1:], nil
}

func (j jenkins) Builds() ([]Build, error) {
	resp, err := http.Get(j.url() + "/ajaxExecutors")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return parseExecutors(resp.Body)
}

func (j jenkins) NodeInfo(node string) (NodeInfo, error) {
	user, pass, err := j.auth()
	if err != nil {
		return NodeInfo{}, err
	}
	req, err := http.NewRequest("GET", j.url()+"/computer/"+node+"/log", nil)
	if err != nil {
		return NodeInfo{}, err
	}
	req.SetBasicAuth(user, pass)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return NodeInfo{}, err
	}
	defer resp.Body.Close()
	return parseComputer(resp.Body)
}
