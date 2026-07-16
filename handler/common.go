package handler

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/unfoldingWord/go-rc2sb/rc"
	"github.com/unfoldingWord/go-rc2sb/sb"
)

// defaultLicense is the embedded CC BY-SA 4.0 LICENSE.md used as a fallback
// when the RC repository does not include its own LICENSE.md file.
//
//go:embed default_license.md
var defaultLicense []byte

// CopyFile copies a file from src to dst, creating any necessary directories.
func CopyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("creating directory for %s: %w", dst, err)
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating destination %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copying %s to %s: %w", src, dst, err)
	}

	return out.Close()
}

// CopyFileAndComputeIngredient copies a file and computes its ingredient entry.
// Returns the ingredient key (relative path in SB) and the Ingredient.
func CopyFileAndComputeIngredient(src, outDir, ingredientKey string) (sb.Ingredient, error) {
	dst := filepath.Join(outDir, ingredientKey)
	if err := CopyFile(src, dst); err != nil {
		return sb.Ingredient{}, err
	}
	return sb.ComputeIngredient(dst)
}

// CopyFileWithScope copies a file and computes its ingredient entry with scope.
func CopyFileWithScope(src, outDir, ingredientKey string, scope map[string][]string) (sb.Ingredient, error) {
	dst := filepath.Join(outDir, ingredientKey)
	if err := CopyFile(src, dst); err != nil {
		return sb.Ingredient{}, err
	}
	return sb.ComputeIngredientWithScope(dst, scope)
}

// IDAuthorityDCS is the idAuthorities key used for the DCS instance that
// hosts the source repository and mints the "owner/repo" primary identifier.
const IDAuthorityDCS = "dcs"

// DefaultDCSURL is the ID authority URL used when none is provided or detected.
const DefaultDCSURL = "https://git.door43.org"

// BuildBaseMetadata creates a base SB Metadata from an RC manifest with common fields set.
// The primary identifier is "owner/repo" under the DCS ID authority. Owner, repo name,
// DCS URL, and revision come from opts (resolved by Convert from explicit options or the
// git clone metadata); anything missing falls back to the manifest: dublin_core.publisher
// for the owner, the "<language>_<identifier>" DCS naming convention for the repo name,
// and dublin_core.version (then "1") for the revision.
func BuildBaseMetadata(manifest *rc.Manifest, opts Options, abbreviation string) *sb.Metadata {
	m := sb.NewMetadata()

	now := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	m.Meta.DateCreated = now

	dc := manifest.DublinCore

	// Set ID authority
	dcsURL := opts.DCSURL
	if dcsURL == "" {
		dcsURL = DefaultDCSURL
	}
	m.IDAuthorities[IDAuthorityDCS] = sb.IDAuthority{
		ID:   dcsURL,
		Name: map[string]string{"en": "Door43 Content Service"},
	}

	// Set identification
	abbr := abbreviation
	if abbr == "" {
		abbr = strings.ToUpper(dc.Identifier)
	}

	owner := opts.Owner
	if owner == "" {
		// Publisher is the closest manifest field to the DCS owner; sanitize it
		// into a slug-safe form (e.g. "unfoldingWord®" -> "unfoldingWord").
		owner = strings.ReplaceAll(strings.TrimSpace(strings.ReplaceAll(dc.Publisher, "®", "")), " ", "-")
	}
	repoName := opts.RepoName
	if repoName == "" && dc.Language.Identifier != "" && dc.Identifier != "" {
		repoName = dc.Language.Identifier + "_" + dc.Identifier
	}
	primaryID := repoName
	if owner != "" && repoName != "" {
		primaryID = owner + "/" + repoName
	}
	if primaryID == "" {
		primaryID = abbr
	}
	revision := opts.Revision
	if revision == "" {
		revision = dc.Version
	}
	if revision == "" {
		revision = "1"
	}

	m.Identification = sb.Identification{
		Primary: map[string]map[string]sb.PrimaryEntry{
			IDAuthorityDCS: {
				primaryID: {
					Revision:  revision,
					Timestamp: now,
				},
			},
		},
		Name:         map[string]string{"en": dc.Title},
		Description:  map[string]string{"en": dc.Title},
		Abbreviation: map[string]string{"en": abbr},
	}

	// Set language
	m.Languages = []sb.LanguageEntry{
		{
			Tag:             dc.Language.Identifier,
			Name:            map[string]string{"en": dc.Language.Title},
			ScriptDirection: dc.Language.Direction,
		},
	}

	return m
}

