package jenkins

import (
	"errors"
	"io"
	"os"
)

func parseComputer(rdr io.Reader) (NodeInfo, error) {
	f, err := os.Create("computer_test.html")
	if err != nil {
		return NodeInfo{}, err
	}
	io.Copy(f, rdr)
	f.Close()
	return NodeInfo{}, errors.New("Not Implemented")
}
