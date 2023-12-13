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
	"time"

	"vitess.io/vitess-releaser/go/interactive/state"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

const (
	stateReadingItem = iota
	stateReadingBackport
	stateReadingReleaseBlockerIssue
	stateReadingCodeFreezeItem
	stateReadindCreateReleasePRItem
	stateReadingNewMilestoneItem
)

const (
	markdownItemDone = "- [x]"
	markdownItemToDo = "- [ ]"

	// Divers
	dateItem = "This release is scheduled for"

	// Prerequisites
	preSlackAnnouncementItem = "Notify the community on Slack."
	checkSummaryItem         = "Make sure the release notes summary is prepared and clean."
	backportItem             = "Make sure backport Pull Requests are merged, list below."
	releaseBlockerItem       = "Make sure release blocker Issues are closed, list below."

	// Pre-Release
	codeFreezeItem      = "Code Freeze."
	createReleasePRItem = "Create Release PR."
	newMilestoneItem    = "Create new GitHub Milestone."

	// Post-Release
	postSlackAnnouncementItem = "Notify the community on Slack for the new release."
)

type (
	ItemWithLink struct {
		Done bool

		// URL can use two formats:
		// 	- GH links:		"#111"
		//  - HTTP links:	"https://github.com...."
		URL string
	}

	ParentOfItems struct {
		Items []ItemWithLink
	}

	Issue struct {
		Date time.Time

		// Prerequisites
		SlackPreRequisite bool
		CheckSummary      bool
		CheckBackport     ParentOfItems
		ReleaseBlocker    ParentOfItems

		// Pre-Release
		CodeFreeze         ItemWithLink
		CreateReleasePR    ItemWithLink
		NewGitHubMilestone ItemWithLink

		// Post-Release
		SlackPostRelease bool
	}
)

const (
	releaseIssueTemplate = `This release is scheduled for {{fmtDate .Date }}

### Prerequisites for Release

- [{{fmtStatus .SlackPreRequisite}}] Notify the community on Slack.
- [{{fmtStatus .CheckSummary}}] Make sure the release notes summary is prepared and clean.
- [{{fmtStatus .CheckBackport.Done}}] Make sure backport Pull Requests are merged, list below.
{{- range $item := .CheckBackport.Items }}
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
- [{{fmtStatus .CreateReleasePR.Done}}] Create Release PR.
{{- if .CreateReleasePR.URL }}
  - {{ .CreateReleasePR.URL }}
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

func (ctx *State) LoadIssue() {
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
			if strings.HasPrefix(line, dateItem) {
				nline := strings.TrimSpace(line[len(dateItem):])
				parsedDate, err := time.Parse("Mon _2 Jan 2006", nline)
				if err != nil {
					log.Fatal(err)
				}
				newIssue.Date = parsedDate
			}
			if strings.Contains(line, preSlackAnnouncementItem) {
				newIssue.SlackPreRequisite = strings.HasPrefix(line, markdownItemDone)
			}
			if strings.Contains(line, checkSummaryItem) {
				newIssue.CheckSummary = strings.HasPrefix(line, markdownItemDone)
			}
			if strings.Contains(line, backportItem) && isNextLineAList(lines, i) {
				s = stateReadingBackport
			}
			if strings.Contains(line, releaseBlockerItem) && isNextLineAList(lines, i) {
				s = stateReadingReleaseBlockerIssue
			}
			if strings.Contains(line, codeFreezeItem) {
				newIssue.CodeFreeze.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					s = stateReadingCodeFreezeItem
				}
			}
			if strings.Contains(line, createReleasePRItem) {
				newIssue.CreateReleasePR.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					s = stateReadindCreateReleasePRItem
				}
			}
			if strings.Contains(line, newMilestoneItem) {
				newIssue.NewGitHubMilestone.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					s = stateReadingNewMilestoneItem
				}
			}
			if strings.Contains(line, postSlackAnnouncementItem) {
				newIssue.SlackPostRelease = strings.HasPrefix(line, markdownItemDone)
			}
		case stateReadingBackport:
			newIssue.CheckBackport.Items = append(newIssue.CheckBackport.Items, handleNewListItem(lines, i, &s))
		case stateReadingReleaseBlockerIssue:
			newIssue.ReleaseBlocker.Items = append(newIssue.ReleaseBlocker.Items, handleNewListItem(lines, i, &s))
		case stateReadingCodeFreezeItem:
			newIssue.CodeFreeze.URL = handleSingleTextItem(line, &s)
		case stateReadindCreateReleasePRItem:
			newIssue.CreateReleasePR.URL = handleSingleTextItem(line, &s)
		case stateReadingNewMilestoneItem:
			newIssue.NewGitHubMilestone.URL = handleSingleTextItem(line, &s)
		}
	}
	ctx.Issue = newIssue
}

func handleNewListItem(lines []string, i int, s *int) ItemWithLink {
	line := strings.TrimSpace(lines[i])
	newItem := ItemWithLink{
		Done: strings.HasPrefix(line, markdownItemDone),
		URL:  strings.TrimSpace(line[len(markdownItemDone):]),
	}
	if i+1 == len(lines) || !strings.HasPrefix(lines[i+1], "  -") {
		*s = stateReadingItem
	}
	return newItem
}

func handleSingleTextItem(line string, s *int) string {
	line = strings.TrimSpace(line)
	if line[0] == '-' {
		line = strings.TrimSpace(line[1:])
	}
	*s = stateReadingItem
	return line
}

func isNextLineAList(lines []string, i int) bool {
	return len(lines) > i+1 && strings.HasPrefix(lines[i+1], "  -")
}

func (ctx *State) UploadIssue() (*logging.ProgressLogging, func() string) {
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

func CreateReleaseIssue(state *State) (*logging.ProgressLogging, func() (int, string)) {
	pl := &logging.ProgressLogging{
		TotalSteps: 2,
	}

	return pl, func() (int, string) {
		CorrectCleanRepo(state.VitessRepo)
		newRelease, _ := FindNextRelease(state.MajorRelease)

		pl.NewStepf("Create Release Issue on GitHub")
		newIssue := github.Issue{
			Title:    fmt.Sprintf("Release of v%s", newRelease),
			Body:     state.Issue.toString(),
			Labels:   []github.Label{{Name: "Component: General"}, {Name: "Type: Release"}},
			Assignee: "@me",
		}

		link := newIssue.Create(state.VitessRepo)
		nb := github.URLToNb(link)
		state.IssueLink = link
		state.IssueNbGH = nb
		pl.NewStepf("Issue created: %s", link)
		return nb, link
	}
}

func (i *Issue) toString() string {
	tmpl := template.New("release-issue")
	tmpl = tmpl.Funcs(template.FuncMap{
		"fmtStatus": state.FmtMd,
		"fmtDate": func(d time.Time) string {
			return d.Format("Mon _2 Jan 2006")
		},
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
