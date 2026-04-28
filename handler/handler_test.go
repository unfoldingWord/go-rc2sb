package handler_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unfoldingWord/go-rc2sb/handler"
	"github.com/unfoldingWord/go-rc2sb/rc"
	"github.com/unfoldingWord/go-rc2sb/sb"

	// Register all handlers so Lookup works.
	_ "github.com/unfoldingWord/go-rc2sb/handler/subjects"
)

// --- CopyCommonRootFiles tests ---

func TestCopyCommonRootFiles_CopiesREADMEAndGitignore(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create README.md and .gitignore in the input directory
	if err := os.WriteFile(filepath.Join(inDir, "README.md"), []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inDir, ".gitignore"), []byte("*.tmp\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m := sb.NewMetadata()
	if err := handler.CopyCommonRootFiles(inDir, outDir, m); err != nil {
		t.Fatalf("CopyCommonRootFiles failed: %v", err)
	}

	// Verify README.md was copied and is not in ingredients
	if _, err := os.Stat(filepath.Join(outDir, "README.md")); os.IsNotExist(err) {
		t.Error("README.md was not copied to outDir")
	}
	if _, ok := m.Ingredients["README.md"]; ok {
		t.Error("README.md should not be in metadata ingredients")
	}

	// Verify .gitignore was copied and is not in ingredients
	if _, err := os.Stat(filepath.Join(outDir, ".gitignore")); os.IsNotExist(err) {
		t.Error(".gitignore was not copied to outDir")
	}
	if _, ok := m.Ingredients[".gitignore"]; ok {
		t.Error(".gitignore should not be in metadata ingredients")
	}
}

func TestCopyCommonRootFiles_CopiesGiteaDir(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create .gitea/ directory with a file
	giteaDir := filepath.Join(inDir, ".gitea")
	if err := os.MkdirAll(giteaDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(giteaDir, "auto_merge.yaml"), []byte("auto_merge: true\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m := sb.NewMetadata()
	if err := handler.CopyCommonRootFiles(inDir, outDir, m); err != nil {
		t.Fatalf("CopyCommonRootFiles failed: %v", err)
	}

	// Verify the .gitea directory was copied
	if _, err := os.Stat(filepath.Join(outDir, ".gitea", "auto_merge.yaml")); os.IsNotExist(err) {
		t.Error(".gitea/auto_merge.yaml was not copied to outDir")
	}
	if _, ok := m.Ingredients[".gitea/auto_merge.yaml"]; ok {
		t.Error(".gitea/auto_merge.yaml should not be in metadata ingredients")
	}
}

func TestCopyCommonRootFiles_CopiesGithubDir(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create .github/ directory with nested structure
	workflowsDir := filepath.Join(inDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte("name: CI\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m := sb.NewMetadata()
	if err := handler.CopyCommonRootFiles(inDir, outDir, m); err != nil {
		t.Fatalf("CopyCommonRootFiles failed: %v", err)
	}

	// Verify the .github directory was copied recursively
	if _, err := os.Stat(filepath.Join(outDir, ".github", "workflows", "ci.yml")); os.IsNotExist(err) {
		t.Error(".github/workflows/ci.yml was not copied to outDir")
	}
	if _, ok := m.Ingredients[".github/workflows/ci.yml"]; ok {
		t.Error(".github/workflows/ci.yml should not be in metadata ingredients")
	}
}

func TestCopyCommonRootFiles_SkipsMissingFiles(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// No files at all — should succeed without copying anything
	m := sb.NewMetadata()
	if err := handler.CopyCommonRootFiles(inDir, outDir, m); err != nil {
		t.Fatalf("CopyCommonRootFiles should not fail when no root files exist: %v", err)
	}

	if len(m.Ingredients) != 0 {
		t.Errorf("Expected 0 ingredients for empty inDir, got %d", len(m.Ingredients))
	}
}

func TestCopyCommonRootFiles_DoesNotCopyGitDir(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create .git/ directory (should NOT be copied)
	gitDir := filepath.Join(inDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte("[core]\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m := sb.NewMetadata()
	if err := handler.CopyCommonRootFiles(inDir, outDir, m); err != nil {
		t.Fatalf("CopyCommonRootFiles failed: %v", err)
	}

	// Verify .git was NOT copied
	if _, err := os.Stat(filepath.Join(outDir, ".git")); !os.IsNotExist(err) {
		t.Error(".git directory should NOT be copied to outDir")
	}
	for key := range m.Ingredients {
		if strings.HasPrefix(key, ".git/") {
			t.Errorf("ingredient key %q should not start with .git/", key)
		}
	}
}

func TestCopyCommonRootFiles_DoesNotAddRootFilesToIngredients(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	content := []byte("Hello, World!")
	if err := os.WriteFile(filepath.Join(inDir, "README.md"), content, 0644); err != nil {
		t.Fatal(err)
	}

	m := sb.NewMetadata()
	if err := handler.CopyCommonRootFiles(inDir, outDir, m); err != nil {
		t.Fatalf("CopyCommonRootFiles failed: %v", err)
	}

	if len(m.Ingredients) != 0 {
		t.Errorf("Expected no root-file ingredient entries, got %d", len(m.Ingredients))
	}
}

// --- Bible subject alias tests ---

func TestBibleSubjectAliases_AllRegistered(t *testing.T) {
	subjects := []string{
		"Aligned Bible",
		"Bible",
		"Hebrew Old Testament",
		"Greek New Testament",
	}

	for _, subject := range subjects {
		t.Run(subject, func(t *testing.T) {
			h, err := handler.Lookup(subject)
			if err != nil {
				t.Fatalf("Lookup(%q) failed: %v", subject, err)
			}
			if h.Subject() != subject {
				t.Errorf("Subject() = %q; want %q", h.Subject(), subject)
			}
		})
	}
}

func TestBibleSubjectAliases_AbbreviationFromIdentifier(t *testing.T) {
	tests := []struct {
		subject    string
		identifier string
		wantAbbr   string
	}{
		{"Aligned Bible", "ult", "ULT"},
		{"Bible", "ust", "UST"},
		{"Hebrew Old Testament", "uhb", "UHB"},
		{"Greek New Testament", "ugnt", "UGNT"},
	}

	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			h, err := handler.Lookup(tt.subject)
			if err != nil {
				t.Fatalf("Lookup(%q) failed: %v", tt.subject, err)
			}

			// Create a minimal RC structure to test abbreviation derivation
			inDir := t.TempDir()
			outDir := t.TempDir()

			// Write a minimal manifest
			manifest := &rc.Manifest{
				DublinCore: rc.DublinCore{
					Subject:    tt.subject,
					Identifier: tt.identifier,
					Title:      "Test " + tt.subject,
					Issued:     "2024-01-01",
					Publisher:  "test",
					Rights:     "CC BY-SA 4.0",
					Language: rc.Language{
						Identifier: "en",
						Title:      "English",
						Direction:  "ltr",
					},
				},
			}

			// Create a minimal USFM file for the handler to process
			os.MkdirAll(filepath.Join(inDir, "content"), 0755)
			os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte("License"), 0644)

			metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
			if err != nil {
				t.Fatalf("Convert failed: %v", err)
			}

			gotAbbr := metadata.Identification.Abbreviation["en"]
			if gotAbbr != tt.wantAbbr {
				t.Errorf("Abbreviation = %q; want %q", gotAbbr, tt.wantAbbr)
			}
		})
	}
}

// --- Localized book names tests ---

func TestBible_LocalizedNamesFromUSFM(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create a USFM file with Hindi toc markers
	usfmContent := "\\id GEN\n\\usfm 3.0\n\\h \u0909\u0924\u094d\u092a\u0924\u094d\u0924\u093f\n\\toc1 \u0909\u0924\u094d\u092a\u0924\u094d\u0924\u093f \u0915\u0940 \u092a\u0941\u0938\u094d\u0924\u0915\n\\toc2 \u0909\u0924\u094d\u092a\u0924\u094d\u0924\u093f\n\\toc3 \u0909\u0924\u094d\u092a\n\\mt1 \u0909\u0924\u094d\u092a\u0924\u094d\u0924\u093f\n\\c 1\n\\v 1 Test\n"
	os.WriteFile(filepath.Join(inDir, "01-GEN.usfm"), []byte(usfmContent), 0644)
	os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte("License"), 0644)

	manifest := &rc.Manifest{
		DublinCore: rc.DublinCore{
			Subject:    "Bible",
			Identifier: "irv",
			Title:      "Hindi IRV",
			Issued:     "2024-01-01",
			Publisher:  "test",
			Rights:     "CC BY-SA 4.0",
			Language: rc.Language{
				Identifier: "hi",
				Title:      "Hindi",
				Direction:  "ltr",
			},
		},
		Projects: []rc.Project{
			{
				Identifier: "gen",
				Path:       "./01-GEN.usfm",
				Sort:       1,
				Title:      "\u0909\u0924\u094d\u092a\u0924\u094d\u0924\u093f",
			},
		},
	}

	h, err := handler.Lookup("Bible")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	ln, ok := metadata.LocalizedNames["book-gen"]
	if !ok {
		t.Fatal("book-gen not found in localizedNames")
	}

	// Should have Hindi long name from \toc1
	if ln.Long["hi"] != "\u0909\u0924\u094d\u092a\u0924\u094d\u0924\u093f \u0915\u0940 \u092a\u0941\u0938\u094d\u0924\u0915" {
		t.Errorf("Long[hi] = %q; want Hindi toc1 value", ln.Long["hi"])
	}
	// Should have Hindi short name from \toc2
	if ln.Short["hi"] != "\u0909\u0924\u094d\u092a\u0924\u094d\u0924\u093f" {
		t.Errorf("Short[hi] = %q; want Hindi toc2 value", ln.Short["hi"])
	}
	// Should still have English fallback
	if ln.Long["en"] != "The Book of Genesis" {
		t.Errorf("Long[en] = %q; want English fallback", ln.Long["en"])
	}
}

func TestTN_LocalizedNamesFromManifestTitle(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create a TN TSV file
	tsvContent := "Reference\tID\tTags\tSupportReference\tQuote\tOccurrence\tNote\n1:1\tabcd\t\t\tword\t1\tA note\n"
	os.WriteFile(filepath.Join(inDir, "tn_GEN.tsv"), []byte(tsvContent), 0644)
	os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte("License"), 0644)

	manifest := &rc.Manifest{
		DublinCore: rc.DublinCore{
			Subject:    "TSV Translation Notes",
			Identifier: "tn",
			Title:      "Hindi TN",
			Issued:     "2024-01-01",
			Publisher:  "test",
			Rights:     "CC BY-SA 4.0",
			Language: rc.Language{
				Identifier: "hi",
				Title:      "Hindi",
				Direction:  "ltr",
			},
		},
		Projects: []rc.Project{
			{
				Identifier: "gen",
				Path:       "./tn_GEN.tsv",
				Sort:       1,
				Title:      "\u0909\u0924\u094d\u092a\u0924\u094d\u0924\u093f",
			},
		},
	}

	h, err := handler.Lookup("TSV Translation Notes")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	// No USFMPath — should use manifest project title
	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	ln, ok := metadata.LocalizedNames["book-gen"]
	if !ok {
		t.Fatal("book-gen not found in localizedNames")
	}

	// Should have Hindi name from manifest title
	if ln.Long["hi"] != "\u0909\u0924\u094d\u092a\u0924\u094d\u0924\u093f" {
		t.Errorf("Long[hi] = %q; want manifest project title", ln.Long["hi"])
	}
	if ln.Short["hi"] != "\u0909\u0924\u094d\u092a\u0924\u094d\u0924\u093f" {
		t.Errorf("Short[hi] = %q; want manifest project title", ln.Short["hi"])
	}
	// English fallback should still be present
	if ln.Long["en"] != "The Book of Genesis" {
		t.Errorf("Long[en] = %q; want English fallback", ln.Long["en"])
	}
}

func TestTN_LocalizedNamesFromUSFMPath(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	usfmDir := t.TempDir()

	// Create a USFM file in the USFMPath directory
	usfmContent := "\\id GEN\n\\toc1 Livre de la Genese\n\\toc2 Genese\n\\toc3 Gen\n"
	os.WriteFile(filepath.Join(usfmDir, "01-GEN.usfm"), []byte(usfmContent), 0644)

	// Create a TN TSV file
	tsvContent := "Reference\tID\tTags\tSupportReference\tQuote\tOccurrence\tNote\n1:1\tabcd\t\t\tword\t1\tA note\n"
	os.WriteFile(filepath.Join(inDir, "tn_GEN.tsv"), []byte(tsvContent), 0644)
	os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte("License"), 0644)

	manifest := &rc.Manifest{
		DublinCore: rc.DublinCore{
			Subject:    "TSV Translation Notes",
			Identifier: "tn",
			Title:      "French TN",
			Issued:     "2024-01-01",
			Publisher:  "test",
			Rights:     "CC BY-SA 4.0",
			Language: rc.Language{
				Identifier: "fr",
				Title:      "French",
				Direction:  "ltr",
			},
		},
		Projects: []rc.Project{
			{
				Identifier: "gen",
				Path:       "./tn_GEN.tsv",
				Sort:       1,
				Title:      "Genese",
			},
		},
	}

	h, err := handler.Lookup("TSV Translation Notes")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	opts := handler.Options{USFMPath: usfmDir}
	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, opts)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	ln, ok := metadata.LocalizedNames["book-gen"]
	if !ok {
		t.Fatal("book-gen not found in localizedNames")
	}

	// Should have French names from USFM (overrides manifest title)
	if ln.Long["fr"] != "Livre de la Genese" {
		t.Errorf("Long[fr] = %q; want USFM toc1 value", ln.Long["fr"])
	}
	if ln.Short["fr"] != "Genese" {
		t.Errorf("Short[fr] = %q; want USFM toc2 value", ln.Short["fr"])
	}
	if ln.Abbr["fr"] != "Gen" {
		t.Errorf("Abbr[fr] = %q; want USFM toc3 value", ln.Abbr["fr"])
	}
}

