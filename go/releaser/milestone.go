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

package releaser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vitessio/vitess-releaser/go/releaser/utils"
)

func FindVersionAfterNextRelease(state *State) string {
	segments := strings.Split(RemoveRCFromReleaseTitle(state.VitessRelease.Release), ".")
	if len(segments) != 3 {
		return ""
	}

	segmentInts := make([]int, 0, len(segments))
	for _, segment := range segments {
		v, err := strconv.Atoi(segment)
		if err != nil {
			utils.BailOut(err, "failed to convert release number segment to number (%s)", segment)
		}
		segmentInts = append(segmentInts, v)
	}

	// if it is an RC release
	if state.Issue.RC >= 1 && segmentInts[1] == 0 && segmentInts[2] == 0 {
		return fmt.Sprintf("%d.0.0", segmentInts[0]+1)
	}
	// if a patch or GA release
	return fmt.Sprintf("%d.%d.%d", segmentInts[0], segmentInts[1], segmentInts[2]+1)
}
