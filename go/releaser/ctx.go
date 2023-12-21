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

package releaser

import (
	"context"
)

var (
	skey = new(string)
)

func UnwrapState(ctx context.Context) *State {
	return ctx.Value(skey).(*State)
}

func WrapState(ctx context.Context, s *State) context.Context {
	return context.WithValue(ctx, skey, s)
}

type State struct {
	VitessRepo      string
	MajorRelease    string
	IssueNbGH       int
	IssueLink       string
	Issue           Issue
	ReleaseBranch   string
	IsLatestRelease bool
	Release         string
}
