/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/logicossoftware/go-mdocx"
	"github.com/spf13/cobra"
)

// inspectCmd represents the inspect command
var inspectCmd = &cobra.Command{
	Use:   "inspect <file>",
	Short: "Inspect an .mdocx bundle without extracting",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOut, _ := cmd.Flags().GetBool("json")
		input := args[0]

		f, err := os.Open(input)
		if err != nil {
			return fmt.Errorf("open: %w", err)
		}
		defer f.Close()

		doc, err := mdocx.Decode(f)
		if err != nil {
			return fmt.Errorf("decode: %w", err)
		}

		header, _ := readHeaderInfo(input)
		summary := buildInspectSummary(doc, header)

		if jsonOut {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(summary)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Metadata keys: %v\n", summary.MetadataKeys)
		fmt.Fprintf(cmd.OutOrStdout(), "Markdown files (%d): %v\n", len(summary.MarkdownFiles), summary.MarkdownFiles)
		fmt.Fprintf(cmd.OutOrStdout(), "Media IDs (%d): %v\n", len(summary.MediaIDs), summary.MediaIDs)
		if len(summary.MediaPaths) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Media paths: %v\n", summary.MediaPaths)
		}
		if header != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Header: version=%d flags=0x%04x metadata_len=%d\n", header.Version, header.HeaderFlags, header.MetadataLength)
		}
		return nil
	},
}

type inspectSummary struct {
	Header        *headerInfo `json:"header,omitempty"`
	MetadataKeys  []string    `json:"metadata_keys"`
	MarkdownFiles []string    `json:"markdown_files"`
	MediaIDs      []string    `json:"media_ids"`
	MediaPaths    []string    `json:"media_paths"`
}

func buildInspectSummary(doc *mdocx.Document, header *headerInfo) inspectSummary {
	var s inspectSummary
	s.Header = header
	for k := range doc.Metadata {
		s.MetadataKeys = append(s.MetadataKeys, k)
	}
	for _, mf := range doc.Markdown.Files {
		s.MarkdownFiles = append(s.MarkdownFiles, mf.Path)
	}
	for _, mi := range doc.Media.Items {
		s.MediaIDs = append(s.MediaIDs, mi.ID)
		if mi.Path != "" {
			s.MediaPaths = append(s.MediaPaths, mi.Path)
		}
	}
	sort.Strings(s.MetadataKeys)
	sort.Strings(s.MarkdownFiles)
	sort.Strings(s.MediaIDs)
	sort.Strings(s.MediaPaths)
	return s
}

func init() {
	rootCmd.AddCommand(inspectCmd)

	inspectCmd.Flags().Bool("json", false, "output machine-readable JSON")
}