// --- TWL handler tests ---

func writeTWLManifest(t *testing.T, inDir string) *rc.Manifest {
	t.Helper()
	return &rc.Manifest{
		DublinCore: rc.DublinCore{
			Subject:    "TSV Translation Words Links",
			Identifier: "twl",
			Title:      "Test TWL",
			Issued:     "2024-01-01",
			Publisher:  "test",
			Rights:     "CC BY-SA 4.0",
			Language: rc.Language{
				Identifier: "en",
				Title:      "English",
				Direction:  "ltr",
			},
		},
		Projects: []rc.Project{
			{
				Identifier: "gen",
				Path:       "./twl_GEN.tsv",
				Sort:       1,
				Title:      "Genesis",
			},
		},
	}
}

func TestTWL_AutoDetectsPayload(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	manifest := writeTWLManifest(t, inDir)

	// Create the TWL TSV file with an rc:// link
	tsvContent := "Reference\tID\tTags\tOrigWords\tOccurrence\tTWLink\n" +
		"1:1\tabcd\t\tword\t1\trc://*/tw/dict/bible/names/adam\n"
	os.WriteFile(filepath.Join(inDir, "twl_GEN.tsv"), []byte(tsvContent), 0644)
	os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte("License"), 0644)

	// Create the en_tw/bible/ directory (auto-detection target)
	twBibleDir := filepath.Join(inDir, "en_tw", "bible", "names")
	os.MkdirAll(twBibleDir, 0755)
	os.WriteFile(filepath.Join(twBibleDir, "adam.md"), []byte("# Adam\n\nThe first man."), 0644)

	h, err := handler.Lookup("TSV Translation Words Links")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Verify payload was auto-detected and copied
	if _, ok := metadata.Ingredients["ingredients/payload/names/adam.md"]; !ok {
		t.Error("Payload article ingredients/payload/names/adam.md not found; auto-detection failed")
	}

	// Verify TSV was rewritten
	data, err := os.ReadFile(filepath.Join(outDir, "ingredients", "GEN.tsv"))
	if err != nil {
		t.Fatalf("Reading output TSV: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "rc://") {
		t.Error("TSV still contains rc:// links after auto-detection rewrite")
	}
	if !strings.Contains(content, "./payload/names/adam.md") {
		t.Error("TSV does not contain expected ./payload/names/adam.md path")
	}
}

