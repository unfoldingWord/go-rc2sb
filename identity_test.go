package rc2sb

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGitRemote(t *testing.T) {
	tests := []struct {
		url     string
		baseURL string
		owner   string
		repo    string
		ok      bool
	}{
		{"https://git.door43.org/unfoldingWord/en_tw.git", "https://git.door43.org", "unfoldingWord", "en_tw", true},
		{"https://git.door43.org/unfoldingWord/en_tw", "https://git.door43.org", "unfoldingWord", "en_tw", true},
		{"http://qa.door43.org/Door43-Catalog/en_obs.git", "http://qa.door43.org", "Door43-Catalog", "en_obs", true},
		{"ssh://git@git.door43.org/unfoldingWord/en_ult.git", "https://git.door43.org", "unfoldingWord", "en_ult", true},
		{"ssh://git@git.door43.org:22/unfoldingWord/en_ult.git", "https://git.door43.org", "unfoldingWord", "en_ult", true},
		{"git@git.door43.org:unfoldingWord/en_tn.git", "https://git.door43.org", "unfoldingWord", "en_tn", true},
		{"", "", "", "", false},
		{"https://git.door43.org/", "", "", "", false},
		{"https://git.door43.org/onlyowner", "", "", "", false},
		{"/local/path/repo", "", "", "", false},
	}
	for _, tt := range tests {
		baseURL, owner, repo, ok := parseGitRemote(tt.url)
		if baseURL != tt.baseURL || owner != tt.owner || repo != tt.repo || ok != tt.ok {
			t.Errorf("parseGitRemote(%q) = (%q, %q, %q, %v); want (%q, %q, %q, %v)",
				tt.url, baseURL, owner, repo, ok, tt.baseURL, tt.owner, tt.repo, tt.ok)
		}
	}
}

func TestGitOriginURL(t *testing.T) {
	dir := t.TempDir()
	if got := gitOriginURL(dir); got != "" {
		t.Errorf("gitOriginURL on non-clone = %q; want \"\"", got)
	}

	gitDir := filepath.Join(dir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	config := `[core]
	repositoryformatversion = 0
[remote "upstream"]
	url = https://example.org/other/repo.git
[remote "origin"]
	url = https://git.door43.org/unfoldingWord/en_tw.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`
	if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}
	want := "https://git.door43.org/unfoldingWord/en_tw.git"
	if got := gitOriginURL(dir); got != want {
		t.Errorf("gitOriginURL = %q; want %q", got, want)
	}
}

func TestGitHeadCommit(t *testing.T) {
	looseSHA := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	packedSHA := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	detachedSHA := "cccccccccccccccccccccccccccccccccccccccc"

	writeGit := func(t *testing.T, files map[string]string) string {
		t.Helper()
		dir := t.TempDir()
		for name, content := range files {
			path := filepath.Join(dir, ".git", filepath.FromSlash(name))
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}

	if got := gitHeadCommit(t.TempDir()); got != "" {
		t.Errorf("gitHeadCommit on non-clone = %q; want \"\"", got)
	}

	dir := writeGit(t, map[string]string{
		"HEAD":              "ref: refs/heads/master\n",
		"refs/heads/master": looseSHA + "\n",
	})
	if got := gitHeadCommit(dir); got != looseSHA {
		t.Errorf("gitHeadCommit loose ref = %q; want %q", got, looseSHA)
	}

	dir = writeGit(t, map[string]string{
		"HEAD": "ref: refs/heads/master\n",
		"packed-refs": "# pack-refs with: peeled fully-peeled sorted\n" +
			packedSHA + " refs/heads/master\n" +
			"^dddddddddddddddddddddddddddddddddddddddd\n",
	})
	if got := gitHeadCommit(dir); got != packedSHA {
		t.Errorf("gitHeadCommit packed ref = %q; want %q", got, packedSHA)
	}

	dir = writeGit(t, map[string]string{
		"HEAD": detachedSHA + "\n",
	})
	if got := gitHeadCommit(dir); got != detachedSHA {
		t.Errorf("gitHeadCommit detached = %q; want %q", got, detachedSHA)
	}
}
