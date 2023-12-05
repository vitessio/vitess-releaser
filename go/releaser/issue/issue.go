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

package issue

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"

	"vitess.io/vitess-releaser/go/interactive/state"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
	"vitess.io/vitess-releaser/go/releaser/steps"
	"vitess.io/vitess-releaser/go/releaser/vitess"
)

const (
	markdownItemDone = "- [x]"
	markdownItemToDo = "- [ ]"

	slackAnnouncementStart = "<!-- SLACK_START -->"
	slackAnnouncementEnd   = "<!-- SLACK_END -->"
	slackAnnouncementItem  = "- [ ] Notify the community on Slack."
	slackAnnouncementFmt   = slackAnnouncementStart + "\n" + slackAnnouncementItem + "\n" + slackAnnouncementEnd

	checkSummaryStart = "<!-- SUMMARY_START -->"
	checkSummaryEnd   = "<!-- SUMMARY_START -->"
	checkSummaryItem  = "- [ ] Make sure the release notes summary is prepared and clean."
	checkSummaryFmt   = checkSummaryStart + "\n" + checkSummaryItem + "\n" + checkSummaryEnd

	// List of backports Pull Requests
	backportStart = "<!-- BACKPORT_START -->"
	backportEnd   = "<!-- BACKPORT_END -->"
	backportItem  = "- Make sure backport Pull Requests are merged, list below."
	backportFmt   = backportStart + "\n" + backportItem + "\n" + backportEnd

	// List of release blocker Issues
	releaseBlockerStart = "<!-- RELEASE_BLOCKER_START -->"
	releaseBlockerEnd   = "<!-- RELEASE_BLOCKER_END -->"
	releaseBlockerItem  = "- Make sure release blocker Issues are closed, list below."
	releaseBlockerFmt   = releaseBlockerStart + "\n" + releaseBlockerItem + "\n" + releaseBlockerEnd
)

type StepMeta struct {
	StartToken   string
	EndToken     string
	IssueItemStr string
}

var (
	stepBindings = map[string][]StepMeta{
		steps.SlackAnnouncement: {{
			StartToken:   slackAnnouncementStart,
			EndToken:     slackAnnouncementEnd,
			IssueItemStr: slackAnnouncementItem,
		}},
		steps.CheckAndAdd: {{
			StartToken:   backportStart,
			EndToken:     backportEnd,
			IssueItemStr: backportItem,
		}, {
			StartToken:   releaseBlockerStart,
			EndToken:     releaseBlockerEnd,
			IssueItemStr: releaseBlockerItem,
		}},
		steps.CheckSummary: {{
			StartToken:   checkSummaryStart,
			EndToken:     checkSummaryEnd,
			IssueItemStr: checkSummaryItem,
		}},
		steps.CodeFreeze:            {},
		steps.CreateMilestone:       {},
		steps.SlackAnnouncementPost: {},
	}
)

var (
	releaseIssueTemplate = fmt.Sprintf(
		`This release is scheduled for: TODO: '.Date' here .

<!-- Please DO NOT modify or remove the comments in this file. -->
<!-- Moreover, DO NOT add text in the middle of an _START and _END comment. -->

### Prerequisites for Release

%s
%s
%s
%s

### Pre-Release

- [ ] Follow Code-Freeze / Pre-Release instructions from release documentation.
- [ ] Create new GitHub Milestone.
- [ ] Create Pre-Release Pull Request.

### Release

- [ ] Follow release instructions.
- [ ] Merge release notes on main.
- [ ] Make sure Docker Images are available.
- [ ] Close previous GitHub Milestone.

### Post-Release

- [ ] Announce new release:
  - [ ] Slack
  - [ ] Twitter
`,
		slackAnnouncementFmt,
		checkSummaryFmt,
		backportFmt,
		releaseBlockerFmt,
	)
)

