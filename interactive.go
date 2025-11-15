package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

func readLine(prompt string) (string, error) {
	fmt.Print(prompt)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func readInt(prompt string, allowEmpty bool, defaultVal int) (int, error) {
	for {
		s, err := readLine(prompt)
		if err != nil {
			return 0, err
		}
		if allowEmpty && s == "" {
			return defaultVal, nil
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			fmt.Println("Please enter a valid number.")
			continue
		}
		return n, nil
	}
}

func readExistingPath(prompt string) (string, error) {
	for {
		s, err := readLine(prompt)
		if err != nil {
			return "", err
		}
		if s == "" {
			fmt.Println("Path cannot be empty.")
			continue
		}
		// Expand ~
		if strings.HasPrefix(s, "~") {
			home, err := os.UserHomeDir()
			if err == nil {
				s = filepath.Join(home, strings.TrimPrefix(s, "~"))
			}
		}
		if _, err := os.Stat(s); err != nil {
			if os.IsNotExist(err) {
				fmt.Println("This path does not exist. Please enter an existing path.")
				continue
			}
			fmt.Printf("Error checking path: %v\n", err)
			continue
		}
		return s, nil
	}
}

func readPathCreateIfMissing(prompt string) (string, error) {
	for {
		s, err := readLine(prompt)
		if err != nil {
			return "", err
		}
		if s == "" {
			fmt.Println("Path cannot be empty.")
			continue
		}
		// Expand ~
		if strings.HasPrefix(s, "~") {
			home, err := os.UserHomeDir()
			if err == nil {
				s = filepath.Join(home, strings.TrimPrefix(s, "~"))
			}
		}

		info, err := os.Stat(s)
		if err == nil {
			if info.IsDir() {
				return s, nil
			}
			fmt.Println("Path exists but is not a directory. Please enter a directory.")
			continue
		}
		if os.IsNotExist(err) {
			fmt.Printf("Directory does not exist: %s\n", s)
			answer, _ := readLine("Create it? (y/n): ")
			if strings.ToLower(answer) == "y" {
				if err := os.MkdirAll(s, 0755); err != nil {
					fmt.Printf("Failed to create directory: %v\n", err)
					continue
				}
				return s, nil
			}
			continue
		}

		fmt.Printf("Error checking path: %v\n", err)
	}
}

func ShowMenu(cfg *Config) error {
	for {
		fmt.Println("=== Snapfox – Backup Manager ===")
		fmt.Println("1) List backup jobs")
		fmt.Println("2) Add backup job")
		fmt.Println("3) Manage jobs (edit/delete)")
		fmt.Println("4) Run ALL backups now")
		fmt.Println("5) Exit")

		choice, err := readLine("Choose option: ")
		if err != nil {
			return err
		}

		switch choice {
		case "1":
			listJobs(cfg)
		case "2":
			if err := addJobInteractive(cfg); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "3":
			if err := manageJobs(cfg); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "4":
			if err := RunAllBackups(cfg); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "5":
			fmt.Println("Bye.")
			return nil
		default:
			fmt.Println("Invalid choice.")
		}

		fmt.Println()
	}
}

func listJobs(cfg *Config) {
	if len(cfg.Jobs) == 0 {
		fmt.Println("No jobs configured yet.")
		return
	}
	fmt.Println("Configured backup jobs:")
	for i, j := range cfg.Jobs {
		fmt.Printf("  [%d] %s\n", i+1, j.Name)
		fmt.Printf("      Source:        %s\n", j.Source)
		fmt.Printf("      Destination:   %s\n", j.Destination)
		fmt.Printf("      Interval:      %d hours (0 = manual)\n", j.IntervalHours)
		fmt.Printf("      Max snapshots: %d (0 = unlimited)\n", j.MaxSnapshots)
		if j.LastRun != "" {
			fmt.Printf("      Last run:      %s\n", j.LastRun)
		} else {
			fmt.Printf("      Last run:      never\n")
		}
		fmt.Println()
	}
}

func addJobInteractive(cfg *Config) error {
	fmt.Println("== Add new backup job ==")

	name, err := readLine("Name for this backup (e.g. 'Seafile config'): ")
	if err != nil {
		return err
	}
	if name == "" {
		name = "Unnamed job"
	}

	source, err := readExistingPath("Source directory (what to backup): ")
	if err != nil {
		return err
	}

	dest, err := readPathCreateIfMissing("Destination directory (where snapshots are stored): ")
	if err != nil {
		return err
	}

	interval, err := readInt("Interval in hours (0 = manual only): ", true, 0)
	if err != nil {
		return err
	}

	maxSnap, err := readInt("Max snapshots to keep (0 = unlimited): ", true, 0)
	if err != nil {
		return err
	}

	job := BackupJob{
		ID:            fmt.Sprintf("job-%d", len(cfg.Jobs)+1),
		Name:          name,
		Source:        source,
		Destination:   dest,
		IntervalHours: interval,
		MaxSnapshots:  maxSnap,
	}

	cfg.AddJob(job)
	if err := SaveConfig(cfg); err != nil {
		return err
	}

	fmt.Println("Backup job added.")
	return nil
}

func manageJobs(cfg *Config) error {
	if len(cfg.Jobs) == 0 {
		fmt.Println("No jobs to manage.")
		return nil
	}

	listJobs(cfg)
	idx, err := readInt("Select job number to manage (0 = cancel): ", true, 0)
	if err != nil {
		return err
	}
	if idx == 0 {
		return nil
	}
	idx-- // to 0-based

	if idx < 0 || idx >= len(cfg.Jobs) {
		fmt.Println("Invalid job number.")
		return nil
	}

	job := &cfg.Jobs[idx]

	for {
		fmt.Printf("\nManaging job: %s\n", job.Name)
		fmt.Println("1) Edit job")
		fmt.Println("2) Delete job")
		fmt.Println("3) Back")

		choice, err := readLine("Choose option: ")
		if err != nil {
			return err
		}

		switch choice {
		case "1":
			if err := editJob(job); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				if err := SaveConfig(cfg); err != nil {
					fmt.Printf("Failed to save config: %v\n", err)
				}
			}
		case "2":
			confirm, _ := readLine("Are you sure you want to delete this job? (y/n): ")
			if strings.ToLower(confirm) == "y" {
				if err := cfg.DeleteJobByIndex(idx); err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					if err := SaveConfig(cfg); err != nil {
						fmt.Printf("Failed to save config: %v\n", err)
					} else {
						fmt.Println("Job deleted.")
					}
				}
				return nil
			}
		case "3":
			return nil
		default:
			fmt.Println("Invalid choice.")
		}
	}
}

