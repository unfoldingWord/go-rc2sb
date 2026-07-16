package handler

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/unfoldingWord/go-rc2sb/rc"
	"github.com/unfoldingWord/go-rc2sb/sb"
)

// obsTSVConfig holds the configuration for a specific OBS TSV variant.
type obsTSVConfig struct {
	subject      string // e.g., "TSV OBS Study Notes"
	flavorName   string // e.g., "x-obsnotes"
	abbreviation string // e.g., "OBSSN"
	tsvPrefix    string // e.g., "sn_"
}

// obsTSVHandler handles conversion for OBS TSV variants.
type obsTSVHandler struct {
	config obsTSVConfig
}

func (h *obsTSVHandler) Subject() string {
	return h.config.subject
}

func (h *obsTSVHandler) Convert(ctx context.Context, manifest *rc.Manifest, inDir, outDir string, opts Options) (*sb.Metadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m := BuildBaseMetadata(manifest, opts, h.config.abbreviation)

	// Set type
	m.Type = sb.Type{
		FlavorType: sb.FlavorType{
			Name: "peripheral",
			Flavor: sb.Flavor{
				Name: h.config.flavorName,
			},
		},
	}

	// Set copyright
	m.Copyright = BuildCopyright(manifest, false)

	// Set OBS localized names
	m.LocalizedNames = map[string]sb.LocalizedName{
		"book-obs": {
			Abbr:  map[string]string{"en": "OBS"},
			Short: map[string]string{"en": "OBS"},
			Long:  map[string]string{"en": "OBS"},
		},
	}

	// Find the TSV file from projects
	if len(manifest.Projects) == 0 {
		return nil, fmt.Errorf("no projects found in manifest for %s", h.config.subject)
	}

	project := manifest.Projects[0]
	tsvPath := filepath.Join(inDir, strings.TrimPrefix(project.Path, "./"))

	// The SB ingredient key strips the prefix (e.g., "sn_OBS.tsv" -> "OBS.tsv")
	tsvFilename := filepath.Base(tsvPath)
	sbFilename := strings.TrimPrefix(tsvFilename, h.config.tsvPrefix)
	ingredientKey := "ingredients/" + sbFilename

	// Copy TSV file
	ing, err := CopyFileAndComputeIngredient(tsvPath, outDir, ingredientKey)
	if err != nil {
		return nil, fmt.Errorf("copying TSV file: %w", err)
	}
	m.Ingredients[ingredientKey] = ing

	// Copy common root files (README.md, .gitignore, .gitea, .github)
	if err := CopyCommonRootFiles(inDir, outDir, m); err != nil {
		return nil, err
	}

	// Copy LICENSE.md
	licIng, err := CopyLicenseIngredient(inDir, outDir)
	if err != nil {
		return nil, fmt.Errorf("copying LICENSE.md: %w", err)
	}
	m.Ingredients["ingredients/LICENSE.md"] = licIng

	return m, nil
}

// NewOBSTSVHandler creates a new handler for an OBS TSV variant.
func NewOBSTSVHandler(subject, flavorName, abbreviation, tsvPrefix string) Handler {
	return &obsTSVHandler{
		config: obsTSVConfig{
			subject:      subject,
			flavorName:   flavorName,
			abbreviation: abbreviation,
			tsvPrefix:    tsvPrefix,
		},
	}
}
