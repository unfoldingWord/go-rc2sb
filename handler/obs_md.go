package handler

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/unfoldingWord/go-rc2sb/rc"
	"github.com/unfoldingWord/go-rc2sb/sb"
)

// OBSMDLayout identifies the Markdown layout used by a legacy (pre-TSV) OBS resource.
type OBSMDLayout int

const (
	// OBSLayoutFrameQuestions is content/<story>/<frame>.md where each level-1
	// heading is a question and the body that follows it is the answer.
	// Used by "OBS Translation Questions".
	OBSLayoutFrameQuestions OBSMDLayout = iota

	// OBSLayoutFrameNotes is content/<story>/<frame>.md where each level-1
	// heading is the quoted phrase and the body that follows it is the note.
	// Used by "OBS Translation Notes" and "OBS Study Notes".
	OBSLayoutFrameNotes

	// OBSLayoutStoryQuestions is content/<story>.md where level-2 headings
	// divide the file into sections of numbered question/answer pairs followed
	// by a closing summary. Used by "OBS Study Questions".
	OBSLayoutStoryQuestions
)

// obsStoryFrames holds the canonical number of frames in each of the 50 Open
// Bible Stories, indexed by story number minus one. The frame structure is
// fixed across all OBS translations; these counts were taken from
// unfoldingWord/en_obs and match the story-wide references used by the
// en sq_OBS.tsv summary rows for all 50 stories.
//
// Story-questions sources reference the story as a whole rather than a
// specific frame, so this table supplies the "<story>:1-<lastFrame>" reference
// for summary rows and for questions whose answer links to no frame.
var obsStoryFrames = [50]int{
	16, 12, 16, 9, 10, 7, 10, 15, 15, 12,
	8, 14, 15, 15, 13, 18, 14, 13, 18, 13,
	15, 7, 10, 9, 8, 10, 11, 10, 9, 9,
	8, 16, 9, 10, 13, 7, 11, 15, 12, 9,
	8, 11, 13, 9, 13, 10, 14, 14, 18, 17,
}

// obsMDConfig holds the configuration for a specific legacy OBS Markdown variant.
type obsMDConfig struct {
	subject      string      // e.g., "OBS Translation Questions"
	flavorName   string      // e.g., "x-obsquestions"
	abbreviation string      // e.g., "OBSTQ"
	tsvPrefix    string      // e.g., "tq_"
	layout       OBSMDLayout // source Markdown layout
}

// obsMDHandler converts a legacy OBS resource stored as a tree of Markdown
// files into the single OBS.tsv ingredient produced by the TSV OBS handlers.
// The resulting burrito is structurally identical to the one the corresponding
// "TSV OBS ..." subject produces: the same flavor, abbreviation, localized
// names, and the same ingredients/OBS.tsv plus ingredients/LICENSE.md.
type obsMDHandler struct {
	config obsMDConfig
}

func (h *obsMDHandler) Subject() string {
	return h.config.subject
}

func (h *obsMDHandler) Convert(ctx context.Context, manifest *rc.Manifest, inDir, outDir string, opts Options) (*sb.Metadata, error) {
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

	// Some repos declare a legacy Markdown subject but have already been
	// converted to TSV upstream. Copy the existing TSV through unchanged
	// rather than regenerating it from Markdown that may no longer be present.
	if tsvPath := h.findExistingTSV(manifest, inDir); tsvPath != "" {
		sbFilename := strings.TrimPrefix(filepath.Base(tsvPath), h.config.tsvPrefix)
		ingredientKey := "ingredients/" + sbFilename
		ing, err := CopyFileAndComputeIngredient(tsvPath, outDir, ingredientKey)
		if err != nil {
			return nil, fmt.Errorf("copying TSV file: %w", err)
		}
		m.Ingredients[ingredientKey] = ing
	} else {
		contentDir := h.contentDir(manifest, inDir)
		rows, err := h.buildRows(ctx, contentDir)
		if err != nil {
			return nil, fmt.Errorf("converting %s Markdown in %s: %w", h.config.subject, contentDir, err)
		}
		if len(rows) == 0 {
			return nil, fmt.Errorf("no %s Markdown content found in %s", h.config.subject, contentDir)
		}
		assignIDs(rows, h.config.layout)

		const ingredientKey = "ingredients/OBS.tsv"
		dst := filepath.Join(outDir, ingredientKey)
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return nil, fmt.Errorf("creating directory for %s: %w", dst, err)
		}
		if err := os.WriteFile(dst, renderTSV(rows, h.config.layout), 0644); err != nil {
			return nil, fmt.Errorf("writing %s: %w", ingredientKey, err)
		}
		ing, err := sb.ComputeIngredient(dst)
		if err != nil {
			return nil, err
		}
		m.Ingredients[ingredientKey] = ing
	}

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

