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

package git

import (
	"log"
	"os/exec"
	"strings"
)

func CheckCurrentRepo(repoWanted string) bool {
	out, err := exec.Command("git", "remote", "-v").CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
	outStr := string(out)
	return strings.Contains(outStr, repoWanted)
}

func CleanLocalState() bool {
	out, err := exec.Command("git", "status", "-s").CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
	return len(out) == 0
}

func Checkout(branch string) {
	out, err := exec.Command("git", "checkout", branch).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
}
