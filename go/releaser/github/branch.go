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
	"fmt"
	"os"

	"github.com/vitessio/vitess-releaser/go/releaser/utils"
)

// Types to match what we receive from GitHub when fetching the branch protection rules.

type fetchBranchProtectionRules struct {
	AllowDeletions                 fetchEnabled              `json:"allow_deletions"`
	AllowForcePushes               fetchEnabled              `json:"allow_force_pushes"`
	AllowForkSyncing               fetchEnabled              `json:"allow_fork_syncing"`
	BlockCreations                 fetchEnabled              `json:"block_creations"`
	EnforceAdmins                  fetchEnabledURL           `json:"enforce_admins"`
	LockBranch                     fetchEnabled              `json:"lock_branch"`
	RequiredConversationResolution fetchEnabled              `json:"required_conversation_resolution"`
	RequiredLinearHistory          fetchEnabled              `json:"required_linear_history"`
	RequiredPullRequestReviews     fetchPullRequestReviews   `json:"required_pull_request_reviews"`
	RequiredSignatures             fetchEnabledURL           `json:"required_signatures"`
	RequiredStatusChecks           fetchRequiredStatusChecks `json:"required_status_checks"`
	Restrictions                   fetchRestrictions         `json:"restrictions"`
	Url                            string                    `json:"url"`
}

type fetchEnabled struct {
	Enabled bool `json:"enabled"`
}

type fetchEnabledURL struct {
	Enabled bool   `json:"enabled"`
	Url     string `json:"url"`
}

type fetchStatusCheck struct {
	AppId   int    `json:"app_id"`
	Context string `json:"context"`
}

type fetchRequiredStatusChecks struct {
	Checks      []fetchStatusCheck `json:"checks"`
	Contexts    []string           `json:"contexts"`
	ContextsUrl string             `json:"contexts_url"`
	Strict      bool               `json:"strict"`
	Url         string             `json:"url"`
}

type fetchRestrictions struct {
	Apps     []interface{} `json:"apps"`
	AppsUrl  string        `json:"apps_url"`
	Teams    []fetchTeam   `json:"teams"`
	TeamsUrl string        `json:"teams_url"`
	Url      string        `json:"url"`
	Users    []interface{} `json:"users"`
	UsersUrl string        `json:"users_url"`
}

type fetchTeam struct {
	Description         string      `json:"description"`
	HtmlUrl             string      `json:"html_url"`
	Id                  int         `json:"id"`
	MembersUrl          string      `json:"members_url"`
	Name                string      `json:"name"`
	NodeId              string      `json:"node_id"`
	NotificationSetting string      `json:"notification_setting"`
	Parent              interface{} `json:"parent"`
	Permission          string      `json:"permission"`
	Privacy             string      `json:"privacy"`
	RepositoriesUrl     string      `json:"repositories_url"`
	Slug                string      `json:"slug"`
	Url                 string      `json:"url"`
}

type fetchPullRequestReviews struct {
	DismissStaleReviews          bool   `json:"dismiss_stale_reviews"`
	RequireCodeOwnerReviews      bool   `json:"require_code_owner_reviews"`
	RequireLastPushApproval      bool   `json:"require_last_push_approval"`
	RequiredApprovingReviewCount int    `json:"required_approving_review_count"`
	Url                          string `json:"url"`
}

// Types declaration to match the JSON sent to GitHub to create a new branch protection rule

type updateUpdateBranchProtectionRulesPayload struct {
	RequiredStatusChecks           updateRequiredStatusChecks       `json:"required_status_checks"`
	EnforceAdmins                  bool                             `json:"enforce_admins"`
	RequiredPullRequestReviews     updateRequiredPullRequestReviews `json:"required_pull_request_reviews"`
	Restrictions                   updateUsersTeamsApps             `json:"restrictions"`
	RequiredLinearHistory          bool                             `json:"required_linear_history"`
	AllowForcePushes               bool                             `json:"allow_force_pushes"`
	AllowDeletions                 bool                             `json:"allow_deletions"`
	BlockCreations                 bool                             `json:"block_creations"`
	RequiredConversationResolution bool                             `json:"required_conversation_resolution"`
	LockBranch                     bool                             `json:"lock_branch"`
	AllowForkSyncing               bool                             `json:"allow_fork_syncing"`
}

