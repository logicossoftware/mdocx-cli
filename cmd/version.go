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
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "mdocx %s\n", Version)
		fmt.Fprintf(out, "Commit: %s\n", GitCommit)
		fmt.Fprintf(out, "Built:  %s\n", BuildDate)
		fmt.Fprintln(out)
		fmt.Fprintln(out, "© 2026, Logicos Software")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
