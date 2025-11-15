package main

import (
	"fmt"
	"time"
)

type Job struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Source        string `json:"source"`
	Destination   string `json:"destination"`
	IntervalHours int    `json:"interval_hours"`
	KeepLast      int    `json:"keep_last"`
	LastRun       string `json:"last_run"` // RFC3339
}

func (j *Job) LastRunTime() (time.Time, bool) {
	if j.LastRun == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, j.LastRun)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func (j *Job) NextRunTime() (time.Time, bool) {
	last, ok := j.LastRunTime()
	if !ok || j.IntervalHours <= 0 {
		return time.Time{}, false
	}
	return last.Add(time.Duration(j.IntervalHours) * time.Hour), true
}

func (j *Job) IsDueNow(now time.Time) bool {
	next, ok := j.NextRunTime()
	if !ok {
		// Never ran or no interval -> treat as "due"
		return true
	}
	return !now.Before(next)
}

func (j *Job) Summary() string {
	last, hasLast := j.LastRunTime()
	lastStr := "never"
	if hasLast {
		lastStr = last.Format("2006-01-02 15:04")
	}

	return fmt.Sprintf(
		"[%d] %s\n  from: %s\n    to: %s\n  every: %d h   keep: %d   last: %s",
		j.ID,
		j.Name,
		j.Source,
		j.Destination,
		j.IntervalHours,
		j.KeepLast,
		lastStr,
	)
}