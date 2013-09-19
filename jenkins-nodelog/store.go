package main

import (
	"errors"
	sqlite "github.com/gwenn/gosqlite"
)

type Store struct{ conn *sqlite.Conn }

func OpenStore(path string) (Store, error) {
	db, err := sqlite.Open(path)
	if err != nil {
		return Store{}, err
	}
	err = db.Exec("create table if not exists builds(job, number, start, duration, host, primary key(job, number))")
	if err != nil {
		return Store{}, err
	}
	err = db.Exec("create table if not exists jobs(name primary key, url)")
	if err != nil {
		return Store{}, err
	}
	return Store{db}, nil
}

func (s Store) Close() {
	s.conn.Close()
}

func (s Store) GetJobs() ([]Job, error) {
	stmt, err := s.conn.Prepare("select name, url from jobs")
	if err != nil {
		return nil, err
	}
	defer stmt.Finalize()
	var jobs []Job
	err = stmt.Select(func(s *sqlite.Stmt) error {
		name, _ := s.ScanText(0)
		url, _ := s.ScanText(1)
		jobs = append(jobs, Job{name, url})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (s Store) GetJob(name string) (Job, error) {
	stmt, err := s.conn.Prepare("select name, url from jobs where name = ?")
	if err != nil {
		return Job{}, err
	}
	defer stmt.Finalize()
	var job Job
	found := false
	err = stmt.Select(func(s *sqlite.Stmt) error {
		name, _ := s.ScanText(0)
		url, _ := s.ScanText(1)
		job = Job{name, url}
		found = true
		return nil
	}, name)
	if err != nil {
		return Job{}, err
	}
	if !found {
		return Job{}, errors.New("Job not found " + name)
	}
	return job, nil
}

func (s Store) PutJob(job Job) error {
	stmt, err := s.conn.Prepare("insert into jobs values (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Finalize()
	_, err = stmt.Insert(job.Name, job.Url)
	return err
}

func (s Store) GetBuilds(name string) ([]Build, error) {
	stmt, err := s.conn.Prepare("select job, number, start, duration, host from builds where job = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Finalize()
	var builds []Build
	err = stmt.Select(func(s *sqlite.Stmt) error {
		job, _ := s.ScanText(0)
		number, _, err := s.ScanInt(1)
		if err != nil {
			return err
		}
		start, _, err := s.ScanInt64(2)
		if err != nil {
			return err
		}
		duration, _, err := s.ScanInt64(3)
		if err != nil {
			return err
		}
		host, _ := s.ScanText(4)
		builds = append(builds, Build{job, number, start, duration, host})
		return nil
	}, name)
	if err != nil {
		return nil, err
	}
	return builds, nil
}

func (s Store) PutBuild(build Build) error {
	stmt, err := s.conn.Prepare("insert into builds values (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Finalize()
	_, err = stmt.Insert(build.Job, build.Number, build.Start, build.Duration, build.Host)
	return err
}
