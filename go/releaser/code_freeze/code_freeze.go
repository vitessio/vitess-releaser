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

package code_freeze

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vitessio/vitess-releaser/go/releaser"
	"github.com/vitessio/vitess-releaser/go/releaser/git"
	"github.com/vitessio/vitess-releaser/go/releaser/github"
	"github.com/vitessio/vitess-releaser/go/releaser/logging"
	"github.com/vitessio/vitess-releaser/go/releaser/utils"
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
		pl.NewStepf("Waiting for the PR to be merged. You must enable bypassing the branch protection rules in: https://github.com/%s/settings/branches", state.VitessRelease.Repo)
	outer:
		for {
			select {
			case <-time.After(5 * time.Second):
				if github.IsPRMerged(state.VitessRelease.Repo, nb) {
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

		if state.Issue.RC == 1 {
			pl.NewStepf("Fetch from git remote and create branch %s", state.VitessRelease.ReleaseBranch)
		} else {
			pl.NewStepf("Fetch from git remote")
		}

		git.CorrectCleanRepo(state.VitessRelease.Repo)

		// For RC-1 we need to create two branches, the new release branch ("release-20.0")
		// and the rc release branch ("release-20.0-rc")
		if state.Issue.RC == 1 {
			git.ResetHard(state.VitessRelease.Remote, "main")

			if err := git.CreateBranchAndCheckout(state.VitessRelease.BaseReleaseBranch, fmt.Sprintf("%s/main", state.VitessRelease.Remote)); err != nil {
				git.Checkout(state.VitessRelease.BaseReleaseBranch)
			} else {
				git.Push(state.VitessRelease.Remote, state.VitessRelease.BaseReleaseBranch)
			}
		} else {
			git.ResetHard(state.VitessRelease.Remote, state.VitessRelease.ReleaseBranch)
		}

		codeFreezePRName := fmt.Sprintf("[%s] Code Freeze for `v%s`", state.VitessRelease.ReleaseBranch, state.VitessRelease.Release)

		// look for existing code freeze PRs
		pl.NewStepf("Look for an existing Code Freeze Pull Request named '%s'", codeFreezePRName)
		if nb, url = github.FindPR(state.VitessRelease.Repo, codeFreezePRName); url != "" {
			pl.TotalSteps = 7 // only 7 total steps in this situation
			pl.NewStepf("An opened Code Freeze Pull Request was found: %s", url)
			waitForPRToBeMerged(nb)
			done = true
			return url
		}

		// check if the branch is already frozen or not
		pl.NewStepf("Check if branch %s is already frozen", state.VitessRelease.ReleaseBranch)
		if isCurrentBranchFrozen() {
			pl.TotalSteps = 6 // only 6 total steps in this situation
			pl.NewStepf("Branch %s is already frozen, no action needed", state.VitessRelease.ReleaseBranch)
			done = true
			return ""
		}

		pl.NewStepf("Create new branch based on %s/%s", state.VitessRelease.Remote, state.VitessRelease.ReleaseBranch)
		newBranchName := git.FindNewGeneratedBranch(state.VitessRelease.Remote, state.VitessRelease.ReleaseBranch, "code-freeze")

		pl.NewStepf("Turn on code freeze on branch %s", newBranchName)
		activateCodeFreeze()

		pl.NewStepf("Commit and push to branch %s", newBranchName)
		if git.CommitAll(fmt.Sprintf("Code Freeze of %s", state.VitessRelease.ReleaseBranch)) {
			pl.TotalSteps = 9 // only 9 total steps in this situation
			pl.NewStepf("Nothing to commit, seems like code freeze is already done")
			done = true
			return ""
		}
		git.Push(state.VitessRelease.Remote, newBranchName)

		pl.NewStepf("Create Pull Request")
		pr := github.PR{
			Title:  codeFreezePRName,
			Body:   fmt.Sprintf("This Pull Request freezes the branch `%s` for `v%s`", state.VitessRelease.ReleaseBranch, state.VitessRelease.Release),
			Branch: newBranchName,
			Base:   state.VitessRelease.ReleaseBranch,
			Labels: []github.Label{{Name: "Component: General"}, {Name: "Type: Release"}},
		}
		nb, url = pr.Create(state.IssueLink, state.VitessRelease.Repo)
		pl.NewStepf("Pull Request created %s", url)
		waitForPRToBeMerged(nb)
		done = true
		return url
	}
}

func isCurrentBranchFrozen() bool {
	b, err := os.ReadFile(codeFreezeWorkflowFile)
	if err != nil {
		utils.BailOut(err, "failed to read file %s", codeFreezeWorkflowFile)
	}
	str := string(b)
	return strings.Contains(str, "exit 1")
}

func activateCodeFreeze() {
	changeCodeFreezeWorkflow(codeFreezeActivated)
}

func DeactivateCodeFreeze() {
	changeCodeFreezeWorkflow(codeFreezeDeactivated)
}

func changeCodeFreezeWorkflow(s codeFreezeStatus) {
	// sed -i.bak -E "s/exit (.*)/exit 0/g" $code_freeze_workflow
	utils.Exec("sed", "-i.bak", "-E", fmt.Sprintf("s/exit (.*)/exit %d/g", s), codeFreezeWorkflowFile)
	// remove backup file left by the sed command
	utils.Exec("rm", "-f", fmt.Sprintf("%s.bak", codeFreezeWorkflowFile))
}