func TestTWL_ExplicitPayloadPath(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	payloadDir := t.TempDir() // Separate directory for payload

	manifest := writeTWLManifest(t, inDir)

	// Create the TWL TSV file with an rc:// link
	tsvContent := "Reference\tID\tTags\tOrigWords\tOccurrence\tTWLink\n" +
		"1:1\tabcd\t\tword\t1\trc://*/tw/dict/bible/kt/god\n"
	os.WriteFile(filepath.Join(inDir, "twl_GEN.tsv"), []byte(tsvContent), 0644)
	os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte("License"), 0644)

	// Create the TW directory at the explicit payload path
	twBibleDir := filepath.Join(payloadDir, "bible", "kt")
	os.MkdirAll(twBibleDir, 0755)
	os.WriteFile(filepath.Join(twBibleDir, "god.md"), []byte("# God\n\nThe creator."), 0644)

	h, err := handler.Lookup("TSV Translation Words Links")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	opts := handler.Options{PayloadPath: payloadDir}
	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, opts)
	if err != nil {
		t.Fatalf("Convert with PayloadPath failed: %v", err)
	}

	// Verify payload from explicit path was copied
	if _, ok := metadata.Ingredients["ingredients/payload/kt/god.md"]; !ok {
		t.Error("Payload article ingredients/payload/kt/god.md not found; explicit PayloadPath failed")
	}

	// Verify TSV was rewritten
	data, err := os.ReadFile(filepath.Join(outDir, "ingredients", "GEN.tsv"))
	if err != nil {
		t.Fatalf("Reading output TSV: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "rc://") {
		t.Error("TSV still contains rc:// links after PayloadPath rewrite")
	}
	if !strings.Contains(content, "./payload/kt/god.md") {
		t.Error("TSV does not contain expected ./payload/kt/god.md path")
	}
}

