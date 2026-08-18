package handler_test

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/unfoldingWord/go-rc2sb/handler"
	"github.com/unfoldingWord/go-rc2sb/rc"
)

// writeOBSMarkdownManifest creates a manifest for a legacy OBS Markdown repo
// whose content lives in ./content.
func writeOBSMarkdownManifest(t *testing.T, inDir, subject, identifier string) *rc.Manifest {
	t.Helper()

	m := &rc.Manifest{}
	m.DublinCore.Subject = subject
	m.DublinCore.Identifier = identifier
	m.DublinCore.Title = "Test " + subject
	m.DublinCore.Publisher = "unfoldingWord"
	m.DublinCore.Rights = "CC BY-SA 4.0"
	m.DublinCore.Issued = "2024-01-15"
	m.DublinCore.Version = "1"
	m.DublinCore.Language.Identifier = "fr"
	m.DublinCore.Language.Title = "Français"
	m.DublinCore.Language.Direction = "ltr"
	m.Projects = []rc.Project{{Identifier: "obs", Path: "./content"}}
	return m
}

// writeContentFile writes a file under inDir, creating parent directories.
func writeContentFile(t *testing.T, inDir, relPath, content string) {
	t.Helper()

	path := filepath.Join(inDir, relPath)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// readTSVRows returns the header and data rows of the generated OBS.tsv.
func readTSVRows(t *testing.T, outDir string) ([]string, [][]string) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(outDir, "ingredients", "OBS.tsv"))
	if err != nil {
		t.Fatalf("reading generated OBS.tsv: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(lines) == 0 {
		t.Fatal("generated OBS.tsv is empty")
	}

	header := strings.Split(lines[0], "\t")
	rows := make([][]string, 0, len(lines)-1)
	for _, line := range lines[1:] {
		rows = append(rows, strings.Split(line, "\t"))
	}
	return header, rows
}

// convertOBSMarkdown runs the handler for subject against inDir.
func convertOBSMarkdown(t *testing.T, subject string, m *rc.Manifest, inDir, outDir string) {
	t.Helper()

	h, err := handler.Lookup(subject)
	if err != nil {
		t.Fatalf("Lookup(%q) failed: %v", subject, err)
	}
	if _, err := h.Convert(context.Background(), m, inDir, outDir, handler.Options{}); err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
}

// --- OBS Translation Questions (content/<story>/<frame>.md) ---

func TestOBSMarkdownQuestions_ProducesTSVRows(t *testing.T) {
	inDir, outDir := t.TempDir(), t.TempDir()
	m := writeOBSMarkdownManifest(t, inDir, "OBS Translation Questions", "obs-tq")

	writeContentFile(t, inDir, "content/01/01.md", `# Who created everything?

God created everything.

# How long did it take?

Six days.
`)
	// Frame numbers are sparse in real repos, and stories sort numerically.
	writeContentFile(t, inDir, "content/02/04.md", "# What happened next?\n\nThey hid.\n")

	convertOBSMarkdown(t, "OBS Translation Questions", m, inDir, outDir)

	header, rows := readTSVRows(t, outDir)
	wantHeader := []string{"Reference", "ID", "Tags", "Quote", "Occurrence", "Question", "Response"}
	if strings.Join(header, "\t") != strings.Join(wantHeader, "\t") {
		t.Errorf("header = %v; want %v", header, wantHeader)
	}

	if len(rows) != 3 {
		t.Fatalf("got %d rows; want 3: %v", len(rows), rows)
	}

	// Reference strips the zero padding; Tags, Quote and Occurrence stay empty.
	wantRefs := []string{"1:1", "1:1", "2:4"}
	for i, want := range wantRefs {
		if rows[i][0] != want {
			t.Errorf("row %d Reference = %q; want %q", i, rows[i][0], want)
		}
		for _, col := range []int{2, 3, 4} {
			if rows[i][col] != "" {
				t.Errorf("row %d column %d = %q; want empty", i, col, rows[i][col])
			}
		}
	}

	if rows[0][5] != "Who created everything?" || rows[0][6] != "God created everything." {
		t.Errorf("row 0 question/response = %q / %q", rows[0][5], rows[0][6])
	}
	if rows[2][5] != "What happened next?" || rows[2][6] != "They hid." {
		t.Errorf("row 2 question/response = %q / %q", rows[2][5], rows[2][6])
	}
}

// --- OBS Translation Notes / OBS Study Notes (content/<story>/<frame>.md) ---

