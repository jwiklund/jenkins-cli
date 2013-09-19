package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type Job struct {
	Name string
	Url  string
}

func (j Job) String() string {
	return "Job with name " + j.Name + " at " + j.Url
}

type Jenkins struct {
	Jobs []Job
}

func getJson(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return errors.New("Could not fetch json: " + err.Error())
	}
	defer resp.Body.Close()
	jsonBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("Could not read json: " + err.Error())
	}
	err = json.Unmarshal(jsonBytes, target)
	if err != nil {
		return errors.New("Could not parse json: " + err.Error())
	}
	return nil
}

func GetJobs(prefix string) ([]Job, error) {
	var jenkins Jenkins
	err := getJson("http://jenkins/jenkins/api/json", &jenkins)
	if err != nil {
		return nil, err
	}
	var jobs []Job
	for _, job := range jenkins.Jobs {
		if strings.Index(job.Name, prefix) == 0 {
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

type Build struct {
	Job      string
	Number   int
	Start    int64
	Duration int64
	Host     string
}

type JsonBuild struct {
	FullDisplayName string
	Url             string
	Duration        int64
	Timestamp       int64
	Number          int
}

type JsonJob struct {
	DisplayName string
	Builds      []JsonBuild
}

func itoa(i int64) string {
	var bytes []byte
	return string(strconv.AppendInt(bytes, i, 10))
}

func (b Build) String() string {
	return b.Job + " " + strconv.Itoa(b.Number) + " started " + itoa(b.Start) + ", duration " + itoa(b.Duration) + " at " + b.Host
}

func (j Job) GetBuilds() ([]Build, error) {
	var details JsonJob
	err := getJson(j.Url+"/api/json?tree=builds[number,url,duration,timestamp,fullDisplayName]", &details)
	if err != nil {
		return nil, errors.New("Could not fetch json: " + err.Error())
	}
	var res []Build
	for _, build := range details.Builds {
		host, err := GetHost(build.Url)
		if err != nil {
			host = "failure: " + err.Error()
		}
		res = append(res, Build{j.Name, build.Number, build.Timestamp, build.Duration, host})
	}
	return res, nil
}

func GetHost(url string) (string, error) {
	console, err := http.Get(url + "/consoleText")
	if err != nil {
		return "", err
	}
	defer console.Body.Close()
	r := bufio.NewReader(console.Body)
	line, err := r.ReadString('\n')
	for err == nil {
		if strings.Index(line, "Node Controller:") == 0 {
			ctrl := strings.Trim(strings.Split(line, ":")[1], " \r\n\t")
			if ctrl == "" {
				return "", errors.New("Empty Host")
			}
			if ctrl == "Host key verification failed." {
				return "", errors.New(ctrl)
			}
			return ctrl, nil
		}
		line, err = r.ReadString('\n')
	}
	return "", errors.New("No Host")
}
