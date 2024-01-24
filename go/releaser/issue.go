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
	"strconv"
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
	stateReadingUpdateSnapshotOnMainItem
	stateReadingCreateReleasePRItem
	stateReadingNewMilestoneItem
	stateReadingMergedReleasePRItem
	stateReadingTagReleaseItem
	stateReadingReleaseNotesMainItem
	stateReadingBackToDevModeItem
	stateReadingWebsiteDocsItem
	stateReadingCloseMilestoneItem
	stateReadingVtopUpdateGo
	stateReadingVtopCreateReleasePR
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
	codeFreezeItem                = "Code Freeze."
	copyBranchProtectionRulesItem = "Copy branch protection rules."
	updateSnapshotOnMainItem      = "Update the SNAPSHOT version on main."
	createReleasePRItem           = "Create Release PR."
	newMilestoneItem              = "Create new GitHub Milestone."
	vtopCreateBranchItem          = "Create vitess-operator release branch."
	vtopUpdateGoItem              = "Update vitess-operator Golang version."
	vtopUpdateCompTableItem       = "Update vitess-operator compatibility table."
	vtopCreateReleasePRItem       = "Create vitess-operator Release PR."

	// Release
	mergeReleasePRItem   = "Merge the Release PR."
	tagReleaseItem       = "Tag the release."
	releaseNotesMainItem = "Update release notes on main."
	backToDevItem        = "Go back to dev mode on the release branch."
	websiteDocItem       = "Update the website documentation."
	benchmarkedItem      = "Make sure the release is benchmarked by arewefastyet."
	dockerImagesItem     = "Docker Images available on DockerHub."
	closeMilestoneItem   = "Close current GitHub Milestone."

	// Post-Release
	postSlackAnnouncementItem = "Notify the community on Slack for the new release."
	twitterItem               = "Twitter announcement."
	closeReleaseItem          = "Close this Issue."
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
		Date   time.Time
		RC     int
		DoVtOp bool

		// Prerequisites
		SlackPreRequisite bool
		CheckSummary      bool
		CheckBackport     ParentOfItems
		ReleaseBlocker    ParentOfItems

		// Pre-Release
		CodeFreeze                   ItemWithLink
		CopyBranchProtectionRules    bool
		UpdateSnapshotOnMain         ItemWithLink
		CreateReleasePR              ItemWithLink
		NewGitHubMilestone           ItemWithLink
		VtopCreateBranch             bool
		VtopUpdateGolang             ItemWithLink
		VtopUpdateCompatibilityTable bool
		VtopCreateReleasePR          ItemWithLink

		// Release
		MergeReleasePR       ItemWithLink
		TagRelease           ItemWithLink
		ReleaseNotesOnMain   ItemWithLink
		BackToDevMode        ItemWithLink
		WebsiteDocumentation ItemWithLink
		Benchmarked          bool
		DockerImages         bool
		CloseMilestone       ItemWithLink

		// Post-Release
		SlackPostRelease bool
		Twitter          bool
		CloseIssue       bool
	}
)

