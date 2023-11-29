/*
Copyright 2023 The Vitess Authors.

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

package git

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

var (
	ErrBranchExists = fmt.Errorf("branch already exists")
)

func CheckCurrentRepo(repoWanted string) bool {
	out, err := exec.Command("git", "remote", "-v").CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
	outStr := string(out)
	return strings.Contains(outStr, repoWanted)
}

func CleanLocalState() bool {
	out, err := exec.Command("git", "status", "-s").CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
	return len(out) == 0
}

func Checkout(branch string) {
	out, err := exec.Command("git", "checkout", branch).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
}

func Pull(remote, branch string) {
	out, err := exec.Command("git", "pull", remote, branch).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
}

func ResetHard(remote, branch string) {
	out, err := exec.Command("git", "reset", "--hard", remote+"/"+branch).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
}

func CreateBranchAndCheckout(branch, base string) error {
	out, err := exec.Command("git", "checkout", "-b", branch, base).CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), fmt.Sprintf("a branch named '%s' already exists", branch)) {
			return ErrBranchExists
		}
		log.Fatalf("%s: %s", err, out)
	}
	return nil
}

func Push(remote, branch string) {
	out, err := exec.Command("git", "push", remote, branch).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
}

func CommitAll(msg string) {
	out, err := exec.Command("git", "add", "--all").CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}

	out, err = exec.Command("git", "commit", "-n", "-s", "-m", msg).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
}

// FindRemoteName takes the output of `git remote -v` and a repository name,
// and returns the name of the remote associated with that repository.
// If no remote is found, an empty string is returned.
func FindRemoteName(repository string) string {
	out, err := exec.Command("git", "remote", "-v").CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
	gitRemoteOutput := string(out)

	lines := strings.Split(gitRemoteOutput, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			remoteName, remoteURL := parts[0], parts[1]
			if strings.Contains(remoteURL, repository) {
				return remoteName
			}
		}
	}
	return ""
}
