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

// NewOBSTWLHandler creates a new TSV OBS Translation Words Links handler.
//
// OBS Translation Words Links are converted like the Bible TSV Translation Words
// Links handler (parascriptural/x-bcvarticles, with a Translation Words payload and
// rc:// link rewriting). The difference is that OBS is treated as a single book —
// each story is a chapter and each frame a verse — so the scope is keyed on the
// "OBS" book code and the lone project is twl_OBS.tsv -> OBS.tsv.
func NewOBSTWLHandler() Handler {
	return &obsTWLHandler{}
}

type obsTWLHandler struct{}

func (h *obsTWLHandler) Subject() string {
	return "TSV OBS Translation Words Links"
}

func (h *obsTWLHandler) Convert(ctx context.Context, manifest *rc.Manifest, inDir, outDir string, opts Options) (*sb.Metadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m := BuildBaseMetadata(manifest, "BurritoTruck", "OBSTWL")

	// Same parascriptural/x-bcvarticles type as the Bible TWL handler.
	m.Type = sb.Type{
		FlavorType: sb.FlavorType{
			Name: "parascriptural",
			Flavor: sb.Flavor{
				Name: "x-bcvarticles",
			},
		},
	}

	// OBS uses its own copyright format.
	m.Copyright = BuildCopyright(manifest, true)

	// OBS is treated as a single book.
	m.LocalizedNames = map[string]sb.LocalizedName{
		"book-obs": {
			Abbr:  map[string]string{"en": "OBS"},
			Short: map[string]string{"en": "OBS"},
			Long:  map[string]string{"en": "OBS"},
		},
	}

	lang := manifest.DublinCore.Language.Identifier

	// Determine payload source: explicit PayloadPath option, or auto-detect <lang>_tw/ in inDir.
	var twBibleDir string
	if opts.PayloadPath != "" {
		twBibleDir = filepath.Join(opts.PayloadPath, "bible")
	} else {
		twBibleDir = filepath.Join(inDir, lang+"_tw", "bible")
	}

	_, twDirErr := os.Stat(twBibleDir)
	hasPayload := twDirErr == nil

	// If payload exists, copy the TW bible/ tree to ingredients/payload/.
	if hasPayload {
		if err := copyTreeToIngredients(twBibleDir, outDir, "ingredients/payload", m); err != nil {
			return nil, fmt.Errorf("copying TW payload: %w", err)
		}
	}

	// Process the OBS TWL project(s) — normally a single twl_OBS.tsv.
	currentScope := make(map[string][]string)
	for _, project := range manifest.Projects {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		srcPath := filepath.Join(inDir, strings.TrimPrefix(project.Path, "./"))
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}
		srcFilename := filepath.Base(srcPath)

		// Strip "twl_" prefix: "twl_OBS.tsv" -> "OBS.tsv"
		destFilename := strings.TrimPrefix(srcFilename, "twl_")
		ingredientKey := "ingredients/" + destFilename

		// OBS is the book; stories map to chapters and frames to verses.
		bookCode := books.CodeFromProjectID(strings.ToLower(project.Identifier))
		scope := map[string][]string{bookCode: {}}
		currentScope[bookCode] = []string{}

		if hasPayload {
			// Copy TSV with rc:// link rewriting to ./payload/ paths.
			ing, err := copyTSVWithLinkRewrite(srcPath, outDir, ingredientKey, scope)
			if err != nil {
				return nil, fmt.Errorf("copying %s with link rewrite: %w", srcFilename, err)
			}
			m.Ingredients[ingredientKey] = ing
		} else {
			// Copy TSV as-is (no payload, no link rewriting).
			ing, err := CopyFileWithScope(srcPath, outDir, ingredientKey, scope)
			if err != nil {
				return nil, fmt.Errorf("copying %s: %w", srcFilename, err)
			}
			m.Ingredients[ingredientKey] = ing
		}
	}

	m.Type.FlavorType.CurrentScope = currentScope

	// Copy common root files (README.md, .gitignore, .gitea, .github).
	if err := CopyCommonRootFiles(inDir, outDir, m); err != nil {
		return nil, err
	}

	// Copy LICENSE.md to ingredients/.
	licIng, err := CopyLicenseIngredient(inDir, outDir)
	if err != nil {
		return nil, fmt.Errorf("copying LICENSE.md: %w", err)
	}
	m.Ingredients["ingredients/LICENSE.md"] = licIng

	return m, nil
}
