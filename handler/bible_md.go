package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/unfoldingWord/go-rc2sb/books"
	"github.com/unfoldingWord/go-rc2sb/rc"
	"github.com/unfoldingWord/go-rc2sb/sb"
)

// bibleMDHandler converts the legacy Bible-based helps whose content is a tree
// of Markdown files, <book>/<chapter>/<chunk>.md, into the per-book TSVs of
// their "TSV ..." counterpart, so both convert to the same burrito.
type bibleMDHandler struct {
	tsv    bibleTSVConfig
	layout OBSMDLayout // OBSLayoutFrameNotes or OBSLayoutFrameQuestions: the notes or questions TSV columns
}

func (h *bibleMDHandler) Subject() string {
	return h.tsv.subject
}

func (h *bibleMDHandler) Convert(ctx context.Context, manifest *rc.Manifest, inDir, outDir string, opts Options) (*sb.Metadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m := BuildBaseMetadata(manifest, opts, h.tsv.abbreviation)
	m.Type = sb.Type{
		FlavorType: sb.FlavorType{
			Name:   "parascriptural",
			Flavor: sb.Flavor{Name: h.tsv.flavorName},
		},
	}
	m.Copyright = BuildCopyright(manifest, false)

	lang := manifest.DublinCore.Language.Identifier
	currentScope := make(map[string][]string)
	for _, project := range manifest.Projects {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		srcPath := filepath.Join(inDir, strings.TrimPrefix(project.Path, "./"))
		info, err := os.Stat(srcPath)
		if err != nil {
			continue
		}

		bookID := strings.ToLower(project.Identifier)
		bookCode := books.CodeFromProjectID(bookID)
		ingredientKey := "ingredients/" + bookCode + ".tsv"
		scope := map[string][]string{bookCode: {}}

		var ing sb.Ingredient
		if info.IsDir() {
			rows, err := h.buildRows(ctx, srcPath)
			if err != nil {
				return nil, fmt.Errorf("converting %s Markdown in %s: %w", h.tsv.subject, srcPath, err)
			}
			if len(rows) == 0 {
				continue
			}
			assignIDs(rows, h.layout)
			dst := filepath.Join(outDir, ingredientKey)
			if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
				return nil, fmt.Errorf("creating directory for %s: %w", dst, err)
			}
			if err := os.WriteFile(dst, renderTSV(rows, h.layout), 0644); err != nil {
				return nil, fmt.Errorf("writing %s: %w", ingredientKey, err)
			}
			if ing, err = sb.ComputeIngredientWithScope(dst, scope); err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(srcPath, ".tsv") {
			// The repo declares the Markdown subject but already holds the TSV
			if ing, err = CopyFileWithScope(srcPath, outDir, ingredientKey, scope); err != nil {
				return nil, fmt.Errorf("copying %s: %w", filepath.Base(srcPath), err)
			}
		} else {
			continue
		}
		m.Ingredients[ingredientKey] = ing
		currentScope[bookCode] = []string{}

		var usfmNames *books.LocalizedBookNames
		if opts.USFMPath != "" {
			if usfmFile := books.FindUSFMFile(opts.USFMPath, bookID); usfmFile != "" {
				usfmNames = books.ParseUSFMBookNames(usfmFile)
			}
		}
		if key, localizedName := books.LocalizedNameEntryWithNames(bookID, lang, project.Title, usfmNames); key != "" {
			m.LocalizedNames[key] = localizedName
		}
	}
	if len(currentScope) == 0 {
		return nil, fmt.Errorf("no %s content found in any project", h.tsv.subject)
	}
	m.Type.FlavorType.CurrentScope = currentScope

	if err := CopyCommonRootFiles(inDir, outDir, m); err != nil {
		return nil, err
	}
	licIng, err := CopyLicenseIngredient(inDir, outDir)
	if err != nil {
		return nil, fmt.Errorf("copying LICENSE.md: %w", err)
	}
	m.Ingredients["ingredients/LICENSE.md"] = licIng

	return m, nil
}

