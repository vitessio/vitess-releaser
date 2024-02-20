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

package release

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/interactive/ui"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/release"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

func JavaReleaseItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	return &ui.MenuItem{
		State:  state,
		Name:   steps.JavaRelease,
		Act:    javaReleaseAct,
		Update: javaReleaseUpdate,
		IsDone: state.Issue.JavaRelease,
		Ignore: !state.VitessRelease.GA,
	}
}

type javaReleaseUrl string

func javaReleaseUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(javaReleaseUrl)
	if !ok {
		return mi, nil
	}

	mi.IsDone = mi.State.Issue.JavaRelease
	return mi, nil
}

func javaReleaseAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, j := release.JavaRelease(mi.State)
	return mi, tea.Batch(func() tea.Msg {
		return javaReleaseUrl(j())
	}, ui.PushDialog(ui.NewProgressDialog(steps.JavaRelease, pl)))
}
