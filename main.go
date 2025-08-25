/*
Copyright 2024 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime/debug"
	"time"

	"github.com/hashicorp/go-version"

	"github.com/vitessio/vitess-releaser/go/cmd"
)

// Struct to hold the GitHub API response.
type githubRelease struct {
	TagName string `json:"tag_name"`
}

// getLatestVersionFromGitHub queries the GitHub API for the latest release version.
func getLatestVersionFromGitHub() (string, error) {
	const url = "https://api.github.com/repos/vitessio/vitess-releaser/releases/latest"

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("error fetching latest release info: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch latest version, status code: %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	return release.TagName, nil
}

// compareVersions checks if the current version is the latest and prints a message.
func compareVersions() {
	latestVersionStr, err := getLatestVersionFromGitHub()
	if err != nil {
		fmt.Println("Could not check for updates:", err.Error())
		os.Exit(1)
	}

	currentVersion, err := version.NewVersion(cmd.VERSION[1:])
	if err != nil {
		log.Fatalf("Error parsing current version: %v", err)
	}

	latestVersion, err := version.NewVersion(latestVersionStr[1:])
	if err != nil {
		log.Fatalf("Error parsing latest version: %v", err)
	}

	if currentVersion.LessThan(latestVersion) {
		fmt.Printf("A new version of the tool is available: %s (you have %s)\n", latestVersionStr, cmd.VERSION)
		fmt.Println("Please update to the latest version.")
		fmt.Println("\n\tgo install github.com/vitessio/vitess-releaser@latest")
		os.Exit(1)
	}
}

// On some shells the terminal is left in a bad state possibly because the debug output is large or
// it contains control characters. This function restores the terminal to a sane state on a panic.
func restoreTerminal() {
	cmdSane := exec.Command("stty", "sane")
	cmdSane.Stdin = os.Stdin

	if err := cmdSane.Run(); err != nil {
		fmt.Println("Failed to restore terminal to sane state:", err)
	}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Vitess Releaser Panic:\n", r)
			debug.PrintStack()
			restoreTerminal()
		}
	}()

	compareVersions()

	cmd.Execute()
}
