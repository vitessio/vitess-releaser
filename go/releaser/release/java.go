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
	"os"
	"os/exec"
	"path"
	"strings"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/logging"
	"vitess.io/vitess-releaser/go/releaser/utils"
)

func JavaRelease(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 5,
	}

	return pl, func() string {
		lowerCaseRelease := "v" + strings.ToLower(state.VitessRelease.Release)

		pl.NewStepf("Fetch from git remote")
		git.CorrectCleanRepo(state.VitessRelease.Repo)
		git.Checkout(lowerCaseRelease)

		if !strings.Contains(state.VitessRelease.Repo, "vitessio/vitess") {
			pl.NewStepf("Do the Java release")
			cmd := exec.Command("/bin/sh", "-c", "eval $(gpg-agent --daemon --no-grab --write-env-file $HOME/.gpg-agent-info); export GPG_TTY=$(tty); export GPG_AGENT_INFO; export MAVEN_OPTS=\"--add-opens=java.base/java.util=ALL-UNNAMED --add-opens=java.base/java.lang.reflect=ALL-UNNAMED --add-opens=java.base/java.text=ALL-UNNAMED --add-opens=java.desktop/java.awt.font=ALL-UNNAMED\"; mvn clean deploy -P release -DskipTests;",
			)
			pwd, err := os.Getwd()
			if err != nil {
				utils.LogPanic(err, "failed to get current working directory")
			}
			cmd.Dir = path.Join(pwd, "/java")
			out, err := cmd.CombinedOutput()
			if err != nil {
				utils.LogPanic(err, "failed to execute: %s, got: %s", cmd.String(), string(out))
			}
		} else {
			pl.NewStepf("Running in non-live mode, skipping the actual Java release.")
		}
		pl.NewStepf("Done")

		state.Issue.JavaRelease = true
		pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
		_, fn := state.UploadIssue()
		issueLink := fn()

		pl.NewStepf("Issue updated, see: %s", issueLink)
		return ""
	}
}
