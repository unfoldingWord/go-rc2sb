# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go library (`github.com/unfoldingWord/go-rc2sb`) for converting Resource Container (RC) repositories to Scripture Burrito (SB) repositories. RC is a format used by unfoldingWord for Bible translation resources (spec: rc0.2). SB (Scripture Burrito) is the newer standardized format (spec: 1.0.0).

## Build & Test Commands

```bash
go build ./...           # Build all packages
go test ./...            # Run all tests
go test ./... -v         # Run tests with verbose output
go test -run TestName    # Run a specific test
go vet ./...             # Lint/static analysis

# Run only unit tests (no samples needed)
go test ./rc/... ./sb/... ./books/...

# Run integration tests (requires samples/ directory)
go test -run TestConvert -v
```

## Architecture

### Public API

The single entry point is `Convert()` in `convert.go`:

```go
func Convert(ctx context.Context, inDir string, outDir string, opts Options) (Result, error)
```

- Reads `manifest.yaml` from `inDir`, determines the subject, looks up the handler, runs conversion, writes `metadata.json` to `outDir`.
- `Options.PayloadPath` specifies an explicit path to a `<lang>_tw` directory for TWL payload creation. If empty, auto-detects `<lang>_tw/` inside `inDir`.
- `Options.TWLPath` specifies an explicit path to a `<lang>_twl` directory for TW conversion. If empty, auto-detects `<lang>_twl/` inside `inDir`.
- `Options.USFMPath` specifies a directory containing USFM files for localized Bible book names (used by TSV handlers). Bible handlers read USFM directly from their own input files.
- `Options.Owner`, `Options.RepoName`, `Options.DCSURL` specify the DCS repo identity used for the burrito's primary identifier (`owner/repo` under the `dcs` ID authority). If empty, detected from the git remote `origin` URL of `inDir`, then from the manifest (publisher + `<lang>_<identifier>`).
- `Options.Revision` specifies the converted source's revision (release number, tag, or commit SHA) for `identification.primary`. If empty, uses the git HEAD commit SHA of `inDir`, then `dublin_core.version`, then `"1"`.

### Package Structure

```
go-rc2sb/
├── convert.go              # Public Convert() function, orchestration
├── options.go              # Options and Result types
├── cmd/rc2sb/
│   └── main.go             # CLI wrapper
├── rc/
│   └── manifest.go         # RC manifest.yaml parsing (DublinCore, projects)
├── sb/
│   ├── metadata.go         # SB metadata.json types and JSON serialization
│   └── ingredient.go       # Ingredient computation (MD5, MIME type, size)
├── books/
│   └── books.go            # Bible book data (66 books, localized names, codes, USFM parsing)
├── handler/
│   ├── handler.go          # Handler interface definition
│   ├── registry.go         # Subject -> handler registry (Register/Lookup)
│   ├── common.go           # Shared helpers (file copy, metadata building, copyright)
│   ├── obs.go              # Open Bible Stories handler
│   ├── aligned_bible.go    # Bible/USFM handler (Aligned Bible, Bible, Hebrew OT, Greek NT)
│   ├── tw.go               # Translation Words handler
│   ├── ta.go               # Translation Academy handler
│   ├── bible_tsv.go        # Generic Bible TSV handler (TN, TQ, SN, SQ)
│   ├── twl.go              # TSV Translation Words Links handler (with payload)
│   ├── obs_tsv.go          # Generic OBS TSV handler (4 variants)
│   ├── obs_twl.go          # TSV OBS Translation Words Links handler
│   └── subjects/
│       └── register.go     # Registers all 17 handlers via init()
```

### Key Design Patterns

- **Handler pattern**: Each subject type implements the `Handler` interface (`Subject() string`, `Convert(...)`). Handlers are registered in `handler/subjects/register.go` via `init()`.
- **Blank import for registration**: `convert.go` imports `_ "github.com/unfoldingWord/go-rc2sb/handler/subjects"` to trigger handler registration.
- **Shared helpers in `handler/common.go`**: `BuildBaseMetadata()`, `BuildCopyright()`, `CopyFileAndComputeIngredient()`, `CopyFileWithScope()`, `CopyLicenseIngredient()`, `CopyCommonRootFiles()`.
- **Root file copying**: All handlers call `CopyCommonRootFiles()` which copies README.md, .gitignore, .gitea/, .github/ (but NOT .git/) from the RC repo if they exist.

### Subject -> SB Type Mapping

All subjects use the single `dcs` ID authority (the DCS instance hosting the source repo, default `https://git.door43.org`). The primary identifier is `owner/repo` (e.g., `unfoldingWord/en_tw`), resolved by priority: (1) `Options.Owner`/`Options.RepoName`, (2) the git remote `origin` URL of `inDir` (parsed from `.git/config`, see `identity.go`), (3) manifest fallback — `dublin_core.publisher` for the owner and `<language>_<identifier>` for the repo name. The primary entry's `revision` is resolved by priority: (1) `Options.Revision`, (2) the git HEAD commit SHA of `inDir`, (3) `dublin_core.version`, (4) `"1"`.

