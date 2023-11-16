package beforerelease

import (
	"github.com/spf13/cobra"
)

func BeforeRelease() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "before-release",
		Short: "Prepares a release",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	cmd.AddCommand(createIssue)
	return cmd
}