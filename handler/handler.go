// Package handler provides the interface and registry for subject-specific RC-to-SB conversion handlers.
package handler

import (
	"context"

	"github.com/unfoldingWord/go-rc2sb/rc"
	"github.com/unfoldingWord/go-rc2sb/sb"
)

// Options holds conversion options passed to handlers.
type Options struct {
	// PayloadPath is the path to a Translation Words directory for TWL conversion.
	// See rc2sb.Options.PayloadPath for details.
	PayloadPath string

	// TWLPath is the path to a TSV Translation Words Links directory for TW conversion.
	// See rc2sb.Options.TWLPath for details.
	TWLPath string

	// USFMPath is the path to a directory containing USFM files for localized book names.
	// See rc2sb.Options.USFMPath for details.
	USFMPath string

	// Owner is the repository owner (user or organization) on DCS, used to build
	// the primary identifier "owner/repo". See rc2sb.Options.Owner for details.
	Owner string

	// RepoName is the repository name on DCS (e.g., "en_tw").
	// See rc2sb.Options.RepoName for details.
	RepoName string

	// DCSURL is the base URL of the DCS instance acting as the ID authority.
	// See rc2sb.Options.DCSURL for details.
	DCSURL string

	// Revision identifies the converted state of the source repository.
	// See rc2sb.Options.Revision for details.
	Revision string
}

// Handler is the interface that each subject-specific converter implements.
type Handler interface {
	// Subject returns the RC subject string this handler supports.
	Subject() string

	// Convert performs the conversion from RC to SB.
	// It reads from inDir (RC repo), writes files to outDir (SB output),
	// and returns the SB metadata to be written as metadata.json.
	Convert(ctx context.Context, manifest *rc.Manifest, inDir, outDir string, opts Options) (*sb.Metadata, error)
}
