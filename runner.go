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

func sanitizeName(name string) string {
	if name == "" {
		return "job"
	}
	replacer := strings.NewReplacer(" ", "_", ":", "-", "/", "-", "\\", "-")
	return replacer.Replace(name)
}

func RunJob(job *BackupJob) error {
	now := time.Now()
	ts := now.Format("2006-01-02_15-04-05")

	baseName := sanitizeName(job.Name)
	if baseName == "" {
		baseName = sanitizeName(job.ID)
	}
	targetDir := filepath.Join(job.Destination, fmt.Sprintf("%s_%s", baseName, ts))

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target dir: %w", err)
	}

	// Ensure source ends with slash (rsync convention)
	src := job.Source
	if !strings.HasSuffix(src, "/") {
		src = src + "/"
	}
	dst := targetDir + "/"

	fmt.Printf("Running backup '%s'\n", job.Name)
	fmt.Printf("  From: %s\n", job.Source)
	fmt.Printf("  To:   %s\n", targetDir)

	cmd := exec.Command("rsync", "-a", src, dst)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync failed: %w", err)
	}

	// Update last run time
	job.LastRun = now.Format(time.RFC3339)

	// Enforce retention if set
	if job.MaxSnapshots > 0 {
		if err := enforceRetention(job); err != nil {
			fmt.Printf("Retention warning for '%s': %v\n", job.Name, err)
		}
	}

	fmt.Println("Backup finished.")
	return nil
}

func enforceRetention(job *BackupJob) error {
	entries, err := os.ReadDir(job.Destination)
	if err != nil {
		return err
	}

	baseName := sanitizeName(job.Name)
	if baseName == "" {
		baseName = sanitizeName(job.ID)
	}

	var snapshots []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), baseName+"_") {
			snapshots = append(snapshots, e)
		}
	}

	if len(snapshots) <= job.MaxSnapshots {
		return nil
	}

	// sort by name (contains timestamp, so works reasonably)
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Name() < snapshots[j].Name()
	})

	toDelete := len(snapshots) - job.MaxSnapshots
	for i := 0; i < toDelete; i++ {
		path := filepath.Join(job.Destination, snapshots[i].Name())
		fmt.Printf("Deleting old snapshot: %s\n", path)
		if err := os.RemoveAll(path); err != nil {
			fmt.Printf("Failed to delete %s: %v\n", path, err)
		}
	}

	return nil
}

func RunAllBackups(cfg *Config) error {
	if len(cfg.Jobs) == 0 {
		fmt.Println("No jobs configured.")
		return nil
	}

	for i := range cfg.Jobs {
		job := &cfg.Jobs[i]
		fmt.Printf("\n=== Job %d: %s ===\n", i+1, job.Name)
		if err := RunJob(job); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	return SaveConfig(cfg)
}

func RunDueBackups(cfg *Config) error {
	now := time.Now()
	ranAny := false

	for i := range cfg.Jobs {
		job := &cfg.Jobs[i]
		if job.IsDue(now) {
			fmt.Printf("\n=== Running due job: %s ===\n", job.Name)
			if err := RunJob(job); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			ranAny = true
		}
	}

	if !ranAny {
		fmt.Println("No jobs are due at this time.")
		return nil
	}

	return SaveConfig(cfg)
}