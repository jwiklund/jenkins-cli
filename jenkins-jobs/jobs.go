package main

import (
	"flag"
	"fmt"
	"github.com/jwiklund/jenkins"
	"regexp"
	"strings"
)

func nameMatch(name string, match []string) bool {
	for _, part := range match {
		ind := strings.Index(strings.ToLower(name), strings.ToLower(part))
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
	p := flag.String("pattern", "", "Pattern to restrict which jobs to report")
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Println("No fields, fields to fetch argument is required")
		return
	}
	var pattern *regexp.Regexp
	if *p != "" {
		pattern = regexp.MustCompile(".*" + *p + ".*")
	}
	jobs, err := j.Jobs()
	if err != nil {
		fmt.Println("Could not list jobs " + err.Error())
		return
	}
	for _, job := range jobs {
		if *p != "" && !pattern.MatchString(job.Name) {
			continue
		}
		cfg, err := j.JobInfo(job.Name)
		if err != nil {
			fmt.Printf("Could not fetch job %s due to %s", job, err.Error())
			continue
		}

		for _, name := range flag.Args() {
			fmt.Printf("%s\t%s\t%s\n", job.Name, name, cfg[name])
		}
	}
}
