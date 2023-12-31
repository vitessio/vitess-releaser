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

package github

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/cli/go-gh"
)

type Milestone struct {
	URL    string `json:"url"`
	Number int    `json:"number"`
}

func GetMilestonesByName(repo, name string) []Milestone {
	stdOut, _, err := gh.Exec(
		"milestone", "list",
		"--query", name,
		"--repo", repo,
		"--json", "url,number",
		"--state", "all",
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	str := stdOut.String()
	str = str[strings.Index(str, "]")+1:]
	var ms []Milestone
	err = json.Unmarshal([]byte(str), &ms)
	if err != nil {
		log.Fatal(err.Error())
	}
	return ms
}

func CreateNewMilestone(repo, name string) string {
	stdOut, _, err := gh.Exec(
		"milestone", "create",
		"--repo", repo,
		"--title", name,
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	out := strings.ReplaceAll(stdOut.String(), "\n", "")
	idx := strings.LastIndex(out, fmt.Sprintf("https://github.com/%s/milestone/", repo))
	return out[idx:]
}

func CloseMilestone(repo, name string) string {
	ms := GetMilestonesByName(repo, name)
	if len(ms) != 1 {
		log.Fatalf("expected to find one milestone found %d", len(ms))
	}

	stdOut, _, err := gh.Exec(
		"milestone", "edit",
		strconv.Itoa(ms[0].Number),
		"--repo", repo,
		"--state", "closed",
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	out := strings.ReplaceAll(stdOut.String(), "\n", "")
	idx := strings.LastIndex(out, fmt.Sprintf("https://github.com/%s/milestone/", repo))
	return out[idx:]
}
