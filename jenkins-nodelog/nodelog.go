package main

import (
	"flag"
	"fmt"
)

var storeLocation = "data"

func SaveJobs(filter string, args []string) {
	if len(args) == 0 {
		jobs, err := GetJobs(filter)
		if err != nil {
			fmt.Println("Could not list jobs ", err)
			return
		}
		for _, job := range jobs {
			fmt.Println(job.Name)
		}
		return
	}
	jobs, err := GetJobs("")
	if err != nil {
		fmt.Println("Could not list jobs ", err)
		return
	}
	store, err := OpenStore(storeLocation)
	if err != nil {
		fmt.Println("Could not open store ", err)
		return
	}
	defer store.Close()
	for _, arg := range args {
		found := false
		for _, job := range jobs {
			if arg == job.Name {
				found = true
				_, err = store.GetJob(job.Name)
				if err != nil {
					err = store.PutJob(job)
					fmt.Println("Added job "+job.Name, err)
				} else {
					fmt.Println("Job already added " + job.Name)
				}
			}
		}
		if !found {
			fmt.Println("Job not found " + arg)
		}
	}
}

func GetBuilds(jobNames []string) {
	jobs, err := GetJobs("")
	if err != nil {
		fmt.Println("Could not list jobs ", err)
		return
	}
	for _, jobName := range jobNames {
		for _, job := range jobs {
			if job.Name == jobName {
				builds, err := job.GetBuilds()
				if err != nil {
					fmt.Println("Could not get builds for "+jobName+" ", err)
					continue
				}
				for _, build := range builds {
					fmt.Println(build.String())
				}
			}
		}
	}
}

func RefreshBuilds() {
	store, err := OpenStore(storeLocation)
	if err != nil {
		fmt.Println("Could not open store ", err)
		return
	}
	defer store.Close()
	jobs, err := store.GetJobs()
	for _, job := range jobs {
		builds, err := job.GetBuilds()
		if err != nil {
			fmt.Println("Could not refresh "+job.Name+", ", err)
			continue
		}
		current, err := store.GetBuilds(job.Name)
		if err != nil {
			fmt.Println("Could not get builds from store ", err)
			continue
		}
		existing := make(map[int]bool)
		for _, build := range current {
			existing[build.Number] = true
		}
		for _, build := range builds {
			_, ok := existing[build.Number]
			if !ok {
				err = store.PutBuild(build)
				fmt.Println("Added build "+build.String(), err)
			}
		}
	}
}

func main() {
	save := flag.Bool("store", false, "Store new jobs")
	refresh := flag.Bool("refresh", false, "Update job builds")
	builds := flag.Bool("builds", false, "Get builds for job")
	filter := flag.String("filter", "", "Jobs list filter")
	flag.Parse()
	if *save {
		SaveJobs(*filter, flag.Args())
	} else if *refresh {
		RefreshBuilds()
	} else if *builds {
		GetBuilds(flag.Args())
	} else {
		fmt.Println("Don't know what to do (run -help)")
	}
}