func TestTWL_NoPayloadCopiesAsIs(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	manifest := writeTWLManifest(t, inDir)

	// Create the TWL TSV file with an rc:// link — but NO en_tw/ directory
	tsvContent := "Reference\tID\tTags\tOrigWords\tOccurrence\tTWLink\n" +
		"1:1\tabcd\t\tword\t1\trc://*/tw/dict/bible/names/adam\n"
	os.WriteFile(filepath.Join(inDir, "twl_GEN.tsv"), []byte(tsvContent), 0644)
	os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte("License"), 0644)

	h, err := handler.Lookup("TSV Translation Words Links")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert without payload failed: %v", err)
	}

	// Verify no payload ingredients
	for key := range metadata.Ingredients {
		if strings.HasPrefix(key, "ingredients/payload/") {
			t.Errorf("Unexpected payload ingredient %s when no TW directory exists", key)
		}
	}

	// Verify TSV was copied as-is (rc:// links preserved)
	data, err := os.ReadFile(filepath.Join(outDir, "ingredients", "GEN.tsv"))
	if err != nil {
		t.Fatalf("Reading output TSV: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "rc://") {
		t.Error("TSV should preserve rc:// links when no payload exists")
	}
	if strings.Contains(content, "./payload/") {
		t.Error("TSV should NOT contain ./payload/ paths when no payload exists")
	}
}

func TestTWL_LinkRewriteMultipleLinks(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	manifest := writeTWLManifest(t, inDir)

	// Create a TSV with multiple rc:// links across several rows
	tsvContent := "Reference\tID\tTags\tOrigWords\tOccurrence\tTWLink\n" +
		"1:1\ta001\t\tword1\t1\trc://*/tw/dict/bible/names/adam\n" +
		"1:2\ta002\t\tword2\t1\trc://*/tw/dict/bible/kt/god\n" +
		"1:3\ta003\t\tword3\t1\trc://en/tw/dict/bible/other/creation\n"
	os.WriteFile(filepath.Join(inDir, "twl_GEN.tsv"), []byte(tsvContent), 0644)
	os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte("License"), 0644)

	// Create the en_tw/bible/ directory
	for _, path := range []string{"names/adam.md", "kt/god.md", "other/creation.md"} {
		fullPath := filepath.Join(inDir, "en_tw", "bible", path)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte("# Article\n"), 0644)
	}

	h, err := handler.Lookup("TSV Translation Words Links")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Verify all three payload articles were copied
	expectedPayload := []string{
		"ingredients/payload/names/adam.md",
		"ingredients/payload/kt/god.md",
		"ingredients/payload/other/creation.md",
	}
	for _, key := range expectedPayload {
		if _, ok := metadata.Ingredients[key]; !ok {
			t.Errorf("Missing payload ingredient: %s", key)
		}
	}

	// Verify all rc:// links were rewritten
	data, err := os.ReadFile(filepath.Join(outDir, "ingredients", "GEN.tsv"))
	if err != nil {
		t.Fatalf("Reading output TSV: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "rc://") {
		t.Error("TSV still contains rc:// links — not all were rewritten")
	}

	// Verify specific rewrites
	expectedPaths := []string{
		"./payload/names/adam.md",
		"./payload/kt/god.md",
		"./payload/other/creation.md",
	}
	for _, p := range expectedPaths {
		if !strings.Contains(content, p) {
			t.Errorf("TSV missing expected rewritten path: %s", p)
		}
	}
}

func TestTWL_StripsTWLPrefix(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	manifest := writeTWLManifest(t, inDir)

	tsvContent := "Reference\tID\tTags\tOrigWords\tOccurrence\tTWLink\n" +
		"1:1\ta001\t\tword1\t1\trc://*/tw/dict/bible/names/adam\n"
	os.WriteFile(filepath.Join(inDir, "twl_GEN.tsv"), []byte(tsvContent), 0644)
	os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte("License"), 0644)

	h, err := handler.Lookup("TSV Translation Words Links")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Verify twl_ prefix was stripped: "twl_GEN.tsv" -> "ingredients/GEN.tsv"
	if _, ok := metadata.Ingredients["ingredients/GEN.tsv"]; !ok {
		t.Error("Expected ingredient key 'ingredients/GEN.tsv' (twl_ prefix should be stripped)")
	}

	// Verify the file exists on disk with the stripped name
	if _, err := os.Stat(filepath.Join(outDir, "ingredients", "GEN.tsv")); os.IsNotExist(err) {
		t.Error("ingredients/GEN.tsv file does not exist on disk")
	}
}

func TestTWL_CopiesRootFilesWithoutIngredientEntries(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	manifest := writeTWLManifest(t, inDir)

	tsvContent := "Reference\tID\tTags\tOrigWords\tOccurrence\tTWLink\n" +
		"1:1\ta001\t\tword1\t1\trc://*/tw/dict/bible/names/adam\n"
	os.WriteFile(filepath.Join(inDir, "twl_GEN.tsv"), []byte(tsvContent), 0644)
	os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte("License"), 0644)
	os.WriteFile(filepath.Join(inDir, "README.md"), []byte("# TWL Readme"), 0644)
	os.WriteFile(filepath.Join(inDir, ".gitignore"), []byte("*.tmp\n"), 0644)

	h, err := handler.Lookup("TSV Translation Words Links")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Verify root files are not in ingredients metadata
	if _, ok := metadata.Ingredients["README.md"]; ok {
		t.Error("README.md should not be present in TWL metadata ingredients")
	}
	if _, ok := metadata.Ingredients[".gitignore"]; ok {
		t.Error(".gitignore should not be present in TWL metadata ingredients")
	}

	// Verify files exist on disk
	if _, err := os.Stat(filepath.Join(outDir, "README.md")); os.IsNotExist(err) {
		t.Error("README.md was not copied to TWL output")
	}
	if _, err := os.Stat(filepath.Join(outDir, ".gitignore")); os.IsNotExist(err) {
		t.Error(".gitignore was not copied to TWL output")
	}
}