func editJob(job *BackupJob) error {
	fmt.Printf("Editing job '%s'. Leave field empty to keep current value.\n", job.Name)

	name, err := readLine(fmt.Sprintf("Name [%s]: ", job.Name))
	if err != nil {
		return err
	}
	if name != "" {
		job.Name = name
	}

	fmt.Printf("Current source: %s\n", job.Source)
	newSource, err := readLine("New source (empty = keep): ")
	if err != nil {
		return err
	}
	if newSource != "" {
		// validate
		readerBackup := reader
		reader = bufio.NewReader(os.Stdin)
		defer func() { reader = readerBackup }()

		// Reuse validation
		reader = bufio.NewReader(os.Stdin)
		tmp := newSource
		if strings.HasPrefix(tmp, "~") {
			home, e := os.UserHomeDir()
			if e == nil {
				tmp = filepath.Join(home, strings.TrimPrefix(tmp, "~"))
			}
		}
		if _, err := os.Stat(tmp); err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Path does not exist. Source not changed.")
			} else {
				fmt.Printf("Error checking path: %v. Source not changed.\n", err)
			}
		} else {
			job.Source = tmp
		}
	}

	fmt.Printf("Current destination: %s\n", job.Destination)
	newDest, err := readLine("New destination (empty = keep): ")
	if err != nil {
		return err
	}
	if newDest != "" {
		tmp := newDest
		if strings.HasPrefix(tmp, "~") {
			home, e := os.UserHomeDir()
			if e == nil {
				tmp = filepath.Join(home, strings.TrimPrefix(tmp, "~"))
			}
		}
		info, err := os.Stat(tmp)
		if err == nil && info.IsDir() {
			job.Destination = tmp
		} else if os.IsNotExist(err) {
			fmt.Printf("Directory does not exist: %s\n", tmp)
			answer, _ := readLine("Create it? (y/n): ")
			if strings.ToLower(answer) == "y" {
				if err := os.MkdirAll(tmp, 0755); err != nil {
					fmt.Printf("Failed to create directory: %v. Destination not changed.\n", err)
				} else {
					job.Destination = tmp
				}
			}
		} else if err != nil {
			fmt.Printf("Error checking path: %v. Destination not changed.\n", err)
		} else {
			fmt.Println("Path exists but is not a directory. Destination not changed.")
		}
	}

	fmt.Printf("Current interval hours: %d\n", job.IntervalHours)
	intervalStr, err := readLine("New interval hours (empty = keep): ")
	if err != nil {
		return err
	}
	if intervalStr != "" {
		n, err := strconv.Atoi(intervalStr)
		if err != nil {
			fmt.Println("Invalid number, interval not changed.")
		} else {
			job.IntervalHours = n
		}
	}

	fmt.Printf("Current max snapshots: %d\n", job.MaxSnapshots)
	maxStr, err := readLine("New max snapshots (empty = keep): ")
	if err != nil {
		return err
	}
	if maxStr != "" {
		n, err := strconv.Atoi(maxStr)
		if err != nil {
			fmt.Println("Invalid number, max snapshots not changed.")
		} else {
			job.MaxSnapshots = n
		}
	}

	fmt.Println("Job updated.")
	return nil
}