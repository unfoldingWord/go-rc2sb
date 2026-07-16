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

// NewBibleHandler creates a new Bible handler for the given subject.
// This handler works for any USFM-based Bible subject: "Aligned Bible",
// "Bible", "Hebrew Old Testament", "Greek New Testament", etc.
// The abbreviation is derived from the RC manifest's dublin_core.identifier
// (uppercased), e.g. "ult" → "ULT", "uhb" → "UHB", "ugnt" → "UGNT".
func NewBibleHandler(subject string) Handler {
	return &bibleHandler{subject: subject}
}

type bibleHandler struct {
	subject string
}

func (h *bibleHandler) Subject() string {
	return h.subject
}

func (h *bibleHandler) Convert(ctx context.Context, manifest *rc.Manifest, inDir, outDir string, opts Options) (*sb.Metadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m := BuildBaseMetadata(manifest, opts, "")

	// Set type - scripture/textTranslation
	currentScope := make(map[string][]string)
	m.Type = sb.Type{
		FlavorType: sb.FlavorType{
			Name: "scripture",
			Flavor: sb.Flavor{
				Name:            "textTranslation",
				USFMVersion:     "3.0",
				TranslationType: "revision",
				Audience:        "common",
				ProjectType:     "standard",
			},
		},
	}

	m.Copyright = BuildCopyright(manifest, false)

	lang := manifest.DublinCore.Language.Identifier

	// Process each project
	for _, project := range manifest.Projects {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// Get the source file path
		srcPath := filepath.Join(inDir, strings.TrimPrefix(project.Path, "./"))
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}
		srcFilename := filepath.Base(srcPath)

		// Convert filename: "01-GEN.usfm" -> "GEN.usfm"
		bookCode := extractBookCode(srcFilename)
		destFilename := bookCode + ".usfm"
		ingredientKey := "ingredients/" + destFilename

		// Determine scope
		bookID := strings.ToLower(project.Identifier)
		var scope map[string][]string

		if books.IsBookID(bookID) {
			code := books.CodeFromProjectID(bookID)
			scope = map[string][]string{code: {}}
			currentScope[code] = []string{}

			// Parse USFM file for localized book names (\toc1, \toc2, \toc3)
			usfmNames := books.ParseUSFMBookNames(srcPath)

			// Add localized name using: USFM > manifest project title > English fallback
			key, localizedName := books.LocalizedNameEntryWithNames(bookID, lang, project.Title, usfmNames)
			if key != "" {
				m.LocalizedNames[key] = localizedName
			}
		}

		// Copy file with scope
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

// extractBookCode extracts the book code from a USFM filename.
// "01-GEN.usfm" -> "GEN", "A0-FRT.usfm" -> "FRT"
func extractBookCode(filename string) string {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	parts := strings.SplitN(name, "-", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return name
}