func TestTA_DoesNotCopyManifestOrMediaToRoot(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	manifest := &rc.Manifest{
		DublinCore: rc.DublinCore{
			Subject:    "Translation Academy",
			Identifier: "ta",
			Title:      "Test TA",
			Issued:     "2024-01-01",
			Publisher:  "test",
			Rights:     "CC BY-SA 4.0",
			Language: rc.Language{
				Identifier: "en",
				Title:      "English",
				Direction:  "ltr",
			},
		},
		Projects: []rc.Project{
			{Identifier: "intro"},
		},
	}

	if err := os.MkdirAll(filepath.Join(inDir, "intro"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inDir, "intro", "01.md"), []byte("# Intro"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte("License"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inDir, "manifest.yaml"), []byte("dublin_core: {}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inDir, "media.yaml"), []byte("projects: []"), 0644); err != nil {
		t.Fatal(err)
	}

	h, err := handler.Lookup("Translation Academy")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "LICENSE.md")); os.IsNotExist(err) {
		t.Error("LICENSE.md should be copied to TA output root")
	}
	if _, err := os.Stat(filepath.Join(outDir, "manifest.yaml")); !os.IsNotExist(err) {
		t.Error("manifest.yaml should not be copied to TA output root")
	}
	if _, err := os.Stat(filepath.Join(outDir, "media.yaml")); !os.IsNotExist(err) {
		t.Error("media.yaml should not be copied to TA output root")
	}
	if _, ok := metadata.Ingredients["ingredients/LICENSE.md"]; !ok {
		t.Error("ingredients/LICENSE.md should exist in TA metadata ingredients")
	}
}

func TestOBS_DoesNotCopyManifestOrMediaToRoot(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	manifest := &rc.Manifest{
		DublinCore: rc.DublinCore{
			Subject:    "Open Bible Stories",
			Identifier: "obs",
			Title:      "Test OBS",
			Issued:     "2024-01-01",
			Publisher:  "test",
			Rights:     "CC BY-SA 4.0",
			Language: rc.Language{
				Identifier: "en",
				Title:      "English",
				Direction:  "ltr",
			},
		},
	}

	if err := os.MkdirAll(filepath.Join(inDir, "content"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inDir, "content", "01.md"), []byte("# Story"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte("License"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inDir, "manifest.yaml"), []byte("dublin_core: {}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inDir, "media.yaml"), []byte("projects: []"), 0644); err != nil {
		t.Fatal(err)
	}

	h, err := handler.Lookup("Open Bible Stories")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "LICENSE.md")); os.IsNotExist(err) {
		t.Error("LICENSE.md should be copied to OBS output root")
	}
	if _, err := os.Stat(filepath.Join(outDir, "manifest.yaml")); !os.IsNotExist(err) {
		t.Error("manifest.yaml should not be copied to OBS output root")
	}
	if _, err := os.Stat(filepath.Join(outDir, "media.yaml")); !os.IsNotExist(err) {
		t.Error("media.yaml should not be copied to OBS output root")
	}
	if _, ok := metadata.Ingredients["ingredients/LICENSE.md"]; !ok {
		t.Error("ingredients/LICENSE.md should exist in OBS metadata ingredients")
	}
}

// --- Registry tests ---

func TestLookup_AllRegisteredSubjects(t *testing.T) {
	expectedSubjects := []string{
		"Open Bible Stories",
		"Aligned Bible",
		"Bible",
		"Hebrew Old Testament",
		"Greek New Testament",
		"Translation Words",
		"Translation Academy",
		"TSV Translation Notes",
		"TSV Translation Questions",
		"TSV Translation Words Links",
		"TSV OBS Study Notes",
		"TSV OBS Study Questions",
		"TSV OBS Translation Notes",
		"TSV OBS Translation Questions",
	}

	for _, subject := range expectedSubjects {
		t.Run(subject, func(t *testing.T) {
			h, err := handler.Lookup(subject)
			if err != nil {
				t.Fatalf("Lookup(%q) failed: %v", subject, err)
			}
			if h.Subject() != subject {
				t.Errorf("Subject() = %q; want %q", h.Subject(), subject)
			}
		})
	}
}

func TestSupportedSubjects_Count(t *testing.T) {
	subjects := handler.SupportedSubjects()
	if len(subjects) != 14 {
		t.Errorf("SupportedSubjects() returned %d subjects; want 14. Got: %v", len(subjects), subjects)
	}
}

func TestLookup_UnsupportedSubject(t *testing.T) {
	_, err := handler.Lookup("Nonexistent Subject")
	if err == nil {
		t.Fatal("expected error for unsupported subject")
	}
	if !strings.Contains(err.Error(), "unsupported subject") {
		t.Errorf("error should mention 'unsupported subject': %v", err)
	}
}

// --- Missing LICENSE.md / README.md tests ---

func TestCopyLicenseIngredient_MissingLicenseUsesDefault(t *testing.T) {
	inDir := t.TempDir() // No LICENSE.md
	outDir := t.TempDir()

	ing, err := handler.CopyLicenseIngredient(inDir, outDir)
	if err != nil {
		t.Fatalf("CopyLicenseIngredient should not fail when LICENSE.md is missing: %v", err)
	}

	// Verify the default LICENSE.md was written
	dst := filepath.Join(outDir, "ingredients", "LICENSE.md")
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Fatal("ingredients/LICENSE.md should exist using default license")
	}

	// Verify the ingredient has valid checksum and size
	if ing.Size == 0 {
		t.Error("default LICENSE.md ingredient size should be > 0")
	}
	if ing.Checksum.MD5 == "" {
		t.Error("default LICENSE.md ingredient should have MD5 checksum")
	}

	// Verify the content contains CC BY-SA 4.0 text
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("reading default LICENSE.md: %v", err)
	}
	if !strings.Contains(string(data), "Creative Commons Attribution-ShareAlike 4.0") {
		t.Error("default LICENSE.md should contain CC BY-SA 4.0 text")
	}
}

