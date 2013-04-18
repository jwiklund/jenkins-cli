package jenkins

import (
	"os"
	"testing"
)

func TestParseComputer(t *testing.T) {
	f, err := os.Open("computer_test.html")
	if err != nil {
		t.Fatalf("Could not open test file %s", err.Error())
	}
	defer f.Close()
	c, err := parseComputer(f)
	if err != nil {
		t.Fatalf("Could not parse computer %s", err.Error())
	}
	t.Logf("Ip %s", c.Ip)
	if "192.168.100.40" != c.Ip {
		t.Fatal("Wrong ip expected 192.168.100.40")
	}
}
