package main

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"strconv"
)

type Store struct{ db *sql.DB }

func OpenStore(path string) (Store, error) {
	db, err := sql.Open("sqlite3", "data")
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
	_, err := s.db.Exec("create table if not exists version(version)")
	if err != nil {
		return err
	}
	vrows, err := s.db.Query("select version from version")
	if err != nil {
		return err
	}
	defer vrows.Close()
	version := 0
	if vrows.Next() {
		err = vrows.Scan(&version)
		if err != nil {
			return err
		}
	}
	if version == 0 {
		trows, err := s.db.Query("select name from sqlite_master where type = 'table' and name = 'builds'")
		if err != nil {
			return err
		}
		if trows.Next() {
			version = 1
		}
		trows.Close()
		_, err = s.db.Exec("insert into versions values (?)", version)
		return err
	}
	return nil
}

func (s Store) version() (int, error) {
	rows, err := s.db.Query("select version from version")
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	if rows.Next() {
		version := 0
		err = rows.Scan(&version)
		if err != nil {
			return -1, err
		}
		return version, nil
	}
	return -1, errors.New("version table not initialized")
}

func (s Store) ensureBuildTable() error {
	version, err := s.version()
	if err != nil {
		return err
	}
	switch version {
	case 0:
		if _, err := s.db.Exec("create table if not exists builds(job, number, start, duration, host, result, failed, total, primary key(job, number))"); err != nil {
			return err
		}
		_, err = s.db.Exec("update version set version = '2'")
		return err
	case 1:
		if _, err := s.db.Exec("alter table builds add column failed"); err != nil {
			return err
		}
		if _, err := s.db.Exec("alter table builds add column total"); err != nil {
			return err
		}
		_, err = s.db.Exec("update version set version = '2'")
		return err
	case 2:
		return nil
	}
	return errors.New("Unsupported version " + strconv.Itoa(version))
}

func (s Store) ensureJobTable() error {
	_, err := s.db.Exec("create table if not exists jobs(name primary key, url)")
	return err
}

func (s Store) Close() {
	s.db.Close()
}

func (s Store) GetJobs() ([]Job, error) {
	rows, err := s.db.Query("select name, url from jobs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []Job
	for rows.Next() {
		var name, url string
		if err := rows.Scan(&name, &url); err != nil {
			return nil, err
		}
		jobs = append(jobs, Job{name, url})
	}
	return jobs, nil
}

func (s Store) GetJob(name string) (Job, error) {
	rows, err := s.db.Query("select name, url from jobs where name = ?", name)
	if err != nil {
		return Job{}, err
	}
	defer rows.Close()
	if rows.Next() {
		var name, url string
		if err := rows.Scan(&name, &url); err != nil {
			return Job{}, err
		}
		return Job{name, url}, nil
	}
	return Job{}, errors.New("Job not found " + name)
}

func (s Store) PutJob(job Job) error {
	_, err := s.db.Exec("insert into jobs values (?, ?)", job.Name, job.Url)
	return err
}

func conv64(in sql.NullString) (int64, error) {
	if !in.Valid {
		return -1, nil
	}
	i, err := strconv.ParseInt(in.String, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func conv(in sql.NullString) (int, error) {
	if !in.Valid {
		return -1, nil
	}
	i, err := strconv.Atoi(in.String)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (s Store) GetBuilds(name string) ([]Build, error) {
	rows, err := s.db.Query("select job, number, start, duration, host, result, failed, total from builds where job = ?", name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var builds []Build
	for rows.Next() {
		var job, snumber, sstart, sduration, host, result, sfailed, stotal sql.NullString
		if err := rows.Scan(&job, &snumber, &sstart, &sduration, &host, &result, &sfailed, &stotal); err != nil {
			return nil, err
		}
		number, err := conv(snumber)
		if err != nil {
			return nil, err
		}
		start, err := conv64(sstart)
		if err != nil {
			return nil, err
		}
		duration, err := conv64(sduration)
		if err != nil {
			return nil, err
		}
		failed, err := conv(sfailed)
		if err != nil {
			return nil, err
		}
		total, err := conv(stotal)
		if err != nil {
			return nil, err
		}
		builds = append(builds, Build{job.String, number, start, duration, host.String, result.String, failed, total})
	}
	return builds, nil
}

func (s Store) InsertBuild(build Build) error {
	_, err := s.db.Exec("insert into builds values (?, ?, ?, ?, ?, ?, ?, ?)",
		build.Job, build.Number, build.Start, build.Duration, build.Host, build.Result, build.Failed, build.Total)
	return err
}

func (s Store) UpdateBuild(build Build) error {
	_, err := s.db.Exec("update builds set start = ?, duration = ?, host = ?, result = ?, failed = ?, total = ? where job = ? and number = ?",
		build.Start, build.Duration, build.Host, build.Result, build.Failed, build.Total, build.Job, build.Number)
	return err
}
