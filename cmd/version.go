package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information - set via ldflags at build time
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print the version, build date, and commit hash of the mdocx CLI.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mdocx %s\n", Version)
		fmt.Printf("Commit: %s\n", GitCommit)
		fmt.Printf("Built:  %s\n", BuildDate)
		fmt.Println()
		fmt.Println("© 2026, Logicos Software")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
