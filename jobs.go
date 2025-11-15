package main

import (
	"fmt"
	"time"
)

type BackupJob struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Source        string `json:"source"`
	Destination   string `json:"destination"`
	IntervalHours int    `json:"interval_hours"` // 0 = manual only
	LastRun       string `json:"last_run,omitempty"`
	MaxSnapshots  int    `json:"max_snapshots"` // 0 = unlimited
}

func (j *BackupJob) NextRunTime() *time.Time {
	if j.IntervalHours <= 0 {
		return nil
	}
	if j.LastRun == "" {
		// never ran = due now
		t := time.Time{}
		return &t
	}

	t, err := time.Parse(time.RFC3339, j.LastRun)
	if err != nil {
		t2 := time.Time{}
		return &t2
	}
	next := t.Add(time.Duration(j.IntervalHours) * time.Hour)
	return &next
}

func (j *BackupJob) IsDue(now time.Time) bool {
	if j.IntervalHours <= 0 {
		return false
	}
	next := j.NextRunTime()
	if next == nil {
		return false
	}
	return !next.After(now)
}

func (cfg *Config) AddJob(job BackupJob) {
	cfg.Jobs = append(cfg.Jobs, job)
}

func (cfg *Config) DeleteJobByIndex(idx int) error {
	if idx < 0 || idx >= len(cfg.Jobs) {
		return fmt.Errorf("invalid job index")
	}
	cfg.Jobs = append(cfg.Jobs[:idx], cfg.Jobs[idx+1:]...)
	return nil
}