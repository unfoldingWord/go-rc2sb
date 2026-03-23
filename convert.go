// Package rc2sb converts Resource Container (RC) repositories to Scripture Burrito (SB) format.
package rc2sb

import (
	"context"
	"fmt"
	"os"

	"github.com/unfoldingWord/go-rc2sb/handler"
	"github.com/unfoldingWord/go-rc2sb/rc"

	// Import all handlers to register them.
	_ "github.com/unfoldingWord/go-rc2sb/handler/subjects"
)

// Convert converts an RC repository at inDir to SB format, writing output to outDir.
func Convert(ctx context.Context, inDir string, outDir string, opts Options) (Result, error) {
	// Check context
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf("context error: %w", err)
	}

	// Load the RC manifest
	manifest, err := rc.LoadManifest(inDir)
	if err != nil {
		return Result{}, err
	}

	subject := manifest.DublinCore.Subject

	// Look up the handler for this subject
	h, err := handler.Lookup(subject)
	if err != nil {
		return Result{}, err
	}

	// Ensure the output directory exists
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return Result{}, fmt.Errorf("creating output directory: %w", err)
	}

	// Run the handler
	handlerOpts := handler.Options{
		PayloadPath: opts.PayloadPath,
		TWLPath:     opts.TWLPath,
		USFMPath:    opts.USFMPath,
	}
	metadata, err := h.Convert(ctx, manifest, inDir, outDir, handlerOpts)
	if err != nil {
		return Result{}, fmt.Errorf("converting %s: %w", subject, err)
	}

	// Write metadata.json
	if err := metadata.WriteToFile(outDir); err != nil {
		return Result{}, err
	}

	return Result{
		Subject:     subject,
		Identifier:  manifest.DublinCore.Identifier,
		InDir:       inDir,
		OutDir:      outDir,
		Ingredients: len(metadata.Ingredients),
	}, nil
}
