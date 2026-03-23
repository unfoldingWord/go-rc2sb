package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/unfoldingWord/go-rc2sb/books"
	"github.com/unfoldingWord/go-rc2sb/rc"
	"github.com/unfoldingWord/go-rc2sb/sb"
)

// NewTWHandler creates a new Translation Words handler.
func NewTWHandler() Handler {
	return &twHandler{}
}

type twHandler struct{}

func (h *twHandler) Subject() string {
	return "Translation Words"
}

func (h *twHandler) Convert(ctx context.Context, manifest *rc.Manifest, inDir, outDir string, opts Options) (*sb.Metadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m := BuildBaseMetadata(manifest, "uWBurritos", "TW")

	// Set type - parascriptural/x-bcvarticles (same as TWL)
	currentScope := make(map[string][]string)
	m.Type = sb.Type{
		FlavorType: sb.FlavorType{
			Name: "parascriptural",
			Flavor: sb.Flavor{
				Name: "x-bcvarticles",
			},
		},
	}

	m.Copyright = BuildCopyright(manifest, false)
	m.LocalizedNames = map[string]sb.LocalizedName{}

	lang := manifest.DublinCore.Language.Identifier

	// Always copy TW bible/ to ingredients/payload/
	bibleDir := filepath.Join(inDir, "bible")
	if err := copyTreeToIngredients(bibleDir, outDir, "ingredients/payload", m); err != nil {
		return nil, fmt.Errorf("copying TW bible to payload: %w", err)
	}

	// Determine TWL source: explicit TWLPath option, or auto-detect <lang>_twl/ in inDir
	var twlDir string
	if opts.TWLPath != "" {
		twlDir = opts.TWLPath
	} else {
		twlDir = filepath.Join(inDir, lang+"_twl")
	}

	_, twlDirErr := os.Stat(twlDir)
	hasTWL := twlDirErr == nil

	if hasTWL {
		// Find all twl_*.tsv files in the TWL directory
		tsvFiles, err := filepath.Glob(filepath.Join(twlDir, "twl_*.tsv"))
		if err != nil {
			return nil, fmt.Errorf("finding TWL TSV files: %w", err)
		}

		for _, srcPath := range tsvFiles {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			srcFilename := filepath.Base(srcPath)
			// Strip "twl_" prefix: "twl_GEN.tsv" -> "GEN.tsv"
			destFilename := strings.TrimPrefix(srcFilename, "twl_")
			ingredientKey := "ingredients/" + destFilename

			// Derive book code from filename: "GEN.tsv" -> "GEN"
			bookCode := strings.TrimSuffix(destFilename, ".tsv")
			scope := map[string][]string{bookCode: {}}
			currentScope[bookCode] = []string{}

			// Add localized name
			bookID := strings.ToLower(bookCode)
			var usfmNames *books.LocalizedBookNames
			if opts.USFMPath != "" {
				if usfmFile := books.FindUSFMFile(opts.USFMPath, bookID); usfmFile != "" {
					usfmNames = books.ParseUSFMBookNames(usfmFile)
				}
			}
			key, localizedName := books.LocalizedNameEntryWithNames(bookID, lang, "", usfmNames)
			if key != "" {
				m.LocalizedNames[key] = localizedName
			}

			// Copy TSV file with rc:// link rewriting
			ing, err := copyTSVWithLinkRewrite(srcPath, outDir, ingredientKey, scope)
			if err != nil {
				return nil, fmt.Errorf("copying %s with link rewrite: %w", srcFilename, err)
			}
			m.Ingredients[ingredientKey] = ing
		}
	}

	// Set the currentScope
	m.Type.FlavorType.CurrentScope = currentScope

	// Copy common root files (README.md, .gitignore, .gitea, .github)
	if err := CopyCommonRootFiles(inDir, outDir, m); err != nil {
		return nil, err
	}

	// Copy LICENSE.md to root (uses embedded default if RC doesn't have one).
	if err := CopyLicenseToRoot(inDir, outDir); err != nil {
		return nil, fmt.Errorf("copying root LICENSE.md: %w", err)
	}

	// Copy LICENSE.md to ingredients/
	licIng, err := CopyLicenseIngredient(inDir, outDir)
	if err != nil {
		return nil, fmt.Errorf("copying ingredients/LICENSE.md: %w", err)
	}
	m.Ingredients["ingredients/LICENSE.md"] = licIng

	return m, nil
}

// copyTreeToIngredients recursively copies a directory tree into the ingredients directory.
func copyTreeToIngredients(srcDir, outDir, destPrefix string, m *sb.Metadata) error {
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

		ingredientKey := destPrefix + "/" + filepath.ToSlash(relPath)

		ing, err := CopyFileAndComputeIngredient(path, outDir, ingredientKey)
		if err != nil {
			return fmt.Errorf("copying %s: %w", relPath, err)
		}
		m.Ingredients[ingredientKey] = ing

		return nil
	})
}