func TestOBSMarkdownNotes_TitleTagQuoteAndOccurrence(t *testing.T) {
	inDir, outDir := t.TempDir(), t.TempDir()
	m := writeOBSMarkdownManifest(t, inDir, "OBS Translation Notes", "obs-tn")

	// <story>/00.md holds the note on the story title.
	writeContentFile(t, inDir, "content/01/00.md", "# The Creation\n\nThis title can also be translated as ...\n")
	writeContentFile(t, inDir, "content/01/01.md", "# the beginning\n\nBefore anything existed except God.\n")

	convertOBSMarkdown(t, "OBS Translation Notes", m, inDir, outDir)

	header, rows := readTSVRows(t, outDir)
	wantHeader := []string{"Reference", "ID", "Tags", "SupportReference", "Quote", "Occurrence", "Note"}
	if strings.Join(header, "\t") != strings.Join(wantHeader, "\t") {
		t.Errorf("header = %v; want %v", header, wantHeader)
	}

	if len(rows) != 2 {
		t.Fatalf("got %d rows; want 2: %v", len(rows), rows)
	}

	if rows[0][0] != "1:0" || rows[0][2] != "title" {
		t.Errorf("title row = reference %q tags %q; want \"1:0\" / \"title\"", rows[0][0], rows[0][2])
	}
	if rows[0][4] != "The Creation" {
		t.Errorf("title row Quote = %q; want %q", rows[0][4], "The Creation")
	}

	if rows[1][0] != "1:1" || rows[1][2] != "" {
		t.Errorf("note row = reference %q tags %q; want \"1:1\" / \"\"", rows[1][0], rows[1][2])
	}
	for i, r := range rows {
		if r[5] != "1" {
			t.Errorf("row %d Occurrence = %q; want %q", i, r[5], "1")
		}
	}
}

func TestOBSMarkdownNotes_ExtractsSupportReference(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantNote    string
		wantSupport string
	}{
		{
			name:        "trailing parenthetical",
			body:        "This is a direct quotation. (See: [[rc://fr/ta/man/translate/figs-quotations]])",
			wantNote:    "This is a direct quotation.",
			wantSupport: "rc://*/ta/man/translate/figs-quotations",
		},
		{
			name:        "escaped brackets",
			body:        "Это прямая речь. (См.: \\[\\[rc://en/ta/man/translate/figs-quotations\\]\\])",
			wantNote:    "Это прямая речь.",
			wantSupport: "rc://*/ta/man/translate/figs-quotations",
		},
		{
			name:        "link already language neutral",
			body:        "A note. (See: [[rc://*/ta/man/translate/figs-idiom]])",
			wantNote:    "A note.",
			wantSupport: "rc://*/ta/man/translate/figs-idiom",
		},
		{
			// Cutting a mid-sentence link out would damage the note, so it stays put.
			name:        "link not in a trailing parenthetical",
			body:        "See (See: [[rc://fr/ta/man/translate/figs-idiom]]) and then more text.",
			wantNote:    "See (See: [[rc://fr/ta/man/translate/figs-idiom]]) and then more text.",
			wantSupport: "",
		},
		{
			name:        "no link at all",
			body:        "Just a plain note.",
			wantNote:    "Just a plain note.",
			wantSupport: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inDir, outDir := t.TempDir(), t.TempDir()
			m := writeOBSMarkdownManifest(t, inDir, "OBS Study Notes", "obs-sn")
			writeContentFile(t, inDir, "content/01/01.md", "# quote\n\n"+tc.body+"\n")

			convertOBSMarkdown(t, "OBS Study Notes", m, inDir, outDir)

			_, rows := readTSVRows(t, outDir)
			if len(rows) != 1 {
				t.Fatalf("got %d rows; want 1: %v", len(rows), rows)
			}
			if rows[0][3] != tc.wantSupport {
				t.Errorf("SupportReference = %q; want %q", rows[0][3], tc.wantSupport)
			}
			if rows[0][6] != tc.wantNote {
				t.Errorf("Note = %q; want %q", rows[0][6], tc.wantNote)
			}
		})
	}
}