// contentDir resolves the directory holding the Markdown content, taking it
// from the first project's path and falling back to "content".
func (h *obsMDHandler) contentDir(manifest *rc.Manifest, inDir string) string {
	if len(manifest.Projects) > 0 {
		if p := strings.TrimPrefix(manifest.Projects[0].Path, "./"); p != "" {
			return filepath.Join(inDir, p)
		}
	}
	return filepath.Join(inDir, "content")
}

// findExistingTSV returns the path of a TSV already present in the RC repo, or
// "" when the repo really does hold Markdown. It checks the first project's
// path, then the conventional "<prefix>OBS.tsv" at the repo root (some repos
// point their project path at a "./content" directory that no longer exists).
func (h *obsMDHandler) findExistingTSV(manifest *rc.Manifest, inDir string) string {
	var candidates []string
	if len(manifest.Projects) > 0 {
		if p := strings.TrimPrefix(manifest.Projects[0].Path, "./"); strings.HasSuffix(p, ".tsv") {
			candidates = append(candidates, filepath.Join(inDir, p))
		}
	}
	candidates = append(candidates, filepath.Join(inDir, h.config.tsvPrefix+"OBS.tsv"))

	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return c
		}
	}
	return ""
}

// buildRows parses the Markdown content into output rows, in story then frame order.
func (h *obsMDHandler) buildRows(ctx context.Context, contentDir string) ([]obsRow, error) {
	if h.config.layout == OBSLayoutStoryQuestions {
		return buildStoryQuestionRows(ctx, contentDir)
	}
	return h.buildFrameRows(ctx, contentDir)
}

// buildFrameRows handles the content/<story>/<frame>.md layouts, where every
// level-1 heading starts a new row.
func (h *obsMDHandler) buildFrameRows(ctx context.Context, contentDir string) ([]obsRow, error) {
	files, err := frameFiles(contentDir)
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
		reference := fmt.Sprintf("%d:%d", f.story, f.frame)

		for _, block := range parseMDBlocks(string(data)) {
			if h.config.layout == OBSLayoutFrameQuestions {
				rows = append(rows, obsRow{
					Reference: reference,
					Question:  block.heading,
					Response:  block.body,
				})
				continue
			}

			// Notes: an "Important Terms" block is a bare list of Translation
			// Words links with no note text. The TSV notes format has no
			// column for it, so it is dropped.
			if isImportantTermsBlock(block) {
				continue
			}
			note, support := extractSupportReference(block.body)
			tags := ""
			if f.frame == 0 {
				// <story>/00.md holds the note on the story title.
				tags = "title"
			}
			rows = append(rows, obsRow{
				Reference:        reference,
				Tags:             tags,
				SupportReference: support,
				Quote:            block.heading,
				Occurrence:       "1",
				Note:             note,
			})
		}
	}
	return rows, nil
}

