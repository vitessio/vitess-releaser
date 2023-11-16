package interactive

import (
	"github.com/spf13/cobra"
	m "vitess.io/vitess-releaser/go/cmd/model"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:     "interactive",
		Aliases: []string{"i"},
		Short:   "Runs the releaser in interactive mode",
		Run: func(cmd *cobra.Command, args []string) {
			m.CorrectCleanRepo()
			mainScreen()
		},
	}
}
