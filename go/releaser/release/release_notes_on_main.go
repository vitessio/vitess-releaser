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

package release

import (
	"fmt"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
	"vitess.io/vitess-releaser/go/releaser/pre_release"
)

func ReleaseNotesOnMain(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 9,
	}

	var done bool
	var url string
	return pl, func() string {
		defer func() {
			state.Issue.ReleaseNotesOnMain.Done = done
			state.Issue.ReleaseNotesOnMain.URL = url
			pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
			_, fn := state.UploadIssue()
			issueLink := fn()

			pl.NewStepf("Issue updated, see: %s", issueLink)
		}()

		pl.NewStepf("Fetch from git remote")
		git.CorrectCleanRepo(state.VitessRepo)
		git.ResetHard(state.Remote, state.ReleaseBranch)

		git.Checkout("main")
		git.ResetHard(state.Remote, "main")

		prName := fmt.Sprintf("Copy `v%s` release notes on `main`", state.Release)

		pl.NewStepf("Look for an existing Pull Request named '%s'", prName)
		if _, url = github.FindPR(state.VitessRepo, prName); url != "" {
			pl.TotalSteps = 5 // only 5 total steps in this situation
			pl.NewStepf("An opened Pull Request was found: %s", url)
			done = true
			return url
		}

		pl.NewStepf("Create new branch based on %s/main", state.Remote)
		newBranchName := git.FindNewGeneratedBranch(state.Remote, "main", "release-notes-main")

		pl.NewStepf("Copy release notes from %s/%s", state.Remote, state.ReleaseBranch)
		releaseNotesPath := pre_release.GetReleaseNotesDirPathForMajor(state.Release)
		git.CheckoutPath(state.Remote, state.ReleaseBranch, releaseNotesPath)

		pl.NewStepf("Commit and push to branch %s", newBranchName)
		if git.CommitAll(fmt.Sprintf("Copy release notes from %s into main", state.ReleaseBranch)) {
			pl.TotalSteps = 8 // only 8 total steps in this situation
			pl.NewStepf("Nothing to commit, seems like the release notes have already been copied")
			done = true
			return ""
		}
		git.Push(state.Remote, newBranchName)

		pl.NewStepf("Create Pull Request")
		pr := github.PR{
			Title:  prName,
			Body:   fmt.Sprintf("This Pull Request copies the release notes found on `%s` to keep release notes up-to-date after the `v%s` release.", state.ReleaseBranch, state.Release),
			Branch: newBranchName,
			Base:   "main",
			Labels: []github.Label{{Name: "Component: General"}, {Name: "Type: Release"}},
		}
		_, url = pr.Create(state.VitessRepo)
		pl.NewStepf("Pull Request created %s", url)
		done = true
		return url
	}
}
