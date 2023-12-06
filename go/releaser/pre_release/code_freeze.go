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
	"errors"
	"fmt"
	"log"
	"os/exec"
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
func CodeFreeze(ctx *releaser.Context) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 10,
	}

	return pl, func() string {
		git.CorrectCleanRepo(ctx.VitessRepo)
		nextRelease, branchName := releaser.FindNextRelease(ctx.MajorRelease)

		pl.NewStepf("Fetch from git remote")
		remote := git.FindRemoteName(ctx.VitessRepo)
		git.ResetHard(remote, branchName)

		pl.NewStepf("Create new branch based on %s/%s", remote, branchName)
		newBranchName := findNewBranchForCodeFreeze(remote, branchName)

		pl.NewStepf("Turn on code freeze on branch %s", newBranchName)
		activateCodeFreeze()

		pl.NewStepf("Commit and push to branch %s", newBranchName)
		if git.CommitAll(fmt.Sprintf("Code Freeze of %s", branchName)) {
			pl.TotalSteps = 5
			pl.NewStepf("Nothing to commit, seems like code freeze is already done.", newBranchName)
			return ""
		}
		git.Push(remote, newBranchName)

		pl.NewStepf("Create Pull Request")
		pr := github.PR{
			Title:  fmt.Sprintf("[%s] Code Freeze for `v%s`", branchName, nextRelease),
			Body:   fmt.Sprintf("This Pull Request freezes the branch `%s` for `v%s`", branchName, nextRelease),
			Branch: newBranchName,
			Base:   branchName,
			Labels: []github.Label{{Name: "Component: General"}, {Name: "Type: Release"}},
		}
		nb, url := pr.Create(ctx.VitessRepo)
		pl.NewStepf("PR created %s", url)
		pl.NewStepf("Waiting for the PR to be merged. You must enable bypassing the branch protection rules in: https://github.com/vitessio/vitess/settings/branches")
	outer:
		for {
			select {
			case <-time.After(5 * time.Second):
				if github.IsPRMerged(ctx.VitessRepo, nb) {
					break outer
				}
			}
		}
		pl.NewStepf("PR has been merged")

		ctx.Issue.CodeFreeze.Done = true
		ctx.Issue.CodeFreeze.URL = url
		pl.NewStepf("Update Issue %s on GitHub", ctx.IssueLink)
		_, fn := ctx.UploadIssue()
		issueLink := fn()

		pl.NewStepf("Issue updated, see: %s", issueLink)
		return url
	}
}

func findNewBranchForCodeFreeze(remote, baseBranch string) string {
	remoteAndBase := fmt.Sprintf("%s/%s", remote, baseBranch)

	var newBranch string
	for i := 1; ; i++ {
		newBranch = fmt.Sprintf("%s-code-freeze-%d", baseBranch, i)
		err := git.CreateBranchAndCheckout(newBranch, remoteAndBase)
		if err != nil {
			if errors.Is(err, git.ErrBranchExists) {
				continue
			}
			log.Fatal(err)
		}
		break
	}
	return newBranch
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
