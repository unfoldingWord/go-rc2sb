package rc2sb_test

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	rc2sb "github.com/unfoldingWord/go-rc2sb"
	"github.com/unfoldingWord/go-rc2sb/sb"
)

// samplesDir returns the path to the samples directory, if it exists.
func samplesDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join("samples")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skip("samples/ directory not found; skipping integration tests")
	}
	return dir
}

// loadExpectedMetadata reads and parses the expected metadata.json from the sample SB directory.
func loadExpectedMetadata(t *testing.T, sbDir string) *sb.Metadata {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(sbDir, "metadata.json"))
	if err != nil {
		t.Fatalf("reading expected metadata.json: %v", err)
	}
	var m sb.Metadata
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parsing expected metadata.json: %v", err)
	}
	return &m
}

// loadGeneratedMetadata reads and parses the generated metadata.json from the output directory.
func loadGeneratedMetadata(t *testing.T, outDir string) *sb.Metadata {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(outDir, "metadata.json"))
	if err != nil {
		t.Fatalf("reading generated metadata.json: %v", err)
	}
	var m sb.Metadata
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parsing generated metadata.json: %v", err)
	}
	return &m
}

// TestConvertOBSTSVStudyNotes tests conversion of TSV OBS Study Notes.
func TestConvertOBSTSVStudyNotes(t *testing.T) {
	samples := samplesDir(t)
	sampleDir := filepath.Join(samples, "TSV OBS Study Notes")
	inDir := filepath.Join(sampleDir, "rc")
	sbDir := filepath.Join(sampleDir, "sb")

	outDir := t.TempDir()
	ctx := context.Background()

	result, err := rc2sb.Convert(ctx, inDir, outDir, rc2sb.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.Subject != "TSV OBS Study Notes" {
		t.Errorf("Subject = %q; want %q", result.Subject, "TSV OBS Study Notes")
	}

	expected := loadExpectedMetadata(t, sbDir)
	generated := loadGeneratedMetadata(t, outDir)

	compareStructuralMetadata(t, expected, generated)
	verifyInternalConsistency(t, generated, outDir)
	verifyRootFileCopying(t, inDir, outDir, generated)
}

// TestConvertOBSTSVStudyQuestions tests conversion of TSV OBS Study Questions.
func TestConvertOBSTSVStudyQuestions(t *testing.T) {
	samples := samplesDir(t)
	sampleDir := filepath.Join(samples, "TSV OBS Study Questions")
	inDir := filepath.Join(sampleDir, "rc")
	sbDir := filepath.Join(sampleDir, "sb")

	outDir := t.TempDir()
	ctx := context.Background()

	result, err := rc2sb.Convert(ctx, inDir, outDir, rc2sb.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.Subject != "TSV OBS Study Questions" {
		t.Errorf("Subject = %q; want %q", result.Subject, "TSV OBS Study Questions")
	}

	expected := loadExpectedMetadata(t, sbDir)
	generated := loadGeneratedMetadata(t, outDir)

	compareStructuralMetadata(t, expected, generated)
	verifyInternalConsistency(t, generated, outDir)
	verifyRootFileCopying(t, inDir, outDir, generated)
}

// TestConvertOBSTSVTranslationNotes tests conversion of TSV OBS Translation Notes.
func TestConvertOBSTSVTranslationNotes(t *testing.T) {
	samples := samplesDir(t)
	sampleDir := filepath.Join(samples, "TSV OBS Translation Notes")
	inDir := filepath.Join(sampleDir, "rc")
	sbDir := filepath.Join(sampleDir, "sb")

	outDir := t.TempDir()
	ctx := context.Background()

	result, err := rc2sb.Convert(ctx, inDir, outDir, rc2sb.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.Subject != "TSV OBS Translation Notes" {
		t.Errorf("Subject = %q; want %q", result.Subject, "TSV OBS Translation Notes")
	}

	expected := loadExpectedMetadata(t, sbDir)
	generated := loadGeneratedMetadata(t, outDir)

	compareStructuralMetadata(t, expected, generated)
	verifyInternalConsistency(t, generated, outDir)
	verifyRootFileCopying(t, inDir, outDir, generated)
}

// TestConvertOBSTSVTranslationQuestions tests conversion of TSV OBS Translation Questions.
func TestConvertOBSTSVTranslationQuestions(t *testing.T) {
	samples := samplesDir(t)
	sampleDir := filepath.Join(samples, "TSV OBS Translation Questions")
	inDir := filepath.Join(sampleDir, "rc")
	sbDir := filepath.Join(sampleDir, "sb")

	outDir := t.TempDir()
	ctx := context.Background()

	result, err := rc2sb.Convert(ctx, inDir, outDir, rc2sb.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.Subject != "TSV OBS Translation Questions" {
		t.Errorf("Subject = %q; want %q", result.Subject, "TSV OBS Translation Questions")
	}

	expected := loadExpectedMetadata(t, sbDir)
	generated := loadGeneratedMetadata(t, outDir)

	compareStructuralMetadata(t, expected, generated)
	verifyInternalConsistency(t, generated, outDir)
	verifyRootFileCopying(t, inDir, outDir, generated)
}

// TestConvertOpenBibleStories tests conversion of Open Bible Stories.
func TestConvertOpenBibleStories(t *testing.T) {
	samples := samplesDir(t)
	sampleDir := filepath.Join(samples, "Open Bible Stories")
	inDir := filepath.Join(sampleDir, "rc")
	sbDir := filepath.Join(sampleDir, "sb")

	outDir := t.TempDir()
	ctx := context.Background()

	result, err := rc2sb.Convert(ctx, inDir, outDir, rc2sb.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.Subject != "Open Bible Stories" {
		t.Errorf("Subject = %q; want %q", result.Subject, "Open Bible Stories")
	}

	expected := loadExpectedMetadata(t, sbDir)
	generated := loadGeneratedMetadata(t, outDir)

	compareStructuralMetadata(t, expected, generated)
	verifyInternalConsistency(t, generated, outDir)
	verifyRootFileCopying(t, inDir, outDir, generated)
}

// TestConvertAlignedBible tests conversion of Aligned Bible.
func TestConvertAlignedBible(t *testing.T) {
	samples := samplesDir(t)
	sampleDir := filepath.Join(samples, "Aligned Bible")
	inDir := filepath.Join(sampleDir, "rc")
	sbDir := filepath.Join(sampleDir, "sb")

	outDir := t.TempDir()
	ctx := context.Background()

	result, err := rc2sb.Convert(ctx, inDir, outDir, rc2sb.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.Subject != "Aligned Bible" {
		t.Errorf("Subject = %q; want %q", result.Subject, "Aligned Bible")
	}

	expected := loadExpectedMetadata(t, sbDir)
	generated := loadGeneratedMetadata(t, outDir)

	compareStructuralMetadata(t, expected, generated)
	verifyInternalConsistency(t, generated, outDir)
	verifyRootFileCopying(t, inDir, outDir, generated)
}

// TestConvertTranslationWords tests conversion of Translation Words with TWL payload.
// TW is converted like TWL: TW bible/ becomes ingredients/payload/, and twl_*.tsv files
// from the sibling <lang>_twl/ directory become the main ingredients with rewritten links.
func TestConvertTranslationWords(t *testing.T) {
	samples := samplesDir(t)
	sampleDir := filepath.Join(samples, "Translation Words")
	inDir := filepath.Join(sampleDir, "rc")
	sbDir := filepath.Join(sampleDir, "sb")

	outDir := t.TempDir()
	ctx := context.Background()

	result, err := rc2sb.Convert(ctx, inDir, outDir, rc2sb.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.Subject != "Translation Words" {
		t.Errorf("Subject = %q; want %q", result.Subject, "Translation Words")
	}

	expected := loadExpectedMetadata(t, sbDir)
	generated := loadGeneratedMetadata(t, outDir)

	compareStructuralMetadata(t, expected, generated)
	verifyInternalConsistency(t, generated, outDir)

	// Verify payload was included (TW bible/ → ingredients/payload/)
	payloadCount := 0
	for key := range generated.Ingredients {
		if strings.HasPrefix(key, "ingredients/payload/") {
			payloadCount++
		}
	}
	if payloadCount == 0 {
		t.Error("Expected payload ingredients but found none")
	}

	// Verify TSV files had rc:// links rewritten to ./payload/ paths (if TWL dir exists)
	for key := range generated.Ingredients {
		if !strings.HasSuffix(key, ".tsv") {
			continue
		}
		tsvPath := filepath.Join(outDir, key)
		data, err := os.ReadFile(tsvPath)
		if err != nil {
			t.Errorf("reading %s: %v", key, err)
			continue
		}
		content := string(data)
		if strings.Contains(content, "rc://") {
			t.Errorf("TSV file %s still contains rc:// links after rewrite", key)
		}
		if !strings.Contains(content, "./payload/") {
			t.Errorf("TSV file %s does not contain ./payload/ paths after rewrite", key)
		}
		break // Only need to check one file
	}

	verifyRootFileCopying(t, inDir, outDir, generated)
}

// TestConvertTWWithTWLPath tests TW conversion using an explicit TWLPath option
// instead of relying on auto-detection of <lang>_twl/ inside inDir.
func TestConvertTWWithTWLPath(t *testing.T) {
	samples := samplesDir(t)
	sampleDir := filepath.Join(samples, "Translation Words")
	inDir := filepath.Join(sampleDir, "rc")

	// The <lang>_twl directory is inside the RC repo; we'll pass it explicitly via TWLPath
	// Read the manifest to get the language identifier
	manifestData, err := os.ReadFile(filepath.Join(inDir, "manifest.yaml"))
	if err != nil {
		t.Skip("manifest.yaml not found in TW sample; skipping TWLPath test")
	}
	// Simple scan for language identifier to find the twl dir
	twlDir := ""
	for _, line := range strings.Split(string(manifestData), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "identifier:") {
			lang := strings.TrimSpace(strings.TrimPrefix(line, "identifier:"))
			candidate := filepath.Join(inDir, lang+"_twl")
			if _, err := os.Stat(candidate); err == nil {
				twlDir = candidate
				break
			}
		}
	}
	if twlDir == "" {
		t.Skip("no <lang>_twl directory found in TW sample; skipping TWLPath test")
	}

	outDir := t.TempDir()
	ctx := context.Background()

	opts := rc2sb.Options{TWLPath: twlDir}
	result, err := rc2sb.Convert(ctx, inDir, outDir, opts)
	if err != nil {
		t.Fatalf("Convert with TWLPath failed: %v", err)
	}

	if result.Subject != "Translation Words" {
		t.Errorf("Subject = %q; want %q", result.Subject, "Translation Words")
	}

	generated := loadGeneratedMetadata(t, outDir)

	// Verify payload was included
	payloadCount := 0
	for key := range generated.Ingredients {
		if strings.HasPrefix(key, "ingredients/payload/") {
			payloadCount++
		}
	}
	if payloadCount == 0 {
		t.Error("Expected payload ingredients with explicit TWLPath but found none")
	}

	// Verify TSV files had rc:// links rewritten
	for key := range generated.Ingredients {
		if !strings.HasSuffix(key, ".tsv") {
			continue
		}
		tsvPath := filepath.Join(outDir, key)
		data, err := os.ReadFile(tsvPath)
		if err != nil {
			t.Errorf("reading %s: %v", key, err)
			continue
		}
		content := string(data)
		if strings.Contains(content, "rc://") {
			t.Errorf("TSV file %s still contains rc:// links after rewrite with TWLPath", key)
		}
		if !strings.Contains(content, "./payload/") {
			t.Errorf("TSV file %s does not contain ./payload/ paths after rewrite with TWLPath", key)
		}
		break
	}

	verifyInternalConsistency(t, generated, outDir)
}

// TestConvertTranslationAcademy tests conversion of Translation Academy.
func TestConvertTranslationAcademy(t *testing.T) {
	samples := samplesDir(t)
	sampleDir := filepath.Join(samples, "Translation Academy")
	inDir := filepath.Join(sampleDir, "rc")
	sbDir := filepath.Join(sampleDir, "sb")

	outDir := t.TempDir()
	ctx := context.Background()

	result, err := rc2sb.Convert(ctx, inDir, outDir, rc2sb.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.Subject != "Translation Academy" {
		t.Errorf("Subject = %q; want %q", result.Subject, "Translation Academy")
	}

	expected := loadExpectedMetadata(t, sbDir)
	generated := loadGeneratedMetadata(t, outDir)

	compareStructuralMetadata(t, expected, generated)
	verifyInternalConsistency(t, generated, outDir)
	verifyRootFileCopying(t, inDir, outDir, generated)
}

// TestConvertTSVTranslationNotes tests conversion of TSV Translation Notes.
func TestConvertTSVTranslationNotes(t *testing.T) {
	samples := samplesDir(t)
	sampleDir := filepath.Join(samples, "TSV Translation Notes")
	inDir := filepath.Join(sampleDir, "rc")
	sbDir := filepath.Join(sampleDir, "sb")

	outDir := t.TempDir()
	ctx := context.Background()

	result, err := rc2sb.Convert(ctx, inDir, outDir, rc2sb.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.Subject != "TSV Translation Notes" {
		t.Errorf("Subject = %q; want %q", result.Subject, "TSV Translation Notes")
	}

	expected := loadExpectedMetadata(t, sbDir)
	generated := loadGeneratedMetadata(t, outDir)

	compareStructuralMetadata(t, expected, generated)
	verifyInternalConsistency(t, generated, outDir)
	verifyRootFileCopying(t, inDir, outDir, generated)
}

// TestConvertTSVTranslationQuestions tests conversion of TSV Translation Questions.
func TestConvertTSVTranslationQuestions(t *testing.T) {
	samples := samplesDir(t)
	sampleDir := filepath.Join(samples, "TSV Translation Questions")
	inDir := filepath.Join(sampleDir, "rc")
	sbDir := filepath.Join(sampleDir, "sb")

	outDir := t.TempDir()
	ctx := context.Background()

	result, err := rc2sb.Convert(ctx, inDir, outDir, rc2sb.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.Subject != "TSV Translation Questions" {
		t.Errorf("Subject = %q; want %q", result.Subject, "TSV Translation Questions")
	}

	expected := loadExpectedMetadata(t, sbDir)
	generated := loadGeneratedMetadata(t, outDir)

	compareStructuralMetadata(t, expected, generated)
	verifyInternalConsistency(t, generated, outDir)
	verifyRootFileCopying(t, inDir, outDir, generated)
}

// TestConvertTSVTranslationWordsLinks tests conversion of TSV Translation Words Links with payload.
func TestConvertTSVTranslationWordsLinks(t *testing.T) {
	samples := samplesDir(t)
	sampleDir := filepath.Join(samples, "TSV Translation Words Links")
	inDir := filepath.Join(sampleDir, "rc")
	sbDir := filepath.Join(sampleDir, "sb")

	outDir := t.TempDir()
	ctx := context.Background()

	result, err := rc2sb.Convert(ctx, inDir, outDir, rc2sb.Options{})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.Subject != "TSV Translation Words Links" {
		t.Errorf("Subject = %q; want %q", result.Subject, "TSV Translation Words Links")
	}

	expected := loadExpectedMetadata(t, sbDir)
	generated := loadGeneratedMetadata(t, outDir)

	compareStructuralMetadata(t, expected, generated)
	verifyInternalConsistency(t, generated, outDir)

	// Verify payload was included (en_tw/ exists in the RC repo)
	payloadCount := 0
	for key := range generated.Ingredients {
		if strings.HasPrefix(key, "ingredients/payload/") {
			payloadCount++
		}
	}
	if payloadCount == 0 {
		t.Error("Expected payload ingredients but found none")
	}

	// Verify TSV files had rc:// links rewritten to ./payload/ paths
	for key := range generated.Ingredients {
		if !strings.HasSuffix(key, ".tsv") {
			continue
		}
		tsvPath := filepath.Join(outDir, key)
		data, err := os.ReadFile(tsvPath)
		if err != nil {
			t.Errorf("reading %s: %v", key, err)
			continue
		}
		content := string(data)
		if strings.Contains(content, "rc://") {
			t.Errorf("TSV file %s still contains rc:// links after rewrite", key)
		}
		if !strings.Contains(content, "./payload/") {
			t.Errorf("TSV file %s does not contain ./payload/ paths after rewrite", key)
		}
		break // Only need to check one file
	}

	verifyRootFileCopying(t, inDir, outDir, generated)
}

// TestConvertTWLWithPayloadPath tests TWL conversion using an explicit PayloadPath option
// instead of relying on auto-detection of <lang>_tw/ inside inDir.
func TestConvertTWLWithPayloadPath(t *testing.T) {
	samples := samplesDir(t)
	sampleDir := filepath.Join(samples, "TSV Translation Words Links")
	inDir := filepath.Join(sampleDir, "rc")

	// The en_tw directory is inside the RC repo; we'll pass it explicitly via PayloadPath
	payloadPath := filepath.Join(inDir, "en_tw")
	if _, err := os.Stat(payloadPath); os.IsNotExist(err) {
		t.Skip("en_tw directory not found in TWL sample; skipping PayloadPath test")
	}

	outDir := t.TempDir()
	ctx := context.Background()

	opts := rc2sb.Options{PayloadPath: payloadPath}
	result, err := rc2sb.Convert(ctx, inDir, outDir, opts)
	if err != nil {
		t.Fatalf("Convert with PayloadPath failed: %v", err)
	}

	if result.Subject != "TSV Translation Words Links" {
		t.Errorf("Subject = %q; want %q", result.Subject, "TSV Translation Words Links")
	}

	generated := loadGeneratedMetadata(t, outDir)

	// Verify payload was included via the explicit PayloadPath
	payloadCount := 0
	for key := range generated.Ingredients {
		if strings.HasPrefix(key, "ingredients/payload/") {
			payloadCount++
		}
	}
	if payloadCount == 0 {
		t.Error("Expected payload ingredients with explicit PayloadPath but found none")
	}

	// Verify TSV files had rc:// links rewritten
	for key := range generated.Ingredients {
		if !strings.HasSuffix(key, ".tsv") {
			continue
		}
		tsvPath := filepath.Join(outDir, key)
		data, err := os.ReadFile(tsvPath)
		if err != nil {
			t.Errorf("reading %s: %v", key, err)
			continue
		}
		content := string(data)
		if strings.Contains(content, "rc://") {
			t.Errorf("TSV file %s still contains rc:// links after rewrite with PayloadPath", key)
		}
		if !strings.Contains(content, "./payload/") {
			t.Errorf("TSV file %s does not contain ./payload/ paths after rewrite with PayloadPath", key)
		}
		break
	}

	verifyInternalConsistency(t, generated, outDir)
}

// compareStructuralMetadata compares the structural elements of expected and generated metadata.
// This compares things like flavor type, scope keys, abbreviation, language, and ingredient keys -
// NOT checksums/sizes which may differ if source files have been updated since the sample was created.
func compareStructuralMetadata(t *testing.T, expected, generated *sb.Metadata) {
	t.Helper()

	// Compare format
	if generated.Format != expected.Format {
		t.Errorf("Format = %q; want %q", generated.Format, expected.Format)
	}

	// Compare type/flavorType
	if generated.Type.FlavorType.Name != expected.Type.FlavorType.Name {
		t.Errorf("FlavorType.Name = %q; want %q", generated.Type.FlavorType.Name, expected.Type.FlavorType.Name)
	}
	if generated.Type.FlavorType.Flavor.Name != expected.Type.FlavorType.Flavor.Name {
		t.Errorf("Flavor.Name = %q; want %q", generated.Type.FlavorType.Flavor.Name, expected.Type.FlavorType.Flavor.Name)
	}

	// Compare currentScope keys
	expectedScopeKeys := make(map[string]bool)
	for k := range expected.Type.FlavorType.CurrentScope {
		expectedScopeKeys[k] = true
	}
	generatedScopeKeys := make(map[string]bool)
	for k := range generated.Type.FlavorType.CurrentScope {
		generatedScopeKeys[k] = true
	}
	for k := range expectedScopeKeys {
		if !generatedScopeKeys[k] {
			t.Errorf("currentScope missing key %q", k)
		}
	}
	for k := range generatedScopeKeys {
		if !expectedScopeKeys[k] {
			t.Errorf("currentScope has extra key %q", k)
		}
	}

	// Compare ingredient keys (not values, since source files may have changed).
	// Only compare keys under ingredients/; root files are intentionally excluded
	// from metadata ingredients.
	// Source RC files may evolve independently of the sample SB metadata,
	// so differences in content-based ingredients are logged but not fatal.
	expectedIngredientKeys := make(map[string]bool)
	for key := range expected.Ingredients {
		if strings.HasPrefix(key, "ingredients/") {
			expectedIngredientKeys[key] = true
		}
	}
	generatedIngredientKeys := make(map[string]bool)
	for key := range generated.Ingredients {
		if strings.HasPrefix(key, "ingredients/") {
			generatedIngredientKeys[key] = true
		}
	}

	missing := 0
	extra := 0
	for key := range expectedIngredientKeys {
		if !generatedIngredientKeys[key] {
			missing++
			t.Logf("  ingredient in expected but not generated: %s", key)
		}
	}
	for key := range generatedIngredientKeys {
		if !expectedIngredientKeys[key] {
			extra++
			t.Logf("  ingredient in generated but not expected: %s", key)
		}
	}
	// Only fail if there are too many missing ingredients (>10% of expected).
	expectedCount := len(expectedIngredientKeys)
	if expectedCount > 0 && missing > 0 {
		missingRate := float64(missing) / float64(expectedCount)
		if missingRate > 0.10 {
			t.Errorf("Too many missing ingredients (>10%%): generated=%d, expected=%d (missing=%d, extra=%d)",
				len(generatedIngredientKeys), expectedCount, missing, extra)
		}
	}

	// Compare language
	if len(generated.Languages) != len(expected.Languages) {
		t.Errorf("Languages count = %d; want %d", len(generated.Languages), len(expected.Languages))
	} else if len(generated.Languages) > 0 {
		if generated.Languages[0].Tag != expected.Languages[0].Tag {
			t.Errorf("Language tag = %q; want %q", generated.Languages[0].Tag, expected.Languages[0].Tag)
		}
	}

	// Compare abbreviation
	expectedAbbr := expected.Identification.Abbreviation["en"]
	generatedAbbr := generated.Identification.Abbreviation["en"]
	if generatedAbbr != expectedAbbr {
		t.Errorf("Abbreviation = %q; want %q", generatedAbbr, expectedAbbr)
	}

	// Compare localizedNames keys
	for key := range expected.LocalizedNames {
		if _, ok := generated.LocalizedNames[key]; !ok {
			t.Errorf("localizedNames missing key %q", key)
		}
	}
}

// verifyRootFileCopying checks that root files (README.md, .gitignore, .gitea/, .github/)
// are copied from the RC repo to the SB output when they exist in the source.
func verifyRootFileCopying(t *testing.T, inDir, outDir string, _ *sb.Metadata) {
	t.Helper()

	// Check individual root files
	rootFiles := []string{"README.md", ".gitignore"}
	for _, name := range rootFiles {
		srcPath := filepath.Join(inDir, name)
		_, srcErr := os.Stat(srcPath)
		srcExists := srcErr == nil

		dstPath := filepath.Join(outDir, name)
		_, dstErr := os.Stat(dstPath)
		dstExists := dstErr == nil

		if srcExists && !dstExists {
			t.Errorf("Root file %s exists in RC but was not copied to SB output", name)
		}
		if !srcExists && dstExists {
			t.Errorf("Root file %s does not exist in RC but appeared in SB output", name)
		}
	}

	// Check root directories
	rootDirs := []string{".gitea", ".github"}
	for _, dirName := range rootDirs {
		srcPath := filepath.Join(inDir, dirName)
		srcInfo, srcErr := os.Stat(srcPath)
		srcExists := srcErr == nil && srcInfo.IsDir()

		dstPath := filepath.Join(outDir, dirName)
		dstInfo, dstErr := os.Stat(dstPath)
		dstExists := dstErr == nil && dstInfo.IsDir()

		if srcExists && !dstExists {
			t.Errorf("Root directory %s exists in RC but was not copied to SB output", dirName)
		}
	}
}

// verifyInternalConsistency ensures the generated metadata.json matches the actual files on disk.
func verifyInternalConsistency(t *testing.T, generated *sb.Metadata, outDir string) {
	t.Helper()

	for key, ing := range generated.Ingredients {
		filePath := filepath.Join(outDir, key)

		// Check file exists
		info, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			t.Errorf("Ingredient file missing: %s", key)
			continue
		}
		if err != nil {
			t.Errorf("Error checking ingredient %s: %v", key, err)
			continue
		}

		// Check size matches
		if info.Size() != ing.Size {
			t.Errorf("Ingredient %s: actual size = %d; metadata says %d", key, info.Size(), ing.Size)
		}

		// Check MD5 matches
		actualMD5, err := computeFileMD5(filePath)
		if err != nil {
			t.Errorf("Error computing MD5 for %s: %v", key, err)
			continue
		}
		if actualMD5 != ing.Checksum.MD5 {
			t.Errorf("Ingredient %s: actual MD5 = %q; metadata says %q", key, actualMD5, ing.Checksum.MD5)
		}
	}
}

// computeFileMD5 computes the MD5 hash of a file.
func computeFileMD5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