func TestOBSMarkdownNotes_DropsImportantTermsBlock(t *testing.T) {
	inDir, outDir := t.TempDir(), t.TempDir()
	m := writeOBSMarkdownManifest(t, inDir, "OBS Translation Notes", "obs-tn")

	// The heading is in the resource's own language, so the block is
	// recognized by its body: nothing but Translation Words links.
	writeContentFile(t, inDir, "content/01/01.md", `# au commencement

Le commencement est le moment quand Dieu a entrepris la création.

# Termes Importants

* [[rc://fr/tw/dict/bible/kt/god]]
* [[rc://fr/tw/dict/bible/kt/holyspirit]]
`)

	convertOBSMarkdown(t, "OBS Translation Notes", m, inDir, outDir)

	_, rows := readTSVRows(t, outDir)
	if len(rows) != 1 {
		t.Fatalf("got %d rows; want 1 (the terms block should be dropped): %v", len(rows), rows)
	}
	if rows[0][4] != "au commencement" {
		t.Errorf("Quote = %q; want %q", rows[0][4], "au commencement")
	}
	for _, r := range rows {
		if strings.Contains(strings.Join(r, "\t"), "/tw/") {
			t.Errorf("row still carries a Translation Words link: %v", r)
		}
	}
}

// --- OBS Study Questions (content/<story>.md) ---

func TestOBSMarkdownStoryQuestions_SectionsTagsAndReferences(t *testing.T) {
	inDir, outDir := t.TempDir(), t.TempDir()
	m := writeOBSMarkdownManifest(t, inDir, "OBS Study Questions", "obs-sq")

	writeContentFile(t, inDir, "content/00.md", "# A Guide to Using the Stories\n\n## The Reason\n\nBecause.\n")
	writeContentFile(t, inDir, "content/01.md", `# 01. The Creation

## What the Story Says

1. What did God create?

God created the universe.

1. What two trees did God plant?

He planted two special trees ([01:11](01/11)).

## What the Story Means

1. Why did God rest?

He had finished ([01:15](01/15)), ([01:16](01/16)).

## Summary

God created everything that exists.
`)

	convertOBSMarkdown(t, "OBS Study Questions", m, inDir, outDir)

	header, rows := readTSVRows(t, outDir)
	wantHeader := []string{"Reference", "ID", "Tags", "Quote", "Occurrence", "Question", "Response"}
	if strings.Join(header, "\t") != strings.Join(wantHeader, "\t") {
		t.Errorf("header = %v; want %v", header, wantHeader)
	}

	if len(rows) != 5 {
		t.Fatalf("got %d rows; want 5: %v", len(rows), rows)
	}

	// content/00.md becomes the front-matter intro, newlines folded to \n.
	if rows[0][0] != "front" || rows[0][2] != "intro" {
		t.Errorf("front row = reference %q tags %q; want \"front\" / \"intro\"", rows[0][0], rows[0][2])
	}
	if rows[0][5] != "" {
		t.Errorf("front row Question = %q; want empty", rows[0][5])
	}
	if !strings.Contains(rows[0][6], `\n`) || strings.Contains(rows[0][6], "\n") {
		t.Errorf("front row Response should fold newlines to the literal sequence: %q", rows[0][6])
	}

	// Section order sets the tag, since the headings are localized.
	wantTags := []string{"intro", "meaning", "meaning", "application", "summary"}
	for i, want := range wantTags {
		if rows[i][2] != want {
			t.Errorf("row %d Tags = %q; want %q", i, rows[i][2], want)
		}
	}

	// A question whose answer links no frame covers the whole story;
	// story 1 has 16 frames.
	if rows[1][0] != "1:1-16" {
		t.Errorf("unlinked question Reference = %q; want %q", rows[1][0], "1:1-16")
	}
	if rows[2][0] != "1:11" {
		t.Errorf("single-frame question Reference = %q; want %q", rows[2][0], "1:11")
	}
	if rows[3][0] != "1:15-16" {
		t.Errorf("multi-frame question Reference = %q; want %q", rows[3][0], "1:15-16")
	}
	if rows[4][0] != "1:1-16" {
		t.Errorf("summary Reference = %q; want %q", rows[4][0], "1:1-16")
	}
	if rows[4][5] != "" || rows[4][6] != "God created everything that exists." {
		t.Errorf("summary row question/response = %q / %q", rows[4][5], rows[4][6])
	}
}

// --- Repos that declare a Markdown subject but already ship a TSV ---

func TestOBSMarkdown_PassesThroughExistingTSV(t *testing.T) {
	inDir, outDir := t.TempDir(), t.TempDir()
	m := writeOBSMarkdownManifest(t, inDir, "OBS Translation Questions", "obs-tq")
	m.Projects = []rc.Project{{Identifier: "obs", Path: "./tq_OBS.tsv"}}

	content := "Reference\tID\tTags\tQuote\tOccurrence\tQuestion\tResponse\n1:1\tes4e\t\t\t\tWhy?\tBecause.\n"
	writeContentFile(t, inDir, "tq_OBS.tsv", content)

	convertOBSMarkdown(t, "OBS Translation Questions", m, inDir, outDir)

	got, err := os.ReadFile(filepath.Join(outDir, "ingredients", "OBS.tsv"))
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	if string(got) != content {
		t.Errorf("existing TSV was not copied through unchanged:\ngot:  %q\nwant: %q", got, content)
	}
}

