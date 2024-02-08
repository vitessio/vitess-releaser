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
	"path"
	"strings"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
	"vitess.io/vitess-releaser/go/releaser/pre_release"
	"vitess.io/vitess-releaser/go/releaser/utils"
)

func TagRelease(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 6,
	}

	return pl, func() string {
		pl.NewStepf("Fetch from git remote")
		git.CorrectCleanRepo(state.VitessRelease.Repo)
		git.ResetHard(state.VitessRelease.Remote, state.VitessRelease.ReleaseBranch)

		// We want to transform the release name into lower case in case the release is an RC
		// Example: we will go from v19.0.0-RC1 to v19.0.0-rc1 which is a better format for our tags
		lowerCaseRelease := strings.ToLower(state.VitessRelease.Release)

		pl.NewStepf("Create and push the tags")
		gitTag := fmt.Sprintf("v%s", lowerCaseRelease)
		git.TagAndPush(state.VitessRelease.Remote, gitTag)

		// we also need to tag and push the Go doc tag
		// i.e. if we release v17.0.1, we also want to tag: v0.17.1
		nextReleaseSplit := strings.Split(lowerCaseRelease, ".")
		if len(nextReleaseSplit) != 3 {
			utils.LogPanic(nil, "%s was not formated x.x.x", state.VitessRelease.Release)
		}
		gdocGitTag := fmt.Sprintf("v0.%s.%s", nextReleaseSplit[0], nextReleaseSplit[2])
		git.TagAndPush(state.VitessRelease.Remote, gdocGitTag)

		pl.NewStepf("Create the release on the GitHub UI")
		releaseNotesPath := path.Join(pre_release.GetReleaseNotesDirPath(releaser.RemoveRCFromReleaseTitle(state.VitessRelease.Release)), "release_notes.md")
		url := github.CreateRelease(state.VitessRelease.Repo, gitTag, releaseNotesPath, state.VitessRelease.IsLatestRelease && state.Issue.RC == 0, state.Issue.RC > 0)

		pl.NewStepf("Done %s", url)
		state.Issue.TagRelease.Done = true
		state.Issue.TagRelease.URL = url
		pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
		_, fn := state.UploadIssue()
		issueLink := fn()

		pl.NewStepf("Issue updated, see: %s", issueLink)
		return url
	}
}
