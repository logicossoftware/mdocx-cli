# Changelog

## Unreleased

### Bug Fixes

- **SHA256 not computed for media items** — `collectMediaItems` now computes `sha256.Sum256` on media data so integrity hashes are populated per the MDOCX v1 spec.
- **Duplicate media IDs silently accepted** — `makeIDFromPath` collapses special characters to `_`, so paths like `a-b.png` and `a_b.png` would both produce ID `a_b_png`. `collectMediaItems` now detects and rejects ID collisions with a clear error.
- **Pack leaves orphaned output file on encode failure** — `os.Create` was called before encoding; if `Encode()` failed the partial file was left on disk. A deferred cleanup now removes the file on error.
- **Container paths not validated before encoding** — Malformed paths (traversal, absolute, backslash) would silently produce invalid bundles. `pack` now validates all container paths via `validateContainerPaths()` before encoding.
- **Validate didn't actually validate** — `buildValidationResult` only counted files; it never checked `BundleVersion`, unique markdown paths, or unique media IDs. All three are now enforced.
- **Validate didn't check header integrity** — Magic bytes, version, fixed header size (32), and reserved bytes (must be zero) are now verified.
- **Version command bypassed Cobra output** — `fmt.Printf` was used instead of `cmd.OutOrStdout()`, making the version command untestable and ignoring output redirection. Fixed to use `fmt.Fprintf` with the Cobra writer.
- **Nil slices serialized as `null` in JSON** — Inspect's `MetadataKeys`, `MarkdownFiles`, `MediaIDs`, and `MediaPaths` were nil slices, producing `null` in JSON output. Now initialized as empty slices so they serialize as `[]`.
- **`scaleImage` division by zero** — Zero-dimension images would cause a panic. Added a guard to return early for degenerate images.
- **Glamour renderer conflicting options** — `WithAutoStyle()` and `WithStylePath(theme)` were both applied simultaneously. Now only the explicit theme is used when provided, falling back to auto-style otherwise.

### Enhancements

- **Inspect shows RootPath and BundleVersion** — `inspect` output (text and JSON) now includes `root_path`, `markdown_bundle_version`, and `media_bundle_version`.
- **Inspect shows total sizes** — `inspect` output now includes `total_markdown_bytes` and `total_media_bytes` in both text and JSON output.
- **Reserved header bytes check** — `headerInfo` now includes a `ReservedClean` field indicating whether bytes 20–31 are all zero per the v1 spec. Shown in `inspect`, `browse`, and `validate`.
- **Hash verification on write** — `pack` now uses `WithVerifyHashesOnWrite(true)` to catch data corruption during encoding.
- **Human-readable file sizes** — `inspect` and `validate` text output now shows sizes like `35.32 KiB` / `255.31 KiB` alongside file/item counts.
- **Overwrite protection for `unpack`** — Added `--force` / `-f` flag. `unpack` no longer silently overwrites existing files; a clear error message directs users to use `--force`.
