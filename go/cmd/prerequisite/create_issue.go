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

package prerequisite

import (
	"bytes"
	"fmt"
	"log"
	"text/template"

	"github.com/spf13/cobra"
	"vitess.io/vitess-releaser/go/cmd/flags"
	"vitess.io/vitess-releaser/go/git"
	"vitess.io/vitess-releaser/go/github"
	"vitess.io/vitess-releaser/go/vitess"
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

// Create issue:
// - Make sure we are in the vitess repo
// - Make sure the git state is clean
// - Figure out the new release number
// - Create the issue for the corresponding release number
var createIssue = &cobra.Command{
	Use:   "create-issue",
	Short: "Create the release issue",
	Run: func(cmd *cobra.Command, args []string) {
		if !git.CheckCurrentRepo("vitessio/vitess.git") {
			log.Fatal("the tool should be run from the vitessio/vitess repository directory")
		}
		if !git.CleanLocalState() {
			log.Fatal("the vitess repository should have a clean state")
		}

		majorRelease := cmd.Flags().Lookup(flags.MajorRelease).Value.String()
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
		fmt.Println("Link to the new GitHub Issue: ", link)
	},
}
