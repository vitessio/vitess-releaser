package cmd

import (
	"fmt"
	"os"
	beforerelease "systay/vitess-releaser/go/cmd/before_release"

	"github.com/spf13/cobra"
)

var (
	releaseVersion string
)

var rootCmd = &cobra.Command{
	Use:   "vitess-releaser",
	Short: "vitess-releaser - a tool for releasing vitess",
	Long:  "vitess-releaser - a tool for releasing vitess",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&releaseVersion, "release", "r", "", "Number of the major release on which we want to create a new release.")
	err := rootCmd.MarkPersistentFlagRequired("release")
	if err != nil {
		panic(err)
	}

	rootCmd.AddCommand(beforerelease.BeforeRelease())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
