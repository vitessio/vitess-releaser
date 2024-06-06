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

package releaser

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	"vitess.io/vitess-releaser/go/interactive/state"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
	"vitess.io/vitess-releaser/go/releaser/utils"
)

const (
	stateReadingItem = iota
	stateReadingGeneral
	stateReadingBackport
	stateReadingReleaseBlockerIssue
	stateReadingCodeFreezeItem
	stateReadingCreateBackportLabelItem
	stateReadingUpdateSnapshotOnMainItem
	stateReadingCreateReleasePRItem
	stateReadingNewMilestoneItem
	stateReadingMergedReleasePRItem
	stateReadingTagReleaseItem
	stateReadingReleaseNotesMainItem
	stateReadingReleaseNotesReleaseBranchItem
	stateReadingBackToDevModeItem
	stateReadingBackToDevModeBaseBranchItem
	stateReadingCloseMilestoneItem
	stateReadingVtopUpdateGo
	stateReadingVtopCreateReleasePR
	stateReadingVtopBumpVersionOnMainPR
)

const (
	markdownItemDone = "- [x]"

	// Divers
	dateItem = "> This release is scheduled for"

	// Prerequisites
	generalPrerequisitesItem = "General prerequisites."
	preSlackAnnouncementItem = "Notify the community on Slack."
	checkSummaryItem         = "Make sure the release notes summary is prepared and clean."
	backportItem             = "Make sure important Pull Requests are merged, list below."
	releaseBlockerItem       = "Make sure release blocker Issues are closed, list below."
	draftBlogPostItem        = "Draft the release blog post."
	crossBlogPostItem        = "Send requests to cross-post the blog post (CNCF, PlanetScale)."

	// Pre-Release
	codeFreezeItem                = "Code Freeze."
	copyBranchProtectionRulesItem = "Copy branch protection rules."
	createBackportToLabelItem     = "Create the Backport to labels."
	updateSnapshotOnMainItem      = "Update the SNAPSHOT version on main."
	createReleasePRItem           = "Create Release PR."
	newMilestoneItem              = "Create new GitHub Milestone."
	vtopCreateBranchItem          = "Create vitess-operator release branch."
	vtopBumpVersionOnMain         = "Bump the version vitess-operator main."
	vtopUpdateGoItem              = "Update vitess-operator Golang version."
	vtopUpdateCompTableItem       = "Update vitess-operator compatibility table."
	createBlogPostPRItem          = "Open a Pull Request on the website repository for the blog post."

	// Release
	mergeReleasePRItem            = "Merge the Release PR."
	tagReleaseItem                = "Tag the release."
	javaRelease                   = "Java release."
	vtopCreateReleasePRItem       = "Create vitess-operator Release PR."
	vtopManualUpdateItem          = "Manual update of vitess-operator test code."
	releaseNotesMainItem          = "Update release notes on main."
	releaseNotesReleaseBranchItem = "Update release notes on the release branch."
	backToDevItem                 = "Go back to dev mode on the release branch."
	backToDevBaseBranchItem       = "Go back to dev mode on the base of the release branch."
	websiteDocItem                = "Update the website documentation."
	benchmarkedItem               = "Make sure the release is benchmarked by arewefastyet."
	dockerImagesItem              = "Docker Images available on DockerHub."
	closeMilestoneItem            = "Close current GitHub Milestone."
	mergeBlogPostItem             = "Merge the blog post Pull Request on the website repository."

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

	ItemWithLinks struct {
		Done bool
		URLs []string
	}

	ParentOfItems struct {
		Items []ItemWithLink
	}

	Issue struct {
		Date        time.Time
		RC          int
		DoVtOp      bool
		VtopRelease string
		GA          bool

		// Prerequisites
		General                  ParentOfItems
		SlackPreRequisite        bool
		CheckSummary             bool
		DraftBlogPost            bool
		RequestCrossPostBlogPost bool
		CheckBackport            ParentOfItems
		ReleaseBlocker           ParentOfItems

		// Pre-Release
		CodeFreeze                   ItemWithLink
		CopyBranchProtectionRules    bool
		CreateBackportToLabel        ItemWithLink
		UpdateSnapshotOnMain         ItemWithLink
		CreateReleasePR              ItemWithLink
		NewGitHubMilestone           ItemWithLink
		VtopCreateBranch             bool
		VtopBumpMainVersion          ItemWithLink
		VtopUpdateGolang             ItemWithLink
		VtopUpdateCompatibilityTable bool
		VtopCreateReleasePR          ItemWithLinks
		VtopManualUpdate             bool
		CreateBlogPostPR             bool

		// Release
		MergeReleasePR              ItemWithLink
		TagRelease                  ItemWithLink
		JavaRelease                 bool
		ReleaseNotesOnMain          ItemWithLink
		ReleaseNotesOnReleaseBranch ItemWithLink
		BackToDevMode               ItemWithLink
		BackToDevModeBaseBranch     ItemWithLink
		MergeBlogPostPR             bool
		WebsiteDocumentation        bool
		Benchmarked                 bool
		DockerImages                bool
		CloseMilestone              ItemWithLink

		// Post-Release
		SlackPostRelease bool
		Twitter          bool
		CloseIssue       bool
	}
)

