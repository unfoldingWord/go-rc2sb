package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/unfoldingWord/go-rc2sb/rc"
	"github.com/unfoldingWord/go-rc2sb/sb"
)

// NewTAHandler creates a new Translation Academy handler.
func NewTAHandler() Handler {
	return &taHandler{}
}

type taHandler struct{}

func (h *taHandler) Subject() string {
	return "Translation Academy"
}

func (h *taHandler) Convert(ctx context.Context, manifest *rc.Manifest, inDir, outDir string, opts Options) (*sb.Metadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m := BuildBaseMetadata(manifest, opts, "TA")

	// Set type - peripheral/x-peripheralArticles
	m.Type = sb.Type{
		FlavorType: sb.FlavorType{
			Name: "peripheral",
			Flavor: sb.Flavor{
				Name: "x-peripheralArticles",
			},
		},
	}

	m.Copyright = BuildCopyright(manifest, false)
	m.LocalizedNames = map[string]sb.LocalizedName{}

	// Copy common root files (README.md, .gitignore, .gitea, .github)
	if err := CopyCommonRootFiles(inDir, outDir, m); err != nil {
		return nil, err
	}

	// Copy LICENSE.md to root (uses embedded default if RC doesn't have one).
	if err := CopyLicenseToRoot(inDir, outDir); err != nil {
		return nil, fmt.Errorf("copying root LICENSE.md: %w", err)
	}

	// Copy each project directory to ingredients/
	// Projects are: intro, process, translate, checking
	for _, project := range manifest.Projects {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		projectDir := filepath.Join(inDir, project.Identifier)
		if _, err := os.Stat(projectDir); os.IsNotExist(err) {
			continue
		}

		destPrefix := "ingredients/" + project.Identifier
		if err := copyTreeToIngredients(projectDir, outDir, destPrefix, m); err != nil {
			return nil, fmt.Errorf("copying project %s: %w", project.Identifier, err)
		}
	}

	// Copy LICENSE.md to ingredients/LICENSE.md (uses embedded default if RC doesn't have one).
	licIng, err := CopyLicenseIngredient(inDir, outDir)
	if err != nil {
		return nil, fmt.Errorf("copying ingredients/LICENSE.md: %w", err)
	}
	m.Ingredients["ingredients/LICENSE.md"] = licIng

	return m, nil
}
