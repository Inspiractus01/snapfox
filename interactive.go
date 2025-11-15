package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Menu struct {
	Config  *Config
	CfgPath string
	scanner *bufio.Scanner
}

func NewMenu(cfg *Config, cfgPath string) *Menu {
	return &Menu{
		Config:  cfg,
		CfgPath: cfgPath,
		scanner: bufio.NewScanner(os.Stdin),
	}
}

func (m *Menu) readLine(prompt string) string {
	fmt.Print(prompt)
	if !m.scanner.Scan() {
		return ""
	}
	return strings.TrimSpace(m.scanner.Text())
}

func (m *Menu) ShowMainMenu() {
	for {
		fmt.Println("=== Snapfox – Backup Manager ===")
		fmt.Println("1) List backup jobs")
		fmt.Println("2) Add backup job")
		fmt.Println("3) Edit backup job")
		fmt.Println("4) Run ALL backups now")
		fmt.Println("5) Exit")

		choice := m.readLine("Choose option: ")
		fmt.Println()

		switch choice {
		case "1":
			m.listJobs()
		case "2":
			m.addJobInteractive()
		case "3":
			m.editJobInteractive()
		case "4":
			if err := RunAllJobs(m.Config, m.CfgPath); err != nil {
				fmt.Printf("Error running backups: %v\n", err)
			}
		case "5":
			fmt.Println("Bye.")
			return
		default:
			fmt.Println("Unknown option, choose 1–5.")
		}

		fmt.Println()
	}
}

func (m *Menu) listJobs() {
	if len(m.Config.Jobs) == 0 {
		fmt.Println("No backup jobs defined yet.")
		return
	}
	for _, j := range m.Config.Jobs {
		fmt.Println(j.Summary())
		fmt.Println()
	}
}

func (m *Menu) addJobInteractive() {
	fmt.Println("Add new backup job")
	fmt.Println("-------------------")

	name := m.readNonEmpty("Job name: ")

	src := m.readExistingDir("Source directory: ")

	dest := m.readExistingDirOrCreate("Destination root directory (snapshots will go inside): ")

	interval := m.readIntWithDefault(
		"Interval in hours between runs: ",
		24,
	)

	keepLast := m.readIntWithDefault(
		"How many snapshots to keep: ",
		10,
	)

	fmt.Println()
	fmt.Println("Job summary:")
	tmp := Job{
		Name:          name,
		Source:        src,
		Destination:   dest,
		IntervalHours: interval,
		KeepLast:      keepLast,
	}
	fmt.Println(tmp.Summary())
	fmt.Println()

	confirm := strings.ToLower(m.readLine("Save this job? [y/N]: "))
	if confirm != "y" && confirm != "yes" {
		fmt.Println("Job discarded.")
		return
	}

	j := m.Config.AddJob(tmp)
	if err := m.Config.Save(m.CfgPath); err != nil {
		fmt.Printf("Failed to save config: %v\n", err)
		return
	}
	fmt.Printf("Job [%d] saved.\n", j.ID)
}

func (m *Menu) editJobInteractive() {
	if len(m.Config.Jobs) == 0 {
		fmt.Println("No jobs to edit.")
		return
	}

	m.listJobs()
	idStr := m.readLine("Enter job ID to edit: ")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Invalid ID.")
		return
	}

	job, _ := m.Config.FindJobByID(id)
	if job == nil {
		fmt.Println("Job not found.")
		return
	}

	fmt.Println()
	fmt.Printf("Editing job [%d] %s\n", job.ID, job.Name)
	fmt.Println("Press ENTER to keep current value.")

	// Name
	newName := m.readLine(fmt.Sprintf("Name [%s]: ", job.Name))
	if strings.TrimSpace(newName) != "" {
		job.Name = newName
	}

	// Source
	for {
		inp := m.readLine(fmt.Sprintf("Source directory [%s]: ", job.Source))
		if strings.TrimSpace(inp) == "" {
			break
		}
		resolved, err := resolveAndCheckDir(inp)
		if err != nil {
			fmt.Printf("  %v\n", err)
			continue
		}
		job.Source = resolved
		break
	}

	// Destination
	for {
		inp := m.readLine(fmt.Sprintf("Destination root [%s]: ", job.Destination))
		if strings.TrimSpace(inp) == "" {
			break
		}
		resolved, err := resolveDirOrCreate(inp)
		if err != nil {
			fmt.Printf("  %v\n", err)
			continue
		}
		job.Destination = resolved
		break
	}

	// IntervalHours
	for {
		inp := m.readLine(fmt.Sprintf("Interval hours [%d]: ", job.IntervalHours))
		if strings.TrimSpace(inp) == "" {
			break
		}
		v, err := strconv.Atoi(inp)
		if err != nil || v < 0 {
			fmt.Println("  Please enter a positive number.")
			continue
		}
		job.IntervalHours = v
		break
	}

	// KeepLast
	for {
		inp := m.readLine(fmt.Sprintf("Keep last snapshots [%d]: ", job.KeepLast))
		if strings.TrimSpace(inp) == "" {
			break
		}
		v, err := strconv.Atoi(inp)
		if err != nil || v <= 0 {
			fmt.Println("  Please enter a positive number.")
			continue
		}
		job.KeepLast = v
		break
	}

	if err := m.Config.Save(m.CfgPath); err != nil {
		fmt.Printf("Failed to save config: %v\n", err)
		return
	}
	fmt.Println("Job updated.")
}

// helpers

func (m *Menu) readNonEmpty(prompt string) string {
	for {
		s := m.readLine(prompt)
		if strings.TrimSpace(s) != "" {
			return s
		}
		fmt.Println("  Please enter a value.")
	}
}

func (m *Menu) readIntWithDefault(prompt string, def int) int {
	for {
		s := m.readLine(fmt.Sprintf("%s [%d]: ", prompt, def))
		if strings.TrimSpace(s) == "" {
			return def
		}
		v, err := strconv.Atoi(s)
		if err != nil || v < 0 {
			fmt.Println("  Please enter a positive number.")
			continue
		}
		return v
	}
}

func (m *Menu) readExistingDir(prompt string) string {
	for {
		path := m.readLine(prompt)
		resolved, err := resolveAndCheckDir(path)
		if err != nil {
			fmt.Printf("  %v\n", err)
			continue
		}
		return resolved
	}
}

func (m *Menu) readExistingDirOrCreate(prompt string) string {
	for {
		path := m.readLine(prompt)
		resolved, err := resolveDirOrCreate(path)
		if err != nil {
			fmt.Printf("  %v\n", err)
			continue
		}
		return resolved
	}
}

func resolveAndCheckDir(p string) (string, error) {
	if strings.TrimSpace(p) == "" {
		return "", fmt.Errorf("path cannot be empty")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("path does not exist: %s", abs)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", abs)
	}
	return abs, nil
}

func resolveDirOrCreate(p string) (string, error) {
	if strings.TrimSpace(p) == "" {
		return "", fmt.Errorf("path cannot be empty")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path: %w", err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return "", fmt.Errorf("cannot create directory: %w", err)
	}
	return abs, nil
}