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
	"log"
	"path"
	"strings"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
	"vitess.io/vitess-releaser/go/releaser/pre_release"
)

func TagRelease(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 6,
	}

	return pl, func() string {
		pl.NewStepf("Fetch from git remote")
		git.CorrectCleanRepo(state.VitessRepo)
		remote := git.FindRemoteName(state.VitessRepo)
		git.ResetHard(remote, state.ReleaseBranch)

		pl.NewStepf("Create and push the tags")
		gitTag := fmt.Sprintf("v%s", state.Release)
		git.TagAndPush(remote, gitTag)
		// we also need to tag and push the Go doc tag
		// i.e. if we release v17.0.1, we also want to tag: v0.17.1
		nextReleaseSplit := strings.Split(state.Release, ".")
		if len(nextReleaseSplit) != 3 {
			log.Fatalf("%s was not formated x.x.x", state.Release)
		}
		git.TagAndPush(remote, fmt.Sprintf("v0.%s.%s", nextReleaseSplit[0], nextReleaseSplit[2]))

		pl.NewStepf("Create the release on the GitHub UI")
		releaseNotesPath := path.Join(pre_release.GetReleaseNotesDirPath(state.Release), "release_notes.md")
		url := github.CreateRelease(state.VitessRepo, gitTag, releaseNotesPath, state.IsLatestRelease)

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
