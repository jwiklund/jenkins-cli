package jenkins

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

func parseComputer(rdr io.Reader) (NodeInfo, error) {
	r := bufio.NewReader(rdr)
	for {
		l, err := r.ReadString('\n')
		if err != nil {
			return NodeInfo{}, err
		}
		if strings.Contains(l, "Connecting to ") {
			ip := l[len("Connecting to "):]
			ind := strings.Index(ip, " ")
			if ind > 0 {
				ip = ip[0:ind]
			}
			return NodeInfo{Node: "", Ip: ip}, nil
		}
	}
	return NodeInfo{}, errors.New("Not Implemented")
}
