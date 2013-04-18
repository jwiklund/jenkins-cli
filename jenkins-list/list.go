package main

import (
	"fmt"
	"github.com/jwiklund/jenkins"
)

func main() {
	j, err := jenkins.NewFromConfig()
	if err != nil {
		fmt.Println("Could not configure jenkins: " + err.Error())
		return
	}
	builds, err := j.Builds()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for _, value := range builds {
		fmt.Println(value.String())
	}
}
