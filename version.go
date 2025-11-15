package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// zvýš to keď vydáš novú verziu
const Version = "0.1.0"

// TODO: nastav na svoj GitHub repo
// napr. "https://raw.githubusercontent.com/michalmucha/snapfox/main/latest.txt"
const latestVersionURL = "https://raw.githubusercontent.com/YOUR_GH_USER/snapfox/main/latest.txt"

func parseVersion(v string) []int {
	parts := strings.Split(v, ".")
	res := make([]int, 3)
	for i := 0; i < len(res) && i < len(parts); i++ {
		var n int
		fmt.Sscanf(parts[i], "%d", &n)
		res[i] = n
	}
	return res
}

func isNewerVersion(latest, current string) bool {
	l := parseVersion(latest)
	c := parseVersion(current)
	for i := 0; i < 3; i++ {
		if l[i] > c[i] {
			return true
		}
		if l[i] < c[i] {
			return false
		}
	}
	return false
}

// volaj iba v interaktívnom režime, nie v run-due (systemd)
func maybeCheckForUpdates() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestVersionURL, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		return
	}
	latest := strings.TrimSpace(string(body))
	if latest == "" {
		return
	}

	if isNewerVersion(latest, Version) {
		fmt.Println()
		fmt.Printf(">> New Snapfox version available: %s (you have %s)\n", latest, Version)
		fmt.Println(">> Update with:")
		fmt.Println("   curl -fsSL https://raw.githubusercontent.com/Inspiractus01/snapfox/main/install.sh | sh")
		fmt.Println()
	}
}