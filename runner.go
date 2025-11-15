package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func RunAllJobs(cfg *Config, cfgPath string) error {
	now := time.Now()
	for i := range cfg.Jobs {
		if err := runSingleJob(&cfg.Jobs[i], now); err != nil {
			fmt.Printf("Job %d (%s) failed: %v\n", cfg.Jobs[i].ID, cfg.Jobs[i].Name, err)
		}
	}
	return cfg.Save(cfgPath)
}

// still nechávam aj RunDueJobs, môžeš si ho časom zavolať z cronu
func RunDueJobs(cfg *Config, cfgPath string) error {
	now := time.Now()
	for i := range cfg.Jobs {
		if cfg.Jobs[i].IsDueNow(now) {
			if err := runSingleJob(&cfg.Jobs[i], now); err != nil {
				fmt.Printf("Job %d (%s) failed: %v\n", cfg.Jobs[i].ID, cfg.Jobs[i].Name, err)
			}
		}
	}
	return cfg.Save(cfgPath)
}

func runSingleJob(j *Job, now time.Time) error {
	if j.KeepLast <= 0 {
		j.KeepLast = 10
	}

	src := j.Source
	dstRoot := j.Destination

	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("source not accessible: %w", err)
	}
	if err := os.MkdirAll(dstRoot, 0o755); err != nil {
		return fmt.Errorf("failed to create destination root: %w", err)
	}

	ts := now.Format("20060102-150405")
	slug := slugify(j.Name)
	dstSnap := filepath.Join(dstRoot, fmt.Sprintf("%s-%s", slug, ts))

	if err := os.MkdirAll(dstSnap, 0o755); err != nil {
		return fmt.Errorf("failed to create snapshot dir: %w", err)
	}

	fmt.Printf("→ Running job [%d] %s\n", j.ID, j.Name)
	fmt.Printf("  from %s\n  to   %s\n", src, dstSnap)

	// rsync src/ -> dstSnap/
	cmd := exec.Command(
		"rsync",
		"-a",
		"--delete",
		src+string(os.PathSeparator),
		dstSnap+string(os.PathSeparator),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync failed: %w", err)
	}

	j.LastRun = now.Format(time.RFC3339)

	if err := pruneSnapshots(j); err != nil {
		fmt.Printf("  ! failed to prune snapshots: %v\n", err)
	}

	fmt.Println("  done.")
	return nil
}

type snapshotInfo struct {
	path string
	time time.Time
}

func pruneSnapshots(j *Job) error {
	if j.KeepLast <= 0 {
		return nil
	}

	entries, err := os.ReadDir(j.Destination)
	if err != nil {
		return err
	}

	prefix := slugify(j.Name) + "-"
	var snaps []snapshotInfo

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		tsPart := strings.TrimPrefix(name, prefix)
		t, err := time.Parse("20060102-150405", tsPart)
		if err != nil {
			continue
		}
		snaps = append(snaps, snapshotInfo{
			path: filepath.Join(j.Destination, name),
			time: t,
		})
	}

	if len(snaps) <= j.KeepLast {
		return nil
	}

	sort.Slice(snaps, func(i, k int) bool {
		return snaps[i].time.Before(snaps[k].time)
	})

	toDelete := snaps[0 : len(snaps)-j.KeepLast]
	for _, s := range toDelete {
		fmt.Printf("  → removing old backup %s\n", s.path)
		if err := os.RemoveAll(s.path); err != nil {
			fmt.Printf("    ! failed to remove: %v\n", err)
		}
	}

	return nil
}

func slugify(name string) string {
	s := strings.TrimSpace(strings.ToLower(name))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")
	return s
}