func TestOBSMarkdown_PassesThroughRootTSVWhenProjectPathIsStale(t *testing.T) {
	inDir, outDir := t.TempDir(), t.TempDir()
	// Some repos still point at ./content after moving to a root-level TSV.
	m := writeOBSMarkdownManifest(t, inDir, "OBS Translation Notes", "obs-tn")

	content := "Reference\tID\tTags\tSupportReference\tQuote\tOccurrence\tNote\n1:1\tlm48\t\t\tquote\t1\tA note.\n"
	writeContentFile(t, inDir, "tn_OBS.tsv", content)

	convertOBSMarkdown(t, "OBS Translation Notes", m, inDir, outDir)

	got, err := os.ReadFile(filepath.Join(outDir, "ingredients", "OBS.tsv"))
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	if string(got) != content {
		t.Errorf("root TSV was not copied through unchanged:\ngot:  %q\nwant: %q", got, content)
	}
}

// --- Metadata and IDs ---

func TestOBSMarkdown_MetadataMatchesTSVCounterpart(t *testing.T) {
	tests := []struct {
		subject      string
		identifier   string
		flavor       string
		abbreviation string
	}{
		{"OBS Translation Questions", "obs-tq", "x-obsquestions", "OBSTQ"},
		{"OBS Translation Notes", "obs-tn", "x-obsnotes", "OBSTN"},
		{"OBS Study Notes", "obs-sn", "x-obsnotes", "OBSSN"},
		{"OBS Study Questions", "obs-sq", "x-obsquestions", "OBSSQ"},
	}

	for _, tc := range tests {
		t.Run(tc.subject, func(t *testing.T) {
			inDir, outDir := t.TempDir(), t.TempDir()
			m := writeOBSMarkdownManifest(t, inDir, tc.subject, tc.identifier)

			if tc.subject == "OBS Study Questions" {
				writeContentFile(t, inDir, "content/01.md", "# 01. Title\n\n## Says\n\n1. Q?\n\nA.\n")
			} else {
				writeContentFile(t, inDir, "content/01/01.md", "# heading\n\nbody\n")
			}

			h, err := handler.Lookup(tc.subject)
			if err != nil {
				t.Fatalf("Lookup(%q) failed: %v", tc.subject, err)
			}
			meta, err := h.Convert(context.Background(), m, inDir, outDir, handler.Options{})
			if err != nil {
				t.Fatalf("Convert failed: %v", err)
			}

			if meta.Type.FlavorType.Name != "peripheral" {
				t.Errorf("flavorType = %q; want %q", meta.Type.FlavorType.Name, "peripheral")
			}
			if meta.Type.FlavorType.Flavor.Name != tc.flavor {
				t.Errorf("flavor = %q; want %q", meta.Type.FlavorType.Flavor.Name, tc.flavor)
			}
			if meta.Identification.Abbreviation["en"] != tc.abbreviation {
				t.Errorf("abbreviation = %q; want %q", meta.Identification.Abbreviation["en"], tc.abbreviation)
			}
			if _, ok := meta.LocalizedNames["book-obs"]; !ok {
				t.Errorf("localizedNames missing book-obs: %v", meta.LocalizedNames)
			}

			wantIngredients := []string{"ingredients/OBS.tsv", "ingredients/LICENSE.md"}
			if len(meta.Ingredients) != len(wantIngredients) {
				t.Errorf("got %d ingredients; want %d: %v", len(meta.Ingredients), len(wantIngredients), meta.Ingredients)
			}
			for _, key := range wantIngredients {
				if _, ok := meta.Ingredients[key]; !ok {
					t.Errorf("missing ingredient %q; got %v", key, meta.Ingredients)
				}
			}
			if ing := meta.Ingredients["ingredients/OBS.tsv"]; ing.MimeType != "text/tab-separated-values" {
				t.Errorf("OBS.tsv mimeType = %q; want %q", ing.MimeType, "text/tab-separated-values")
			}
		})
	}
}

// tsvID is the 4-character ID form the unfoldingWord TSV resources use.
var tsvID = regexp.MustCompile(`^[a-z][a-z0-9]{3}$`)

