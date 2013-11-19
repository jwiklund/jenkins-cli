package jenkins

import (
	"os"
	"testing"
)

func TestConfigParse(t *testing.T) {
	f, err := os.Open("config_test.xml")
	if err != nil {
		t.Fatalf("Could not open test file", err)
		return
	}
	defer f.Close()
	lookup, err := parseConfig(f)
	if err != nil {
		t.Fatalf("Could not parse config", err)
	}
	t.Logf("Assigned node is '%s'", lookup["assignedNode"])
	if "10.0_websphere-6.1_oracle-11.2_jdk-1.5_linux-2.6" != lookup["assignedNode"] {
		t.Error("Wrong assignedNode")
	}
}
