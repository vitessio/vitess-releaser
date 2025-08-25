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

package pre_release

import (
	"fmt"

	"github.com/vitessio/vitess-releaser/go/releaser"
)

func CobraDocs(state *releaser.State) []string {
	return []string{
		"Regenerate cobra cli docs by running the following in the root of the website repo:\n",
		fmt.Sprintf("\t$> export COBRADOC_VERSION_PAIRS=\"v%s:%s.0\"",
			releaser.RemoveRCFromReleaseTitle(state.VitessRelease.Release), state.VitessRelease.MajorRelease),
		"\t$> make generated-docs",
		"",
	}
}
