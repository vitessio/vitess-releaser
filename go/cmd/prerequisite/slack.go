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
	"fmt"

	"github.com/spf13/cobra"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/slack"
)

var slackAnnouncement = &cobra.Command{
	Use:   "slack",
	Short: "Prompts the Slack announcement for the prerequisite step",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := releaser.UnwrapCtx(cmd.Context())
		msg := slack.AnnouncementMessage(ctx)
		fmt.Print("Please post this message in the #general and #releases channels:\n\n")
		fmt.Println("\t"+msg)
	},
}