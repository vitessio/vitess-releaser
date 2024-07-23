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

import "fmt"

func CheckDockerMessage(majorRelease int, repo string, vtopRepo string) []string {
	msg := []string{
		"Make sure the Docker Images are being built by GitHub Actions.",
		"This can be done by visiting the following links, our new release should appear in either green (done) or yellow (building / pending build):",
		"",
	}

	// Hack: versions < v20 and versions >= v20 use different GitHub Actions workflows to build the Docker images.
	if majorRelease < 20 {
		msg = append(msg,
			fmt.Sprintf("\t- https://github.com/%s/actions/workflows/docker_build_vttestserver.yml", repo),
			fmt.Sprintf("\t- https://github.com/%s/actions/workflows/docker_build_base.yml", repo),
			fmt.Sprintf("\t- https://github.com/%s/actions/workflows/docker_build_lite.yml", repo),
		)
	} else if majorRelease == 20 {
		msg = append(msg, fmt.Sprintf("\t- https://github.com/%s/actions/workflows/docker_build_vttestserver.yml", repo))
		msg = append(msg, fmt.Sprintf("\t- https://github.com/%s/actions/workflows/docker_build_images.yml", repo))
	} else {
		// this links to the newer GitHub Actions workflow that was introduced in v21 by https://github.com/vitessio/vitess/pull/16339
		msg = append(msg, fmt.Sprintf("\t- https://github.com/%s/vitess/actions/workflows/build_docker_images.yml", repo))
	}

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
