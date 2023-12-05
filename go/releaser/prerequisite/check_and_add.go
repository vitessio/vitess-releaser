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

package prerequisite

import (
	"fmt"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

func CheckAndAddPRsIssues(ctx *releaser.Context) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 3,
	}

	return pl, func() string {
		pl.NewStepf("Check and add Pull Requests")
		nbPRs, url := releaser.AddBackportPRs(ctx)

		pl.NewStepf("Check and add Release Blocker Issues")
		nbIssues, _ := releaser.AddReleaseBlockerIssues(ctx)

		msg := fmt.Sprintf("Up to date, see: %s", url)
		if nbPRs > 0 || nbIssues > 0 {
			msg = fmt.Sprintf("Found %d PRs and %d issues, see: %s", nbPRs, nbIssues, url)
		}
		pl.NewStepf(msg)
		return msg
	}
}
