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

package github

import (
	"log"
	"strings"

	gh "github.com/cli/go-gh"

	"vitess.io/vitess-releaser/go/releaser/state"
)

type PR struct {
	Title  string
	Body   string
	Branch string
	Base   string
	Labels []string
}

func (p *PR) Create() string {
	stdOut, _, err := gh.Exec(
		"pr", "create",
		"--repo", state.VitessRepo,
		"--title", p.Title,
		"--body", p.Body,
		"--label", strings.Join(p.Labels, ","),
		"--head", p.Branch,
		"--base", p.Base,
	)
	if err != nil {
		log.Fatal(err)
	}
	return strings.ReplaceAll(stdOut.String(), "\n", "")
}
