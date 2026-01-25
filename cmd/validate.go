/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/logicossoftware/go-mdocx"
	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate an .mdocx bundle",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOut, _ := cmd.Flags().GetBool("json")
		strict, _ := cmd.Flags().GetBool("strict")
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
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				_ = enc.Encode(validationResult{Valid: false, Error: err.Error()})
			}
			return fmt.Errorf("decode: %w", err)
		}

		result := buildValidationResult(doc)
		if jsonOut {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Valid MDOCX: markdown=%d media=%d\n", result.MarkdownFileCount, result.MediaItemCount)
		return nil
	},
}

type validationResult struct {
	Valid              bool   `json:"valid"`
	MarkdownFileCount  int    `json:"markdown_file_count"`
	MediaItemCount     int    `json:"media_item_count"`
	TotalMarkdownBytes int    `json:"total_markdown_bytes"`
	TotalMediaBytes    int    `json:"total_media_bytes"`
	Error              string `json:"error,omitempty"`
}

func buildValidationResult(doc *mdocx.Document) validationResult {
	result := validationResult{Valid: true}
	result.MarkdownFileCount = len(doc.Markdown.Files)
	result.MediaItemCount = len(doc.Media.Items)
	for _, mf := range doc.Markdown.Files {
		result.TotalMarkdownBytes += len(mf.Content)
	}
	for _, mi := range doc.Media.Items {
		result.TotalMediaBytes += len(mi.Data)
	}
	return result
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().Bool("json", false, "output machine-readable JSON")
	validateCmd.Flags().Bool("strict", true, "fail on any spec violation")
}
