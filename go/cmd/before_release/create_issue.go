package beforerelease

import (
	"fmt"
	"systay/vitess-releaser/go/git"

	"github.com/spf13/cobra"
)

var createIssue = &cobra.Command{
	Use:   "create-issue",
	Short: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !git.CheckCurrentRepo("vitessio/vitess.git") {
			return fmt.Errorf("The tool should be run from the vitessio/vitess repository directory")
		}
		if !git.CleanLocalState() {
			return fmt.Errorf("The vitess repository should have a clean state")
		}
		return nil
	},
}