// BuildCopyright generates a copyright statement from the RC manifest.
// Uses the format "© {publisher} {year}, {rights}" for most types,
// or "Copyright © {year} by {publisher}" for OBS.
func BuildCopyright(manifest *rc.Manifest, isOBS bool) sb.Copyright {
	dc := manifest.DublinCore
	year := dc.Issued
	if len(year) >= 4 {
		year = year[:4]
	}

	if isOBS {
		return sb.Copyright{
			ShortStatements: []sb.CopyrightStatement{
				{
					Statement: fmt.Sprintf("Copyright \u00a9 %s by %s", year, dc.Publisher),
				},
			},
		}
	}

	return sb.Copyright{
		ShortStatements: []sb.CopyrightStatement{
			{
				Statement: fmt.Sprintf("\u00a9 %s %s, %s", dc.Publisher, year, dc.Rights),
				MimeType:  "text/plain",
				Lang:      "en",
			},
		},
	}
}

// CopyLicenseIngredient copies LICENSE.md from the RC repo to ingredients/LICENSE.md
// and returns the ingredient. If the RC repo does not contain a LICENSE.md file,
// the embedded default CC BY-SA 4.0 license is used instead.
func CopyLicenseIngredient(inDir, outDir string) (sb.Ingredient, error) {
	src := filepath.Join(inDir, "LICENSE.md")
	if _, err := os.Stat(src); os.IsNotExist(err) {
		// Use the embedded default LICENSE.md
		return writeDefaultLicenseIngredient(outDir)
	}
	return CopyFileAndComputeIngredient(src, outDir, "ingredients/LICENSE.md")
}

// writeDefaultLicenseIngredient writes the embedded default LICENSE.md
// to ingredients/LICENSE.md and computes its ingredient entry.
func writeDefaultLicenseIngredient(outDir string) (sb.Ingredient, error) {
	dst := filepath.Join(outDir, "ingredients", "LICENSE.md")
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return sb.Ingredient{}, fmt.Errorf("creating directory for %s: %w", dst, err)
	}
	if err := os.WriteFile(dst, defaultLicense, 0644); err != nil {
		return sb.Ingredient{}, fmt.Errorf("writing default LICENSE.md: %w", err)
	}
	return sb.ComputeIngredient(dst)
}

// CopyLicenseToRoot copies LICENSE.md from the RC repo to the SB output root directory.
// If the RC repo does not contain a LICENSE.md file, the embedded default is used instead.
func CopyLicenseToRoot(inDir, outDir string) error {
	src := filepath.Join(inDir, "LICENSE.md")
	dst := filepath.Join(outDir, "LICENSE.md")
	if _, err := os.Stat(src); os.IsNotExist(err) {
		// Use the embedded default LICENSE.md
		return os.WriteFile(dst, defaultLicense, 0644)
	}
	return CopyFile(src, dst)
}

// CopyRootFile copies a root-level file from RC to SB root and returns the ingredient.
func CopyRootFile(inDir, outDir, filename string) (sb.Ingredient, error) {
	src := filepath.Join(inDir, filename)
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return sb.Ingredient{}, nil // File doesn't exist, skip silently
	}
	return CopyFileAndComputeIngredient(src, outDir, filename)
}

// CopyCommonRootFiles copies common root-level files from the RC repo to the SB output
// if they exist: README.md, .gitea, .github, .gitignore (but NOT .git).
// Files are copied to the SB root but are intentionally NOT added to metadata ingredients.
func CopyCommonRootFiles(inDir, outDir string, _ *sb.Metadata) error {
	// Individual files to copy
	files := []string{"README.md", ".gitignore"}
	for _, name := range files {
		src := filepath.Join(inDir, name)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}
		if err := CopyFile(src, filepath.Join(outDir, name)); err != nil {
			return fmt.Errorf("copying root file %s: %w", name, err)
		}
	}

	// Directories to copy recursively
	dirs := []string{".gitea", ".github"}
	for _, dirName := range dirs {
		src := filepath.Join(inDir, dirName)
		info, err := os.Stat(src)
		if os.IsNotExist(err) || !info.IsDir() {
			continue
		}
		if err := copyTree(src, outDir, dirName); err != nil {
			return fmt.Errorf("copying root directory %s: %w", dirName, err)
		}
	}

	return nil
}

// copyTree recursively copies srcDir into outDir/destPrefix without adding metadata entries.
func copyTree(srcDir, outDir, destPrefix string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(outDir, destPrefix, relPath)
		if err := CopyFile(path, dstPath); err != nil {
			return fmt.Errorf("copying %s: %w", relPath, err)
		}
		return nil
	})
}
