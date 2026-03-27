# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build ./cmd/pyrgear/

# Run
go run ./cmd/pyrgear/ <subcommand>

# Run all tests
go test ./...

# Run a single test
go test ./internal/comands/ -run TestName

# Run with verbose output
go test ./internal/comands/ -v
```

## Architecture

This is a Go CLI tool built with [cobra](https://github.com/spf13/cobra). Entry point is `cmd/pyrgear/main.go`, which calls `comands.Execute()`.

All commands live in `internal/comands/`:
- `root.go` — defines `RootCmd` and registers subcommands
- `rename.go` — `rename` subcommand: batch file renaming via regex patterns or predefined rules
- `exif.go` — `exif` subcommand: reads EXIF metadata from JPEG/TIFF images using `goexif`

### rename command rules

The `--rule` flag dispatches to different rename strategies:
- `timestamp` — prepends file modification time (`YYYYMMDD_HHMMSS_`) to filename
- `sequence` — renames files to `<name>_001.ext`, `<name>_002.ext`, ... (default prefix `file`, override with `--sequence-name`)
- `lowercase` — converts filenames to lowercase
- `prefix` — prepends `--prefix` string to all entries
- `foldername-rename` — renames all files in a dir to `foldername_001.ext`, etc.; use `--pdir` for batch mode across subdirs
- `wx-exporter` — copies images from `<source-path>/<subdir>/assets/` to `--output-dir`, named `<sourcename>_<subdir>_001.ext`; override source name with `--pre-name`

The `wx-exporter` and `foldername-rename` rules are handled before the standard rule dispatch in `RenameCmd.Run`.

### Note on package name

The commands package is named `comands` (single `m`) — this is intentional in the codebase.
