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

		// Validate header first.
		header, headerErr := readHeaderInfo(input)

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

		result := buildValidationResult(doc, header, headerErr)
		if jsonOut {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		if !result.Valid {
			for _, w := range result.Warnings {
				fmt.Fprintf(cmd.OutOrStdout(), "WARNING: %s\n", w)
			}
			return fmt.Errorf("validation failed: markdown=%d media=%d", result.MarkdownFileCount, result.MediaItemCount)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Valid MDOCX: markdown=%d (%s) media=%d (%s)\n",
			result.MarkdownFileCount, humanSize(result.TotalMarkdownBytes),
			result.MediaItemCount, humanSize(result.TotalMediaBytes))
		return nil
	},
}

type validationResult struct {
	Valid              bool     `json:"valid"`
	MarkdownFileCount  int      `json:"markdown_file_count"`
	MediaItemCount     int      `json:"media_item_count"`
	TotalMarkdownBytes int      `json:"total_markdown_bytes"`
	TotalMediaBytes    int      `json:"total_media_bytes"`
	Warnings           []string `json:"warnings,omitempty"`
	Error              string   `json:"error,omitempty"`
}

func buildValidationResult(doc *mdocx.Document, header *headerInfo, headerErr error) validationResult {
	result := validationResult{Valid: true}
	result.MarkdownFileCount = len(doc.Markdown.Files)
	result.MediaItemCount = len(doc.Media.Items)
	for _, mf := range doc.Markdown.Files {
		result.TotalMarkdownBytes += len(mf.Content)
	}
	for _, mi := range doc.Media.Items {
		result.TotalMediaBytes += len(mi.Data)
	}

	// Validate header
	if headerErr != nil {
		result.Valid = false
		result.Warnings = append(result.Warnings, fmt.Sprintf("header read error: %v", headerErr))
	} else if header != nil {
		if !header.MagicValid {
			result.Valid = false
			result.Warnings = append(result.Warnings, "invalid magic bytes")
		}
		if header.Version != 1 {
			result.Valid = false
			result.Warnings = append(result.Warnings, fmt.Sprintf("header version is %d, expected 1", header.Version))
		}
		if header.FixedHdrSize != 32 {
			result.Valid = false
			result.Warnings = append(result.Warnings, fmt.Sprintf("fixed header size is %d, expected 32", header.FixedHdrSize))
		}
		if !header.ReservedClean {
			result.Valid = false
			result.Warnings = append(result.Warnings, "reserved header bytes are not zero")
		}
	}

	// Validate BundleVersion
	if doc.Markdown.BundleVersion != 1 {
		result.Valid = false
		result.Warnings = append(result.Warnings, fmt.Sprintf("markdown BundleVersion is %d, expected 1", doc.Markdown.BundleVersion))
	}
	if doc.Media.BundleVersion != 1 {
		result.Valid = false
		result.Warnings = append(result.Warnings, fmt.Sprintf("media BundleVersion is %d, expected 1", doc.Media.BundleVersion))
	}

	// Check unique Markdown paths
	seenPaths := make(map[string]struct{})
	for _, mf := range doc.Markdown.Files {
		if _, ok := seenPaths[mf.Path]; ok {
			result.Valid = false
			result.Warnings = append(result.Warnings, fmt.Sprintf("duplicate markdown path: %q", mf.Path))
		}
		seenPaths[mf.Path] = struct{}{}
	}

	// Check unique Media IDs
	seenIDs := make(map[string]struct{})
	for _, mi := range doc.Media.Items {
		if _, ok := seenIDs[mi.ID]; ok {
			result.Valid = false
			result.Warnings = append(result.Warnings, fmt.Sprintf("duplicate media ID: %q", mi.ID))
		}
		seenIDs[mi.ID] = struct{}{}
	}

	return result
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().Bool("json", false, "output machine-readable JSON")
	validateCmd.Flags().Bool("strict", true, "fail on any spec violation")
}
