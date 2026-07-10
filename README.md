# go-rc2sb

Go library for converting [Resource Container](https://resource-container.readthedocs.io/) (RC) repositories to [Scripture Burrito](https://docs.burrito.bible/) (SB) format.

## Installation

```bash
go get github.com/unfoldingWord/go-rc2sb
```

## Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    rc2sb "github.com/unfoldingWord/go-rc2sb"
)

func main() {
    ctx := context.Background()

    result, err := rc2sb.Convert(ctx, "/path/to/rc-repo", "/path/to/sb-output", rc2sb.Options{})
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Converted %s (%s) with %d ingredients\n",
        result.Subject, result.Identifier, result.Ingredients)
}
```

### TWL Payload

For TSV Translation Words Links repos, a payload directory can be created from a Translation Words source. When a payload is available, the handler will:
1. Copy `bible/*` from the TW directory to `ingredients/payload/` in the SB output
2. Rewrite `rc://*/tw/dict/bible/...` links in the TSV files to `./payload/...` paths

There are two ways to provide the TW source:

**Option 1: Explicit path via `PayloadPath`** — Use this when the TW directory is stored separately from the TWL repo:

```go
opts := rc2sb.Options{
    PayloadPath: "/path/to/en_tw",  // path to the <lang>_tw directory
}
result, err := rc2sb.Convert(ctx, "/path/to/en_twl", "/path/to/output", opts)
```

**Option 2: Auto-detection** — If the RC repo contains a `<lang>_tw/` subdirectory (where `<lang>` matches the manifest's language identifier, e.g., `en_tw/`), it is detected automatically:

```go
// The en_twl RC repo contains en_tw/ as a subdirectory — auto-detected
result, err := rc2sb.Convert(ctx, "/path/to/en_twl", "/path/to/output", rc2sb.Options{})
```

If neither `PayloadPath` is set nor a `<lang>_tw/` subdirectory exists, the TSV files are copied as-is without payload or link rewriting.

### Localized Book Names

Bible book names in the SB `localizedNames` are resolved using this priority:

1. **USFM `\toc1`/`\toc2`/`\toc3` markers** — For Bible/USFM repos, these are read directly from the USFM files in the input. For TSV repos, use `USFMPath` to point to a USFM directory.
2. **Manifest `projects[].title`** — The `title` field from the RC `manifest.yaml` project entries.
3. **English fallback** — Hardcoded English names from `books/books.go`.

For non-English repos, this ensures book names like "उत्पत्ति" (Hindi for Genesis) appear in the metadata instead of only English names.

```go
// Convert a Hindi TN repo with book names from a Hindi Bible USFM repo
opts := rc2sb.Options{
    USFMPath: "/path/to/hi_irv",  // directory containing NN-CODE.usfm files
}
result, err := rc2sb.Convert(ctx, "/path/to/hi_tn", "/path/to/output", opts)
```

### CLI Tool

A simple CLI wrapper is available at `cmd/rc2sb/`:

```bash
go run ./cmd/rc2sb /path/to/rc-repo /path/to/sb-output

# With localized book names from a USFM directory (for TSV repos)
go run ./cmd/rc2sb --usfm /path/to/hi_irv /path/to/hi_tn /path/to/sb-output