const (
	releaseIssueTemplate = `> [!NOTE]  
> This release is scheduled for {{fmtDate .Date }}.
{{- if .DoVtOp }}
> The release of vitess-operator **v{{.VtopRelease}}** is also planned.
{{- end }}
> Release team: @vitessio/release

> [!IMPORTANT]  
> Please **do not** edit the content of the Issue's body manually.
> The **vitess-releaser** tool is managing and handling this issue.
> You can however click on the check boxes to mark them as done/not done, and write comments.

### Prerequisites _(~2 weeks before)_

- [{{fmtStatus .General.Done}}] General prerequisites.
{{- range $item := .General.Items }}
  - [{{fmtStatus $item.Done}}] {{$item.URL}}
{{- end }}
- [{{fmtStatus .SlackPreRequisite}}] Notify the community on Slack.
- [{{fmtStatus .CheckSummary}}] Make sure the release notes summary is prepared and clean.
- Make sure important Pull Requests are merged, list below.
{{- range $item := .CheckBackport.Items }}
  - [{{fmtStatus $item.Done}}] {{$item.URL}}
{{- end }}
- Make sure release blocker Issues are closed, list below.
{{- range $item := .ReleaseBlocker.Items }}
  - [{{fmtStatus $item.Done}}] {{$item.URL}}
{{- end }}
{{- if .GA }}
- [{{fmtStatus .DraftBlogPost}}] Draft the release blog post.
- [{{fmtStatus .RequestCrossPostBlogPost}}] Send requests to cross-post the blog post (CNCF, PlanetScale).
{{- end }}

{{- if not (or (gt .RC 1) (.GA))}}
### Code Freeze {{if eq .RC 1}}_(1 week before)_{{else}}_(~1-3 days before)_{{end}}
- [{{fmtStatus .CodeFreeze.Done}}] Code Freeze.
{{- if .CodeFreeze.URL }}
  - {{ .CodeFreeze.URL }}
{{- end }}
{{- if eq .RC 1 }}
- [{{fmtStatus .CopyBranchProtectionRules}}] Copy branch protection rules.
- [{{fmtStatus .CreateBackportToLabel.Done}}] Create the Backport to labels.
{{- if .CreateBackportToLabel.URL }}
  - {{ .CreateBackportToLabel.URL }}
{{- end }}
- [{{fmtStatus .UpdateSnapshotOnMain.Done}}] Update the SNAPSHOT version on main.
{{- if .UpdateSnapshotOnMain.URL }}
  - {{ .UpdateSnapshotOnMain.URL }}
{{- end }}
{{- end }}
- [{{fmtStatus .NewGitHubMilestone.Done}}] Create new GitHub Milestone.
{{- if .NewGitHubMilestone.URL }}
  - {{ .NewGitHubMilestone.URL }}
{{- end }}
{{- end }}

### Pre-Release _(~1-3 days before)_

- [{{fmtStatus .CreateReleasePR.Done}}] Create Release PR.
{{- if .CreateReleasePR.URL }}
  - {{ .CreateReleasePR.URL }}
{{- end }}
{{- if .DoVtOp }}
{{- if eq .RC 1 }}
- [{{fmtStatus .VtopCreateBranch}}] Create vitess-operator release branch.
- [{{fmtStatus .VtopBumpMainVersion.Done}}] Bump the version vitess-operator main.
{{- if .VtopBumpMainVersion.URL }}
  - {{ .VtopBumpMainVersion.URL }}
{{- end }}
{{- end }}
- [{{fmtStatus .VtopUpdateGolang.Done}}] Update vitess-operator Golang version.
{{- if .VtopUpdateGolang.URL }}
  - {{ .VtopUpdateGolang.URL }}
{{- end }}
{{- if eq .RC 1 }}
- [{{fmtStatus .VtopUpdateCompatibilityTable}}] Update vitess-operator compatibility table.
{{- end }}
{{- end }}
{{- if .GA }}
- [{{fmtStatus .CreateBlogPostPR}}] Open a Pull Request on the website repository for the blog post.
{{- end }}

### Release _({{fmtShortDate .Date }})_

- [{{fmtStatus .MergeReleasePR.Done}}] Merge the Release PR.
{{- if .MergeReleasePR.URL }}
  - {{ .MergeReleasePR.URL }}
{{- end }}
- [{{fmtStatus .TagRelease.Done}}] Tag the release.
{{- if .TagRelease.URL }}
  - {{ .TagRelease.URL }}
{{- end }}
{{- if .GA }}
- [{{fmtStatus .JavaRelease}}] Java release.
{{- end }}
{{- if .DoVtOp }}
- [{{fmtStatus .VtopCreateReleasePR.Done}}] Create vitess-operator Release PR.
{{- range $item := .VtopCreateReleasePR.URLs }}
  - {{$item}}
{{- end }}
- [{{fmtStatus .VtopManualUpdate}}] Manual update of vitess-operator test code.
{{- end }}
- [{{fmtStatus .ReleaseNotesOnMain.Done}}] Update release notes on main.
{{- if .ReleaseNotesOnMain.URL }}
  - {{ .ReleaseNotesOnMain.URL }}
{{- end }}
{{- if or (gt .RC 0) (.GA) }}
- [{{fmtStatus .ReleaseNotesOnReleaseBranch.Done}}] Update release notes on the release branch.
{{- if .ReleaseNotesOnReleaseBranch.URL }}
  - {{ .ReleaseNotesOnReleaseBranch.URL }}
{{- end }}
{{- end }}
- [{{fmtStatus .BackToDevMode.Done}}] Go back to dev mode on the release branch.
{{- if .BackToDevMode.URL }}
  - {{ .BackToDevMode.URL }}
{{- end }}
{{- if .GA }}
- [{{fmtStatus .BackToDevModeBaseBranch.Done}}] Go back to dev mode on the base of the release branch.
{{- if .BackToDevModeBaseBranch.URL }}
  - {{ .BackToDevModeBaseBranch.URL }}
{{- end }}
- [{{fmtStatus .MergeBlogPostPR}}] Merge the blog post Pull Request on the website repository.
{{- end }}
- [{{fmtStatus .WebsiteDocumentation}}] Update the website documentation.
- [{{fmtStatus .Benchmarked}}] Make sure the release is benchmarked by arewefastyet.
- [{{fmtStatus .DockerImages}}] Docker Images available on DockerHub.
{{- if eq .RC 0 }}
- [{{fmtStatus .CloseMilestone.Done}}] Close current GitHub Milestone.
{{- if .CloseMilestone.URL }}
  - {{ .CloseMilestone.URL }}
{{- end }}
{{- end }}


### Post-Release _({{fmtShortDate .Date }})_
- [{{fmtStatus .SlackPostRelease}}] Notify the community on Slack for the new release.
- [{{fmtStatus .Twitter}}] Twitter announcement.
- [{{fmtStatus .CloseIssue}}] Close this Issue.
`
)

