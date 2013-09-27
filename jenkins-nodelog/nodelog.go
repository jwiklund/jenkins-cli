package main

import (
	"flag"
	"fmt"
	"regexp"
)

var storeLocation = "data"

func ListJobs(filter string) {
	jobs, err := GetJobs(filter)
	if err != nil {
		fmt.Println("Could not list jobs ", err)
		return
	}
	for _, job := range jobs {
		fmt.Println(job.Name)
	}
}

func SaveJobs(args []string) {
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

func RefreshBuilds(update bool) {
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
				err = store.InsertBuild(build)
				fmt.Println("Added build "+build.String(), err)
			} else if update {
				err = store.UpdateBuild(build)
				fmt.Println("Updated build "+build.String(), err)
			}
		}
	}
}

func ExportBuilds(filter string) {
	store, err := OpenStore(storeLocation)
	if err != nil {
		fmt.Println("Could not open store ", err)
		return
	}
	defer store.Close()
	jobs, err := store.GetJobs()
	if err != nil {
		fmt.Println("Could not load jobs ", err)
		return
	}
	fmt.Println("Job,Number,Host,Duration,Start,Result")
	for _, job := range jobs {
		if filter != "" {
			matched, err := regexp.MatchString(filter, job.Name)
			if err != nil {
				panic(err)
			}
			if !matched {
				continue
			}
		}
		builds, err := store.GetBuilds(job.Name)
		if err != nil {
			fmt.Println("Could not load builds ", err)
			return
		}
		for _, build := range builds {
			fmt.Printf("%s,%d,%s,%d,%d,%s\n", build.Job, build.Number, build.Host, build.Duration, build.Start, build.Result)
		}
	}
}

func main() {
	list := flag.Bool("list", false, "List existing job (possibly filtered)")
	save := flag.Bool("store", false, "Store new jobs")
	refresh := flag.Bool("refresh", false, "Update job builds")
	update := flag.Bool("update", false, "Update existing jobs")
	builds := flag.Bool("builds", false, "Get builds for job")
	export := flag.Bool("export", false, "Export to CSV (possibly filtered)")
	filter := flag.String("filter", "", "Jobs list filter (a regular expression)")
	flag.Parse()
	if *list {
		ListJobs(*filter)
	} else if *save {
		SaveJobs(flag.Args())
	} else if *refresh {
		RefreshBuilds(*update)
	} else if *builds {
		GetBuilds(flag.Args())
	} else if *export {
		ExportBuilds(*filter)
	} else {
		fmt.Println("Don't know what to do (run -help)")
	}
}
