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

package interactive

import (
	"github.com/spf13/cobra"
	"vitess.io/vitess-releaser/go/interactive"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/git"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:     "interactive",
		Aliases: []string{"i"},
		Short:   "Runs the releaser in interactive mode",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			state := releaser.UnwrapState(ctx)
			git.CorrectCleanRepo(state.VitessRepo)

			// TODO: The assumption that the Release Manager won't be
			// modifying the release issue while using vitess-releaser
			// is made here, perhaps there is a better way of doing it
			state.LoadIssue()

			interactive.MainScreen(ctx)
		},
	}
}
