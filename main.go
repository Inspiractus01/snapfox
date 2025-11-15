package main

import (
	"fmt"
	"os"
)

const ConfigPath = "snapfox.json"

func main() {
	cfg, err := LoadConfig(ConfigPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Optional CLI shortcuts
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "run-all":
			if err := RunAllJobs(cfg, ConfigPath); err != nil {
				fmt.Printf("Error running jobs: %v\n", err)
				os.Exit(1)
			}
			return
		case "run-due":
			if err := RunDueJobs(cfg, ConfigPath); err != nil {
				fmt.Printf("Error running jobs: %v\n", err)
				os.Exit(1)
			}
			return
		case "list":
			m := NewMenu(cfg, ConfigPath)
			m.listJobs()
			return
		}
	}

	menu := NewMenu(cfg, ConfigPath)
	menu.ShowMainMenu()
}