// buildStoryQuestionRows handles the content/<story>.md layout used by
// "OBS Study Questions": content/00.md is the front-matter intro, and each
// story file holds level-2 sections of numbered question/answer pairs
// followed by a summary section.
func buildStoryQuestionRows(ctx context.Context, contentDir string) ([]obsRow, error) {
	entries, err := os.ReadDir(contentDir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", contentDir, err)
	}

	stories := make([]int, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if n, ok := numericName(e.Name(), ".md"); ok {
			stories = append(stories, n)
		}
	}
	sort.Ints(stories)

	var rows []obsRow
	for _, story := range stories {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		path := filepath.Join(contentDir, fmt.Sprintf("%02d.md", story))
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
		text := string(data)

		// content/00.md is the front matter, carried through whole.
		if story == 0 {
			if body := strings.TrimSpace(text); body != "" {
				rows = append(rows, obsRow{
					Reference: "front",
					Tags:      "intro",
					Response:  body,
				})
			}
			continue
		}

		storyRef := storyReference(story)
		questionSections := 0
		for _, section := range splitSections(text) {
			items := splitNumberedItems(section)
			if len(items) == 0 {
				// A section with no numbered items is the closing summary.
				if body := strings.TrimSpace(section); body != "" {
					rows = append(rows, obsRow{
						Reference: storyRef,
						Tags:      "summary",
						Response:  body,
					})
				}
				continue
			}

			// The first question-bearing section asks what the story says;
			// later ones ask what it means for the reader.
			tags := "application"
			if questionSections == 0 {
				tags = "meaning"
			}
			questionSections++

			for _, item := range items {
				question, answer := splitParagraph(item)
				rows = append(rows, obsRow{
					Reference: frameRangeReference(story, answer),
					Tags:      tags,
					Question:  question,
					Response:  answer,
				})
			}
		}
	}
	return rows, nil
}

// obsRow is a single row of the generated TSV. Which fields are emitted
// depends on the layout; see obsRow.cells.
type obsRow struct {
	Reference        string
	ID               string
	Tags             string
	SupportReference string
	Quote            string
	Occurrence       string
	Question         string
	Response         string
	Note             string
}

// cells returns the row's columns in output order for the given layout.
func (r obsRow) cells(layout OBSMDLayout) []string {
	if layout == OBSLayoutFrameNotes {
		return []string{r.Reference, r.ID, r.Tags, r.SupportReference, r.Quote, r.Occurrence, r.Note}
	}
	return []string{r.Reference, r.ID, r.Tags, r.Quote, r.Occurrence, r.Question, r.Response}
}

// tsvHeader returns the column names for the given layout. These match the
// headers of the TSV that the corresponding "TSV OBS ..." subject ships.
func tsvHeader(layout OBSMDLayout) []string {
	if layout == OBSLayoutFrameNotes {
		return []string{"Reference", "ID", "Tags", "SupportReference", "Quote", "Occurrence", "Note"}
	}
	return []string{"Reference", "ID", "Tags", "Quote", "Occurrence", "Question", "Response"}
}

// renderTSV serializes the rows, escaping cell content so each row stays on a
// single line.
func renderTSV(rows []obsRow, layout OBSMDLayout) []byte {
	var b strings.Builder
	b.WriteString(strings.Join(tsvHeader(layout), "\t"))
	b.WriteByte('\n')
	for _, r := range rows {
		cells := r.cells(layout)
		for i, c := range cells {
			if i > 0 {
				b.WriteByte('\t')
			}
			b.WriteString(escapeTSVCell(c))
		}
		b.WriteByte('\n')
	}
	return []byte(b.String())
}

// escapeTSVCell folds a cell's line breaks into the literal two-character
// sequence \n and its tabs into spaces, the convention the unfoldingWord TSV
// resources use for multi-line cell content.
func escapeTSVCell(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, "\n", `\n`)
	return strings.ReplaceAll(s, "\t", " ")
}

// tsvIDFirst and tsvIDRest are the character sets unfoldingWord TSV IDs use:
// four characters, the first a letter so the ID is never read as a number.
const (
	tsvIDFirst = "abcdefghijklmnopqrstuvwxyz"
	tsvIDRest  = "abcdefghijklmnopqrstuvwxyz0123456789"
	tsvIDLen   = 4
)

// assignIDs gives every row a 4-character ID derived from its own content, so
// converting the same source twice produces a byte-identical TSV (and so a
// stable ingredient checksum). Collisions are broken with an incrementing salt.
func assignIDs(rows []obsRow, layout OBSMDLayout) {
	used := make(map[string]bool, len(rows))
	for i := range rows {
		seed := strings.Join(rows[i].cells(layout), "\x00")
		for salt := 0; ; salt++ {
			id := deriveTSVID(seed, salt)
			if !used[id] {
				used[id] = true
				rows[i].ID = id
				break
			}
		}
	}
}

