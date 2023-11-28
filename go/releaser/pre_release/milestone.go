/*
Copyright 2023 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreedto in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pre_release

import (
	"fmt"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
	"vitess.io/vitess-releaser/go/releaser/vitess"
)

func NewMilestone(ctx *releaser.Context) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 3,
	}

	return pl, func() string {
		pl.NewStepf("Finding the next Milestone")
		nextNextRelease := vitess.FindVersionAfterNextRelease(ctx)
		newMilestone := fmt.Sprintf("v%s", nextNextRelease)

		ms := github.GetMilestonesByName(ctx.VitessRepo, newMilestone)
		if len(ms) > 0 {
			pl.SetTotalStep(2)
			pl.NewStepf("Found an existing Milestone: %s", ms[0].URL)
			return ms[0].URL
		}

		pl.NewStepf("Creating Milestone %s on GitHub", newMilestone)
		link := github.CreateNewMilestone(ctx.VitessRepo, newMilestone)

		pl.NewStepf("New Milestone %s created: %s", newMilestone, link)
		return link
	}
}