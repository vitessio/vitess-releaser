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

	"vitess.io/vitess-releaser/go/releaser/utils"
)

func CurrentUser() string {
	exec := execGh("api", "user")
	x := map[string]any{}

	err := json.Unmarshal([]byte(exec), &x)
	if err != nil {
		utils.BailOut(err, "failed to parse the current user, got: %s", exec)
	}

	return x["login"].(string)
}