func TestCopyLicenseIngredient_ExistingLicensePreferred(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create a custom LICENSE.md
	customContent := "Custom License Content"
	os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte(customContent), 0644)

	_, err := handler.CopyLicenseIngredient(inDir, outDir)
	if err != nil {
		t.Fatalf("CopyLicenseIngredient failed: %v", err)
	}

	// Verify the RC's LICENSE.md was used (not default)
	data, err := os.ReadFile(filepath.Join(outDir, "ingredients", "LICENSE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != customContent {
		t.Errorf("Expected RC's LICENSE.md content, got %q", string(data))
	}
}

func TestCopyLicenseToRoot_MissingLicenseUsesDefault(t *testing.T) {
	inDir := t.TempDir() // No LICENSE.md
	outDir := t.TempDir()

	err := handler.CopyLicenseToRoot(inDir, outDir)
	if err != nil {
		t.Fatalf("CopyLicenseToRoot should not fail when LICENSE.md is missing: %v", err)
	}

	// Verify the default LICENSE.md was written to root
	data, err := os.ReadFile(filepath.Join(outDir, "LICENSE.md"))
	if err != nil {
		t.Fatal("LICENSE.md should exist at SB root using default license")
	}
	if !strings.Contains(string(data), "Creative Commons Attribution-ShareAlike 4.0") {
		t.Error("default root LICENSE.md should contain CC BY-SA 4.0 text")
	}
}

func TestCopyLicenseToRoot_ExistingLicensePreferred(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	customContent := "My Custom License"
	os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte(customContent), 0644)

	err := handler.CopyLicenseToRoot(inDir, outDir)
	if err != nil {
		t.Fatalf("CopyLicenseToRoot failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "LICENSE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != customContent {
		t.Errorf("Expected RC's LICENSE.md content, got %q", string(data))
	}
}

func TestBible_ConvertsWithoutLicenseOrReadme(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create a minimal USFM file — NO LICENSE.md, NO README.md
	usfmContent := "\\id GEN\n\\c 1\n\\v 1 In the beginning.\n"
	os.WriteFile(filepath.Join(inDir, "01-GEN.usfm"), []byte(usfmContent), 0644)

	manifest := &rc.Manifest{
		DublinCore: rc.DublinCore{
			Subject:    "Bible",
			Identifier: "ult",
			Title:      "Test Bible",
			Issued:     "2024-01-01",
			Publisher:  "test",
			Rights:     "CC BY-SA 4.0",
			Language: rc.Language{
				Identifier: "en",
				Title:      "English",
				Direction:  "ltr",
			},
		},
		Projects: []rc.Project{
			{
				Identifier: "gen",
				Path:       "./01-GEN.usfm",
				Sort:       1,
				Title:      "Genesis",
			},
		},
	}

	h, err := handler.Lookup("Bible")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert should not fail without LICENSE.md: %v", err)
	}

	// Verify LICENSE.md exists in ingredients/ with default content
	if _, ok := metadata.Ingredients["ingredients/LICENSE.md"]; !ok {
		t.Error("ingredients/LICENSE.md should exist in metadata using default license")
	}

	data, err := os.ReadFile(filepath.Join(outDir, "ingredients", "LICENSE.md"))
	if err != nil {
		t.Fatal("ingredients/LICENSE.md should exist on disk")
	}
	if !strings.Contains(string(data), "Creative Commons") {
		t.Error("default LICENSE.md should contain Creative Commons text")
	}
}

func TestTN_ConvertsWithoutLicenseOrReadme(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create a TN TSV file — NO LICENSE.md, NO README.md
	tsvContent := "Reference\tID\tTags\tSupportReference\tQuote\tOccurrence\tNote\n1:1\tabcd\t\t\tword\t1\tA note\n"
	os.WriteFile(filepath.Join(inDir, "tn_GEN.tsv"), []byte(tsvContent), 0644)

	manifest := &rc.Manifest{
		DublinCore: rc.DublinCore{
			Subject:    "TSV Translation Notes",
			Identifier: "tn",
			Title:      "Test TN",
			Issued:     "2024-01-01",
			Publisher:  "test",
			Rights:     "CC BY-SA 4.0",
			Language: rc.Language{
				Identifier: "en",
				Title:      "English",
				Direction:  "ltr",
			},
		},
		Projects: []rc.Project{
			{
				Identifier: "gen",
				Path:       "./tn_GEN.tsv",
				Sort:       1,
				Title:      "Genesis",
			},
		},
	}

	h, err := handler.Lookup("TSV Translation Notes")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert should not fail without LICENSE.md: %v", err)
	}

	if _, ok := metadata.Ingredients["ingredients/LICENSE.md"]; !ok {
		t.Error("ingredients/LICENSE.md should exist using default license")
	}
}

func TestOBS_ConvertsWithoutLicenseOrReadme(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create minimal OBS content — NO LICENSE.md, NO README.md
	os.MkdirAll(filepath.Join(inDir, "content"), 0755)
	os.WriteFile(filepath.Join(inDir, "content", "01.md"), []byte("# Story 1\n"), 0644)

	manifest := &rc.Manifest{
		DublinCore: rc.DublinCore{
			Subject:    "Open Bible Stories",
			Identifier: "obs",
			Title:      "Test OBS",
			Issued:     "2024-01-01",
			Publisher:  "test",
			Rights:     "CC BY-SA 4.0",
			Language: rc.Language{
				Identifier: "en",
				Title:      "English",
				Direction:  "ltr",
			},
		},
	}

	h, err := handler.Lookup("Open Bible Stories")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert should not fail without LICENSE.md: %v", err)
	}

	// Verify both root and ingredients LICENSE.md exist with default content
	if _, ok := metadata.Ingredients["ingredients/LICENSE.md"]; !ok {
		t.Error("ingredients/LICENSE.md should exist using default license")
	}

	rootLic, err := os.ReadFile(filepath.Join(outDir, "LICENSE.md"))
	if err != nil {
		t.Fatal("root LICENSE.md should exist using default license")
	}
	if !strings.Contains(string(rootLic), "Creative Commons") {
		t.Error("root LICENSE.md should contain Creative Commons text")
	}
}

