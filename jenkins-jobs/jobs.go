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
	l := flag.Bool("list", false, "List key names for job")
	la := flag.Bool("listall", false, "List key names for all jobs, not just the first that matches")
	flag.Parse()
	if (len(flag.Args()) != 0 && *l) || (len(flag.Args()) == 0 && !*l) {
		fmt.Println("Either specify fields to list or use -list to show field names")
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
		if *l {
			for name, _ := range cfg {
				fmt.Printf("%s\t%s\n", job.Name, name)
			}
			if !*la {
				return
			}
		} else {
			for _, name := range flag.Args() {
				fmt.Printf("%s\t%s\t%s\n", job.Name, name, cfg[name])
			}
		}

	}
}
