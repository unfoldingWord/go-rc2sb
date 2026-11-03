package rc2sb

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveSubject(t *testing.T) {
	for _, tc := range []struct {
		subject  string
		projects []string
		want     string
	}{
		{"Study Notes", []string{"obs"}, "OBS Study Notes"},
		{"Study Notes", []string{"gen", "exo"}, "Study Notes"},
		{"TSV Translation Notes", []string{"obs"}, "TSV OBS Translation Notes"},
		{"TSV Translation Words Links", []string{"OBS"}, "TSV OBS Translation Words Links"},
		{"Translation Questions", nil, "Translation Questions"},
		{"Bible translation comprehension questions", []string{"gen"}, "Translation Questions"},
		{"Bible translation comprehension questions", []string{"obs"}, "OBS Translation Questions"},
		{"Translator Notes", []string{"tit"}, "Translation Notes"},
		{"bible stories", []string{"obs"}, "Open Bible Stories"},
		{"Open Bible Stories", []string{"obs"}, "Open Bible Stories"},
		{"Greek Lexicon", []string{"ugl"}, "Greek Lexicon"},
		{"Open Bible stories", []string{"obs"}, "Open Bible Stories"},
		{"translation notes", []string{"obs"}, "OBS Translation Notes"},
		{"TRANSLATOR notes", []string{"tit"}, "Translation Notes"},
		{"bible Translation Comprehension Questions", []string{"gen"}, "Translation Questions"},
	} {
		if got := ResolveSubject(tc.subject, tc.projects); got != tc.want {
			t.Errorf("ResolveSubject(%q, %q) = %q, want %q", tc.subject, tc.projects, got, tc.want)
		}
	}
	for _, tc := range []struct {
		subject, identifier string
		want                string
	}{
		{"", "ta", "Translation Academy"},
		{"  ", "TA", "Translation Academy"},
		{"", "obs-tn", "OBS Translation Notes"},
		{"Translation Words", "ta", "Translation Words"},
		{"", "ugl", ""},
	} {
		if got := ResolveRCSubject(tc.subject, tc.identifier, nil); got != tc.want {
			t.Errorf("ResolveRCSubject(%q, %q) = %q, want %q", tc.subject, tc.identifier, got, tc.want)
		}
	}
	for _, s := range SupportedSubjects() {
		if got := ResolveSubject(s, []string{"gen"}); got != s {
			t.Errorf("ResolveSubject(%q) = %q, want it unchanged", s, got)
		}
	}
	if !Supports("OBS Study Notes") || !Supports("Study Notes") || Supports("Greek Lexicon") {
		t.Error("Supports does not match the registered handlers")
	}
}

// TestConvert_StudyNotesOfOBS converts an RC like Bussard/fr_obs-sn, whose
// subject is "Study Notes" but whose only project is OBS Markdown frame notes.
func TestConvert_StudyNotesOfOBS(t *testing.T) {
	inDir, outDir := t.TempDir(), t.TempDir()
	write := func(path, content string) {
		path = filepath.Join(inDir, path)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("manifest.yaml", `dublin_core:
  conformsto: rc0.2
  identifier: fr_obs-sn
  subject: 'Study Notes'
  title: OBS Notes d'étude
  language: {identifier: fr, title: French, direction: ltr}
projects:
  - {identifier: obs, title: OBS Notes d'étude, path: ./content}
`)
	write("content/01/01.md", "# Dieu\n\nLe créateur.\n")

	if _, err := Convert(context.Background(), inDir, outDir, Options{Owner: "Bussard", RepoName: "fr_obs-sn", Revision: "1"}); err != nil {
		t.Fatalf("Convert: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(outDir, "metadata.json"))
	if err != nil {
		t.Fatal(err)
	}
	var md struct {
		Type struct {
			FlavorType struct {
				Name   string `json:"name"`
				Flavor struct {
					Name string `json:"name"`
				} `json:"flavor"`
			} `json:"flavorType"`
		} `json:"type"`
	}
	if err := json.Unmarshal(data, &md); err != nil {
		t.Fatal(err)
	}
	if md.Type.FlavorType.Name != "peripheral" || md.Type.FlavorType.Flavor.Name != "x-obsnotes" {
		t.Errorf("flavor = %s/%s, want peripheral/x-obsnotes", md.Type.FlavorType.Name, md.Type.FlavorType.Flavor.Name)
	}
	if _, err := os.Stat(filepath.Join(outDir, "ingredients", "OBS.tsv")); err != nil {
		t.Errorf("ingredients/OBS.tsv: %v", err)
	}
}
