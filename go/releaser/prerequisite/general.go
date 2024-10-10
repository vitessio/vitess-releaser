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

package prerequisite

func General() []string {
	return []string{
		"Please make sure you respect all the following items:",
		"\t- Be part of the vitessio \"Release\" team: https://github.com/orgs/vitessio/teams/release",
		"\t- Be an admin of the \"planetscale/vitess-operator\" repository",
		"\t- Be an admin of the \"vitess\" organization on DockerHub",
		"\t- Have access to Vitess' Java repository and have it working locally: https://github.com/vitessio/vitess/blob/main/doc/internal/release/java-packages.md",
		"\t- Have \"vitessio/vitess\" and \"planetscale/vitess-operator\" cloned in the same parent directory",
	}
}