| Subject | FlavorType/Flavor | Abbreviation |
|---------|-------------------|--------------|
| Open Bible Stories | gloss/textStories | OBS |
| Aligned Bible | scripture/textTranslation | (from RC identifier) |
| Bible | scripture/textTranslation | (from RC identifier) |
| Hebrew Old Testament | scripture/textTranslation | (from RC identifier) |
| Greek New Testament | scripture/textTranslation | (from RC identifier) |
| Translation Words | parascriptural/x-bcvarticles | TW |
| Translation Academy | peripheral/x-peripheralArticles | TA |
| TSV Translation Notes | parascriptural/x-bcvnotes | TN |
| TSV Translation Questions | parascriptural/x-bcvquestions | TQ |
| TSV Study Notes | parascriptural/x-bcvnotes | SN |
| TSV Study Questions | parascriptural/x-bcvquestions | SQ |
| TSV Translation Words Links | parascriptural/x-bcvarticles | TW |
| TSV OBS Study Notes | peripheral/x-obsnotes | OBSSN |
| TSV OBS Study Questions | peripheral/x-obsquestions | OBSSQ |
| TSV OBS Translation Notes | peripheral/x-obsnotes | OBSTN |
| TSV OBS Translation Questions | peripheral/x-obsquestions | OBSTQ |
| TSV OBS Translation Words Links | parascriptural/x-bcvarticles | OBSTWL |

### RC Format (Input)
- **manifest.yaml**: Dublin Core metadata (conformsto: rc0.2), project list, language, versioning
- **media.yaml**: Optional media format definitions (PDF, audio, video URLs)
- **content/**: Resource files in various formats depending on type

### SB Format (Output)
- **metadata.json**: Scripture Burrito metadata with identification, languages, type/flavor, and an `ingredients` map listing every file with its MD5 checksum, MIME type, and size
- **ingredients/**: All content files organized under this directory

### Key Conversion Logic

1. **Metadata**: Transform `manifest.yaml` (Dublin Core) into `metadata.json` (Scripture Burrito schema) — map identifiers, versions, languages, project info
2. **File relocation**: Copy content files into `ingredients/` directory, adjusting paths per resource type (e.g., strip `tn_` prefix from TSV filenames, strip numeric prefix from USFM filenames)
3. **Checksum computation**: SB metadata.json requires MD5 checksums, MIME types, and byte sizes for every ingredient file
4. **Content preservation**: File contents (Markdown, USFM, TSV) are unchanged between formats (except TW/TWL TSV link rewriting)
5. **Root file copying**: README.md, .gitignore, .gitea/, .github/ are copied from RC to SB root if present (not .git/)
6. **TWL payload resolution**: If `Options.PayloadPath` is set or a `<lang>_tw/` subdirectory exists in the RC repo (where `<lang>` = `dublin_core.language.identifier`), copies the TW `bible/*` to `ingredients/payload/` and rewrites `rc://*/tw/dict/bible/{path}` links in TSV files to `./payload/{path}.md`
6b. **TW conversion (like TWL)**: Translation Words repos are converted identically to TWL — the TW `bible/*` is always copied to `ingredients/payload/`. If `Options.TWLPath` is set or a `<lang>_twl/` subdirectory exists in the RC repo, its `twl_*.tsv` files are processed as main ingredients (strip `twl_` prefix, rewrite `rc://` links to `./payload/` paths, set per-book scope).
6c. **OBS content**: Only the story files `01.md`–`50.md` plus `front.md` and `back.md` land in `ingredients/content/` — nothing else. `content/front/intro.md` is flattened to `ingredients/content/front.md` and `content/back/intro.md` to `ingredients/content/back.md` (the `front/`/`back/` directories are never recreated; flat `front.md`/`back.md` in the RC are copied through). `content/front/title.md` is not copied; its trimmed contents become `identification.name` and `identification.description` under the RC language tag, with `"en": "Open Bible Stories"` added as fallback whenever the title doesn't provide an English entry.
7. **Localized book names**: Bible book names in `localizedNames` are resolved by priority: (1) USFM `\toc1`/`\toc2`/`\toc3` markers from the USFM file itself (Bible handlers) or from `Options.USFMPath` (TSV handlers), (2) manifest `projects[].title`, (3) English fallback from `books/books.go`. The `books.ParseUSFMBookNames()` function reads the first 20 lines of a USFM file to extract these markers, falling back to `\mt`/`\h` when toc markers are absent.

### Testing

- **Integration tests** (`convert_test.go`): One test per subject type (14 total, including TWLPath and TWLPath variant tests). Requires `samples/` directory (gitignored) with RC/SB pairs. Tests compare structural metadata (flavor type, scope keys, abbreviation, language, ingredient keys) and verify internal consistency (every ingredient exists on disk with correct MD5 and size).
- **Unit tests**: `rc/manifest_test.go`, `sb/ingredient_test.go`, `sb/metadata_test.go`, `books/books_test.go`
- **Error handling tests** (`error_test.go`): Missing manifest, unsupported subject, cancelled context, invalid YAML

### Dependencies

- `gopkg.in/yaml.v3` — YAML parsing for RC manifest files
