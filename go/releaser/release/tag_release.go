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
	"strings"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

func TagRelease(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 12,
	}

	return pl, func() string {
		git.CorrectCleanRepo(state.VitessRepo)
		nextRelease, branchName := releaser.FindNextRelease(state.MajorRelease)

		pl.NewStepf("Fetch from git remote")
		remote := git.FindRemoteName(state.VitessRepo)
		git.ResetHard(remote, branchName)

		pl.NewStepf("Create and push the tags")
		git.TagAndPush(remote, fmt.Sprintf("v%s", nextRelease))
		// we also need to tag and push the Go doc tag
		// i.e. if we release v17.0.1, we also want to tag: v0.17.1
		nextReleaseSplit := strings.Split(nextRelease, ".")
		if len(nextReleaseSplit) != 3 {
			log.Fatalf("%s was not formated x.x.x", nextRelease)
		}
		git.TagAndPush(remote, fmt.Sprintf("v0.%s.%s", nextReleaseSplit[0], nextReleaseSplit[2]))

		// TODO: create the release on the UI

		return ""
	}
}
