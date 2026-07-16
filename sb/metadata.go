// Package sb provides types for Scripture Burrito (SB) metadata and ingredient computation.
package sb

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Metadata represents the top-level structure of an SB metadata.json file.
type Metadata struct {
	Format         string                     `json:"format"`
	Meta           Meta                       `json:"meta"`
	IDAuthorities  map[string]IDAuthority     `json:"idAuthorities"`
	Identification Identification             `json:"identification"`
	Languages      []LanguageEntry            `json:"languages"`
	Type           Type                       `json:"type"`
	Confidential   bool                       `json:"confidential"`
	LocalizedNames map[string]LocalizedName   `json:"localizedNames,omitempty"`
	Ingredients    map[string]Ingredient      `json:"ingredients"`
	Copyright      Copyright                  `json:"copyright"`
}

// Meta holds the meta section of an SB metadata file.
type Meta struct {
	Version       string    `json:"version"`
	Category      string    `json:"category"`
	Generator     Generator `json:"generator"`
	DefaultLocale string    `json:"defaultLocale"`
	DateCreated   string    `json:"dateCreated"`
	Normalization string    `json:"normalization"`
}

// Generator identifies the software that created the SB.
type Generator struct {
	SoftwareName    string `json:"softwareName"`
	SoftwareVersion string `json:"softwareVersion"`
	UserName        string `json:"userName"`
}

// IDAuthority represents an ID authority entry.
type IDAuthority struct {
	ID   string            `json:"id"`
	Name map[string]string `json:"name"`
}

// Identification holds the identification section.
type Identification struct {
	Primary      map[string]map[string]PrimaryEntry `json:"primary"`
	Name         map[string]string                  `json:"name"`
	Description  map[string]string                  `json:"description"`
	Abbreviation map[string]string                  `json:"abbreviation"`
}

// PrimaryEntry holds a primary identification entry.
type PrimaryEntry struct {
	Revision  string `json:"revision"`
	Timestamp string `json:"timestamp"`
}

// LanguageEntry describes a language in the SB metadata.
type LanguageEntry struct {
	Tag             string            `json:"tag"`
	Name            map[string]string `json:"name"`
	ScriptDirection string            `json:"scriptDirection"`
}

// Type holds the type section with flavorType.
type Type struct {
	FlavorType FlavorType `json:"flavorType"`
}

// FlavorType describes the type and flavor of the SB.
type FlavorType struct {
	Name         string                       `json:"name"`
	Flavor       Flavor                       `json:"flavor"`
	CurrentScope map[string][]string          `json:"currentScope,omitempty"`
}

// Flavor holds the flavor details. Fields vary by type.
type Flavor struct {
	Name            string `json:"name"`
	USFMVersion     string `json:"usfmVersion,omitempty"`
	TranslationType string `json:"translationType,omitempty"`
	Audience        string `json:"audience,omitempty"`
	ProjectType     string `json:"projectType,omitempty"`
}

// LocalizedName holds localized name entries for a book or resource.
type LocalizedName struct {
	Abbr  map[string]string `json:"abbr"`
	Short map[string]string `json:"short"`
	Long  map[string]string `json:"long"`
}

// Ingredient describes a single ingredient file in the SB.
type Ingredient struct {
	Checksum Checksum          `json:"checksum"`
	MimeType string            `json:"mimeType"`
	Size     int64             `json:"size"`
	Scope    map[string][]string `json:"scope,omitempty"`
}

// Checksum holds the checksum(s) for an ingredient.
type Checksum struct {
	MD5 string `json:"md5"`
}

// Copyright holds the copyright information.
type Copyright struct {
	ShortStatements []CopyrightStatement `json:"shortStatements"`
}

// CopyrightStatement holds a single copyright statement.
type CopyrightStatement struct {
	Statement string `json:"statement"`
	MimeType  string `json:"mimetype,omitempty"`
	Lang      string `json:"lang,omitempty"`
}

// NewMetadata creates a new Metadata with standard defaults.
func NewMetadata() *Metadata {
	return &Metadata{
		Format: "scripture burrito",
		Meta: Meta{
			Version:       "1.0.0",
			Category:      "source",
			Generator: Generator{
				SoftwareName:    "go-rc2sb",
				SoftwareVersion: moduleVersion,
				UserName:        "",
			},
			DefaultLocale: "en",
			Normalization: "NFC",
		},
		Confidential:   false,
		IDAuthorities:  make(map[string]IDAuthority),
		Ingredients:    make(map[string]Ingredient),
		LocalizedNames: make(map[string]LocalizedName),
	}
}

// WriteToFile serializes the metadata as JSON and writes it to metadata.json in dir.
func (m *Metadata) WriteToFile(dir string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling metadata.json: %w", err)
	}
	// Add trailing newline
	data = append(data, '\n')

	path := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing metadata.json: %w", err)
	}
	return nil
}