func TestOBSTSV_ConvertsWithoutLicense(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create a OBS TSV file — NO LICENSE.md
	tsvContent := "Reference\tID\tTags\tSupportReference\tQuote\tOccurrence\tNote\n01:01\tabcd\t\t\tword\t1\tA note\n"
	os.WriteFile(filepath.Join(inDir, "sn_OBS.tsv"), []byte(tsvContent), 0644)

	manifest := &rc.Manifest{
		DublinCore: rc.DublinCore{
			Subject:    "TSV OBS Study Notes",
			Identifier: "obs-sn",
			Title:      "Test OBS SN",
			Issued:     "2024-01-01",
			Publisher:  "test",
			Rights:     "CC BY-SA 4.0",
			Language: rc.Language{
				Identifier: "en",
				Title:      "English",
				Direction:  "ltr",
			},
		},
		Projects: []rc.Project{
			{
				Identifier: "obs",
				Path:       "./sn_OBS.tsv",
				Sort:       1,
				Title:      "OBS Study Notes",
			},
		},
	}

	h, err := handler.Lookup("TSV OBS Study Notes")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert should not fail without LICENSE.md: %v", err)
	}

	if _, ok := metadata.Ingredients["ingredients/LICENSE.md"]; !ok {
		t.Error("ingredients/LICENSE.md should exist using default license")
	}
}

// --- OBS root-level content tests ---

func TestOBS_RootLevelContent(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create OBS content in the repo root (path: ".")
	// Includes both flat files and subdirectories
	os.WriteFile(filepath.Join(inDir, "01.md"), []byte("# Story 1\n"), 0644)
	os.WriteFile(filepath.Join(inDir, "02.md"), []byte("# Story 2\n"), 0644)
	os.WriteFile(filepath.Join(inDir, "50.md"), []byte("# Story 50\n"), 0644)
	os.WriteFile(filepath.Join(inDir, "front.md"), []byte("# Front Matter\n"), 0644)
	os.WriteFile(filepath.Join(inDir, "back.md"), []byte("# Back Matter\n"), 0644)
	os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte("License"), 0644)
	os.WriteFile(filepath.Join(inDir, "README.md"), []byte("# OBS Readme"), 0644)
	os.WriteFile(filepath.Join(inDir, "manifest.yaml"), []byte("dublin_core: {}"), 0644)

	manifest := &rc.Manifest{
		DublinCore: rc.DublinCore{
			Subject:    "Open Bible Stories",
			Identifier: "obs",
			Title:      "Test OBS Root",
			Issued:     "2024-01-01",
			Publisher:  "test",
			Rights:     "CC BY-SA 4.0",
			Language: rc.Language{
				Identifier: "sgh",
				Title:      "Shughni",
				Direction:  "ltr",
			},
		},
		Projects: []rc.Project{
			{
				Identifier: "obs",
				Path:       ".",
				Sort:       0,
				Title:      "Open Bible Stories",
			},
		},
	}

	h, err := handler.Lookup("Open Bible Stories")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Verify content files were copied as content/ ingredients
	expectedContent := []string{
		"ingredients/content/01.md",
		"ingredients/content/02.md",
		"ingredients/content/50.md",
		"ingredients/content/front.md",
		"ingredients/content/back.md",
	}
	for _, key := range expectedContent {
		if _, ok := metadata.Ingredients[key]; !ok {
			t.Errorf("Expected ingredient %s not found", key)
		}
		if _, err := os.Stat(filepath.Join(outDir, key)); os.IsNotExist(err) {
			t.Errorf("Expected file %s not found on disk", key)
		}
	}

	// Verify LICENSE.md is in ingredients
	if _, ok := metadata.Ingredients["ingredients/LICENSE.md"]; !ok {
		t.Error("ingredients/LICENSE.md should exist")
	}

	// Verify excluded files were NOT copied to content/ ingredients
	excludedKeys := []string{
		"ingredients/content/LICENSE.md",
		"ingredients/content/README.md",
		"ingredients/content/manifest.yaml",
	}
	for _, key := range excludedKeys {
		if _, ok := metadata.Ingredients[key]; ok {
			t.Errorf("Non-content file should not be in ingredients: %s", key)
		}
	}

	// Verify README.md was copied to root (by CopyCommonRootFiles)
	if _, err := os.Stat(filepath.Join(outDir, "README.md")); os.IsNotExist(err) {
		t.Error("README.md should be copied to output root")
	}
}

func TestOBS_RootLevelContent_WithSubdirectories(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create OBS content with front/ and back/ subdirectories (like en_obs)
	os.WriteFile(filepath.Join(inDir, "01.md"), []byte("# Story 1\n"), 0644)
	os.WriteFile(filepath.Join(inDir, "02.md"), []byte("# Story 2\n"), 0644)

	// front/ directory with nested files
	os.MkdirAll(filepath.Join(inDir, "front"), 0755)
	os.WriteFile(filepath.Join(inDir, "front", "intro.md"), []byte("# Intro\n"), 0644)
	os.WriteFile(filepath.Join(inDir, "front", "title.md"), []byte("# Title\n"), 0644)

	// back/ directory with nested files
	os.MkdirAll(filepath.Join(inDir, "back"), 0755)
	os.WriteFile(filepath.Join(inDir, "back", "intro.md"), []byte("# Back Intro\n"), 0644)

	os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte("License"), 0644)
	os.WriteFile(filepath.Join(inDir, "manifest.yaml"), []byte("dublin_core: {}"), 0644)

	manifest := &rc.Manifest{
		DublinCore: rc.DublinCore{
			Subject:    "Open Bible Stories",
			Identifier: "obs",
			Title:      "Test OBS",
			Issued:     "2024-01-01",
			Publisher:  "test",
			Rights:     "CC BY-SA 4.0",
			Language: rc.Language{
				Identifier: "en",
				Title:      "English",
				Direction:  "ltr",
			},
		},
		Projects: []rc.Project{
			{
				Identifier: "obs",
				Path:       ".",
				Sort:       0,
				Title:      "OBS",
			},
		},
	}

	h, err := handler.Lookup("Open Bible Stories")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Verify stories and subdirectory files are all present
	expectedContent := []string{
		"ingredients/content/01.md",
		"ingredients/content/02.md",
		"ingredients/content/front/intro.md",
		"ingredients/content/front/title.md",
		"ingredients/content/back/intro.md",
	}
	for _, key := range expectedContent {
		if _, ok := metadata.Ingredients[key]; !ok {
			t.Errorf("Expected ingredient %s not found", key)
		}
		if _, err := os.Stat(filepath.Join(outDir, key)); os.IsNotExist(err) {
			t.Errorf("Expected file %s not found on disk", key)
		}
	}

	// Verify excluded files are not in content
	if _, ok := metadata.Ingredients["ingredients/content/manifest.yaml"]; ok {
		t.Error("manifest.yaml should not be in content/ ingredients")
	}
	if _, ok := metadata.Ingredients["ingredients/content/LICENSE.md"]; ok {
		t.Error("LICENSE.md should not be in content/ ingredients")
	}
}

