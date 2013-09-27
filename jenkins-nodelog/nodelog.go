package main

import (
	"flag"
	"fmt"
	"regexp"
	"sync"
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

type PutReq struct {
	Build  Build
	Update bool
}

type GetReq struct {
	Name   string
	Builds chan []Build
}

func StoreHandler(store *Store, puts chan *PutReq, gets chan *GetReq, fini chan bool) {
	if err := store.Begin(); err != nil {
		fmt.Println("Begin failure", err)
	}
	doget := func(get *GetReq) {
		builds, err := store.GetBuilds(get.Name)
		if err != nil {
			fmt.Println("Failed getting "+get.Name, err)
		}
		get.Builds <- builds
	}
	doput := func(put *PutReq) {
		if put.Update {
			err := store.UpdateBuild(put.Build)
			fmt.Println("Updated build "+put.Build.String(), err)
		} else {
			err := store.InsertBuild(put.Build)
			fmt.Println("Added build "+put.Build.String(), err)
		}
	}
	run := true
	for run {
		select {
		case get, ok := <-gets:
			if ok {
				doget(get)
			} else {
				run = false
			}
		case put, ok := <-puts:
			if ok {
				doput(put)
			} else {
				run = false
			}
		}
	}
	for get := range gets {
		doget(get)
	}
	for put := range puts {
		doput(put)
	}
	if err := store.Commit(); err != nil {
		fmt.Println("Commit failure", err)
	}
	close(fini)
}

func RefreshBuilds(update bool) {
	store, err := OpenStore(storeLocation)
	if err != nil {
		fmt.Println("Could not open store ", err)
		return
	}
	defer store.Close()
	jobs, err := store.GetJobs()
	var wg sync.WaitGroup
	puts := make(chan *PutReq, 100)
	gets := make(chan *GetReq, 100)
	fini := make(chan bool)
	go StoreHandler(&store, puts, gets, fini)
	for _, job := range jobs {
		wg.Add(1)
		f := func(job Job) {
			defer wg.Done()
			builds, err := job.GetBuilds()
			if err != nil {
				fmt.Println("Could not refresh "+job.Name+", ", err)
				return
			}
			getreq := GetReq{job.Name, make(chan []Build)}
			gets <- &getreq
			existing := make(map[int]bool)
			for _, build := range <-getreq.Builds {
				existing[build.Number] = true
			}
			for _, build := range builds {
				_, ok := existing[build.Number]
				if !ok {
					puts <- &PutReq{build, false}
				} else if update {
					puts <- &PutReq{build, true}
				}
			}
		}
		go f(job)
	}
	wg.Wait()
	// clean up
	close(puts)
	close(gets)
	// wait until drained
	<-fini
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
