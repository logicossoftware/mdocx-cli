# mdocx-cli

A command-line toolkit for working with MDOCX v1 containers — a binary format for bundling Markdown documents with embedded media.

## Features

- **Pack** Markdown files and media into a single `.mdocx` container
- **Unpack** containers to extract Markdown and media files
- **Inspect** container structure, metadata, and contents
- **Validate** containers against the MDOCX v1 specification
- **Browse** containers interactively with a terminal UI (TUI)

## Installation

### Windows (Installer)

Download the latest `mdocx-cli-x.x.x-windows-amd64-setup.exe` from [Releases](https://github.com/logicossoftware/mdocx-cli/releases). The installer adds `mdocx` to your PATH automatically.

### Build from Source

```bash
go install github.com/logicossoftware/mdocx-cli@latest
```

Or clone and build:

```bash
git clone https://github.com/logicossoftware/mdocx-cli.git
cd mdocx-cli
go build .
```

## Usage

### Pack

Create an `.mdocx` from Markdown files and optional media:

```bash
mdocx pack ./docs --media ./images --output bundle.mdocx
mdocx pack README.md --compression zstd --output readme.mdocx
```

Options:
- `--output, -o` — Output file path (required)
- `--media, -m` — Directory containing media files
- `--metadata` — JSON file with container metadata
- `--compression, -c` — Compression algorithm: `none`, `zip`, `zstd`, `lz4`, `br` (default: `none`)
- `--root` — Root path prefix for files in the bundle

### Unpack

Extract an `.mdocx` to a directory:

```bash
mdocx unpack bundle.mdocx --output ./extracted
mdocx unpack bundle.mdocx -o ./out --strict
```

Options:
- `--output, -o` — Output directory (default: current directory)
- `--strict` — Fail on any spec violation

### Inspect

Display container information without extracting:

```bash
mdocx inspect bundle.mdocx
mdocx inspect bundle.mdocx --json
```

Options:
- `--json` — Output as JSON for scripting

### Validate

Check container integrity and spec compliance:

```bash
mdocx validate bundle.mdocx
```

### Browse

Open an interactive TUI to explore container contents:

```bash
mdocx browse bundle.mdocx
mdocx browse bundle.mdocx --theme dark --no-images
```

Options:
- `--theme` — Glamour theme for Markdown rendering
- `--no-images` — Disable Sixel image rendering
- `--strict` — Fail on any spec violation

## MDOCX Format Overview

MDOCX v1 is a binary container format with:

- **32-byte fixed header** — Magic bytes, version, flags, metadata length
- **Optional metadata block** — UTF-8 JSON object
- **Markdown section** — Gob-encoded bundle of Markdown files with media references
- **Media section** — Gob-encoded bundle of media items with SHA-256 checksums

Supported compression: None, ZIP, Zstandard, LZ4, Brotli.

## License

MIT License — see [LICENSE](LICENSE) for details.

## Author

© 2026 Logicos Software
