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

// packCmd represents the pack command
var packCmd = &cobra.Command{
	Use:   "pack",
	Short: "Create an .mdocx bundle",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		mdDir, _ := cmd.Flags().GetString("markdown-dir")
		mediaDir, _ := cmd.Flags().GetString("media-dir")
		metadataPath, _ := cmd.Flags().GetString("metadata")
		rootPath, _ := cmd.Flags().GetString("root")
		compressionName, _ := cmd.Flags().GetString("compression")
		outputPath, _ := cmd.Flags().GetString("output")

		if mdDir == "" && len(args) == 0 {
			return fmt.Errorf("provide markdown files or --markdown-dir")
		}

		markdownFiles, err := collectMarkdownFiles(args, mdDir)
		if err != nil {
			return fmt.Errorf("collect markdown: %w", err)
		}

		if rootPath != "" {
			found := false
			for _, f := range markdownFiles {
				if f.Path == rootPath {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("root path %q not found in markdown files", rootPath)
			}
		}

		mediaItems, err := collectMediaItems(mediaDir)
		if err != nil {
			return fmt.Errorf("collect media: %w", err)
		}

		compression, err := parseCompression(compressionName)
		if err != nil {
			return err
		}

		var metadata map[string]any
		if metadataPath != "" {
			metadata, err = readMetadataJSON(metadataPath)
			if err != nil {
				return fmt.Errorf("read metadata: %w", err)
			}
		}

		doc := &mdocx.Document{
			Metadata: metadata,
			Markdown: mdocx.MarkdownBundle{
				BundleVersion: mdocx.VersionV1,
				RootPath:      rootPath,
				Files:         markdownFiles,
			},
			Media: mdocx.MediaBundle{
				BundleVersion: mdocx.VersionV1,
				Items:         mediaItems,
			},
		}

		if err := validateContainerPaths(doc); err != nil {
			return fmt.Errorf("invalid bundle: %w", err)
		}

		out, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("create output: %w", err)
		}
		success := false
		defer func() {
			out.Close()
			if !success {
				os.Remove(outputPath)
			}
		}()

		if err := mdocx.Encode(out, doc,
			mdocx.WithMarkdownCompression(compression),
			mdocx.WithMediaCompression(compression),
			mdocx.WithVerifyHashesOnWrite(true),
		); err != nil {
			return fmt.Errorf("encode: %w", err)
		}
		success = true

		fmt.Fprintf(cmd.OutOrStdout(), "Packed %d markdown files and %d media items into %s\n", len(markdownFiles), len(mediaItems), outputPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(packCmd)

	packCmd.Flags().String("markdown-dir", "", "directory containing markdown files")
	packCmd.Flags().String("media-dir", "", "directory containing media files")
	packCmd.Flags().String("metadata", "", "path to metadata JSON file")
	packCmd.Flags().String("root", "", "root markdown path inside the bundle")
	packCmd.Flags().String("compression", "zstd", "compression (none|zip|zstd|lz4|br)")
	packCmd.Flags().StringP("output", "o", "bundle.mdocx", "output .mdocx file")
}
