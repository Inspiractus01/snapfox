package main

import (
	"fmt"
	"os"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// No args = interactive menu
	if len(os.Args) == 1 {
		maybeCheckForUpdates() // <<-- pridané
		if err := ShowMenu(cfg); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	switch os.Args[1] {
	case "menu":
		maybeCheckForUpdates() // <<-- pridané
		if err := ShowMenu(cfg); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "run-due":
		// systemd timer – žiadny network check, nech to je tiché a rýchle
		if err := RunDueBackups(cfg); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "run-all":
		if err := RunAllBackups(cfg); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Println("Snapfox – Simple backup scheduler using rsync")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  snapfox           # interactive menu")
		fmt.Println("  snapfox menu      # interactive menu")
		fmt.Println("  snapfox run-due   # run only due backups (for systemd timer)")
		fmt.Println("  snapfox run-all   # run all configured backups once")
		os.Exit(1)
	}
}