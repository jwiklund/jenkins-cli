package jenkins

import (
	"os"
	"testing"
)

func checkBuild(t *testing.T, node, build string, actual Build) {
	if actual.node != node || actual.build != build {
		t.Fatal("Expected " + Build{node, build}.String() + " but got " + actual.String())
	}
}

func TestParseExecutors(t *testing.T) {
	f, err := os.Open("executors_test.html")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	executors, err := parseExecutors(f)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if len(executors) != 15 {
		t.Fatalf("Expected 15 builds but got %d", len(executors))
	}
	checkBuild(t, "dumslav", "", executors[0])
	checkBuild(t, "euca-jdk-1-6-linux-2-6-782", "VOID_Minutely", executors[2])
}
