// Command rc2sb converts a Resource Container (RC) repository to Scripture Burrito (SB) format.
//
// Usage:
//
//	rc2sb [flags] <inDir> <outDir>
//	rc2sb --payload /path/to/en_tw <inDir> <outDir>
//	rc2sb --usfm /path/to/en_ult <inDir> <outDir>
//
// Flags:
//
//	--payload <dir>   Path to a Translation Words directory (e.g., en_tw) for TWL payload creation.
//	                  If not set, auto-detects <lang>_tw/ inside inDir.
//	--usfm <dir>      Path to a USFM directory for localized Bible book names in TSV repos.
//	                  If not set, uses manifest project titles, then English fallback.
//	--owner <name>    Repository owner on DCS for the primary identifier (owner/repo).
//	                  If not set, detects from the git remote, then manifest publisher.
//	--repo <name>     Repository name on DCS (e.g., en_tw) for the primary identifier.
//	                  If not set, detects from the git remote, then <lang>_<identifier>.
//	--dcs-url <url>   Base URL of the DCS instance used as the ID authority.
//	                  If not set, detects from the git remote, then https://git.door43.org.
//	--revision <rev>  Revision of the converted source (release number, tag, or commit SHA).
//	                  If not set, uses the git HEAD commit, then the manifest version, then "1".
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	rc2sb "github.com/unfoldingWord/go-rc2sb"
)

func main() {
	payload := flag.String("payload", "", "path to a Translation Words directory (e.g., en_tw) for TWL payload creation")
	usfm := flag.String("usfm", "", "path to a USFM directory for localized Bible book names in TSV repos")
	owner := flag.String("owner", "", "repository owner on DCS for the primary identifier (owner/repo)")
	repoName := flag.String("repo", "", "repository name on DCS (e.g., en_tw) for the primary identifier")
	dcsURL := flag.String("dcs-url", "", "base URL of the DCS instance used as the ID authority")
	revision := flag.String("revision", "", "revision of the converted source (release number, tag, or commit SHA)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: rc2sb [flags] <inDir> <outDir>\n\n")
		fmt.Fprintf(os.Stderr, "Converts a Resource Container (RC) repository to Scripture Burrito (SB) format.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  inDir    Path to the RC repository (must contain manifest.yaml)\n")
		fmt.Fprintf(os.Stderr, "  outDir   Path where SB output will be written\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	inDir := flag.Arg(0)
	outDir := flag.Arg(1)

	opts := rc2sb.Options{
		PayloadPath: *payload,
		USFMPath:    *usfm,
		Owner:       *owner,
		RepoName:    *repoName,
		DCSURL:      *dcsURL,
		Revision:    *revision,
	}

	result, err := rc2sb.Convert(context.Background(), inDir, outDir, opts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Converted %s (%s) with %d ingredients\n",
		result.Subject, result.Identifier, result.Ingredients)
}
