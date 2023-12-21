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

package pre_release

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

type codeFreezeStatus int

const (
	codeFreezeDeactivated codeFreezeStatus = iota
	codeFreezeActivated

	codeFreezeWorkflowFile = "./.github/workflows/code_freeze.yml"
)

// CodeFreeze will freeze the branch of the next release we want to release.
// The function returns the URL of the code freeze Pull Request, this Pull
// Request must be forced-merged by a Vitess maintainer, this step cannot be automated.
func CodeFreeze(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 12,
	}

	waitForPRToBeMerged := func(nb int) {
		pl.NewStepf("Waiting for the PR to be merged. You must enable bypassing the branch protection rules in: https://github.com/vitessio/vitess/settings/branches")
	outer:
		for {
			select {
			case <-time.After(5 * time.Second):
				if github.IsPRMerged(state.VitessRepo, nb) {
					break outer
				}
			}
		}
		pl.NewStepf("PR has been merged")
	}

	var done bool
	var url string
	var nb int
	return pl, func() string {
		defer func() {
			state.Issue.CodeFreeze.Done = done
			state.Issue.CodeFreeze.URL = url
			pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
			_, fn := state.UploadIssue()
			issueLink := fn()

			pl.NewStepf("Issue updated, see: %s", issueLink)
		}()

		pl.NewStepf("Fetch from git remote")
		git.CorrectCleanRepo(state.VitessRepo)
		git.ResetHard(state.Remote, state.ReleaseBranch)

		codeFreezePRName := fmt.Sprintf("[%s] Code Freeze for `v%s`", state.ReleaseBranch, state.Release)

		// look for existing code freeze PRs
		pl.NewStepf("Look for an existing Code Freeze Pull Request named '%s'", codeFreezePRName)
		if nb, url = github.FindPR(state.VitessRepo, codeFreezePRName); url != "" {
			pl.TotalSteps = 7 // only 7 total steps in this situation
			pl.NewStepf("An opened Code Freeze Pull Request was found: %s", url)
			waitForPRToBeMerged(nb)
			done = true
			return url
		}

		// check if the branch is already frozen or not
		pl.NewStepf("Check if branch %s is already frozen", state.ReleaseBranch)
		if isCurrentBranchFrozen() {
			pl.TotalSteps = 6 // only 6 total steps in this situation
			pl.NewStepf("Branch %s is already frozen, no action needed.", state.ReleaseBranch)
			done = true
			return ""
		}

		pl.NewStepf("Create new branch based on %s/%s", state.Remote, state.ReleaseBranch)
		newBranchName := git.FindNewGeneratedBranch(state.Remote, state.ReleaseBranch, "code-freeze")

		pl.NewStepf("Turn on code freeze on branch %s", newBranchName)
		activateCodeFreeze()

		pl.NewStepf("Commit and push to branch %s", newBranchName)
		if git.CommitAll(fmt.Sprintf("Code Freeze of %s", state.ReleaseBranch)) {
			pl.TotalSteps = 9 // only 9 total steps in this situation
			pl.NewStepf("Nothing to commit, seems like code freeze is already done.")
			done = true
			return ""
		}
		git.Push(state.Remote, newBranchName)

		pl.NewStepf("Create Pull Request")
		pr := github.PR{
			Title:  codeFreezePRName,
			Body:   fmt.Sprintf("This Pull Request freezes the branch `%s` for `v%s`", state.ReleaseBranch, state.Release),
			Branch: newBranchName,
			Base:   state.ReleaseBranch,
			Labels: []github.Label{{Name: "Component: General"}, {Name: "Type: Release"}},
		}
		nb, url = pr.Create(state.VitessRepo)
		pl.NewStepf("Pull Request created %s", url)
		waitForPRToBeMerged(nb)
		done = true
		return url
	}
}

func isCurrentBranchFrozen() bool {
	b, err := os.ReadFile(codeFreezeWorkflowFile)
	if err != nil {
		log.Fatal(err)
	}
	str := string(b)
	return strings.Contains(str, "exit 1")
}

func activateCodeFreeze() {
	changeCodeFreezeWorkflow(codeFreezeActivated)
}

func deactivateCodeFreeze() {
	changeCodeFreezeWorkflow(codeFreezeDeactivated)
}

func changeCodeFreezeWorkflow(s codeFreezeStatus) {
	// sed -i.bak -E "s/exit (.*)/exit 0/g" $code_freeze_workflow
	out, err := exec.Command("sed", "-i.bak", "-E", fmt.Sprintf("s/exit (.*)/exit %d/g", s), codeFreezeWorkflowFile).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}

	// remove backup file left by the sed command
	out, err = exec.Command("rm", "-f", fmt.Sprintf("%s.bak", codeFreezeWorkflowFile)).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
}