func TestOBS_RootLevelContent_ExcludesOnlyMetadataFiles(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create OBS content plus various file types in root.
	// The exclusion-based approach should only exclude *.yaml, README.md,
	// LICENSE.md, .gitignore, and dot-directories. Everything else is content.
	os.WriteFile(filepath.Join(inDir, "01.md"), []byte("# Story 1\n"), 0644)
	os.WriteFile(filepath.Join(inDir, "front.md"), []byte("# Front\n"), 0644)
	os.WriteFile(filepath.Join(inDir, "notes.md"), []byte("notes"), 0644)     // should be included
	os.WriteFile(filepath.Join(inDir, "extra.txt"), []byte("extra"), 0644)    // should be included
	os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte("License"), 0644) // excluded
	os.WriteFile(filepath.Join(inDir, "README.md"), []byte("# Readme"), 0644) // excluded
	os.WriteFile(filepath.Join(inDir, "manifest.yaml"), []byte("yaml"), 0644) // excluded
	os.WriteFile(filepath.Join(inDir, "media.yaml"), []byte("yaml"), 0644)    // excluded
	os.WriteFile(filepath.Join(inDir, ".gitignore"), []byte("*.tmp\n"), 0644) // excluded

	// Dot-directory should be excluded
	os.MkdirAll(filepath.Join(inDir, ".git"), 0755)
	os.WriteFile(filepath.Join(inDir, ".git", "config"), []byte("[core]\n"), 0644)

	manifest := &rc.Manifest{
		DublinCore: rc.DublinCore{
			Subject:    "Open Bible Stories",
			Identifier: "obs",
			Title:      "Test OBS",
			Issued:     "2024-01-01",
			Publisher:  "test",
			Rights:     "CC BY-SA 4.0",
			Language: rc.Language{
				Identifier: "en",
				Title:      "English",
				Direction:  "ltr",
			},
		},
		Projects: []rc.Project{
			{
				Identifier: "obs",
				Path:       ".",
				Sort:       0,
				Title:      "OBS",
			},
		},
	}

	h, err := handler.Lookup("Open Bible Stories")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Content files that should be included
	included := []string{
		"ingredients/content/01.md",
		"ingredients/content/front.md",
		"ingredients/content/notes.md",
		"ingredients/content/extra.txt",
	}
	for _, key := range included {
		if _, ok := metadata.Ingredients[key]; !ok {
			t.Errorf("Expected content ingredient %s not found", key)
		}
	}

	// Files that should be excluded from content/ ingredients
	excluded := []string{
		"ingredients/content/manifest.yaml",
		"ingredients/content/media.yaml",
		"ingredients/content/README.md",
		"ingredients/content/LICENSE.md",
		"ingredients/content/.gitignore",
	}
	for _, key := range excluded {
		if _, ok := metadata.Ingredients[key]; ok {
			t.Errorf("Excluded file should not be in ingredients: %s", key)
		}
	}

	// Dot-directory content should not appear
	for key := range metadata.Ingredients {
		if strings.Contains(key, ".git/") {
			t.Errorf(".git/ content should not be in ingredients: %s", key)
		}
	}
}

func TestOBS_ContentSubdirectory_StillWorks(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Standard OBS layout with content/ subdirectory (including front/ and back/ dirs)
	os.MkdirAll(filepath.Join(inDir, "content"), 0755)
	os.WriteFile(filepath.Join(inDir, "content", "01.md"), []byte("# Story 1\n"), 0644)
	os.MkdirAll(filepath.Join(inDir, "content", "front"), 0755)
	os.WriteFile(filepath.Join(inDir, "content", "front", "intro.md"), []byte("# Intro\n"), 0644)
	os.MkdirAll(filepath.Join(inDir, "content", "back"), 0755)
	os.WriteFile(filepath.Join(inDir, "content", "back", "intro.md"), []byte("# Back\n"), 0644)
	os.WriteFile(filepath.Join(inDir, "LICENSE.md"), []byte("License"), 0644)

	manifest := &rc.Manifest{
		DublinCore: rc.DublinCore{
			Subject:    "Open Bible Stories",
			Identifier: "obs",
			Title:      "Test OBS",
			Issued:     "2024-01-01",
			Publisher:  "test",
			Rights:     "CC BY-SA 4.0",
			Language: rc.Language{
				Identifier: "en",
				Title:      "English",
				Direction:  "ltr",
			},
		},
		Projects: []rc.Project{
			{
				Identifier: "obs",
				Path:       "./content",
				Sort:       0,
				Title:      "OBS",
			},
		},
	}

	h, err := handler.Lookup("Open Bible Stories")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	metadata, err := h.Convert(context.Background(), manifest, inDir, outDir, handler.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Verify standard content/ path still works with subdirectories
	expected := []string{
		"ingredients/content/01.md",
		"ingredients/content/front/intro.md",
		"ingredients/content/back/intro.md",
	}
	for _, key := range expected {
		if _, ok := metadata.Ingredients[key]; !ok {
			t.Errorf("%s should exist for ./content path", key)
		}
	}
}