func (pi *ParentOfItems) ItemsLeft() int {
	nb := 0
	for _, item := range pi.Items {
		if !item.Done {
			nb++
		}
	}
	return nb
}

func (pi *ParentOfItems) MarkAllAsDone() {
	for i, _ := range pi.Items {
		pi.Items[i].Done = true
	}
}

func (pi *ParentOfItems) MarkAllAsNotDone() {
	for i, _ := range pi.Items {
		pi.Items[i].Done = false
	}
}

func (pi *ParentOfItems) Done() bool {
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

	// Parse the title of the Issue to determine the RC increment if any
	title = strings.ReplaceAll(title, "`", "")
	if idx := strings.Index(title, "-RC"); idx != -1 {
		rc, err := strconv.Atoi(title[idx+len("-RC"):])
		if err != nil {
			utils.LogPanic(err, "failed to parse the RC number from the release issue title (%s)", title)
		}
		newIssue.RC = rc
	}

	newIssue.GA = s.VitessRelease.GA
	newIssue.DoVtOp = s.VtOpRelease.Release != ""
	newIssue.VtopRelease = AddRCToReleaseTitle(s.VtOpRelease.Release, newIssue.RC)

	st := stateReadingItem
	for i, line := range lines {
		switch st {
		case stateReadingItem:
			// divers
			if strings.HasPrefix(line, dateItem) {
				nline := strings.TrimSpace(line[len(dateItem) : len(line)-2])
				parsedDate, err := time.Parse("Mon _2 Jan 2006", nline)
				if err != nil {
					utils.LogPanic(err, "failed to parse the date from the release issue body (%s)", nline)
				}
				newIssue.Date = parsedDate
			}

			switch {
			case strings.Contains(line, generalPrerequisitesItem) && isNextLineAList(lines, i):
				st = stateReadingGeneral
			case strings.Contains(line, draftBlogPostItem):
				newIssue.DraftBlogPost = strings.HasPrefix(line, markdownItemDone)
			case strings.Contains(line, crossBlogPostItem):
				newIssue.RequestCrossPostBlogPost = strings.HasPrefix(line, markdownItemDone)
			case strings.Contains(line, preSlackAnnouncementItem):
				newIssue.SlackPreRequisite = strings.HasPrefix(line, markdownItemDone)
			case strings.Contains(line, checkSummaryItem):
				newIssue.CheckSummary = strings.HasPrefix(line, markdownItemDone)
			case strings.Contains(line, backportItem) && isNextLineAList(lines, i):
				st = stateReadingBackport
			case strings.Contains(line, releaseBlockerItem) && isNextLineAList(lines, i):
				st = stateReadingReleaseBlockerIssue
			case strings.Contains(line, codeFreezeItem):
				newIssue.CodeFreeze.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingCodeFreezeItem
				}
			case strings.Contains(line, copyBranchProtectionRulesItem):
				newIssue.CopyBranchProtectionRules = strings.HasPrefix(line, markdownItemDone)
			case strings.Contains(line, createBackportToLabelItem):
				newIssue.CreateBackportToLabel.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingCreateBackportLabelItem
				}
			case strings.Contains(line, updateSnapshotOnMainItem):
				newIssue.UpdateSnapshotOnMain.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingUpdateSnapshotOnMainItem
				}
			case strings.Contains(line, createReleasePRItem):
				newIssue.CreateReleasePR.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingCreateReleasePRItem
				}
			case strings.Contains(line, newMilestoneItem):
				newIssue.NewGitHubMilestone.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingNewMilestoneItem
				}
			case strings.Contains(line, vtopCreateBranchItem):
				newIssue.VtopCreateBranch = strings.HasPrefix(line, markdownItemDone)
			case strings.Contains(line, vtopBumpVersionOnMain):
				newIssue.VtopBumpMainVersion.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingVtopBumpVersionOnMainPR
				}
			case strings.Contains(line, vtopUpdateGoItem):
				newIssue.VtopUpdateGolang.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingVtopUpdateGo
				}
			case strings.Contains(line, vtopUpdateCompTableItem):
				newIssue.VtopUpdateCompatibilityTable = strings.HasPrefix(line, markdownItemDone)
			case strings.Contains(line, vtopCreateReleasePRItem):
				newIssue.VtopCreateReleasePR.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingVtopCreateReleasePR
				}
			case strings.Contains(line, vtopManualUpdateItem):
				newIssue.VtopManualUpdate = strings.HasPrefix(line, markdownItemDone)
			case strings.Contains(line, mergeReleasePRItem):
				newIssue.MergeReleasePR.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingMergedReleasePRItem
				}
			case strings.Contains(line, tagReleaseItem):
				newIssue.TagRelease.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingTagReleaseItem
				}
			case strings.Contains(line, releaseNotesMainItem):
				newIssue.ReleaseNotesOnMain.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingReleaseNotesMainItem
				}
			case strings.Contains(line, releaseNotesReleaseBranchItem):
				newIssue.ReleaseNotesOnReleaseBranch.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingReleaseNotesReleaseBranchItem
				}
			case strings.Contains(line, backToDevItem):
				newIssue.BackToDevMode.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingBackToDevModeItem
				}
			case strings.Contains(line, backToDevBaseBranchItem):
				newIssue.BackToDevModeBaseBranch.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingBackToDevModeBaseBranchItem
				}
			case strings.Contains(line, websiteDocItem):
				newIssue.WebsiteDocumentation = strings.HasPrefix(line, markdownItemDone)
			case strings.Contains(line, benchmarkedItem):
				newIssue.Benchmarked = strings.HasPrefix(line, markdownItemDone)
			case strings.Contains(line, dockerImagesItem):
				newIssue.DockerImages = strings.HasPrefix(line, markdownItemDone)
			case strings.Contains(line, closeMilestoneItem):
				newIssue.CloseMilestone.Done = strings.HasPrefix(line, markdownItemDone)
				if isNextLineAList(lines, i) {
					st = stateReadingCloseMilestoneItem
				}
			case strings.Contains(line, postSlackAnnouncementItem):
				newIssue.SlackPostRelease = strings.HasPrefix(line, markdownItemDone)
			case strings.Contains(line, twitterItem):
				newIssue.Twitter = strings.HasPrefix(line, markdownItemDone)
			case strings.Contains(line, closeReleaseItem):
				newIssue.CloseIssue = strings.HasPrefix(line, markdownItemDone)
			case strings.Contains(line, createBlogPostPRItem):
				newIssue.CreateBlogPostPR = strings.HasPrefix(line, markdownItemDone)
			case strings.Contains(line, mergeBlogPostItem):
				newIssue.MergeBlogPostPR = strings.HasPrefix(line, markdownItemDone)
			case strings.Contains(line, javaRelease):
				newIssue.JavaRelease = strings.HasPrefix(line, markdownItemDone)
			}
		case stateReadingGeneral:
			newIssue.General.Items = append(newIssue.General.Items, handleNewListItem(lines, i, &st))
		case stateReadingBackport:
			newIssue.CheckBackport.Items = append(newIssue.CheckBackport.Items, handleNewListItem(lines, i, &st))
		case stateReadingReleaseBlockerIssue:
			newIssue.ReleaseBlocker.Items = append(newIssue.ReleaseBlocker.Items, handleNewListItem(lines, i, &st))
		case stateReadingCodeFreezeItem:
			newIssue.CodeFreeze.URL = handleSingleTextItem(line, &st)
		case stateReadingCreateBackportLabelItem:
			newIssue.CreateBackportToLabel.URL = handleSingleTextItem(line, &st)
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
		case stateReadingReleaseNotesReleaseBranchItem:
			newIssue.ReleaseNotesOnReleaseBranch.URL = handleSingleTextItem(line, &st)
		case stateReadingBackToDevModeItem:
			newIssue.BackToDevMode.URL = handleSingleTextItem(line, &st)
		case stateReadingBackToDevModeBaseBranchItem:
			newIssue.BackToDevModeBaseBranch.URL = handleSingleTextItem(line, &st)
		case stateReadingCloseMilestoneItem:
			newIssue.CloseMilestone.URL = handleSingleTextItem(line, &st)
		case stateReadingVtopUpdateGo:
			newIssue.VtopUpdateGolang.URL = handleSingleTextItem(line, &st)
		case stateReadingVtopCreateReleasePR:
			newLine := handleMultipleTextItem(lines, i, &st)
			if newLine != "" {
				newIssue.VtopCreateReleasePR.URLs = append(newIssue.VtopCreateReleasePR.URLs, newLine)
			}
		case stateReadingVtopBumpVersionOnMainPR:
			newIssue.VtopBumpMainVersion.URL = handleSingleTextItem(line, &st)
		}
	}
	s.Issue = newIssue
}

