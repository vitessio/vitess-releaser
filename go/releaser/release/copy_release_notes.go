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
	"github.com/vitessio/vitess-releaser/go/releaser/pre_release"
)

func CopyReleaseNotesToBranch(state *releaser.State, itemToUpdate *releaser.ItemWithLink, branch string) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 9,
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
		git.ResetHard(state.VitessRelease.Remote, state.VitessRelease.ReleaseBranch)

		git.Checkout(branch)
		git.ResetHard(state.VitessRelease.Remote, branch)

		prName := fmt.Sprintf("[%s] Copy `v%s` release notes", branch, state.VitessRelease.Release)

		pl.NewStepf("Look for an existing Pull Request named '%s'", prName)
		if _, url = github.FindPR(state.VitessRelease.Repo, prName); url != "" {
			pl.TotalSteps = 5 // only 5 total steps in this situation
			pl.NewStepf("An opened Pull Request was found: %s", url)
			done = true
			return url
		}

		pl.NewStepf("Create new branch based on %s/%s", state.VitessRelease.Remote, branch)
		newBranchName := git.FindNewGeneratedBranch(state.VitessRelease.Remote, branch, fmt.Sprintf("release-notes-%s", branch))

		pl.NewStepf("Copy release notes from %s/%s", state.VitessRelease.Remote, state.VitessRelease.ReleaseBranch)
		releaseNotesPath := pre_release.GetReleaseNotesDirPathForMajor(releaser.RemoveRCFromReleaseTitle(state.VitessRelease.Release))
		git.CheckoutPath(state.VitessRelease.Remote, state.VitessRelease.ReleaseBranch, releaseNotesPath)

		pl.NewStepf("Commit and push to branch %s", newBranchName)
		if git.CommitAll(fmt.Sprintf("Copy release notes from %s into %s", state.VitessRelease.ReleaseBranch, branch)) {
			pl.TotalSteps = 8 // only 8 total steps in this situation
			pl.NewStepf("Nothing to commit, seems like the release notes have already been copied")
			done = true
			return ""
		}
		git.Push(state.VitessRelease.Remote, newBranchName)

		pl.NewStepf("Create Pull Request")
		pr := github.PR{
			Title:  prName,
			Body:   fmt.Sprintf("This Pull Request copies the release notes found on `%s` to keep release notes up-to-date after the `v%s` release.", state.VitessRelease.ReleaseBranch, state.VitessRelease.Release),
			Branch: newBranchName,
			Base:   branch,
			Labels: []github.Label{{Name: "Component: General"}, {Name: "Type: Release"}},
		}
		_, url = pr.Create(state.IssueLink, state.VitessRelease.Repo)
		pl.NewStepf("Pull Request created %s", url)
		done = true
		return url
	}
}
