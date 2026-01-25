/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/logicossoftware/go-mdocx"
	"github.com/spf13/cobra"
)

// browseCmd represents the browse command
var browseCmd = &cobra.Command{
	Use:   "browse <file>",
	Short: "Browse an .mdocx bundle in a TUI",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		strict, _ := cmd.Flags().GetBool("strict")
		noImages, _ := cmd.Flags().GetBool("no-images")
		theme, _ := cmd.Flags().GetString("theme")
		input := args[0]

		f, err := os.Open(input)
		if err != nil {
			return fmt.Errorf("open: %w", err)
		}
		defer f.Close()

		opts := []mdocx.ReadOption{}
		if !strict {
			opts = append(opts, mdocx.WithVerifyHashes(false))
		}
		doc, err := mdocx.Decode(f, opts...)
		if err != nil {
			return fmt.Errorf("decode: %w", err)
		}
		header, _ := readHeaderInfo(input)

		return runBrowseTUI(doc, header, theme, noImages)
	},
}

func init() {
	rootCmd.AddCommand(browseCmd)

	browseCmd.Flags().Bool("strict", true, "fail on any spec violation")
	browseCmd.Flags().Bool("no-images", false, "disable Sixel rendering")
	browseCmd.Flags().String("theme", "", "Glamour theme name or path")
}