func TestOBSMarkdown_IDsAreWellFormedUniqueAndStable(t *testing.T) {
	build := func(t *testing.T) string {
		t.Helper()
		inDir, outDir := t.TempDir(), t.TempDir()
		m := writeOBSMarkdownManifest(t, inDir, "OBS Translation Questions", "obs-tq")
		for story := 1; story <= 3; story++ {
			for frame := 1; frame <= 4; frame++ {
				writeContentFile(t, inDir,
					filepath.Join("content", pad(story), pad(frame)+".md"),
					"# Question for "+pad(story)+":"+pad(frame)+"?\n\nAnswer.\n")
			}
		}
		convertOBSMarkdown(t, "OBS Translation Questions", m, inDir, outDir)
		data, err := os.ReadFile(filepath.Join(outDir, "ingredients", "OBS.tsv"))
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}

	first := build(t)
	second := build(t)
	if first != second {
		t.Error("converting the same source twice produced different TSV output; IDs must be derived, not random")
	}

	_, rows := readTSVRowsFromString(t, first)
	if len(rows) != 12 {
		t.Fatalf("got %d rows; want 12", len(rows))
	}
	seen := make(map[string]bool, len(rows))
	for i, r := range rows {
		id := r[1]
		if !tsvID.MatchString(id) {
			t.Errorf("row %d ID = %q; want 4 characters matching %s", i, id, tsvID)
		}
		if seen[id] {
			t.Errorf("row %d ID %q is a duplicate", i, id)
		}
		seen[id] = true
	}
}

// TestOBSMarkdown_RowsWithIdenticalContentGetDistinctIDs covers the collision
// path: two rows whose columns are byte-identical still need unique IDs.
func TestOBSMarkdown_RowsWithIdenticalContentGetDistinctIDs(t *testing.T) {
	inDir, outDir := t.TempDir(), t.TempDir()
	m := writeOBSMarkdownManifest(t, inDir, "OBS Translation Questions", "obs-tq")
	writeContentFile(t, inDir, "content/01/01.md", "# Same question?\n\nSame answer.\n\n# Same question?\n\nSame answer.\n")

	convertOBSMarkdown(t, "OBS Translation Questions", m, inDir, outDir)

	_, rows := readTSVRows(t, outDir)
	if len(rows) != 2 {
		t.Fatalf("got %d rows; want 2: %v", len(rows), rows)
	}
	if rows[0][1] == rows[1][1] {
		t.Errorf("identical rows share the ID %q; IDs must be unique", rows[0][1])
	}
}

func TestOBSMarkdown_EmptyContentIsAnError(t *testing.T) {
	inDir, outDir := t.TempDir(), t.TempDir()
	m := writeOBSMarkdownManifest(t, inDir, "OBS Translation Questions", "obs-tq")
	if err := os.MkdirAll(filepath.Join(inDir, "content"), 0755); err != nil {
		t.Fatal(err)
	}

	h, err := handler.Lookup("OBS Translation Questions")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.Convert(context.Background(), m, inDir, outDir, handler.Options{}); err == nil {
		t.Error("expected an error when the content directory holds no Markdown")
	}
}

func TestOBSMarkdown_TabsAndNewlinesAreEscaped(t *testing.T) {
	inDir, outDir := t.TempDir(), t.TempDir()
	m := writeOBSMarkdownManifest(t, inDir, "OBS Translation Notes", "obs-tn")
	writeContentFile(t, inDir, "content/01/01.md", "# quote\n\nFirst paragraph.\n\nSecond\tparagraph.\n")

	convertOBSMarkdown(t, "OBS Translation Notes", m, inDir, outDir)

	_, rows := readTSVRows(t, outDir)
	if len(rows) != 1 {
		t.Fatalf("got %d rows; want 1: %v", len(rows), rows)
	}
	if len(rows[0]) != 7 {
		t.Fatalf("row has %d columns; want 7 (a tab in the body must not split it): %v", len(rows[0]), rows[0])
	}
	want := `First paragraph.\n\nSecond paragraph.`
	if rows[0][6] != want {
		t.Errorf("Note = %q; want %q", rows[0][6], want)
	}
}

// pad renders a story or frame number in the zero-padded form the repos use.
func pad(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

// readTSVRowsFromString splits already-read TSV content into header and rows.
func readTSVRowsFromString(t *testing.T, content string) ([]string, [][]string) {
	t.Helper()

	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	header := strings.Split(lines[0], "\t")
	rows := make([][]string, 0, len(lines)-1)
	for _, line := range lines[1:] {
		rows = append(rows, strings.Split(line, "\t"))
	}
	return header, rows
}
