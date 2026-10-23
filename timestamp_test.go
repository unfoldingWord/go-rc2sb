package rc2sb

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestConvertTimestamp checks that a set Timestamp makes a revision convert to the same metadata.json.
func TestConvertTimestamp(t *testing.T) {
	inDir := t.TempDir()
	manifest := "dublin_core:\n  conformsto: rc0.2\n  identifier: tn\n  language: {identifier: en, title: English, direction: ltr}\n" +
		"  subject: TSV Translation Notes\n  title: Test Notes\nprojects:\n  - {identifier: tit, path: ./tn_TIT.tsv, title: Titus}\n"
	notes := "Reference\tID\tTags\tSupportReference\tQuote\tOccurrence\tNote\n1:1\tabcd\t\t\t\t0\tA note\n"
	for name, content := range map[string]string{"manifest.yaml": manifest, "tn_TIT.tsv": notes} {
		if err := os.WriteFile(filepath.Join(inDir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	stamp := time.Date(2026, 10, 22, 15, 6, 43, 0, time.FixedZone("CDT", -5*60*60))
	var outputs [][]byte
	for range 2 {
		outDir := t.TempDir()
		if _, err := Convert(context.Background(), inDir, outDir, Options{Timestamp: stamp}); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(outDir, "metadata.json"))
		if err != nil {
			t.Fatal(err)
		}
		outputs = append(outputs, data)
		time.Sleep(2 * time.Millisecond) // a clock-based stamp would differ between the two conversions
	}
	if !bytes.Equal(outputs[0], outputs[1]) {
		t.Errorf("metadata.json differs between two conversions with the same Timestamp")
	}
	if want := `"2026-10-22T20:06:43.000Z"`; strings.Count(string(outputs[0]), want) != 2 {
		t.Errorf("metadata.json should stamp %s as dateCreated and the identifier's timestamp:\n%s", want, outputs[0])
	}
}

// TestConvertUpstream checks that Options.Upstream is recorded under identification.upstream.
func TestConvertUpstream(t *testing.T) {
	inDir := t.TempDir()
	manifest := "dublin_core:\n  conformsto: rc0.2\n  identifier: tn\n  language: {identifier: en, title: English, direction: ltr}\n" +
		"  subject: TSV Translation Notes\n  title: Test Notes\nprojects:\n  - {identifier: tit, path: ./tn_TIT.tsv, title: Titus}\n"
	for name, content := range map[string]string{"manifest.yaml": manifest, "tn_TIT.tsv": "Reference\tNote\n1:1\tA note\n"} {
		if err := os.WriteFile(filepath.Join(inDir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	stamp := time.Date(2026, 10, 22, 20, 6, 43, 0, time.UTC)

	read := func(opts Options) map[string]any {
		outDir := t.TempDir()
		if _, err := Convert(context.Background(), inDir, outDir, opts); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(outDir, "metadata.json"))
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatal(err)
		}
		return m["identification"].(map[string]any)
	}

	if _, ok := read(Options{Timestamp: stamp})["upstream"]; ok {
		t.Errorf("identification.upstream should be left out without Options.Upstream")
	}

	identification := read(Options{Timestamp: stamp, Upstream: []Upstream{
		{Owner: "unfoldingWord", RepoName: "en_tw", Revision: "1111111111111111111111111111111111111111", Timestamp: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)},
		{Owner: "unfoldingWord", RepoName: "en_twl", Revision: "2222222222222222222222222222222222222222"},
	}})
	got, _ := json.Marshal(identification["upstream"])
	want := `{"dcs":[{"unfoldingWord/en_tw":{"revision":"1111111111111111111111111111111111111111","timestamp":"2026-10-01T12:00:00.000Z"}},` +
		`{"unfoldingWord/en_twl":{"revision":"2222222222222222222222222222222222222222","timestamp":"2026-10-22T20:06:43.000Z"}}]}`
	if string(got) != want {
		t.Errorf("identification.upstream = %s; want %s", got, want)
	}
}
