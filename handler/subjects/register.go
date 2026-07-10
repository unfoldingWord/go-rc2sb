// Package subjects registers all subject handlers with the handler registry.
// Import this package to make all handlers available.
package subjects

import (
	"github.com/unfoldingWord/go-rc2sb/handler"
)

func init() {
	// Register all subject handlers

	// Open Bible Stories
	handler.Register(handler.NewOBSHandler())

	// Bible / USFM handlers (all share the same conversion logic)
	handler.Register(handler.NewBibleHandler("Aligned Bible"))
	handler.Register(handler.NewBibleHandler("Bible"))
	handler.Register(handler.NewBibleHandler("Hebrew Old Testament"))
	handler.Register(handler.NewBibleHandler("Greek New Testament"))

	// Translation Words
	handler.Register(handler.NewTWHandler())

	// Translation Academy
	handler.Register(handler.NewTAHandler())

	// Bible-based TSV variants (one TSV file per book, shared conversion logic)
	handler.Register(handler.NewBibleTSVHandler(
		"TSV Translation Notes",
		"x-bcvnotes",
		"TN",
		"tn_",
	))
	handler.Register(handler.NewBibleTSVHandler(
		"TSV Translation Questions",
		"x-bcvquestions",
		"TQ",
		"tq_",
	))
	handler.Register(handler.NewBibleTSVHandler(
		"TSV Study Notes",
		"x-bcvnotes",
		"SN",
		"sn_",
	))
	handler.Register(handler.NewBibleTSVHandler(
		"TSV Study Questions",
		"x-bcvquestions",
		"SQ",
		"sq_",
	))

	// TSV Translation Words Links
	handler.Register(handler.NewTWLHandler())

	// OBS TSV variants
	handler.Register(handler.NewOBSTSVHandler(
		"TSV OBS Study Notes",
		"x-obsnotes",
		"OBSSN",
		"sn_",
	))
	handler.Register(handler.NewOBSTSVHandler(
		"TSV OBS Study Questions",
		"x-obsquestions",
		"OBSSQ",
		"sq_",
	))
	handler.Register(handler.NewOBSTSVHandler(
		"TSV OBS Translation Notes",
		"x-obsnotes",
		"OBSTN",
		"tn_",
	))
	handler.Register(handler.NewOBSTSVHandler(
		"TSV OBS Translation Questions",
		"x-obsquestions",
		"OBSTQ",
		"tq_",
	))

	// TSV OBS Translation Words Links — TWL-style payload + rc:// link rewriting
	// (parascriptural/x-bcvarticles), but with OBS treated as a single book.
	handler.Register(handler.NewOBSTWLHandler())
}
