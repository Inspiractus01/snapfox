package main

import (
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	Jobs      []Job `json:"jobs"`
	NextJobID int   `json:"next_job_id"`
}

func LoadConfig(path string) (*Config, error) {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		// New config
		return &Config{
			Jobs:      []Job{},
			NextJobID: 1,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	dec := json.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}

	// Safety: if NextJobID was missing, recalc from Jobs
	if cfg.NextJobID == 0 {
		maxID := 0
		for _, j := range cfg.Jobs {
			if j.ID > maxID {
				maxID = j.ID
			}
		}
		cfg.NextJobID = maxID + 1
	}

	return &cfg, nil
}

func (c *Config) Save(path string) error {
	tmp := path + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(c); err != nil {
		f.Close()
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}

func (c *Config) AddJob(j Job) Job {
	j.ID = c.NextJobID
	c.NextJobID++

	if j.KeepLast <= 0 {
		j.KeepLast = 10
	}

	c.Jobs = append(c.Jobs, j)
	return j
}

func (c *Config) FindJobByID(id int) (*Job, int) {
	for i := range c.Jobs {
		if c.Jobs[i].ID == id {
			return &c.Jobs[i], i
		}
	}
	return nil, -1
}