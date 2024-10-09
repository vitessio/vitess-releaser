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
	"os"
	"strings"

	"github.com/vitessio/vitess-releaser/go/releaser/utils"
)

var (
	errBranchExists = fmt.Errorf("branch already exists")
)

func checkCurrentRepo(repoWanted string) bool {
	out := utils.Exec("git", "remote", "-v")
	return strings.Contains(out, repoWanted)
}

func cleanLocalState() bool {
	out := utils.Exec("git", "status", "-s")
	return len(out) == 0
}

func Checkout(branch string) {
	utils.Exec("git", "checkout", branch)
}

func ResetHard(remote, branch string) {
	utils.Exec("git", "fetch", remote)
	utils.Exec("git", "reset", "--hard", remote+"/"+branch)
}

func CreateBranchAndCheckout(branch, base string) error {
	out, err := utils.ExecWithError("git", "checkout", "-b", branch, base)
	if err != nil {
		if strings.Contains(out, fmt.Sprintf("a branch named '%s' already exists", branch)) {
			return errBranchExists
		}
		utils.BailOut(err, "got: %s", out)
	}
	return nil
}

func Push(remote, branch string) {
	utils.Exec("git", "push", remote, branch)
}

func CommitAll(msg string) (empty bool) {
	utils.Exec("git", "add", "--all")

	out, err := utils.ExecWithError(
		"git",
		"commit",
		"-n",
		"-s",
		"-m",
		msg,
		"-m",
		"This commit was made automatically by the vitess-releaser tool.",
		"-m",
		"See https://github.com/vitessio/vitess-releaser",
	)
	if err != nil {
		if strings.Contains(out, "nothing to commit, working tree clean") {
			return true
		}
		utils.BailOut(err, "got: %s", out)
	}
	return false
}

// FindRemoteName takes the output of `git remote -v` and a repository name,
// and returns the name of the remote associated with that repository.
// If no remote is found, an empty string is returned.
func FindRemoteName(repository string) string {
	out := utils.Exec("git", "remote", "-v")
	lines := strings.Split(out, "\n")
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
		utils.BailOut(nil, "failed to find remote %s in %s", repo, getWorkingDir())
	}
	if !cleanLocalState() {
		utils.BailOut(nil, "the %s repository should have a clean state", getWorkingDir())
	}
}

func getWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		utils.BailOut(err, "failed to find the current working dir")
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
			utils.BailOut(err, "bug should not get here")
		}
		break
	}
	return newBranch
}

func TagAndPush(remote, tag string) (exists bool) {
	out, err := utils.ExecWithError("git", "tag", tag)
	if err != nil {
		if strings.Contains(out, "already exists") {
			return true
		}
		utils.BailOut(err, "got: %s", out)
	}

	utils.Exec("git", "push", remote, tag)
	return false
}

func GetSHAForGitRef(ref string) string {
	out := utils.Exec("git", "rev-parse", ref)
	return strings.ReplaceAll(string(out), "\n", "")
}

func CheckoutPath(remote, branch, path string) {
	utils.Exec("git", "checkout", fmt.Sprintf("%s/%s", remote, branch), path)
}
