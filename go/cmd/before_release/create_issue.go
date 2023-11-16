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

package beforerelease

import (
	"fmt"

	"github.com/spf13/cobra"
	"vitess.io/vitess-releaser/go/cmd/flags"
	"vitess.io/vitess-releaser/go/git"
	"vitess.io/vitess-releaser/go/github"
	"vitess.io/vitess-releaser/go/vitess"
)


// Create issue:
// - Make sure we are in the vitess repo
// - Make sure the git state is clean
// - Figure out the new release number
// - Create the issue for the corresponding release number
var createIssue = &cobra.Command{
	Use:   "create-issue",
	Short: "Create the release issue",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !git.CheckCurrentRepo("vitessio/vitess.git") {
			return fmt.Errorf("the tool should be run from the vitessio/vitess repository directory")
		}
		if !git.CleanLocalState() {
			return fmt.Errorf("the vitess repository should have a clean state")
		}

		majorRelease := cmd.Flags().Lookup(flags.MajorRelease).Value.String()
		newRelease := vitess.FindNextRelease(majorRelease)

		newIssue := github.Issue{
			Title:  fmt.Sprintf("Release of v%s", newRelease),
			Body:   "This is a test.",
			Labels: []string{"Component: General", "Type: Release"},

		}

		newIssue.CreateIssue()
		return nil
	},
}
