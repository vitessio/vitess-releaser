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

package pre_release

import (
	"vitess.io/vitess-releaser/go/releaser"
)

func CheckSummary(ctx *releaser.Context) []string {
	return []string{
		"If the release does not contain significant changes (i.e. a small patch release) then this step can be skipped",
		"One or more Pull Requests have to be submitted in advance to create and update the release summary.",
		"The summary files are located in: ./changelog/*.0/*.*.*/summary.md.",
		"The summary file for a release candidate is the same as the one for the GA release.",
	}
}