// buildRows reads a book directory's chunk files in book order: the front
// matter, then each chapter's intro and chunks.
func (h *bibleMDHandler) buildRows(ctx context.Context, bookDir string) ([]obsRow, error) {
	files, err := bibleChunkFiles(bookDir)
	if err != nil {
		return nil, err
	}

	var rows []obsRow
	for _, f := range files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		data, err := os.ReadFile(f.path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", f.path, err)
		}
		text := strings.TrimSpace(string(data))
		if text == "" {
			continue
		}
		reference := f.chapter + ":" + f.verse

		if h.layout == OBSLayoutFrameQuestions {
			for _, block := range parseMDBlocks(text) {
				rows = append(rows, obsRow{Reference: reference, Question: block.heading, Response: block.body})
			}
			continue
		}

		blocks := parseMDBlocks(text)
		if f.verse == "intro" || len(blocks) == 0 {
			// An intro is one note made of its own headings, as in the TSV notes
			rows = append(rows, obsRow{Reference: reference, Occurrence: "0", Note: text})
			continue
		}
		for _, block := range blocks {
			if isImportantTermsBlock(block) {
				continue
			}
			note, support := extractSupportReference(block.body)
			rows = append(rows, obsRow{
				Reference:        reference,
				SupportReference: support,
				Quote:            block.heading,
				Occurrence:       "1",
				Note:             note,
			})
		}
	}
	return rows, nil
}

// bibleChunkFile is one Markdown file of a book directory, with the reference
// parts it stands for: chapter "front" or a number, verse "intro" or a number.
type bibleChunkFile struct {
	path    string
	chapter string
	verse   string
	order   [2]int
}

// bibleChunkFiles lists a book directory's chunk files in book order. Chapter
// directories are numbers, with "front" or "00" for the book's front matter;
// files are a verse number or "intro". Anything else is not note content.
func bibleChunkFiles(bookDir string) ([]bibleChunkFile, error) {
	chapterDirs, err := os.ReadDir(bookDir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", bookDir, err)
	}

	var files []bibleChunkFile
	for _, cd := range chapterDirs {
		if !cd.IsDir() {
			continue
		}
		chapter, chapterOrder := "front", 0
		if cd.Name() != "front" {
			n, err := strconv.Atoi(cd.Name())
			if err != nil || n < 0 {
				continue
			}
			if n > 0 {
				chapter, chapterOrder = strconv.Itoa(n), n
			}
		}

		entries, err := os.ReadDir(filepath.Join(bookDir, cd.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", filepath.Join(bookDir, cd.Name()), err)
		}
		for _, e := range entries {
			name, ok := strings.CutSuffix(e.Name(), ".md")
			if e.IsDir() || !ok {
				continue
			}
			verse, verseOrder := "intro", 0
			if name != "intro" {
				n, err := strconv.Atoi(name)
				if err != nil || n < 1 || chapter == "front" {
					continue
				}
				verse, verseOrder = strconv.Itoa(n), n
			}
			files = append(files, bibleChunkFile{
				path:    filepath.Join(bookDir, cd.Name(), e.Name()),
				chapter: chapter,
				verse:   verse,
				order:   [2]int{chapterOrder, verseOrder},
			})
		}
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].order[0] != files[j].order[0] {
			return files[i].order[0] < files[j].order[0]
		}
		return files[i].order[1] < files[j].order[1]
	})
	return files, nil
}

// NewBibleMarkdownHandler creates a handler for a legacy Markdown Bible helps
// subject, converting it to the TSV form of the given flavor. notes selects the
// notes columns (with SupportReference and Note) over the questions columns.
func NewBibleMarkdownHandler(subject, flavorName, abbreviation string, notes bool) Handler {
	layout := OBSLayoutFrameQuestions
	if notes {
		layout = OBSLayoutFrameNotes
	}
	return &bibleMDHandler{
		tsv:    bibleTSVConfig{subject: subject, flavorName: flavorName, abbreviation: abbreviation},
		layout: layout,
	}
}
