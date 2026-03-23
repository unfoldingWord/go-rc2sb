package rc2sb

// Options configures the RC to SB conversion.
type Options struct {
	// PayloadPath is the path to a Translation Words directory (e.g., "/path/to/en_tw")
	// used when converting TSV Translation Words Links repos.
	// If set, the bible/ subdirectory within this path is copied to ingredients/payload/
	// in the SB output, and rc:// links in the TWL TSV files are rewritten to
	// relative ./payload/ paths.
	//
	// If empty, the TWL handler auto-detects a <lang>_tw/ subdirectory inside
	// the input RC repo directory (where <lang> is the manifest's language identifier).
	// If neither is found, no payload is created and TSV files are copied as-is.
	PayloadPath string

	// TWLPath is the path to a TSV Translation Words Links directory (e.g., "/path/to/en_twl")
	// used when converting Translation Words repos.
	// If set, the twl_*.tsv files within this path are processed as the main ingredients,
	// rc:// links in the TSV files are rewritten to relative ./payload/ paths, and the
	// TW bible/ content is copied to ingredients/payload/.
	//
	// If empty, the TW handler auto-detects a <lang>_twl/ subdirectory inside
	// the input RC repo directory (where <lang> is the manifest's language identifier).
	// If neither is found, no TSV ingredients are created; only the payload is written.
	TWLPath string

	// USFMPath is the path to a directory containing USFM files for localized
	// Bible book names. This is used by TSV handlers (TN, TQ, TWL) to extract
	// \toc1, \toc2, \toc3 markers for book names in the target language.
	//
	// For Bible/USFM handlers, the USFM files in the input RC repo are used
	// directly, so this option is not needed.
	//
	// If empty, TSV handlers will use project titles from the manifest,
	// falling back to English names from the books package.
	USFMPath string
}

// Result holds information about a completed conversion.
type Result struct {
	// Subject is the RC subject that was converted.
	Subject string

	// Identifier is the RC identifier (e.g., "obs", "ult", "tn").
	Identifier string

	// InDir is the input RC directory that was converted.
	InDir string

	// OutDir is the output SB directory that was created.
	OutDir string

	// Ingredients is the number of ingredient files in the SB output.
	Ingredients int
}
