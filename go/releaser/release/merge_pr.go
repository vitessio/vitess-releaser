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
	"log"
	"strconv"
	"strings"
	"time"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

func MergeReleasePR(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 5,
	}

	return pl, func() string {
		pl.NewStepf("Resolve Release Pull Request URL")
		url := state.Issue.CreateReleasePR.URL
		nb, err := strconv.Atoi(url[strings.LastIndex(url, "/")+1:])
		if err != nil {
			log.Fatal(err)
		}

		pl.NewStepf("Waiting for %s to be merged", url)
	outer:
		for {
			select {
			case <-time.After(5 * time.Second):
				if github.IsPRMerged(state.VitessRelease.Repo, nb) {
					break outer
				}
			}
		}
		pl.NewStepf("Pull Request has been merged")
		state.Issue.MergeReleasePR.Done = true
		state.Issue.MergeReleasePR.URL = url
		pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
		_, fn := state.UploadIssue()
		issueLink := fn()

		pl.NewStepf("Issue updated, see: %s", issueLink)
		return url
	}
}
