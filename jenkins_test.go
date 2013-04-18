package jenkins

import (
	"testing"
)

func TestJenkinsUrl(t *testing.T) {
	j := jenkins("localhost/jenkins")
	t.Logf("Jenkins.url() %s", j.url())
	if "localhost/jenkins" != j.url() {
		t.Fatal("Wrong url")
	}
}

func TestJenkinsAuth(t *testing.T) {
	j := jenkins("http://username:passwd@localhost/jenkins")
	t.Logf("Jenkins.url() %s", j.url())
	if j.url() != "http://localhost/jenkins" {
		t.Fatal("Wrong url expected http://localhost/jenkins")
	}
	user, pass, err := j.auth()
	if err != nil {
		t.Fatalf("Error returned: %s", err.Error())
	}
	t.Logf("Jenkins.auth() %s, %s, nil", user, pass)
	if user != "username" || pass != "passwd" {
		t.Fatal("Wrong user/pass")
	}
}
