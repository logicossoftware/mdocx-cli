/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/logicossoftware/go-mdocx"
	"github.com/spf13/cobra"
)

// unpackCmd represents the unpack command
var unpackCmd = &cobra.Command{
	Use:   "unpack <file>",
	Short: "Extract an .mdocx bundle",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		outDir, _ := cmd.Flags().GetString("output")
		strict, _ := cmd.Flags().GetBool("strict")
		force, _ := cmd.Flags().GetBool("force")

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

		return writeUnpacked(doc, outDir, force, cmd.OutOrStdout())
	},
}

func writeUnpacked(doc *mdocx.Document, outDir string, force bool, out io.Writer) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	// Helper to check for existing files when --force is not set.
	checkOverwrite := func(path string) error {
		if force {
			return nil
		}
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("file already exists: %s (use --force to overwrite)", path)
		}
		return nil
	}

	if doc.Metadata != nil {
		b, err := json.MarshalIndent(doc.Metadata, "", "  ")
		if err != nil {
			return fmt.Errorf("metadata json: %w", err)
		}
		p := filepath.Join(outDir, "metadata.json")
		if err := checkOverwrite(p); err != nil {
			return err
		}
		if err := os.WriteFile(p, b, 0o644); err != nil {
			return fmt.Errorf("write metadata: %w", err)
		}
		fmt.Fprintf(out, "wrote %s\n", p)
	}

	for _, mf := range doc.Markdown.Files {
		p, err := safeJoinOutput(outDir, mf.Path)
		if err != nil {
			return err
		}
		if err := checkOverwrite(p); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return fmt.Errorf("mkdir: %w", err)
		}
		if err := os.WriteFile(p, mf.Content, 0o644); err != nil {
			return fmt.Errorf("write markdown: %w", err)
		}
		fmt.Fprintf(out, "wrote %s\n", p)
	}

	for _, mi := range doc.Media.Items {
		containerPath := mi.Path
		if strings.TrimSpace(containerPath) == "" {
			containerPath = filepath.ToSlash(filepath.Join("media", mi.ID))
		}
		p, err := safeJoinOutput(outDir, containerPath)
		if err != nil {
			return err
		}
		if err := checkOverwrite(p); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return fmt.Errorf("mkdir: %w", err)
		}
		if err := os.WriteFile(p, mi.Data, 0o644); err != nil {
			return fmt.Errorf("write media: %w", err)
		}
		fmt.Fprintf(out, "wrote %s\n", p)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(unpackCmd)

	unpackCmd.Flags().StringP("output", "o", "out", "output directory")
	unpackCmd.Flags().Bool("strict", true, "fail on any spec violation")
	unpackCmd.Flags().BoolP("force", "f", false, "overwrite existing files")
}
