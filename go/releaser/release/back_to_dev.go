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

package release

import (
	"fmt"

	"github.com/vitessio/vitess-releaser/go/releaser"
	"github.com/vitessio/vitess-releaser/go/releaser/git"
	"github.com/vitessio/vitess-releaser/go/releaser/github"
	"github.com/vitessio/vitess-releaser/go/releaser/logging"
)

func BackToDevModeOnBranch(state *releaser.State, itemToUpdate *releaser.ItemWithLink, branch string) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 10,
	}

	var done bool
	var url string
	return pl, func() string {
		defer func() {
			itemToUpdate.Done = done
			itemToUpdate.URL = url
			pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
			_, fn := state.UploadIssue()
			issueLink := fn()

			pl.NewStepf("Issue updated, see: %s", issueLink)
		}()

		pl.NewStepf("Fetch from git remote")
		git.CorrectCleanRepo(state.VitessRelease.Repo)
		git.ResetHard(state.VitessRelease.Remote, branch)

		// If we are releasing an RC release, the next SNAPSHOT version on the release branch
		// will be the same release as the RC but without the RC tag.
		var nextNextRelease string
		if state.Issue.RC > 0 {
			nextNextRelease = releaser.RemoveRCFromReleaseTitle(state.VitessRelease.Release)
		} else {
			nextNextRelease = releaser.FindVersionAfterNextRelease(state)
		}

		devModeRelease := fmt.Sprintf("%s-SNAPSHOT", nextNextRelease)

		backToDevModePRName := fmt.Sprintf("[%s] Bump to `v%s` after the `v%s` release", branch, devModeRelease, state.VitessRelease.Release)

		// look for existing PRs
		pl.NewStepf("Look for an existing Pull Request named '%s'", backToDevModePRName)
		if _, url = github.FindPR(state.VitessRelease.Repo, backToDevModePRName); url != "" {
			pl.TotalSteps = 5 // only 5 total steps in this situation
			pl.NewStepf("An opened Pull Request was found: %s", url)
			done = true
			return url
		}

		pl.NewStepf("Create new branch based on %s/%s", state.VitessRelease.Remote, branch)
		newBranchName := git.FindNewGeneratedBranch(state.VitessRelease.Remote, branch, "back-to-dev-mode")

		pl.NewStepf("Update version.go")
		releaser.UpdateVersionGoFile(devModeRelease)

		pl.NewStepf("Update the Java directory")
		releaser.UpdateJavaDir(devModeRelease)

		pl.NewStepf("Commit and push to branch %s", newBranchName)
		if git.CommitAll(fmt.Sprintf("Back to dev mode: %s", backToDevModePRName)) {
			pl.TotalSteps = 9 // only 9 total steps in this situation
			pl.NewStepf("Nothing to commit, seems like back to dev mode is already done")
			done = true
			return ""
		}
		git.Push(state.VitessRelease.Remote, newBranchName)

		pl.NewStepf("Create Pull Request")
		pr := github.PR{
			Title:  backToDevModePRName,
			Body:   fmt.Sprintf("Includes the changes required to go back into dev mode (v%s) after the release of v%s.", devModeRelease, state.VitessRelease.Release),
			Branch: newBranchName,
			Base:   branch,
			Labels: []github.Label{{Name: "Component: General"}, {Name: "Type: Release"}},
		}
		_, url = pr.Create(state.VitessRelease.Repo)
		pl.NewStepf("Pull Request created %s", url)
		done = true
		return ""
	}
}