# With TWL payload
go run ./cmd/rc2sb --payload /path/to/en_tw /path/to/en_twl /path/to/sb-output
```

## API

### `Convert(ctx, inDir, outDir, opts) (Result, error)`

Converts an RC repository to SB format.

- `ctx` - Context for cancellation
- `inDir` - Path to the RC repository (must contain `manifest.yaml`)
- `outDir` - Path where SB output will be written
- `opts` - Conversion options (see Options below)

Returns a `Result` with conversion metadata, or an error.

### Options

```go
type Options struct {
    // PayloadPath is the path to a Translation Words directory (e.g., "/path/to/en_tw").
    // Used for TWL conversion to create the ingredients/payload/ directory and
    // rewrite rc:// links in TSV files. If empty, auto-detects <lang>_tw/ inside inDir.
    PayloadPath string

    // USFMPath is the path to a directory containing USFM files for localized
    // Bible book names. Used by TSV handlers (TN, TQ, TWL) to extract
    // \toc1, \toc2, \toc3 markers. If empty, uses manifest project titles,
    // then English fallback.
    USFMPath string
}
```

### Result

```go
type Result struct {
    Subject     string // RC subject that was converted
    Identifier  string // RC identifier (e.g., "obs", "ult", "tn")
    InDir       string // Input RC directory
    OutDir      string // Output SB directory
    Ingredients int    // Number of ingredient files
}
```

## Supported Subjects

| Subject | SB Flavor Type | Notes |
|---------|---------------|-------|
| Open Bible Stories | gloss/textStories | Copies story files 01.md-50.md plus front.md/back.md; front/title.md becomes name/description |
| Aligned Bible | scripture/textTranslation | Strips numeric prefix from USFM filenames; abbreviation from RC identifier |
| Bible | scripture/textTranslation | Same as Aligned Bible (e.g., ULT, UST) |
| Hebrew Old Testament | scripture/textTranslation | Same as Aligned Bible (e.g., UHB) |
| Greek New Testament | scripture/textTranslation | Same as Aligned Bible (e.g., UGNT) |
| Translation Words | parascriptural/x-bcvarticles | Copies bible/ to ingredients/payload/; auto-detects `<lang>_twl/` for TSV ingredients |
| Translation Academy | peripheral/x-peripheralArticles | Copies nested markdown hierarchy |
| TSV Translation Notes | parascriptural/x-bcvnotes | Strips tn_ prefix from TSV filenames |
| TSV Translation Questions | parascriptural/x-bcvquestions | Strips tq_ prefix from TSV filenames |
| TSV Study Notes | parascriptural/x-bcvnotes | Strips sn_ prefix from TSV filenames |
| TSV Study Questions | parascriptural/x-bcvquestions | Strips sq_ prefix from TSV filenames |
| TSV Translation Words Links | parascriptural/x-bcvarticles | Auto-detects `<lang>_tw/` for payload; rewrites rc:// links |
| TSV OBS Study Notes | peripheral/x-obsnotes | Single TSV file conversion |
| TSV OBS Study Questions | peripheral/x-obsquestions | Single TSV file conversion |
| TSV OBS Translation Notes | peripheral/x-obsnotes | Single TSV file conversion |
| TSV OBS Translation Questions | peripheral/x-obsquestions | Single TSV file conversion |
| TSV OBS Translation Words Links | parascriptural/x-bcvarticles | TWL-style payload and rc:// rewriting; OBS as a single book |

## Error Handling

- Missing `manifest.yaml` returns an error indicating the directory is not a valid RC repo
- Unsupported subjects return an error listing all supported subjects
- Context cancellation is checked at key points during conversion
- File I/O errors are wrapped with context and returned

## Building

```bash
go build ./...
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run a specific test
go test -run TestConvertOpenBibleStories

# Run only unit tests (no samples needed)
go test ./rc/... ./sb/... ./books/...

# Run integration tests (requires samples/ directory)
go test -run TestConvert -v
```

### Integration Tests

Integration tests use sample RC/SB pairs in the `samples/` directory (gitignored). Each test:

1. Runs `Convert()` on the sample RC input
2. Verifies the output metadata structure matches the expected SB metadata
3. Verifies internal consistency (every ingredient in metadata.json exists on disk with correct MD5 and size)

### Unit Tests

- `rc/manifest_test.go` - Manifest parsing (valid, invalid, missing)
- `sb/ingredient_test.go` - MD5/MIME/size computation
- `sb/metadata_test.go` - Metadata creation, serialization, round-trip
- `books/books_test.go` - Book lookups, localized names, sort order
- `error_test.go` - Error handling (missing manifest, unsupported subject, cancelled context)

## Architecture

```
go-rc2sb/
+-- convert.go              # Public Convert() function
+-- options.go              # Options and Result types
+-- cmd/rc2sb/
|   +-- main.go             # CLI wrapper
+-- rc/
|   +-- manifest.go         # RC manifest.yaml parsing
+-- sb/
|   +-- metadata.go         # SB metadata.json types
|   +-- ingredient.go       # Ingredient computation (MD5, MIME, size)
+-- books/
|   +-- books.go            # Bible book data (66 books, localized names)
+-- handler/
|   +-- handler.go          # Handler interface
|   +-- registry.go         # Subject -> handler registry
|   +-- common.go           # Shared helpers (file copy, metadata building)
|   +-- obs.go              # Open Bible Stories
|   +-- aligned_bible.go    # Bible/USFM handler (Aligned Bible, Bible, Hebrew OT, Greek NT)
|   +-- tw.go               # Translation Words
|   +-- ta.go               # Translation Academy
|   +-- bible_tsv.go        # Bible TSV variants (TN, TQ, SN, SQ)
|   +-- twl.go              # TSV Translation Words Links (with payload)
|   +-- obs_tsv.go          # OBS TSV variants (4 types)
|   +-- obs_twl.go          # TSV OBS Translation Words Links
|   +-- subjects/
|       +-- register.go     # Registers all handlers
```

## License

See [LICENSE](LICENSE) for details.