func handleMultipleTextItem(lines []string, i int, s *int) string {
	line := strings.TrimSpace(lines[i])
	if line[0] == '-' {
		line = strings.TrimSpace(line[1:])
	}
	if i+1 == len(lines) || !strings.HasPrefix(lines[i+1], "  -") {
		*s = stateReadingItem
	}
	return line
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
		state.Issue.General.Items = append(state.Issue.General.Items,
			ItemWithLink{URL: "Be part of the `Release` team in the `vitessio` GitHub organization, [here](https://github.com/orgs/vitessio/teams/release)."},
			ItemWithLink{URL: "Be an admin of the `planetscale/vitess-operator` repository."},
			ItemWithLink{URL: "Have access to Vitess' Java repository and have it working locally, [guide here](https://github.com/vitessio/vitess/blob/main/doc/internal/release/java-packages.md)."},
			ItemWithLink{URL: "Have `vitessio/vitess` and `planetscale/vitess-operator` cloned in the same parent directory."},
		)

		pl.NewStepf("Create Release Issue on GitHub")
		issueTitle := fmt.Sprintf("Release of `v%s`", state.VitessRelease.Release)
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
		"fmtShortDate": func(d time.Time) string {
			return d.Format("Mon _2 Jan")
		},
	})

	parsed, err := tmpl.Parse(releaseIssueTemplate)
	if err != nil {
		utils.LogPanic(err, "failed to parse the release issue template")
	}
	b := bytes.NewBufferString("")
	err = parsed.Execute(b, i)
	if err != nil {
		utils.LogPanic(err, "failed to execute/write the release issue template")
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
	index := strings.Index(release, "-RC")
	if index < 0 {
		return release
	}
	return release[:index]
}

func AddRCToReleaseTitle(release string, rc int) string {
	if rc == 0 {
		return release
	}
	return fmt.Sprintf("%s-RC%d", release, rc)
}
