package rc2sb

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/unfoldingWord/go-rc2sb/handler"
)

// resolveRepoIdentity fills the Owner, RepoName, DCSURL, and Revision handler
// options, preferring explicit Options values, then the git clone metadata of
// the input directory (remote "origin" URL and HEAD commit). Anything still
// unresolved is left empty for the manifest-based fallback in
// handler.BuildBaseMetadata.
func resolveRepoIdentity(inDir string, opts Options, ho *handler.Options) {
	ho.Owner = opts.Owner
	ho.RepoName = opts.RepoName
	ho.DCSURL = opts.DCSURL
	ho.Revision = opts.Revision

	if ho.Revision == "" {
		ho.Revision = gitHeadCommit(inDir)
	}
	if ho.Owner != "" && ho.RepoName != "" && ho.DCSURL != "" {
		return
	}

	baseURL, owner, repo, ok := parseGitRemote(gitOriginURL(inDir))
	if !ok {
		return
	}
	if ho.Owner == "" {
		ho.Owner = owner
	}
	if ho.RepoName == "" {
		ho.RepoName = repo
	}
	if ho.DCSURL == "" {
		ho.DCSURL = baseURL
	}
}

// gitHeadCommit returns the commit SHA that dir/.git/HEAD points to, following
// a symbolic ref through loose refs and packed-refs. Returns "" if dir is not
// a git clone or the SHA cannot be determined.
func gitHeadCommit(dir string) string {
	gitDir := filepath.Join(dir, ".git")
	head, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
	if err != nil {
		return ""
	}

	ref := strings.TrimSpace(string(head))
	if !strings.HasPrefix(ref, "ref: ") {
		// Detached HEAD holds the SHA directly.
		if isCommitSHA(ref) {
			return ref
		}
		return ""
	}
	ref = strings.TrimSpace(strings.TrimPrefix(ref, "ref: "))

	// Loose ref file, e.g. .git/refs/heads/master.
	if data, err := os.ReadFile(filepath.Join(gitDir, filepath.FromSlash(ref))); err == nil {
		if sha := strings.TrimSpace(string(data)); isCommitSHA(sha) {
			return sha
		}
		return ""
	}

	// Packed refs: lines of "<sha> <ref>", plus "#" comments and "^" peeled tags.
	f, err := os.Open(filepath.Join(gitDir, "packed-refs"))
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "^") {
			continue
		}
		if sha, name, found := strings.Cut(line, " "); found && name == ref && isCommitSHA(sha) {
			return sha
		}
	}
	return ""
}

// isCommitSHA reports whether s looks like a full hex commit hash
// (40 chars for SHA-1, 64 for SHA-256 repos).
func isCommitSHA(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// gitOriginURL reads the remote "origin" URL from dir/.git/config.
// Returns "" if dir is not a git clone or no origin remote is configured.
func gitOriginURL(dir string) string {
	f, err := os.Open(filepath.Join(dir, ".git", "config"))
	if err != nil {
		return ""
	}
	defer f.Close()

	inOrigin := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") {
			inOrigin = line == `[remote "origin"]`
			continue
		}
		if !inOrigin {
			continue
		}
		if key, value, found := strings.Cut(line, "="); found && strings.TrimSpace(key) == "url" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// parseGitRemote extracts the base URL, owner, and repository name from a git
// remote URL. Supported forms:
//
//	https://git.door43.org/owner/repo.git
//	http://git.door43.org/owner/repo
//	ssh://git@git.door43.org/owner/repo.git
//	git@git.door43.org:owner/repo.git
//
// SSH forms yield an https base URL, since the ID authority URL must be
// web-addressable.
func parseGitRemote(url string) (baseURL, owner, repo string, ok bool) {
	if url == "" {
		return "", "", "", false
	}

	var host, path string
	switch {
	case strings.HasPrefix(url, "https://"), strings.HasPrefix(url, "http://"):
		scheme, rest, _ := strings.Cut(url, "://")
		host, path, _ = strings.Cut(rest, "/")
		baseURL = scheme + "://" + host
	case strings.HasPrefix(url, "ssh://"):
		rest := strings.TrimPrefix(url, "ssh://")
		host, path, _ = strings.Cut(rest, "/")
		if _, h, found := strings.Cut(host, "@"); found {
			host = h
		}
		host = strings.SplitN(host, ":", 2)[0] // drop optional port
		baseURL = "https://" + host
	case strings.Contains(url, "@") && strings.Contains(url, ":"):
		// scp-like syntax: git@host:owner/repo.git
		_, rest, _ := strings.Cut(url, "@")
		host, path, _ = strings.Cut(rest, ":")
		baseURL = "https://" + host
	default:
		return "", "", "", false
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if host == "" || len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", false
	}
	return baseURL, parts[0], strings.TrimSuffix(parts[1], ".git"), true
}
