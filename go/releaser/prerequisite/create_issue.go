package prerequisite

import (
	"bytes"
	"fmt"
	"log"
	"text/template"

	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/vitess"
)

const (
	releaseIssueTemplate = `This release is scheduled for: TODO: '.Date' here .

### Prerequisites for Release

- [ ] Notify the community on Slack.
- [ ] Make sure the release notes summary is prepared and clean.
- [ ] Make sure backport Pull Requests are merged, list below.

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
`
)


func CreateReleaseIssue(majorRelease string) string {
	vitess.CorrectCleanRepo()

	newRelease := vitess.FindNextRelease(majorRelease)

	tmpl := template.Must(template.New("release-issue").Parse(releaseIssueTemplate))
	b := bytes.NewBuffer(nil)
	err := tmpl.Execute(b, nil)
	if err != nil {
		log.Fatal(err)
	}

	newIssue := github.Issue{
		Title:    fmt.Sprintf("Release of v%s", newRelease),
		Body:     b.String(),
		Labels:   []string{"Component: General", "Type: Release"},
		Assignee: "@me",
	}

	link := newIssue.Create()
	return link
}