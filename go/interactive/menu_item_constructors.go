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

	"vitess.io/vitess-releaser/go/interactive/ui"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/post_release"
	"vitess.io/vitess-releaser/go/releaser/pre_release"
	"vitess.io/vitess-releaser/go/releaser/prerequisite"
	"vitess.io/vitess-releaser/go/releaser/release"
	"vitess.io/vitess-releaser/go/releaser/steps"
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
		pre_release.VtopUpdateCompatibilityTable(state),
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
		release.CheckDockerMessage(state.VitessRelease.MajorReleaseNb, state.VitessRelease.Repo),
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
