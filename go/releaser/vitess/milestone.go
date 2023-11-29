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

package vitess

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"vitess.io/vitess-releaser/go/releaser"
)

func FindVersionAfterNextRelease(ctx *releaser.Context) string {
	CorrectCleanRepo(ctx.VitessRepo)
	nextRelease, _ := FindNextRelease(ctx.MajorRelease)

	if strings.Contains(nextRelease, "rc") {
		panic("RC releases not supported for now")
	}

	segments := strings.Split(nextRelease, ".")
	if len(segments) != 3 {
		return ""
	}

	segmentInts := make([]int, 0, len(segments))
	for _, segment := range segments {
		v, err := strconv.Atoi(segment)
		if err != nil {
			log.Fatal(err.Error())
		}
		segmentInts = append(segmentInts, v)
	}

	// if it is a major release
	if segmentInts[1] == 0 && segmentInts[2] == 0 {
		return fmt.Sprintf("%d.0.0", segmentInts[0]+1)
	}
	// if a patch release
	return fmt.Sprintf("%d.%d.%d", segmentInts[0], segmentInts[1], segmentInts[2]+1)
}