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
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/vitessio/vitess-releaser/go/releaser"
	"github.com/vitessio/vitess-releaser/go/releaser/github"
	"github.com/vitessio/vitess-releaser/go/releaser/logging"
	"github.com/vitessio/vitess-releaser/go/releaser/utils"
)

func VtopMergeReleasePR(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 6,
	}

	return pl, func() string {
		state.GoToVtOp()
		defer state.GoToVitess()

		pl.NewStepf("Resolve Release Pull Request URL")
		url := state.Issue.VtopCreateReleasePR.URL
		nb, err := strconv.Atoi(url[strings.LastIndex(url, "/")+1:])
		if err != nil {
			utils.BailOut(err, "failed to parse the PR number from GitHub URL: %s", url)
		}

		pl.NewStepf("Waiting for %s to be merged", url)

		// The vtop release PR is created at the very last minute when the Vitess release just
		// get released, in this case, CI may not have time to complete before reaching the
		// 'vitess-operator merge release PR'. If the release team runs this step, but the PR
		// is not green or ready to be merged yet, then they will be forced to 'kill' the vitess-releaser
		// process in order to get out of the current step, for this specific situation we are catching
		// interruption signals to allow the release team to cancel this step if they realize they're not ready.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		defer signal.Stop(c)
		pl.NewStepf("If the PR is not ready, this step may be canceled cleanly by doing 'CMD+C'/'CTRL+C'.")

	outer:
		for {
			select {
			case <-time.After(5 * time.Second):
				if github.IsPRMerged(state.VtOpRelease.Repo, nb) {
					break outer
				}
			case <-c:
				pl.TotalSteps -= 2
				pl.NewStepf("Interruption detected, canceling the step.")
				return url
			}
		}

		pl.NewStepf("Pull Request has been merged")
		state.Issue.VtopMergeReleasePR.Done = true
		state.Issue.VtopMergeReleasePR.URL = url
		pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
		_, fn := state.UploadIssue()
		issueLink := fn()

		pl.NewStepf("Issue updated, see: %s", issueLink)
		return url
	}
}
