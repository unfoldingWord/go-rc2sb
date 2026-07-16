package handler

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/unfoldingWord/go-rc2sb/books"
	"github.com/unfoldingWord/go-rc2sb/rc"
	"github.com/unfoldingWord/go-rc2sb/sb"
)

// twLinkRegexp parses RC links like "rc://*/tw/dict/bible/other/creation"
var twLinkRegexp = regexp.MustCompile(`rc://[^/]*/tw/dict/bible/([^/]+)/([^/\t]+)`)

// twLinkReplaceRegexp matches a TWLink column value at end of a TSV line for replacement.
// Matches: \trc://<anything>/tw/dict/bible/<category>/<article> at end of line
var twLinkReplaceRegexp = regexp.MustCompile(`\trc://[^/]+/tw/dict/bible/([^\t]+)$`)

// NewTWLHandler creates a new TSV Translation Words Links handler.
func NewTWLHandler() Handler {
	return &twlHandler{}
}

type twlHandler struct{}

func (h *twlHandler) Subject() string {
	return "TSV Translation Words Links"
}

func (h *twlHandler) Convert(ctx context.Context, manifest *rc.Manifest, inDir, outDir string, opts Options) (*sb.Metadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m := BuildBaseMetadata(manifest, opts, "TW")

	// Set type - parascriptural/x-bcvarticles
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

	lang := manifest.DublinCore.Language.Identifier

	// Determine payload source: explicit PayloadPath option, or auto-detect <lang>_tw/ in inDir
	var twBibleDir string
	if opts.PayloadPath != "" {
		twBibleDir = filepath.Join(opts.PayloadPath, "bible")
	} else {
		twBibleDir = filepath.Join(inDir, lang+"_tw", "bible")
	}

	_, twDirErr := os.Stat(twBibleDir)
	hasPayload := twDirErr == nil

	// If payload exists, copy the TW bible/ tree to ingredients/payload/
	if hasPayload {
		if err := copyTreeToIngredients(twBibleDir, outDir, "ingredients/payload", m); err != nil {
			return nil, fmt.Errorf("copying TW payload: %w", err)
		}
	}

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

		// Strip "twl_" prefix: "twl_GEN.tsv" -> "GEN.tsv"
		destFilename := strings.TrimPrefix(srcFilename, "twl_")
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

		if hasPayload {
			// Copy TSV file with rc:// link rewriting, then compute ingredient
			ing, err := copyTSVWithLinkRewrite(srcPath, outDir, ingredientKey, scope)
			if err != nil {
				return nil, fmt.Errorf("copying %s with link rewrite: %w", srcFilename, err)
			}
			m.Ingredients[ingredientKey] = ing
		} else {
			// Copy TSV file as-is (no payload, no link rewriting)
			ing, err := CopyFileWithScope(srcPath, outDir, ingredientKey, scope)
			if err != nil {
				return nil, fmt.Errorf("copying %s: %w", srcFilename, err)
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

	// Copy LICENSE.md to ingredients/
	licIng, err := CopyLicenseIngredient(inDir, outDir)
	if err != nil {
		return nil, fmt.Errorf("copying LICENSE.md: %w", err)
	}
	m.Ingredients["ingredients/LICENSE.md"] = licIng

	return m, nil
}

// copyTSVWithLinkRewrite copies a TSV file while replacing rc:// TWLink references
// with relative payload paths (e.g., rc://*/tw/dict/bible/names/peter -> ./payload/names/peter.md).
// The ingredient checksum/size is computed after the rewrite.
func copyTSVWithLinkRewrite(srcPath, outDir, ingredientKey string, scope map[string][]string) (sb.Ingredient, error) {
	// Read the source file
	inFile, err := os.Open(srcPath)
	if err != nil {
		return sb.Ingredient{}, fmt.Errorf("opening %s: %w", srcPath, err)
	}
	defer inFile.Close()

	// Create the destination file
	dstPath := filepath.Join(outDir, ingredientKey)
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return sb.Ingredient{}, fmt.Errorf("creating directory for %s: %w", dstPath, err)
	}

	outFile, err := os.Create(dstPath)
	if err != nil {
		return sb.Ingredient{}, fmt.Errorf("creating %s: %w", dstPath, err)
	}
	defer outFile.Close()

	scanner := bufio.NewScanner(inFile)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer for large lines
	writer := bufio.NewWriter(outFile)

	first := true
	for scanner.Scan() {
		line := scanner.Text()

		if !first {
			if _, err := writer.WriteString("\n"); err != nil {
				return sb.Ingredient{}, err
			}
		}
		first = false

		// Replace rc:// links in TWLink column with ./payload/ paths
		rewritten := twLinkReplaceRegexp.ReplaceAllString(line, "\t./payload/$1.md")

		if _, err := writer.WriteString(rewritten); err != nil {
			return sb.Ingredient{}, err
		}
	}

	if err := scanner.Err(); err != nil {
		return sb.Ingredient{}, fmt.Errorf("reading %s: %w", srcPath, err)
	}

	// Write trailing newline if original file had one
	srcInfo, err := os.Stat(srcPath)
	if err == nil && srcInfo.Size() > 0 {
		// Check if original file ends with newline
		f, err := os.Open(srcPath)
		if err == nil {
			buf := make([]byte, 1)
			f.Seek(srcInfo.Size()-1, 0)
			f.Read(buf)
			f.Close()
			if buf[0] == '\n' {
				writer.WriteString("\n")
			}
		}
	}

	if err := writer.Flush(); err != nil {
		return sb.Ingredient{}, err
	}
	if err := outFile.Close(); err != nil {
		return sb.Ingredient{}, err
	}

	// Compute ingredient from the rewritten file
	return sb.ComputeIngredientWithScope(dstPath, scope)
}