func CreateReleaseIssue(ctx *releaser.Context) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 2,
	}

	return pl, func() string {
		vitess.CorrectCleanRepo(ctx.VitessRepo)
		newRelease, _ := vitess.FindNextRelease(ctx.MajorRelease)

		pl.NewStepf("Create Release Issue on GitHub")
		tmpl := template.Must(template.New("release-issue").Parse(releaseIssueTemplate))
		b := bytes.NewBuffer(nil)
		err := tmpl.Execute(b, nil)
		if err != nil {
			log.Fatal(err)
		}

		newIssue := github.Issue{
			Title:    fmt.Sprintf("Release of v%s", newRelease),
			Body:     b.String(),
			Labels:   []github.Label{{Name: "Component: General"}, {Name: "Type: Release"}},
			Assignee: "@me",
		}

		link := newIssue.Create(ctx.VitessRepo)
		pl.NewStepf("Issue created: %s", link)
		return link
	}
}

// func ReadStepStatus(ctx *releaser.Context, step string) string {
// 	binding, ok := stepBindings[step]
// 	if !ok {
// 		log.Fatalf("unknown step: %s", step)
// 	}
//
//
//
// 	return ""
// }

func InverseStepStatus(ctx *releaser.Context, step string) (*logging.ProgressLogging, func()) {
	binding, ok := stepBindings[step]
	if !ok {
		log.Fatalf("unknown step: %s", step)
	}

	pl := &logging.ProgressLogging{TotalSteps: 1 + len(binding)}
	return pl, func() {
		pl.NewStepf("Update status for '%s' on the Release Issue", step)

		issueNb := github.GetReleaseIssueNumber(ctx)
		body := github.GetIssueBody(ctx.VitessRepo, issueNb)

		for _, meta := range binding {
			start, end, err := getIssueTextBetweenTokens(meta.StartToken, meta.EndToken, body)
			if err != nil {
				log.Fatal(err.Error())
			}

			content := body[start:end]

			isDone := getStepStatus(content)

			// proceed to inverse
			if isDone {
				content = markdownItemToDo + content[len(markdownItemDone):]
			} else {
				content = markdownItemDone + content[len(markdownItemToDo):]
			}

			updateSegmentOfIssue(ctx, body, content, start, end, issueNb)

			pl.NewStepf("Item marked as '%s'", state.Fmt(!isDone))
		}
	}
}

func getStepStatus(body string) bool {
	if strings.HasPrefix(body, markdownItemDone) {
		return state.Done
	} else if strings.HasPrefix(body, markdownItemToDo) {
		return state.ToDo
	}
	log.Fatalf("unknown step status: %s", body)
	return false
}

func AddBackportPRs(ctx *releaser.Context) (int, string) {
	issueNb := github.GetReleaseIssueNumber(ctx)
	body := github.GetIssueBody(ctx.VitessRepo, issueNb)

	// we must figure out what is the index of the BACKPORT_START comment
	// in our issue's body, and what is the index of the BACKPORT_END comment too.
	// once we have those, we will be able to get the list of Pull Requests in text,
	// which will then need to be parsed.
	start, end, err := getIssueTextBetweenTokens(backportStart, backportEnd, body)
	if err != nil {
		log.Fatal(err.Error())
	}
	textPullRequest := body[start:end]
	prsInIssue := parseMarkdownCheckboxListWithIssuePRsLinks(ctx.VitessRepo, textPullRequest)
	prsChecked := github.CheckBackportToPRs(ctx)

outer:
	for _, pr := range prsChecked {
		nb := pr.URL[strings.LastIndex(pr.URL, "/")+1:]
		for _, pri := range prsInIssue {
			if pri.nb == nb {
				continue outer
			}
		}
		prsInIssue = append(prsInIssue, prsIssuesListItem{
			nb: nb,
		})
	}

	listURLs := make([]string, 0, len(prsInIssue)+1)
	listURLs = append(listURLs, backportItem)
	prNotDoneCount := 0
	for _, item := range prsInIssue {
		done := "x"
		if !item.done {
			done = " "
			prNotDoneCount++
		}
		listURLs = append(listURLs, fmt.Sprintf("  - [%s] #%s", done, item.nb))
	}

	newList := fmt.Sprintf("\n%s\n", strings.Join(listURLs, "\n"))
	url := updateSegmentOfIssue(ctx, body, newList, start, end, issueNb)
	return prNotDoneCount, url
}

