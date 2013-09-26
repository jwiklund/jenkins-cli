package main

import (
	"errors"
	sqlite "github.com/gwenn/gosqlite"
	"strconv"
)

type Store struct{ conn *sqlite.Conn }

func OpenStore(path string) (Store, error) {
	db, err := sqlite.Open(path)
	if err != nil {
		return Store{}, err
	}
	store := Store{db}
	if err = store.ensureVersionTable(); err != nil {
		store.Close()
		return Store{}, err
	}
	if err = store.ensureBuildTable(); err != nil {
		store.Close()
		return Store{}, err
	}
	if err = store.ensureJobTable(); err != nil {
		store.Close()
		return Store{}, err
	}
	return Store{db}, nil
}

func (s Store) ensureVersionTable() error {
	err := s.conn.Exec("create table if not exists version(version)")
	if err != nil {
		return err
	}
	vstmt, err := s.conn.Prepare("select version from version")
	if err != nil {
		return err
	}
	defer vstmt.Finalize()
	version := 0
	err = vstmt.Select(func(s *sqlite.Stmt) error {
		version, _, err = s.ScanInt(0)
		return err
	})
	if version == 0 {
		tstmt, err := s.conn.Prepare("select name from sqlite_master where type = 'table' and name = 'builds'")
		if err != nil {
			return err
		}
		defer tstmt.Finalize()
		buildsExists := false
		tstmt.Select(func(s *sqlite.Stmt) error {
			buildsExists = true
			return nil
		})
		if buildsExists {
			version = 1
		}
		istmt, err := s.conn.Prepare("insert into version values (?)")
		if err != nil {
			return err
		}
		defer istmt.Finalize()
		_, err = istmt.Insert(version)
		return err
	}
	return nil
}

func (s Store) version() (int, error) {
	stmt, err := s.conn.Prepare("select version from version")
	if err != nil {
		return -1, err
	}
	defer stmt.Finalize()
	version := -1
	err = stmt.Select(func(s *sqlite.Stmt) error {
		version, _, err = s.ScanInt(0)
		return err
	})
	if err != nil {
		return -1, err
	}
	if version == -1 {
		return -1, errors.New("version table not initialized")
	}
	return version, nil
}

func (s Store) ensureBuildTable() error {
	version, err := s.version()
	if err != nil {
		return err
	}
	switch version {
	case 0:
		if err := s.conn.Exec("create table if not exists builds(job, number, start, duration, host, result, errors, primary key(job, number))"); err != nil {
			return err
		}
		return s.conn.Exec("update version set version = '2'")
	case 1:
		if err := s.conn.Exec("alter table builds add column errors"); err != nil {
			return err
		}
		return s.conn.Exec("update version set version = '2'")
	case 2:
		return nil
	}
	return errors.New("Unsupported version " + strconv.Itoa(version))
}

func (s Store) ensureJobTable() error {
	return s.conn.Exec("create table if not exists jobs(name primary key, url)")
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
	stmt, err := s.conn.Prepare("select job, number, start, duration, host, result from builds where job = ?")
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
		result, _ := s.ScanText(5)
		builds = append(builds, Build{job, number, start, duration, host, result})
		return nil
	}, name)
	if err != nil {
		return nil, err
	}
	return builds, nil
}

func (s Store) PutBuild(build Build) error {
	stmt, err := s.conn.Prepare("insert into builds values (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Finalize()
	_, err = stmt.Insert(build.Job, build.Number, build.Start, build.Duration, build.Host, build.Result)
	return err
}
