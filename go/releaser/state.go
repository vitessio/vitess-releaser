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
	"context"
	"path"
	"syscall"

	"vitess.io/vitess-releaser/go/releaser/utils"
)

const (
	pathVitess         = "vitess"
	pathVitessOperator = "vitess-operator"
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

type ReleaseInformation struct {
	Repo   string
	Remote string

	// BaseReleaseBranch is used to refer to the root release branch (i.e. "release-20.0") when doing
	// an RC release or a GA. In this situation the ReleaseBranch will be set to something like "release-20.0-rc".
	// For patch releases, the ReleaseBranch remains as usual i.e. "release-20.0" and BaseReleaseBranch will be empty.
	ReleaseBranch     string
	BaseReleaseBranch string

	MajorRelease    string
	Release         string
	IsLatestRelease bool
	GA              bool
}

type State struct {
	VitessRelease ReleaseInformation
	VtOpRelease   ReleaseInformation

	Issue     Issue
	IssueLink string
	IssueNbGH int

	currentPath string
}

func (s *State) GoToVitess() {
	p := pathVitess
	if s.currentPath != "" {
		p = "../" + p
	}
	s.currentPath = pathVitess
	changeDir(p)
}

func (s *State) GoToVtOp() {
	p := pathVitessOperator
	if s.currentPath != "" {
		p = "../" + p
	}
	s.currentPath = pathVitessOperator
	changeDir(p)
}

func changeDir(p string) {
	cwd, err := syscall.Getwd()
	if err != nil {
		utils.LogPanic(err, "failed to get current working directory")
	}
	p = path.Join(cwd, p)
	err = syscall.Chdir(p)
	if err != nil {
		utils.LogPanic(err, "failed to change directory to %s", p)
	}
}
