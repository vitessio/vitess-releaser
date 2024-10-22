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

package interactive

import (
	"context"
	"fmt"
	"github.com/vitessio/vitess-releaser/go/interactive/ui"
	"github.com/vitessio/vitess-releaser/go/releaser"
	"github.com/vitessio/vitess-releaser/go/releaser/code_freeze"
	"github.com/vitessio/vitess-releaser/go/releaser/post_release"
	"github.com/vitessio/vitess-releaser/go/releaser/prerequisite"
	"github.com/vitessio/vitess-releaser/go/releaser/release"
	"github.com/vitessio/vitess-releaser/go/releaser/steps"
	"github.com/vitessio/vitess-releaser/go/releaser/utils"
	"reflect"
)

func checkSummaryMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	return newBooleanMenu(
		ctx,
		prerequisite.CheckSummary(state),
		steps.CheckSummary,
		func() { state.Issue.CheckSummary = !state.Issue.CheckSummary },
		state.Issue.CheckSummary,
		false)
}

func draftBlogPostMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	return newBooleanMenu(
		ctx,
		releaser.DraftBlogPost(),
		steps.DraftBlogPost,
		func() { state.Issue.DraftBlogPost = !state.Issue.DraftBlogPost },
		state.Issue.DraftBlogPost,
		!state.Issue.GA)
}

func requestCrossPostBlogPostMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	return newBooleanMenu(
		ctx,
		releaser.RequestCrossPostBlogPost(),
		steps.CrossPostBlogPost,
		func() { state.Issue.RequestCrossPostBlogPost = !state.Issue.RequestCrossPostBlogPost },
		state.Issue.RequestCrossPostBlogPost,
		!state.Issue.GA)
}

func vtopUpdateCompatibilityTableMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	return newBooleanMenu(
		ctx,
		code_freeze.VtopUpdateCompatibilityTable(state),
		steps.VtopUpdateCompatibilityTable,
		func() { state.Issue.VtopUpdateCompatibilityTable = !state.Issue.VtopUpdateCompatibilityTable },
		state.Issue.VtopUpdateCompatibilityTable,
		state.VtOpRelease.Release == "" || state.Issue.RC != 1)
}

func websiteDocumentationItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	return newBooleanMenu(
		ctx,
		release.WebsiteDocs(state),
		steps.WebsiteDocumentation,
		func() { state.Issue.WebsiteDocumentation = !state.Issue.WebsiteDocumentation },
		state.Issue.WebsiteDocumentation,
		false)
}

func benchmarkedItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	return newBooleanMenu(
		ctx,
		release.BenchmarkedMessage(),
		steps.Benchmarked,
		func() { state.Issue.Benchmarked = !state.Issue.Benchmarked },
		state.Issue.Benchmarked,
		false)
}

func dockerImagesItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	return newBooleanMenu(
		ctx,
		release.CheckDockerMessage(state),
		steps.DockerImages,
		func() { state.Issue.DockerImages = !state.Issue.DockerImages },
		state.Issue.DockerImages,
		false)
}

func twitterMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	return newBooleanMenu(
		ctx,
		post_release.TwitterAnnouncement(),
		steps.Twitter,
		func() { state.Issue.Twitter = !state.Issue.Twitter },
		state.Issue.Twitter,
		false)
}

func createBlogPostPRMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	return newBooleanMenu(
		ctx,
		releaser.CreateBlogPostPR(),
		steps.DraftBlogPost,
		func() { state.Issue.CreateBlogPostPR = !state.Issue.CreateBlogPostPR },
		state.Issue.CreateBlogPostPR,
		!state.Issue.GA)
}

func mergeBlogPostPRMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	return newBooleanMenu(
		ctx,
		releaser.MergeBlogPostPR(),
		steps.MergeBlogPost,
		func() { state.Issue.MergeBlogPostPR = !state.Issue.MergeBlogPostPR },
		state.Issue.MergeBlogPostPR,
		!state.Issue.GA)
}

func simpleMenuItem(ctx context.Context, issueFieldName string, msgs []string, stepName string, onlyGA bool) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	logMsg := fmt.Sprintf("Menu item %s", stepName)

	fieldVal := getFieldVal(&state.Issue, issueFieldName, logMsg)

	ignore := false
	if onlyGA {
		ignore = !state.Issue.GA
	}

	return newBooleanMenu(
		ctx,
		msgs,
		stepName,
		func() {
			fieldVal.SetBool(!fieldVal.Bool())
		},
		fieldVal.Bool(),
		ignore,
	)
}

func getFieldVal(issue *releaser.Issue, issueFieldName string, logMsg string) reflect.Value {
	v := reflect.ValueOf(issue).Elem()
	fieldVal := v.FieldByName(issueFieldName)
	if !fieldVal.IsValid() {
		utils.BailOut(fmt.Errorf("no such field: %s", issueFieldName), logMsg)
	}
	if fieldVal.Kind() != reflect.Bool {
		utils.BailOut(fmt.Errorf("field %s is not of type bool", issueFieldName), logMsg)
	}

	if !fieldVal.CanSet() {
		utils.BailOut(fmt.Errorf("cannot set field: %s", issueFieldName), logMsg)
	}
	return fieldVal
}
