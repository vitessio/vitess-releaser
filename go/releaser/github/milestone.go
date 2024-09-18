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

package github

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"vitess.io/vitess-releaser/go/releaser/utils"
)

type Milestone struct {
	URL    string `json:"url"`
	Number int    `json:"number"`
}

func GetMilestonesByName(repo, name string) []Milestone {
	str := execGh(
		"milestone", "list",
		"--query", name,
		"--repo", repo,
		"--json", "url,number",
		"--state", "all",
	)

	str = str[strings.Index(str, "]")+1:]
	var ms []Milestone
	err := json.Unmarshal([]byte(str), &ms)
	if err != nil {
		utils.BailOut(err, "failed to parse milestone, got: %s", str)
	}
	return ms
}

func CreateNewMilestone(repo, name string) string {
	stdOut := execGh(
		"milestone", "create",
		"--repo", repo,
		"--title", name,
	)
	out := strings.ReplaceAll(stdOut, "\n", "")
	idx := strings.LastIndex(out, fmt.Sprintf("https://github.com/%s/milestone/", repo))
	return out[idx:]
}

func CloseMilestone(repo, name string) string {
	ms := GetMilestonesByName(repo, name)
	if len(ms) != 1 {
		utils.BailOut(nil, "expected to find one milestone found %d", len(ms))
	}

	stdOut := execGh(
		"milestone", "edit",
		strconv.Itoa(ms[0].Number),
		"--repo", repo,
		"--state", "closed",
	)
	out := strings.ReplaceAll(stdOut, "\n", "")
	idx := strings.LastIndex(out, fmt.Sprintf("https://github.com/%s/milestone/", repo))
	return out[idx:]
}
