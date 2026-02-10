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
	"fmt"
	"strings"

	"github.com/vitessio/vitess-releaser/go/releaser"
)

func CheckDockerMessage(state *releaser.State) []string {
	repo := state.VitessRelease.Repo
	vtopRepo := state.VtOpRelease.Repo
	release := strings.ToLower(state.VitessRelease.Release)
	msg := []string{
		"Make sure the Docker Images are being built by GitHub Actions.",
		"This can be done by visiting the following links, our new release should appear in either green (done) or yellow (building / pending build):",
		"",
	}

	msg = append(msg, fmt.Sprintf("\t- https://github.com/%s/actions/workflows/docker_build_vttestserver.yml", repo))
	msg = append(msg, fmt.Sprintf("\t- https://github.com/%s/actions/workflows/docker_build_images.yml", repo))
	msg = append(msg, fmt.Sprintf("\nCheck that the vttestserver image is pushed at https://hub.docker.com/r/vitess/vttestserver/tags?name=%s.", release))

	if vtopRepo != "" {
		msg = append(msg, []string{
			"",
			"",
			"Please also make sure that the vitess-operator image has been built by DockerHub.",
			"",
			fmt.Sprintf("\t- https://hub.docker.com/repository/docker/%s/builds", vtopRepo),
			"",
		}...)
	}

	return msg
}
