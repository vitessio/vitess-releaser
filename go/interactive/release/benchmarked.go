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

package release

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/interactive/ui"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

func BenchmarkedItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := benchmarkedAct
	if state.Issue.Benchmarked {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.Benchmarked,
		Act:    act,
		Update: benchmarkedUpdate,
		IsDone: state.Issue.Benchmarked,
	}
}

type benchmarkedUrl string

func benchmarkedUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	return mi, nil
}

func benchmarkedAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	return mi, nil
}