const (
	releaseIssueTemplate = `This release is scheduled for {{fmtDate .Date }}

### Prerequisites for Release

- [{{fmtStatus .SlackPreRequisite}}] Notify the community on Slack.
- [{{fmtStatus .CheckSummary}}] Make sure the release notes summary is prepared and clean.
{{- if eq .RC 0 }}
- Make sure backport Pull Requests are merged, list below.
{{- else }}
- Make sure important Pull Requests are merged, list below.
{{- end }}
{{- range $item := .CheckBackport.Items }}
  - [{{fmtStatus $item.Done}}] {{$item.URL}}
{{- end }}
- Make sure release blocker Issues are closed, list below.
{{- range $item := .ReleaseBlocker.Items }}
  - [{{fmtStatus $item.Done}}] {{$item.URL}}
{{- end }}


### Pre-Release

- [{{fmtStatus .CodeFreeze.Done}}] Code Freeze.
{{- if .CodeFreeze.URL }}
  - {{ .CodeFreeze.URL }}
{{- end }}
{{- if eq .RC 1 }}
- [{{fmtStatus .CopyBranchProtectionRules}}] Copy branch protection rules.
- [{{fmtStatus .UpdateSnapshotOnMain.Done}}] Update the SNAPSHOT version on main.
{{- if .UpdateSnapshotOnMain.URL }}
  - {{ .UpdateSnapshotOnMain.URL }}
{{- end }}
{{- end }}
- [{{fmtStatus .CreateReleasePR.Done}}] Create Release PR.
{{- if .CreateReleasePR.URL }}
  - {{ .CreateReleasePR.URL }}
{{- end }}
{{- if lt .RC 2 }}
- [{{fmtStatus .NewGitHubMilestone.Done}}] Create new GitHub Milestone.
{{- if .NewGitHubMilestone.URL }}
  - {{ .NewGitHubMilestone.URL }}
{{- end }}
{{- end }}
{{- if .DoVtOp }}
{{- if eq .RC 1 }}
- [{{fmtStatus .VtopCreateBranch}}] Create vitess-operator release branch.
{{- end }}
- [{{fmtStatus .VtopUpdateGolang.Done}}] Update vitess-operator Golang version.
{{- if .VtopUpdateGolang.URL }}
  - {{ .VtopUpdateGolang.URL }}
{{- end }}
{{- if eq .RC 1 }}
- [{{fmtStatus .VtopUpdateCompatibilityTable}}] Update vitess-operator compatibility table.
{{- end }}
- [{{fmtStatus .VtopCreateReleasePR.Done}}] Create vitess-operator Release PR.
{{- if .VtopCreateReleasePR.URL }}
  - {{ .VtopCreateReleasePR.URL }}
{{- end }}
{{- end }}

### Release

- [{{fmtStatus .MergeReleasePR.Done}}] Merge the Release PR.
{{- if .MergeReleasePR.URL }}
  - {{ .MergeReleasePR.URL }}
{{- end }}
- [{{fmtStatus .TagRelease.Done}}] Tag the release.
{{- if .TagRelease.URL }}
  - {{ .TagRelease.URL }}
{{- end }}
- [{{fmtStatus .ReleaseNotesOnMain.Done}}] Update release notes on main.
{{- if .ReleaseNotesOnMain.URL }}
  - {{ .ReleaseNotesOnMain.URL }}
{{- end }}
- [{{fmtStatus .BackToDevMode.Done}}] Go back to dev mode on the release branch.
{{- if .BackToDevMode.URL }}
  - {{ .BackToDevMode.URL }}
{{- end }}
- [{{fmtStatus .WebsiteDocumentation.Done}}] Update the website documentation.
{{- if .WebsiteDocumentation.URL }}
  - {{ .WebsiteDocumentation.URL }}
{{- end }}
- [{{fmtStatus .Benchmarked}}] Make sure the release is benchmarked by arewefastyet.
- [{{fmtStatus .DockerImages}}] Docker Images available on DockerHub.
{{- if eq .RC 0 }}
- [{{fmtStatus .CloseMilestone.Done}}] Close current GitHub Milestone.
{{- if .CloseMilestone.URL }}
  - {{ .CloseMilestone.URL }}
{{- end }}
{{- end }}


### Post-Release
- [{{fmtStatus .SlackPostRelease}}] Notify the community on Slack for the new release.
- [{{fmtStatus .Twitter}}] Twitter announcement.
- [{{fmtStatus .CloseIssue}}] Close this Issue.
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

func (s *State) LoadIssue() {
	if s.IssueNbGH == 0 {
		// we are in the case where we start vitess-releaser
		// and the Release Issue hasn't been created yet.
		// We simply quit, the issue is left empty, nothing to load.
		return
	}

	title, body := github.GetIssueTitleAndBody(s.VitessRelease.Repo, s.IssueNbGH)

	lines := strings.Split(body, "\n")

	var newIssue Issue

	newIssue.DoVtOp = s.VtOpRelease.Release != ""

	// Parse the title of the Issue to determine the RC increment if any
	if idx := strings.Index(title, "-RC"); idx != -1 {
		rc, err := strconv.Atoi(title[idx+len("-RC"):])
		if err != nil {
			log.Fatal(err)
		}
		newIssue.RC = rc
	}

	st := stateReadingItem
	for i, line := range lines {
		switch st {
		case stateReadingItem:
			// divers
			if strings.HasPrefix(line, dateItem) {
				nline := strings.TrimSpace(line[len(dateItem):])
				parsedDate, err := time.Parse("Mon _2 Jan 2006", nline)
				if err != nil {
					log.Fatal(err)
				}
				newIssue.Date = parsedDate
			}

			// pre-release
			if strings.Contains(line, preSlackAnnouncementItem) {
				newIssue.SlackPreRequisite = strings.HasPrefix(line, markdownItemDone)
			}
			if strings.Contains(line, checkSummaryItem) {
				newIssue.CheckSummary = strings.HasPrefix(line, markdownItemDone)
			}
			if strings.Contains(line, backportItem) && isNextLineAList(lines, i) {
				st = stateReadingBackport
			}
			if strings.Contains(line, releaseBlockerItem) && isNextLineAList(lines, i) {
				st = stateReadingReleaseBlockerIssue
			}
			if strings.Contains(line, codeFreezeItem) {
				newIssue.CodeFreeze.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingCodeFreezeItem
				}
			}
			if strings.Contains(line, copyBranchProtectionRulesItem) {
				newIssue.CopyBranchProtectionRules = strings.HasPrefix(line, markdownItemDone)
			}
			if strings.Contains(line, updateSnapshotOnMainItem) {
				newIssue.UpdateSnapshotOnMain.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingUpdateSnapshotOnMainItem
				}
			}
			if strings.Contains(line, createReleasePRItem) {
				newIssue.CreateReleasePR.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingCreateReleasePRItem
				}
			}
			if strings.Contains(line, newMilestoneItem) {
				newIssue.NewGitHubMilestone.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingNewMilestoneItem
				}
			}
			if strings.Contains(line, vtopCreateBranchItem) {
				newIssue.VtopCreateBranch = strings.HasPrefix(line, markdownItemDone)
			}
			if strings.Contains(line, vtopUpdateGoItem) {
				newIssue.VtopUpdateGolang.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingVtopUpdateGo
				}
			}
			if strings.Contains(line, vtopUpdateCompTableItem) {
				newIssue.VtopUpdateCompatibilityTable = strings.HasPrefix(line, markdownItemDone)
			}
			if strings.Contains(line, vtopCreateReleasePRItem) {
				newIssue.VtopCreateReleasePR.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingVtopCreateReleasePR
				}
			}

			// release
			if strings.Contains(line, mergeReleasePRItem) {
				newIssue.MergeReleasePR.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingMergedReleasePRItem
				}
			}
			if strings.Contains(line, tagReleaseItem) {
				newIssue.TagRelease.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingTagReleaseItem
				}
			}
			if strings.Contains(line, releaseNotesMainItem) {
				newIssue.ReleaseNotesOnMain.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingReleaseNotesMainItem
				}
			}
			if strings.Contains(line, backToDevItem) {
				newIssue.BackToDevMode.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingBackToDevModeItem
				}
			}
			if strings.Contains(line, websiteDocItem) {
				newIssue.WebsiteDocumentation.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingWebsiteDocsItem
				}
			}
			if strings.Contains(line, benchmarkedItem) {
				newIssue.Benchmarked = strings.HasPrefix(line, markdownItemDone)
			}
			if strings.Contains(line, dockerImagesItem) {
				newIssue.DockerImages = strings.HasPrefix(line, markdownItemDone)
			}
			if strings.Contains(line, closeMilestoneItem) {
				newIssue.CloseMilestone.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingCloseMilestoneItem
				}
			}

			// post-release
			if strings.Contains(line, postSlackAnnouncementItem) {
				newIssue.SlackPostRelease = strings.HasPrefix(line, markdownItemDone)
			}
			if strings.Contains(line, twitterItem) {
				newIssue.Twitter = strings.HasPrefix(line, markdownItemDone)
			}
			if strings.Contains(line, closeReleaseItem) {
				newIssue.CloseIssue = strings.HasPrefix(line, markdownItemDone)
			}
		case stateReadingBackport:
			newIssue.CheckBackport.Items = append(newIssue.CheckBackport.Items, handleNewListItem(lines, i, &st))
		case stateReadingReleaseBlockerIssue:
			newIssue.ReleaseBlocker.Items = append(newIssue.ReleaseBlocker.Items, handleNewListItem(lines, i, &st))
		case stateReadingCodeFreezeItem:
			newIssue.CodeFreeze.URL = handleSingleTextItem(line, &st)
		case stateReadingUpdateSnapshotOnMainItem:
			newIssue.UpdateSnapshotOnMain.URL = handleSingleTextItem(line, &st)
		case stateReadingCreateReleasePRItem:
			newIssue.CreateReleasePR.URL = handleSingleTextItem(line, &st)
		case stateReadingNewMilestoneItem:
			newIssue.NewGitHubMilestone.URL = handleSingleTextItem(line, &st)
		case stateReadingMergedReleasePRItem:
			newIssue.MergeReleasePR.URL = handleSingleTextItem(line, &st)
		case stateReadingTagReleaseItem:
			newIssue.TagRelease.URL = handleSingleTextItem(line, &st)
		case stateReadingReleaseNotesMainItem:
			newIssue.ReleaseNotesOnMain.URL = handleSingleTextItem(line, &st)
		case stateReadingBackToDevModeItem:
			newIssue.BackToDevMode.URL = handleSingleTextItem(line, &st)
		case stateReadingWebsiteDocsItem:
			newIssue.WebsiteDocumentation.URL = handleSingleTextItem(line, &st)
		case stateReadingCloseMilestoneItem:
			newIssue.CloseMilestone.URL = handleSingleTextItem(line, &st)
		case stateReadingVtopUpdateGo:
			newIssue.VtopUpdateGolang.URL = handleSingleTextItem(line, &st)
		case stateReadingVtopCreateReleasePR:
			newIssue.VtopCreateReleasePR.URL = handleSingleTextItem(line, &st)
		}
	}
	s.Issue = newIssue
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

func (s *State) UploadIssue() (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 2,
	}

	return pl, func() string {
		pl.NewStepf("Update Issue #%d on GitHub", s.IssueNbGH)
		body := s.Issue.toString()
		issue := github.Issue{Body: body, Number: s.IssueNbGH}
		link := issue.UpdateBody(s.VitessRelease.Repo)
		pl.NewStepf("Issue updated: %s", link)
		return link
	}
}

func CreateReleaseIssue(state *State) (*logging.ProgressLogging, func() (int, string)) {
	pl := &logging.ProgressLogging{
		TotalSteps: 2,
	}

	return pl, func() (int, string) {
		pl.NewStepf("Create Release Issue on GitHub")
		issueTitle := fmt.Sprintf("Release of v%s", state.VitessRelease.Release)
		if state.Issue.RC > 0 {
			issueTitle = fmt.Sprintf("%s-RC%d", issueTitle, state.Issue.RC)
		}
		newIssue := github.Issue{
			Title:    issueTitle,
			Body:     state.Issue.toString(),
			Labels:   []github.Label{{Name: "Component: General"}, {Name: "Type: Release"}},
			Assignee: "@me",
		}

		link := newIssue.Create(state.VitessRelease.Repo)
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

func CloseReleaseIssue(state *State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 4,
	}

	return pl, func() string {
		pl.NewStepf("Closing Release Issue")
		github.CloseReleaseIssue(state.VitessRelease.Repo, state.IssueNbGH)
		state.Issue.CloseIssue = true
		pl.NewStepf("Issue closed: %s", state.IssueLink)

		pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
		_, fn := state.UploadIssue()
		issueLink := fn()
		pl.NewStepf("Issue updated, see: %s", issueLink)
		return state.IssueLink
	}
}

func RemoveRCFromReleaseTitle(release string) string {
	return release[:strings.Index(release, "-RC")]
}
