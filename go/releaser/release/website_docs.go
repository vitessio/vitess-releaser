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

package release

import (
	"vitess.io/vitess-releaser/go/releaser"
)

func WebsiteDocs(state *releaser.State) []string {
	msg := []string{
		"We want to open a Pull Request to update the documentation.",
		"There are several pages we want to update:",
		"\t- https://vitess.io/docs/releases/: we must add the new release to the list with all its information and link.",
		"\t- https://vitess.io/docs/get-started/local/: we must use the proper version increment for this guide and the proper SHA.",
		"We must do a git checkout to the proper release branch after cloning Vitess on the following pages:",
		"\t- https://vitess.io/docs/get-started/operator/#install-the-operator",
		"\t- https://vitess.io/docs/get-started/local-mac/#install-vitess",
		"\t- https://vitess.io/docs/get-started/local-docker/#check-out-the-vitessiovitess-repository",
		"\t- https://vitess.io/docs/get-started/vttestserver-docker-image/#check-out-the-vitessiovitess-repository",
	}
	if state.Issue.RC > 0 {
		msg = append(msg, []string{
			"",
			"Since we are doing an RC release, we must use the ./tools/rc_release.sh script in the website repository to update the documentation even further.",
			"The script creates a new entry in the sidebar which represents the next version on main and mark the version we are releasing as RC.",
		}...)
	}
	return msg
}
