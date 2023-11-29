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
	"github.com/spf13/cobra"
)

func Prerequisite() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "prerequisite",
		Aliases: []string{"pre"},
		Short:   "Runs the prerequisites of a release",
	}

	cmd.AddCommand(createIssue)
	cmd.AddCommand(checkPRs)
	cmd.AddCommand(checkReleaseSummary)
	cmd.AddCommand(addPendingPRsToIssue)
	cmd.AddCommand(addReleaseBlockerIssuesToIssue)
	cmd.AddCommand(slackAnnouncement)
	return cmd
}
