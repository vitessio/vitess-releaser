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

	slackAnnouncementItem = "Notify the community on Slack."
	checkSummaryItem      = "Make sure the release notes summary is prepared and clean."
	backportItem          = "Make sure backport Pull Requests are merged, list below."
	releaseBlockerItem    = "Make sure release blocker Issues are closed, list below."
)

type itemWithLink struct {
	Done bool
	URL  string
}

type parentItem struct {
	Items []itemWithLink
}

type Issue struct {
	SlackPreRequisite bool
	SlackPostRelease  bool
	CheckSummary      bool
	CheckBackports    parentItem
	ReleaseBlocker    parentItem
}

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
`
)

func (pi parentItem) Done() bool {
	for _, item := range pi.Items {
		if !item.Done {
			return false
		}
	}
	return true
}

func (ctx *Context) LoadIssue() {
	body := github.GetIssueBody(ctx.VitessRepo, ctx.IssueNbGH)

	lines := strings.Split(body, "\n")

	var newIssue Issue

	s := stateReadingItem
	for _, line := range lines {
		switch s {
		case stateReadingItem:
			if strings.Contains(line, slackAnnouncementItem) {
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
				break
			}
			// remove indentation from line after we have confirmed it is present
			line = strings.TrimSpace(line)
			newIssue.CheckBackports.Items = append(newIssue.CheckBackports.Items, itemWithLink{
				Done: strings.HasPrefix(line, markdownItemDone),
				URL:  strings.TrimSpace(line[len(markdownItemDone):]),
			})
		case stateReadingReleaseBlockerIssue:
			if !strings.HasPrefix(line, "  -") {
				s = stateReadingItem
				break
			}
			// remove indentation from line after we have confirmed it is present
			line = strings.TrimSpace(line)
			newIssue.ReleaseBlocker.Items = append(newIssue.ReleaseBlocker.Items, itemWithLink{
				Done: strings.HasPrefix(line, markdownItemDone),
				URL:  strings.TrimSpace(line[len(markdownItemDone):]),
			})
		}
	}
	ctx.Issue = newIssue
}

func (ctx *Context) UploadIssue() (*logging.ProgressLogging, func()) {
	pl := &logging.ProgressLogging{
		TotalSteps: 2,
	}

	return pl, func() {
		pl.NewStepf("Update Issue #%d on GitHub", ctx.IssueNbGH)
		body := ctx.Issue.toString()
		issue := github.Issue{Body: body, Number: ctx.IssueNbGH}
		link := issue.UpdateBody(ctx.VitessRepo)
		pl.NewStepf("Issue updated: %s", link)
	}
}

func CreateReleaseIssue(ctx *Context) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 2,
	}

	return pl, func() string {
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
		pl.NewStepf("Issue created: %s", link)
		return link
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
