package wizard

import (
	"github.com/spf13/cobra"
	"log"
	"vitess.io/vitess-releaser/go/releaser"
	wz "vitess.io/vitess-releaser/go/releaser/wizard"
)

// Wizard allows for a next-next-next style of releasing
var Wizard = &cobra.Command{
	Use:     "wizard",
	Aliases: []string{"next"},
	Short:   "Takes the next step in the release process",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := releaser.UnwrapCtx(cmd.Context())
		err := wz.Gogo(ctx)
		if err != nil {
			log.Fatal(err)
		}
	},
}
