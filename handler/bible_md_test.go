package handler_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unfoldingWord/go-rc2sb/handler"
	"github.com/unfoldingWord/go-rc2sb/rc"
)

func bibleMarkdownManifest(subject string, projects ...rc.Project) *rc.Manifest {
	m := &rc.Manifest{Projects: projects}
	m.DublinCore.Subject = subject
	m.DublinCore.Identifier = "tn"
	m.DublinCore.Title = "Test " + subject
	m.DublinCore.Publisher = "Door43"
	m.DublinCore.Rights = "CC BY-SA 4.0"
	m.DublinCore.Version = "1"
	m.DublinCore.Language.Identifier = "as"
	m.DublinCore.Language.Direction = "ltr"
	return m
}

func readBookTSV(t *testing.T, outDir, bookCode string) [][]string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(outDir, "ingredients", bookCode+".tsv"))
	if err != nil {
		t.Fatalf("reading %s.tsv: %v", bookCode, err)
	}
	var rows [][]string
	for _, line := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
		rows = append(rows, strings.Split(line, "\t"))
	}
	return rows
}

func TestBibleMarkdownNotes_ProducesBookTSVInBookOrder(t *testing.T) {
	inDir, outDir := t.TempDir(), t.TempDir()
	writeContentFile(t, inDir, "1co/front/intro.md", "# Introduction to 1 Corinthians\n\n## Part 1\n\nAbout the book.\n")
	writeContentFile(t, inDir, "1co/01/intro.md", "# 1 Corinthians 01 General Notes\n\nChapter notes.\n")
	writeContentFile(t, inDir, "1co/01/04.md", "# grace\n\nGod's kindness.\n")
	writeContentFile(t, inDir, "1co/01/01.md", "# Sosthenes our brother\n\nBoth knew him. (See: [[rc://as/ta/man/translate/translate-names]])\n\n# called to be saints\n\nGod called them.\n")
	writeContentFile(t, inDir, "1co/01/notes.txt", "not content")
	writeContentFile(t, inDir, "1co/10/01.md", "# fathers\n\nAncestors.\n")
	m := bibleMarkdownManifest("Translation Notes", rc.Project{Identifier: "1co", Path: "./1co", Title: "1 Corinthians"})

	h, err := handler.Lookup("Translation Notes")
	if err != nil {
		t.Fatal(err)
	}
	md, err := h.Convert(context.Background(), m, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if md.Type.FlavorType.Name != "parascriptural" || md.Type.FlavorType.Flavor.Name != "x-bcvnotes" {
		t.Errorf("flavor = %s/%s, want parascriptural/x-bcvnotes", md.Type.FlavorType.Name, md.Type.FlavorType.Flavor.Name)
	}
	ing, ok := md.Ingredients["ingredients/1CO.tsv"]
	if !ok || len(ing.Scope["1CO"]) != 0 || ing.Scope == nil {
		t.Errorf("ingredients/1CO.tsv = %+v (present %v), want it scoped to 1CO", ing, ok)
	}

	rows := readBookTSV(t, outDir, "1CO")
	if got := strings.Join(rows[0], "|"); got != "Reference|ID|Tags|SupportReference|Quote|Occurrence|Note" {
		t.Errorf("header = %s", got)
	}
	var refs []string
	for _, r := range rows[1:] {
		refs = append(refs, r[0])
	}
	if got := strings.Join(refs, " "); got != "front:intro 1:intro 1:1 1:1 1:4 10:1" {
		t.Errorf("references = %s", got)
	}
	intro := rows[1]
	if intro[4] != "" || intro[5] != "0" || !strings.Contains(intro[6], `## Part 1\n\nAbout the book.`) {
		t.Errorf("intro row = %q, want one note holding the whole intro with its headings", intro)
	}
	first := rows[3]
	if first[3] != "rc://*/ta/man/translate/translate-names" || first[4] != "Sosthenes our brother" || first[5] != "1" || first[6] != "Both knew him." {
		t.Errorf("note row = %q", first)
	}
}

func TestBibleMarkdownQuestions_HeadingsBecomeQuestions(t *testing.T) {
	inDir, outDir := t.TempDir(), t.TempDir()
	writeContentFile(t, inDir, "tit/01/01.md", "# Who wrote the letter?\n\nPaul.\n\n# To whom?\n\nTitus.\n")
	m := bibleMarkdownManifest("Translation Questions", rc.Project{Identifier: "tit", Path: "./tit"})

	h, err := handler.Lookup("Translation Questions")
	if err != nil {
		t.Fatal(err)
	}
	md, err := h.Convert(context.Background(), m, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if md.Type.FlavorType.Flavor.Name != "x-bcvquestions" {
		t.Errorf("flavor = %s, want x-bcvquestions", md.Type.FlavorType.Flavor.Name)
	}
	rows := readBookTSV(t, outDir, "TIT")
	if got := strings.Join(rows[0], "|"); got != "Reference|ID|Tags|Quote|Occurrence|Question|Response" {
		t.Errorf("header = %s", got)
	}
	if len(rows) != 3 || rows[1][0] != "1:1" || rows[1][5] != "Who wrote the letter?" || rows[1][6] != "Paul." || rows[2][5] != "To whom?" {
		t.Errorf("rows = %q", rows)
	}
}

func TestBibleMarkdown_PassesThroughTSVAndSkipsMissingProjects(t *testing.T) {
	inDir, outDir := t.TempDir(), t.TempDir()
	writeContentFile(t, inDir, "sn_TIT.tsv", "Reference\tID\tTags\tSupportReference\tQuote\tOccurrence\tNote\n1:1\tabcd\t\t\t\t0\tA note\n")
	m := bibleMarkdownManifest("Study Notes",
		rc.Project{Identifier: "tit", Path: "./sn_TIT.tsv"},
		rc.Project{Identifier: "phm", Path: "./phm"})

	h, err := handler.Lookup("Study Notes")
	if err != nil {
		t.Fatal(err)
	}
	md, err := h.Convert(context.Background(), m, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if _, ok := md.Ingredients["ingredients/TIT.tsv"]; !ok {
		t.Error("the existing TSV was not passed through as ingredients/TIT.tsv")
	}
	if _, ok := md.Type.FlavorType.CurrentScope["PHM"]; ok {
		t.Error("a project with no content was put in currentScope")
	}
	scope, _ := json.Marshal(md.Type.FlavorType.CurrentScope)
	if string(scope) != `{"TIT":[]}` {
		t.Errorf("currentScope = %s", scope)
	}
}

func TestBibleMarkdown_NoContentIsAnError(t *testing.T) {
	m := bibleMarkdownManifest("Study Questions", rc.Project{Identifier: "tit", Path: "./tit"})
	h, err := handler.Lookup("Study Questions")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.Convert(context.Background(), m, t.TempDir(), t.TempDir(), handler.Options{}); err == nil {
		t.Error("Convert of an RC with no content succeeded")
	}
}