func AddReleaseBlockerIssues(ctx *releaser.Context) (int, string) {
	issueNb := github.GetReleaseIssueNumber(ctx)
	body := github.GetIssueBody(ctx.VitessRepo, issueNb)

	start, end, err := getIssueTextBetweenTokens(releaseBlockerStart, releaseBlockerEnd, body)
	if err != nil {
		log.Fatal(err.Error())
	}
	textPullRequest := body[start:end]

	issuesInIssue := parseMarkdownCheckboxListWithIssuePRsLinks(ctx.VitessRepo, textPullRequest)
	issuesChecked := github.CheckReleaseBlockerIssues(ctx)

outer:
	for _, issueChecked := range issuesChecked {
		nb := issueChecked.URL[strings.LastIndex(issueChecked.URL, "/")+1:]
		for _, i := range issuesInIssue {
			if i.nb == nb {
				continue outer
			}
		}
		issuesInIssue = append(issuesInIssue, prsIssuesListItem{
			nb: nb,
		})
	}

	listURLs := make([]string, 0, len(issuesInIssue)+1)
	listURLs = append(listURLs, releaseBlockerItem)
	issueNotDone := 0
	for _, item := range issuesInIssue {
		done := "x"
		if !item.done {
			done = " "
			issueNotDone++
		}
		listURLs = append(listURLs, fmt.Sprintf("  - [%s] #%s", done, item.nb))
	}

	newList := fmt.Sprintf("\n%s\n", strings.Join(listURLs, "\n"))
	url := updateSegmentOfIssue(ctx, body, newList, start, end, issueNb)
	return issueNotDone, url
}

func updateSegmentOfIssue(ctx *releaser.Context, body, replaceBy string, startIdx, endIdx, issueNb int) string {
	body = body[:startIdx] + replaceBy + body[endIdx:]

	issue := github.Issue{Body: body, Number: issueNb}
	url := issue.UpdateBody(ctx.VitessRepo)
	return url
}

func getIssueTextBetweenTokens(tokenStart, tokenEnd, body string) (start, end int, err error) {
	start = strings.Index(body, tokenStart)
	if start == -1 {
		return 0, 0, fmt.Errorf("could not parse the issue, %s not found", tokenStart)
	}
	start += len(tokenStart) + 1

	end = strings.Index(body, tokenEnd)
	if end == -1 {
		return 0, 0, fmt.Errorf("could not parse the issue, %s not found", tokenEnd)
	}
	return
}

type prsIssuesListItem struct {
	done bool
	nb   string
}

// parseMarkdownCheckboxListWithIssuePRsLinks takes in a Markdown text that has a list of checkboxes
// parse it and return a slice of prsIssuesListItem. The items in the list must contain links to either Issues or PRs.
// Example value for "body":
//
// - [ ] https://github.com/vitessio/vitess/pull/1
// - [x] https://github.com/vitessio/vitess/issue/1000
func parseMarkdownCheckboxListWithIssuePRsLinks(repo, body string) []prsIssuesListItem {
	lines := strings.Split(body, "\n")

	var lis []prsIssuesListItem
	for _, line := range lines {
		// check that the item begins with a tab and a Markdown checkbox
		const prefix = "  - ["
		if !strings.HasPrefix(line, prefix) || len(line) <= len(prefix)+2 {
			continue
		}
		// move the cursor to the interior of the checkbox
		line = line[len(prefix):]

		var newItem prsIssuesListItem

		// check if the item has been marked as done or not
		if line[0] == 'x' {
			newItem.done = true
		}

		// move cursor after the Markdown checkbox and clear spaces
		line = strings.TrimSpace(line[2:])

		// fetch the number of the referenced issues/pr, it can be in two forms:
		// 		- using an '#' such as: #1
		// 		- using a direct link such as: https://github.com/a/b/pull/1
		if strings.HasPrefix(line, "#") {
			newItem.nb = line[1:]
		} else if strings.HasPrefix(line, "https://") {
			newItem.nb = line[strings.LastIndex(line, "/")+1:]
		}

		lis = append(lis, newItem)
	}
	return lis
}
