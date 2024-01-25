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

package git

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	errBranchExists = fmt.Errorf("branch already exists")
)

func checkCurrentRepo(repoWanted string) bool {
	out, err := exec.Command("git", "remote", "-v").CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
	outStr := string(out)
	return strings.Contains(outStr, repoWanted)
}

func cleanLocalState() bool {
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
	out, err := exec.Command("git", "fetch", remote).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}

	out, err = exec.Command("git", "reset", "--hard", remote+"/"+branch).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
}

func CreateBranchAndCheckout(branch, base string) error {
	out, err := exec.Command("git", "checkout", "-b", branch, base).CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), fmt.Sprintf("a branch named '%s' already exists", branch)) {
			return errBranchExists
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

func CommitAll(msg string) (empty bool) {
	out, err := exec.Command("git", "add", "--all").CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}

	out, err = exec.Command("git", "commit", "-n", "-s", "-m", msg).CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "nothing to commit, working tree clean") {
			return true
		}
		log.Fatalf("%s: %s", err, out)
	}
	return false
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

func CorrectCleanRepo(repo string) {
	if !checkCurrentRepo(repo + ".git") {
		log.Fatalf("failed to find remote %s in %s", repo, getWorkingDir())
	}
	if !cleanLocalState() {
		log.Fatalf("the %s repository should have a clean state", getWorkingDir())
	}
}

func getWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func FindNewGeneratedBranch(remote, baseBranch, branchName string) string {
	remoteAndBase := fmt.Sprintf("%s/%s", remote, baseBranch)

	var newBranch string
	for i := 1; ; i++ {
		newBranch = fmt.Sprintf("%s-%s-%d", baseBranch, branchName, i)
		err := CreateBranchAndCheckout(newBranch, remoteAndBase)
		if err != nil {
			if errors.Is(err, errBranchExists) {
				continue
			}
			log.Fatal(err)
		}
		break
	}
	return newBranch
}

func TagAndPush(remote, tag string) (exists bool) {
	out, err := exec.Command("git", "tag", tag).CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "already exists") {
			return true
		}
		log.Fatalf("%s: %s", err, out)
	}

	out, err = exec.Command("git", "push", remote, tag).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
	return false
}

func GetSHAForGitRef(ref string) string {
	out, err := exec.Command("git", "rev-parse", ref).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
	return strings.ReplaceAll(string(out), "\n", "")
}

func CheckoutPath(remote, branch, path string) {
	out, err := exec.Command("git", "checkout", fmt.Sprintf("%s/%s", remote, branch), path).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
}
