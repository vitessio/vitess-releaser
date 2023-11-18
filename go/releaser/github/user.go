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
	"encoding/json"
	"log"

	gh "github.com/cli/go-gh"
)

func CurrentUser() string {
	exec, _, err := gh.Exec("api", "user")
	if err != nil {
		log.Fatalf(err.Error())
	}
	x := map[string]any{}

	err = json.Unmarshal(exec.Bytes(), &x)
	if err != nil {
		log.Fatalf(err.Error())
	}

	return x["login"].(string)
}
