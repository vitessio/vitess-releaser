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
	"strings"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

func CheckAndAddPRsIssues(ctx *releaser.Context) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 4,
	}

	return pl, func() string {
		pl.NewStepf("Check and add Pull Requests")
		prsOnGH := github.CheckBackportToPRs(ctx.VitessRepo, ctx.MajorRelease)
	outerPR:
		for _, pr := range prsOnGH {
			// separate the PR number from the URL
			nb := pr.URL[strings.LastIndex(pr.URL, "/")+1:]
			markdownURL := fmt.Sprintf("#%s", nb)
			for _, pri := range ctx.Issue.CheckBackports.Items {
				if pri.URL == markdownURL {
					continue outerPR
				}
			}
			ctx.Issue.CheckBackports.Items = append(ctx.Issue.CheckBackports.Items, releaser.ItemWithLink{
				URL: markdownURL,
			})
		}

		pl.NewStepf("Check and add Release Blocker Issues")
		issuesOnGH := github.CheckReleaseBlockerIssues(ctx.VitessRepo, ctx.MajorRelease)
	outerRBI:
		for _, i := range issuesOnGH {
			// separate the Issue number from the URL
			nb := i.URL[strings.LastIndex(i.URL, "/")+1:]
			markdownURL := fmt.Sprintf("#%s", nb)
			for _, rbi := range ctx.Issue.ReleaseBlocker.Items {
				if rbi.URL == markdownURL {
					continue outerRBI
				}
			}
			ctx.Issue.ReleaseBlocker.Items = append(ctx.Issue.ReleaseBlocker.Items, releaser.ItemWithLink{
				URL: markdownURL,
			})
		}

		pl.NewStepf("Update Issue %s on GitHub", ctx.IssueLink)
		_, fn := ctx.UploadIssue()
		link := fn()

		msg := GetCheckAndAddInfoMsg(ctx, link)
		pl.NewStepf(msg)
		return msg
	}
}

func GetCheckAndAddInfoMsg(ctx *releaser.Context, link string) string {
	nbPRs, nbIssues := ctx.Issue.CheckBackports.ItemsLeft(), ctx.Issue.ReleaseBlocker.ItemsLeft()

	msg := fmt.Sprintf("Up to date, see: %s", link)
	if nbPRs > 0 || nbIssues > 0 {
		msg = fmt.Sprintf("Found %d PRs and %d issues, see: %s", nbPRs, nbIssues, link)
	}
	return msg
}
