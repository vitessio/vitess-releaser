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

package release

import "fmt"

func CheckDockerMessage(repo string) []string {
	return []string{
		"Make sure the Docker Images are being built by GitHub Actions.",
		"This can be done by visiting the following links, our new release should appear in either green (done) or yellow (building / pending build):",
		"",
		fmt.Sprintf("\t- https://github.com/%s/actions/workflows/docker_build_base.yml", repo),
		fmt.Sprintf("\t- https://github.com/%s/actions/workflows/docker_build_lite.yml", repo),
		fmt.Sprintf("\t- https://github.com/%s/actions/workflows/docker_build_vttestserver.yml", repo),
	}
}
