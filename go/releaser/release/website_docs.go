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
limitations under the License.®
*/

package release

import (
	"github.com/vitessio/vitess-releaser/go/releaser"
)

func WebsiteDocs(state *releaser.State) []string {
	msg := []string{
		"We want to open a Pull Request to update the documentation.",
		"",
		"There are several pages we want to update:",
		"At the beginning of the following pages, we ask the user to clone Vitess. Please make sure we are doing a 'git checkout' to the proper branch after the 'git clone'.",
		"For RC >= 2 and patch releases it's possible that no change is required if nothing was skipped in the previous releases.",
		"List of pages where we must do a 'git checkout':",
		"\t- https://vitess.io/docs/get-started/operator/#install-the-operator",
		"\t- https://vitess.io/docs/get-started/local-mac/#install-vitess",
		"\t- https://vitess.io/docs/get-started/vttestserver-docker-image/#check-out-the-vitessiovitess-repository",
	}

	if state.Issue.RC == 1 {
		msg = append(msg, []string{
			"",
			"Since we are doing an RC release, we must use the ./tools/rc_release.sh script in the website repository to update the documentation even further.",
			"The script creates a new entry in the sidebar which represents the next version on main and mark the version we are releasing as RC.",
		}...)
	}

	if state.Issue.GA {
		msg = append(msg, []string{
			"",
			"Since we are doing a GA release, we must use the ./tools/ga_release.sh script in the website repository to update the documentation even further.",
			"The script will update the RC version as Stable.",
		}...)
	}

	return msg
}
