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

import (
	"fmt"

	"vitess.io/vitess-releaser/go/releaser"
)

func CheckSummary(state *releaser.State) []string {
	return []string{
		"It is common for small releases (i.e. patch releases) to not have a summary file, in this case this step can be marked as done and ignored.",
		"",
		"We must verify that the summary file makes sense: no grammar or english mistake, that the content is well organized and that the links are not broken.",
		"We must treat this file as if it was a blog post, this is one of the main entrypoints for people wanting to learn more about the release.",
		"",
		"Over the release cycle people add their bits to this file, but no one looks at the file as a whole, this is the time to do it!",
		"",
		fmt.Sprintf("The summary file is located in: './changelog/%s.0/%s/summary.md'.", state.VitessRelease.MajorRelease, releaser.RemoveRCFromReleaseTitle(state.VitessRelease.Release)),
		"",
		"Note: the summary file for a release candidate is the same as the one for the GA release.",
	}
}
