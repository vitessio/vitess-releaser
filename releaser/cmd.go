package releaser

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "vitess-releaser",
	Short: "vitess-releaser - a tool for releasing vitess",
	Long:  "vitess-releaser - a tool for releasing vitess",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var preCmd = &cobra.Command{
	Use:   "pre",
	Short: "Prepares a release",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(42)
	},
}

func init() {
	rootCmd.AddCommand(preCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
