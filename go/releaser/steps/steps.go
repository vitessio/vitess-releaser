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

package steps

const (
	CreateReleaseIssue = "Create Release Issue"
	ReleaseIssue       = "Release Issue"

	// Prerequisite.
	GeneralPrerequisite          = "General"
	SlackAnnouncement            = "Slack Announcement"
	CheckAndAdd                  = "Pending PRs/Issues"
	CheckSummary                 = "Check Release Summary"
	DraftBlogPost                = "Draft Blog Post"
	VtopCreateBranch             = "Create vitess-operator release branch"
	VtopBumpMainVersion          = "Bump version of vitess-operator on main"
	VtopUpdateCompatibilityTable = "Update compatibility table in vitess-operator"
	CrossPostBlogPost            = "Cross-post Blog Post"

	// Pre-Release.
	CodeFreeze                = "Code Freeze"
	CopyBranchProtectionRules = "Copy branch protection rules"
	CreateNewLabels           = "Create new labels"
	UpdateSnapshotOnMain      = "Update SNAPSHOT on main"
	CreateReleasePR           = "Create Release PR"
	CreateMilestone           = "Create Milestone"
	VtopUpdateGolang          = "Update Go version in vitess-operator"
	UpdateCobraDocs           = "Update Cobra Docs"

	// Release.
	MergeReleasePR              = "Merge Release PR"
	TagRelease                  = "Tag Release"
	VtopCreateReleasePR         = "Create vitess-operator Release PR"
	JavaRelease                 = "Java Release"
	ReleaseNotesOnMain          = "Release Notes on main"
	ReleaseNotesOnReleaseBranch = "Release Notes on release branch"
	BackToDev                   = "Back To Dev Mode"
	BackToDevOnBaseBranch       = "Back To Dev Mode on the base branch"
	MergeBlogPost               = "Merge Blog Post"
	WebsiteDocumentation        = "Website Documentation"
	Benchmarked                 = "Benchmarks"
	DockerImages                = "Docker Images"
	CloseMilestone              = "Close Milestone"
	ReleaseArtifacts            = "Release Artifacts"
	VtopMergeReleasePR          = "Merge the vitess-operator Release PR"
	VtopTagRelease              = "Tag the vitess-operator release"
	VtopBackToDev               = "Back To Dev Mode vitess-operator"
	VtopManualUpdate            = "Manual update of vitess-operator tests"

	// Post-Release.
	SlackAnnouncementPost  = "Slack Announcement Post-Release"
	Twitter                = "Twitter"
	CloseIssue             = "Close Issue"
	RemoveBypassProtection = "Remove Bypass Protection"
)
