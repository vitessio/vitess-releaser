package github

import "testing"

// TestCopyBranchProtectionRules is a placeholder for copying branch protection rules to new branches
// for any internal forks you may maintain.
func TestCopyPrivateBranchProtectionRules(t *testing.T) {
	const (
		FORK_OWNER       = "<YOUR_ORG>"
		FORK_REPO        = "<YOUR_REPO>"
		FORK_FROM_BRANCH = "<YOUR_MAIN>"
		FORK_TO_BRANCH   = "<NEW_VERSION_BRANCH>"
	)
	CopyBranchProtectionRules(FORK_OWNER, FORK_REPO, FORK_FROM_BRANCH, FORK_TO_BRANCH)
}
