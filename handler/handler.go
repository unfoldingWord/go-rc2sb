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
