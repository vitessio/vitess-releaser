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

package steps

const (
	CreateReleaseIssue = "Create Release Issue"
	ReleaseIssue       = "Release Issue"

	// Prerequisite
	SlackAnnouncement = "Slack Announcement"
	CheckAndAdd       = "Pending PRs/Issues"
	CheckSummary      = "Check Release Summary"

	// Pre-Release
	CodeFreeze                   = "Code Freeze"
	CopyBranchProtectionRules    = "Copy branch protection rules"
	UpdateSnapshotOnMain         = "Update SNAPSHOT on main"
	CreateReleasePR              = "Create Release PR"
	CreateMilestone              = "Create Milestone"
	VtopCreateBranch             = "Create vitess-operator release branch"
	VtopUpdateGolang             = "Update Go version in vitess-operator"
	VtopUpdateCompatibilityTable = "Update compatibility table in vitess-operator"

	// Release
	MergeReleasePR       = "Merge Release PR"
	TagRelease           = "Tag Release"
	VtopCreateReleasePR  = "Create vitess-operator release PR"
	JavaRelease          = "Java Release"
	ReleaseNotesOnMain   = "Release Notes on Main"
	BackToDev            = "Back To Dev Mode"
	WebsiteDocumentation = "Website Documentation"
	Benchmarked          = "Benchmarks"
	DockerImages         = "Docker Images"
	CloseMilestone       = "Close Milestone"

	// Post-Release
	SlackAnnouncementPost = "Slack Announcement Post-Release"
	Twitter               = "Twitter"
	CloseIssue            = "Close Issue"
)
