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

// bibleTSVConfig holds the configuration for a Bible-based TSV variant
// (one TSV file per Bible book, e.g. tn_GEN.tsv).
type bibleTSVConfig struct {
	subject      string // e.g., "TSV Translation Notes"
	flavorName   string // e.g., "x-bcvnotes"
	abbreviation string // e.g., "TN"
	tsvPrefix    string // e.g., "tn_"
}

// bibleTSVHandler handles conversion for Bible-based TSV variants:
// Translation Notes, Translation Questions, Study Notes, and Study Questions.
type bibleTSVHandler struct {
	config bibleTSVConfig
}

func (h *bibleTSVHandler) Subject() string {
	return h.config.subject
}

func (h *bibleTSVHandler) Convert(ctx context.Context, manifest *rc.Manifest, inDir, outDir string, opts Options) (*sb.Metadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m := BuildBaseMetadata(manifest, "uWBurritos", h.config.abbreviation)

	// Set type - parascriptural with the variant's flavor
	currentScope := make(map[string][]string)
	m.Type = sb.Type{
		FlavorType: sb.FlavorType{
			Name: "parascriptural",
			Flavor: sb.Flavor{
				Name: h.config.flavorName,
			},
		},
	}

	m.Copyright = BuildCopyright(manifest, false)

	lang := manifest.DublinCore.Language.Identifier

	// Process each project (TSV file per book)
	for _, project := range manifest.Projects {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		srcPath := filepath.Join(inDir, strings.TrimPrefix(project.Path, "./"))
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}
		srcFilename := filepath.Base(srcPath)

		// Strip the variant prefix: "tn_GEN.tsv" -> "GEN.tsv"
		destFilename := strings.TrimPrefix(srcFilename, h.config.tsvPrefix)
		ingredientKey := "ingredients/" + destFilename

		// Get book code for scope
		bookID := strings.ToLower(project.Identifier)
		bookCode := books.CodeFromProjectID(bookID)

		scope := map[string][]string{bookCode: {}}
		currentScope[bookCode] = []string{}

		// Add localized name: try USFM from USFMPath, then manifest title, then English
		var usfmNames *books.LocalizedBookNames
		if opts.USFMPath != "" {
			if usfmFile := books.FindUSFMFile(opts.USFMPath, bookID); usfmFile != "" {
				usfmNames = books.ParseUSFMBookNames(usfmFile)
			}
		}
		key, localizedName := books.LocalizedNameEntryWithNames(bookID, lang, project.Title, usfmNames)
		if key != "" {
			m.LocalizedNames[key] = localizedName
		}

		// Copy TSV file with scope
		ing, err := CopyFileWithScope(srcPath, outDir, ingredientKey, scope)
		if err != nil {
			return nil, fmt.Errorf("copying %s: %w", srcFilename, err)
		}
		m.Ingredients[ingredientKey] = ing
	}

	// Set the currentScope
	m.Type.FlavorType.CurrentScope = currentScope

	// Copy common root files (README.md, .gitignore, .gitea, .github)
	if err := CopyCommonRootFiles(inDir, outDir, m); err != nil {
		return nil, err
	}

	// Copy LICENSE.md to ingredients/
	licIng, err := CopyLicenseIngredient(inDir, outDir)
	if err != nil {
		return nil, fmt.Errorf("copying LICENSE.md: %w", err)
	}
	m.Ingredients["ingredients/LICENSE.md"] = licIng

	return m, nil
}

// NewBibleTSVHandler creates a new handler for a Bible-based TSV variant.
func NewBibleTSVHandler(subject, flavorName, abbreviation, tsvPrefix string) Handler {
	return &bibleTSVHandler{
		config: bibleTSVConfig{
			subject:      subject,
			flavorName:   flavorName,
			abbreviation: abbreviation,
			tsvPrefix:    tsvPrefix,
		},
	}
}