// deriveTSVID hashes seed into a 4-character TSV ID.
func deriveTSVID(seed string, salt int) string {
	sum := md5.Sum([]byte(seed + "\x00" + strconv.Itoa(salt)))
	n := binary.BigEndian.Uint64(sum[:8])

	id := make([]byte, tsvIDLen)
	id[0] = tsvIDFirst[n%uint64(len(tsvIDFirst))]
	n /= uint64(len(tsvIDFirst))
	for i := 1; i < tsvIDLen; i++ {
		id[i] = tsvIDRest[n%uint64(len(tsvIDRest))]
		n /= uint64(len(tsvIDRest))
	}
	return string(id)
}

// mdBlock is a level-1 Markdown heading together with the body that follows it.
type mdBlock struct {
	heading string
	body    string
}

var (
	mdHeading1     = regexp.MustCompile(`(?m)^#[ \t]+(.*?)[ \t]*$`)
	mdHeading2     = regexp.MustCompile(`(?m)^##[ \t]+.*$`)
	mdNumberedItem = regexp.MustCompile(`(?m)^[0-9]+\.[ \t]+`)
	// twBulletLine matches a list item that is nothing but a Translation Words link.
	// Some resources escape the link's brackets as \[\[ ... \]\].
	twBulletLine = regexp.MustCompile(`^[-*+][ \t]*\\?\[\\?\[rc://[^\]\\]*/tw/[^\]\\]*\\?\]\\?\]$`)
	// taLinkTrailing matches a trailing parenthetical whose only substance is a
	// Translation Academy link, e.g. "(See: [[rc://*/ta/man/translate/figs-idiom]])",
	// allowing for the escaped \[\[ ... \]\] spelling.
	taLinkTrailing = regexp.MustCompile(`[ \t]*\([^()]*\\?\[\\?\[(rc://[^\]\\]*?/ta/man/[^\]\\]*?)\\?\]\\?\][^()]*\)$`)
	// rcLanguage matches the language segment of an rc:// link.
	rcLanguage = regexp.MustCompile(`^rc://[^/]+/`)
	// frameLink matches a story frame link such as "[01:09](01/09)".
	frameLink = regexp.MustCompile(`\[([0-9]+):([0-9]+)\]\([^)]*\)`)
)

// parseMDBlocks splits Markdown into its level-1 heading blocks. Text before
// the first heading is ignored, as these files always open with one.
func parseMDBlocks(text string) []mdBlock {
	locs := mdHeading1.FindAllStringSubmatchIndex(text, -1)
	blocks := make([]mdBlock, 0, len(locs))
	for i, loc := range locs {
		end := len(text)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		blocks = append(blocks, mdBlock{
			heading: strings.TrimSpace(text[loc[2]:loc[3]]),
			body:    strings.TrimSpace(text[loc[1]:end]),
		})
	}
	return blocks
}

// splitSections splits a story file on its level-2 headings, returning the
// section bodies in order. The heading text itself is dropped: it is written in
// the resource's own language, so section meaning is taken from position
// instead. Text before the first heading (the story title) is dropped too.
func splitSections(text string) []string {
	locs := mdHeading2.FindAllStringIndex(text, -1)
	sections := make([]string, 0, len(locs))
	for i, loc := range locs {
		end := len(text)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		sections = append(sections, text[loc[1]:end])
	}
	return sections
}

// splitNumberedItems splits a section body on its numbered list markers,
// returning one entry per question. Text before the first marker is dropped.
func splitNumberedItems(section string) []string {
	locs := mdNumberedItem.FindAllStringIndex(section, -1)
	items := make([]string, 0, len(locs))
	for i, loc := range locs {
		end := len(section)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		if item := strings.TrimSpace(section[loc[1]:end]); item != "" {
			items = append(items, item)
		}
	}
	return items
}

// splitParagraph splits an item into its first paragraph (the question) and
// everything after it (the answer).
func splitParagraph(item string) (first, rest string) {
	parts := strings.SplitN(strings.TrimSpace(item), "\n\n", 2)
	first = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		rest = strings.TrimSpace(parts[1])
	}
	return first, rest
}

// isImportantTermsBlock reports whether a block is an "Important Terms" list —
// a heading followed by nothing but Translation Words links. The heading is
// written in the resource's own language, so the body is what identifies it.
func isImportantTermsBlock(b mdBlock) bool {
	bullets := 0
	for _, line := range strings.Split(b.body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !twBulletLine.MatchString(line) {
			return false
		}
		bullets++
	}
	return bullets > 0
}

