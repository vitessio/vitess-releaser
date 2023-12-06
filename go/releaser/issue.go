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

package releaser

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"

	"vitess.io/vitess-releaser/go/interactive/state"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

const (
	stateReadingItem = iota
	stateReadingBackport
	stateReadingReleaseBlockerIssue
)

const (
	markdownItemDone = "- [x]"
	markdownItemToDo = "- [ ]"

	// Prerequisites
	preSlackAnnouncementItem = "Notify the community on Slack."
	checkSummaryItem         = "Make sure the release notes summary is prepared and clean."
	backportItem             = "Make sure backport Pull Requests are merged, list below."
	releaseBlockerItem       = "Make sure release blocker Issues are closed, list below."

	// Pre-Release
	codeFreezeItem   = "Code Freeze."
	newMilestoneItem = "Create new GitHub Milestone."

	// Post-Release
	postSlackAnnouncementItem = "Notify the community on Slack for the new release."
)

type (
	ItemWithLink struct {
		Done bool

		// URL always uses the following format: "#111"
		// No http links are used, only Markdown links
		URL string
	}

	ParentOfItems struct {
		Items []ItemWithLink
	}

	Issue struct {
		// Prerequisites
		SlackPreRequisite bool
		CheckSummary      bool
		CheckBackports    ParentOfItems
		ReleaseBlocker    ParentOfItems

		// Pre-Release
		CodeFreeze         ItemWithLink
		NewGitHubMilestone ItemWithLink

		// Post-Release
		SlackPostRelease bool
	}
)

const (
	releaseIssueTemplate = `This release is scheduled for ...

### Prerequisites for Release

- [{{fmtStatus .SlackPreRequisite}}] Notify the community on Slack.
- [{{fmtStatus .CheckSummary}}] Make sure the release notes summary is prepared and clean.
- [{{fmtStatus .CheckBackports.Done}}] Make sure backport Pull Requests are merged, list below.
{{- range $item := .CheckBackports.Items }}
  - [{{fmtStatus $item.Done}}] {{$item.URL}}
{{- end }}
- [{{fmtStatus .ReleaseBlocker.Done}}] Make sure release blocker Issues are closed, list below.
{{- range $item := .ReleaseBlocker.Items }}
  - [{{fmtStatus $item.Done}}] {{$item.URL}}
{{- end }}


### Pre-Release

- [{{fmtStatus .CodeFreeze.Done}}] Code Freeze.
{{- if .CodeFreeze.URL }}
  - {{ .CodeFreeze.URL }}
{{- end }}
- [{{fmtStatus .NewGitHubMilestone.Done}}] Create new GitHub Milestone.
{{- if .NewGitHubMilestone.URL }}
  - {{ .NewGitHubMilestone.URL }}
{{- end }}

### Post-Release
- [{{fmtStatus .SlackPostRelease}}] Notify the community on Slack for the new release.
`
)

func (pi ParentOfItems) ItemsLeft() int {
	nb := 0
	for _, item := range pi.Items {
		if !item.Done {
			nb++
		}
	}
	return nb
}

func (pi ParentOfItems) Done() bool {
	for _, item := range pi.Items {
		if !item.Done {
			return false
		}
	}
	return true
}

