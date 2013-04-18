package main

import (
	"flag"
	"fmt"
	"github.com/jwiklund/jenkins"
	"strings"
)

func nameMatch(name string, match []string) bool {
	for _, part := range match {
		ind := strings.Index(name, part)
		if ind == -1 {
			return false
		}
	}
	return true
}

func main() {
	j, err := jenkins.NewFromConfig()
	if err != nil {
		fmt.Println("Could not configure jenkins: " + err.Error())
		return
	}
	flag.Parse()
	builds, err := j.Builds()
	if err != nil {
		fmt.Println("Could not fetch nodes: " + err.Error())
		return
	}
	for _, build := range builds {
		if nameMatch(build.Node, flag.Args()) {
			info, err := j.NodeInfo(build.Node)
			if err != nil {
				fmt.Println("Could not get info about " + build.Node + ": " + err.Error())
			} else {
				fmt.Println(build.Node + " ip " + info.Ip)
			}
		}
	}
}