// extractSupportReference lifts a trailing Translation Academy link out of a
// note body into the SupportReference column, matching how the TSV notes
// resources store it, and normalizes the link's language segment to "*".
// A body whose TA link is not a trailing parenthetical is left untouched, since
// removing it would cut into the sentence around it.
func extractSupportReference(body string) (note, support string) {
	trimmed := strings.TrimRight(body, " \t\n")
	loc := taLinkTrailing.FindStringSubmatchIndex(trimmed)
	if loc == nil {
		return body, ""
	}
	link := trimmed[loc[2]:loc[3]]
	return strings.TrimSpace(trimmed[:loc[0]]), rcLanguage.ReplaceAllString(link, "rc://*/")
}

// storyReference returns the reference covering a whole story, e.g. "1:1-16".
func storyReference(story int) string {
	if story >= 1 && story <= len(obsStoryFrames) {
		return fmt.Sprintf("%d:1-%d", story, obsStoryFrames[story-1])
	}
	return strconv.Itoa(story)
}

// frameRangeReference derives a question's reference from the frame links in
// its answer: the span of the frames it cites within its own story. An answer
// that cites no frame of its own story is about the story as a whole.
func frameRangeReference(story int, answer string) string {
	low, high := 0, 0
	for _, m := range frameLink.FindAllStringSubmatch(answer, -1) {
		linkStory, err := strconv.Atoi(m[1])
		if err != nil || linkStory != story {
			continue
		}
		frame, err := strconv.Atoi(m[2])
		if err != nil {
			continue
		}
		if low == 0 || frame < low {
			low = frame
		}
		if frame > high {
			high = frame
		}
	}
	switch {
	case low == 0:
		return storyReference(story)
	case low == high:
		return fmt.Sprintf("%d:%d", story, low)
	default:
		return fmt.Sprintf("%d:%d-%d", story, low, high)
	}
}

// frameFile is one content/<story>/<frame>.md file.
type frameFile struct {
	story int
	frame int
	path  string
}

// frameFiles lists the content/<story>/<frame>.md files in story then frame
// order. Directories and files that are not purely numeric are skipped, and
// frame numbers are sparse in real repos.
func frameFiles(contentDir string) ([]frameFile, error) {
	storyDirs, err := os.ReadDir(contentDir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", contentDir, err)
	}

	var files []frameFile
	for _, sd := range storyDirs {
		if !sd.IsDir() {
			continue
		}
		story, ok := numericName(sd.Name(), "")
		if !ok {
			continue
		}
		storyPath := filepath.Join(contentDir, sd.Name())
		frames, err := os.ReadDir(storyPath)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", storyPath, err)
		}
		for _, fe := range frames {
			if fe.IsDir() {
				continue
			}
			frame, ok := numericName(fe.Name(), ".md")
			if !ok {
				continue
			}
			files = append(files, frameFile{
				story: story,
				frame: frame,
				path:  filepath.Join(storyPath, fe.Name()),
			})
		}
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].story != files[j].story {
			return files[i].story < files[j].story
		}
		return files[i].frame < files[j].frame
	})
	return files, nil
}

// numericName parses a zero-padded numeric file or directory name, after
// removing the given suffix. It reports false for any other name.
func numericName(name, suffix string) (int, bool) {
	if suffix != "" {
		if !strings.HasSuffix(name, suffix) {
			return 0, false
		}
		name = strings.TrimSuffix(name, suffix)
	}
	if name == "" {
		return 0, false
	}
	for _, r := range name {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(name)
	if err != nil {
		return 0, false
	}
	return n, true
}

// NewOBSMarkdownHandler creates a handler for a legacy OBS subject whose
// content is a tree of Markdown files rather than a single TSV. The output
// matches what the corresponding "TSV OBS ..." subject produces.
func NewOBSMarkdownHandler(subject, flavorName, abbreviation, tsvPrefix string, layout OBSMDLayout) Handler {
	return &obsMDHandler{
		config: obsMDConfig{
			subject:      subject,
			flavorName:   flavorName,
			abbreviation: abbreviation,
			tsvPrefix:    tsvPrefix,
			layout:       layout,
		},
	}
}