func (ctx *Context) LoadIssue() {
	if ctx.IssueNbGH == 0 {
		// we are in the case where we start vitess-releaser
		// and the Release Issue hasn't been created yet.
		// We simply quit, the issue is left empty, nothing to load.
		return
	}

	body := github.GetIssueBody(ctx.VitessRepo, ctx.IssueNbGH)

	lines := strings.Split(body, "\n")

	var newIssue Issue

	s := stateReadingItem
	for i, line := range lines {
		switch s {
		case stateReadingItem:
			if strings.Contains(line, preSlackAnnouncementItem) {
				newIssue.SlackPreRequisite = strings.HasPrefix(line, markdownItemDone)
			}
			if strings.Contains(line, checkSummaryItem) {
				newIssue.CheckSummary = strings.HasPrefix(line, markdownItemDone)
			}
			if strings.Contains(line, backportItem) {
				s = stateReadingBackport
			}
			if strings.Contains(line, releaseBlockerItem) {
				s = stateReadingReleaseBlockerIssue
			}
		case stateReadingBackport:
			if !strings.HasPrefix(line, "  -") {
				s = stateReadingItem
				continue
			}
			// remove indentation from line after we have confirmed it is present
			line = strings.TrimSpace(line)
			newIssue.CheckBackports.Items = append(newIssue.CheckBackports.Items, ItemWithLink{
				Done: strings.HasPrefix(line, markdownItemDone),
				URL:  strings.TrimSpace(line[len(markdownItemDone):]),
			})
			if i+1 == len(lines) || !strings.HasPrefix(lines[i+1], "  -") {
				s = stateReadingItem
			}
		case stateReadingReleaseBlockerIssue:
			if !strings.HasPrefix(line, "  -") {
				s = stateReadingItem
				continue
			}
			// remove indentation from line after we have confirmed it is present
			line = strings.TrimSpace(line)
			newIssue.ReleaseBlocker.Items = append(newIssue.ReleaseBlocker.Items, ItemWithLink{
				Done: strings.HasPrefix(line, markdownItemDone),
				URL:  strings.TrimSpace(line[len(markdownItemDone):]),
			})
			if i+1 == len(lines) || !strings.HasPrefix(lines[i+1], "  -") {
				s = stateReadingItem
			}
		}
	}
	ctx.Issue = newIssue
}

func (ctx *Context) UploadIssue() (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 2,
	}

	return pl, func() string {
		pl.NewStepf("Update Issue #%d on GitHub", ctx.IssueNbGH)
		body := ctx.Issue.toString()
		issue := github.Issue{Body: body, Number: ctx.IssueNbGH}
		link := issue.UpdateBody(ctx.VitessRepo)
		pl.NewStepf("Issue updated: %s", link)
		return link
	}
}

func CreateReleaseIssue(ctx *Context) (*logging.ProgressLogging, func() (int, string)) {
	pl := &logging.ProgressLogging{
		TotalSteps: 2,
	}

	return pl, func() (int, string) {
		CorrectCleanRepo(ctx.VitessRepo)
		newRelease, _ := FindNextRelease(ctx.MajorRelease)

		var i Issue
		pl.NewStepf("Create Release Issue on GitHub")
		newIssue := github.Issue{
			Title:    fmt.Sprintf("Release of v%s", newRelease),
			Body:     i.toString(),
			Labels:   []github.Label{{Name: "Component: General"}, {Name: "Type: Release"}},
			Assignee: "@me",
		}

		link := newIssue.Create(ctx.VitessRepo)
		nb := github.URLToNb(link)
		pl.NewStepf("Issue created: %s", link)
		return nb, link
	}
}

func (i *Issue) toString() string {
	tmpl := template.New("release-issue")
	tmpl = tmpl.Funcs(template.FuncMap{
		"fmtStatus": state.FmtMd,
	})

	parsed, err := tmpl.Parse(releaseIssueTemplate)
	if err != nil {
		log.Fatal(err)
	}
	b := bytes.NewBufferString("")
	err = parsed.Execute(b, i)
	if err != nil {
		log.Fatal(err)
	}
	return b.String()
}

func InverseStepStatus(step string) (*logging.ProgressLogging, func()) {
	pl := &logging.ProgressLogging{TotalSteps: 1}
	return pl, func() {
		pl.NewStepf("Update status for '%s' on the Release Issue", step)
	}
}

func AddBackportPRs(ctx *Context) (int, string) {
	_ = github.GetIssueBody(ctx.VitessRepo, ctx.IssueNbGH)

	return 0, ""
}

func AddReleaseBlockerIssues(ctx *Context) (int, string) {
	_ = github.GetIssueBody(ctx.VitessRepo, ctx.IssueNbGH)

	return 0, ""
}