type updateRequiredStatusChecks struct {
	Strict   bool     `json:"strict"`
	Contexts []string `json:"contexts"`
}

type updateUsersTeamsApps struct {
	Users []string `json:"users"`
	Teams []string `json:"teams"`
	Apps  []string `json:"apps"`
}

type updateRequiredPullRequestReviews struct {
	DismissStaleReviews          bool `json:"dismiss_stale_reviews"`
	RequireCodeOwnerReviews      bool `json:"require_code_owner_reviews"`
	RequiredApprovingReviewCount int  `json:"required_approving_review_count"`
	RequireLastPushApproval      bool `json:"require_last_push_approval"`
}

func CopyBranchProtectionRules(repo, branch string) {
	originBranchProtectionRules := getBranchProtectionRules(repo)
	destinationBranchProtectionRules := transformBranchProtectionRules(originBranchProtectionRules)
	putBranchProtectionRules(destinationBranchProtectionRules, repo, branch)
}

func getBranchProtectionRules(repo string) fetchBranchProtectionRules {
	stdOut := execGh("api", fmt.Sprintf("repos/%s/branches/main/protection", repo))

	var bpr fetchBranchProtectionRules

	err := json.Unmarshal([]byte(stdOut), &bpr)
	if err != nil {
		utils.BailOut(err, "failed to parse the branch protection rules")
	}

	return bpr
}

func transformBranchProtectionRules(bpr fetchBranchProtectionRules) updateUpdateBranchProtectionRulesPayload {
	ubpr := updateUpdateBranchProtectionRulesPayload{
		RequiredStatusChecks: updateRequiredStatusChecks{
			Strict:   bpr.RequiredStatusChecks.Strict,
			Contexts: bpr.RequiredStatusChecks.Contexts,
		},
		EnforceAdmins: bpr.EnforceAdmins.Enabled,
		RequiredPullRequestReviews: updateRequiredPullRequestReviews{
			DismissStaleReviews:          bpr.RequiredPullRequestReviews.DismissStaleReviews,
			RequireCodeOwnerReviews:      bpr.RequiredPullRequestReviews.RequireCodeOwnerReviews,
			RequiredApprovingReviewCount: bpr.RequiredPullRequestReviews.RequiredApprovingReviewCount,
			RequireLastPushApproval:      bpr.RequiredPullRequestReviews.RequireLastPushApproval,
		},
		RequiredLinearHistory:          bpr.RequiredLinearHistory.Enabled,
		AllowForcePushes:               bpr.AllowForcePushes.Enabled,
		AllowDeletions:                 bpr.AllowDeletions.Enabled,
		BlockCreations:                 bpr.BlockCreations.Enabled,
		RequiredConversationResolution: bpr.RequiredConversationResolution.Enabled,
		LockBranch:                     bpr.LockBranch.Enabled,
		AllowForkSyncing:               bpr.AllowForkSyncing.Enabled,
		Restrictions: updateUsersTeamsApps{
			Users: []string{},
			Teams: []string{},
			Apps:  []string{},
		},
	}

	// The API does not expect 'nil' instead of slices
	if ubpr.RequiredStatusChecks.Contexts == nil {
		ubpr.RequiredStatusChecks.Contexts = []string{}
	}

	for _, team := range bpr.Restrictions.Teams {
		ubpr.Restrictions.Teams = append(ubpr.Restrictions.Teams, team.Name)
	}

	return ubpr
}

func putBranchProtectionRules(ubpr updateUpdateBranchProtectionRulesPayload, repo, branch string) {
	jsonUbpr, err := json.Marshal(ubpr)
	if err != nil {
		utils.BailOut(err, "failed to marshal update branch protection rules")
	}

	f, err := os.CreateTemp("/tmp", "")
	if err != nil {
		utils.BailOut(err, "failed to create a temporary file")
	}

	_, err = f.Write(jsonUbpr)
	if err != nil {
		utils.BailOut(err, "failed to write update branch protection rules to the temporary file")
	}

	_ = execGh("api",
		"--method", "PUT",
		"-H", "Accept: application/vnd.github+json",
		"-H", "X-GitHub-Api-Version: 2022-11-28",
		"--input", f.Name(),
		fmt.Sprintf("repos/%s/branches/%s/protection", repo, branch),
	)
}
