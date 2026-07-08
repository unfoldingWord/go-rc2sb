package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/unfoldingWord/go-rc2sb/rc"
	"github.com/unfoldingWord/go-rc2sb/sb"
)

// obsStoryScopes maps OBS story number (zero-padded, e.g. "01") to its Bible reference scope.
// These are fixed for all OBS repositories across all languages.
var obsStoryScopes = map[string]map[string][]string{
	"01": {"GEN": {"1-2"}},
	"02": {"GEN": {"3"}},
	"03": {"GEN": {"6-8"}},
	"04": {"GEN": {"11-15"}},
	"05": {"GEN": {"16-22"}},
	"06": {"GEN": {"24:1-25:26"}},
	"07": {"GEN": {"25:27-35:29"}},
	"08": {"GEN": {"37-50"}},
	"09": {"EXO": {"1-4"}},
	"10": {"EXO": {"5-10"}},
	"11": {"EXO": {"11:1-12:32"}},
	"12": {"EXO": {"12:33-15:21"}},
	"13": {"EXO": {"19-34"}},
	"14": {"EXO": {"16-17"}, "NUM": {"10-14", "20", "27"}, "DEU": {"34"}},
	"15": {"JOS": {"1-24"}},
	"16": {"JDG": {"1-3", "6-8"}, "1SA": {"1-10"}},
	"17": {"1SA": {"10", "15-19", "24", "31"}, "2SA": {"5", "7", "11-12"}},
	"18": {"1KI": {"1-6", "11-12"}},
	"19": {"1KI": {"16-18"}, "2KI": {"5"}, "JER": {"38"}},
	"20": {"2KI": {"17", "24-25"}, "2CH": {"36"}, "EZR": {"1-10"}, "NEH": {"1-13"}},
	"21": {"GEN": {"3:15", "12:1-3"}, "DEU": {"18:15"}, "2SA": {"7"}, "JER": {"31"}, "ISA": {"59:16", "7:14", "9:1-7", "35:3-5", "61", "53", "50:6"}, "DAN": {"7"}, "MAL": {"4:5"}, "MIC": {"5:2"}, "PSA": {"22:18", "35:19", "69:4", "41:9", "16:10-11"}, "ZEC": {"11:12-13"}},
	"22": {"LUK": {"1"}},
	"23": {"MAT": {"1"}, "LUK": {"2"}},
	"24": {"MAT": {"3"}, "MRK": {"1:9-11"}, "LUK": {"3:1-23"}},
	"25": {"MAT": {"4:1-11"}, "MRK": {"1:12-13"}, "LUK": {"4:1-13"}},
	"26": {"MAT": {"4:12-25"}, "MRK": {"1:14-15", "35-39", "3:13-21"}, "LUK": {"4:14-30", "38-44"}},
	"27": {"LUK": {"10:25-37"}},
	"28": {"MAT": {"19:16-30"}, "MRK": {"10:17-31"}, "LUK": {"18:18-30"}},
	"29": {"MAT": {"18:21-35"}},
	"30": {"MAT": {"14:13-21"}, "MRK": {"6:31-44"}, "LUK": {"9:10-17"}, "JHN": {"6:5-15"}},
	"31": {"MAT": {"14:22-33"}, "MRK": {"6:45-52"}, "JHN": {"6:16-21"}},
	"32": {"MAT": {"8:28-34", "9:20-22"}, "MRK": {"5:1-20", "5:24-34"}, "LUK": {"8:26-39", "8:42-48"}},
	"33": {"MAT": {"13:1-8", "18-23"}, "MRK": {"4:1-8", "13-20"}, "LUK": {"8:4-15"}},
	"34": {"MAT": {"13:31-33", "44-46"}, "MRK": {"4:30-32"}, "LUK": {"13:18-21", "18:9-14"}},
	"35": {"LUK": {"15:11-32"}},
	"36": {"MAT": {"17:1-9"}, "MRK": {"9:2-8"}, "LUK": {"9:28-36"}},
	"37": {"JHN": {"11:1-46"}},
	"38": {"MAT": {"26:14-56"}, "MRK": {"14:10-50"}, "LUK": {"22:1-53"}, "JHN": {"12:6", "18:1-11"}},
	"39": {"MAT": {"26:57-27:26"}, "MRK": {"14:53-15:15"}, "LUK": {"22:54-23:25"}, "JHN": {"18:12-19:16"}},
	"40": {"MAT": {"27:27-61"}, "MRK": {"15:16-47"}, "LUK": {"23:26-56"}, "JHN": {"19:17-42"}},
	"41": {"MAT": {"27:62-28:15"}, "MRK": {"16:1-11"}, "LUK": {"24:1-12"}, "JHN": {"20:1-18"}},
	"42": {"MAT": {"28:16-20"}, "MRK": {"16:12-20"}, "LUK": {"24:13-53"}, "JHN": {"20:19-23"}, "ACT": {"1:1-11"}},
	"43": {"ACT": {"2"}},
	"44": {"ACT": {"3:1-4:22"}},
	"45": {"ACT": {"6:8-8:5", "8:26-40"}},
	"46": {"ACT": {"8:3", "9:1-31", "11:19-26", "13:1-3"}},
	"47": {"ACT": {"16:11-40"}},
	"48": {"GEN": {"1-3", "6", "14", "22"}, "EXO": {"12", "20"}, "2SA": {"7"}, "HEB": {"3:1-6", "4:14-5:10", "7:1-8:13", "9:11-10:18"}, "REV": {"21"}},
	"49": {"ROM": {"3:21-26", "5:1-11"}, "JHN": {"3:16"}, "MRK": {"16:16"}, "COL": {"1:13-14"}, "2CO": {"5:17-21"}, "1JN": {"1:5-10"}},
	"50": {"MAT": {"24:14", "28:18", "13:24-30", "13:36-42", "22:13"}, "JHN": {"15:20", "16:33"}, "REV": {"2:10", "20:10", "21:1-22:21"}, "1TH": {"4:13-5:11"}, "JAS": {"1:12"}},
}

// NewOBSHandler creates a new Open Bible Stories handler.
func NewOBSHandler() Handler {
	return &obsHandler{}
}

type obsHandler struct{}

func (h *obsHandler) Subject() string {
	return "Open Bible Stories"
}

func (h *obsHandler) Convert(ctx context.Context, manifest *rc.Manifest, inDir, outDir string, opts Options) (*sb.Metadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m := BuildBaseMetadata(manifest, "BurritoTruck", "OBS")

	// Build currentScope as the deduplicated union of all OBS story scopes.
	// The SB scope schema requires each book's reference array to have unique
	// items (oneOf: non-empty array with uniqueItems, or empty array), so the
	// same reference contributed by multiple stories (e.g. 2SA "7" from stories
	// 17, 21, and 48) must appear only once. Stories are iterated in sorted
	// order so the resulting reference order is deterministic across runs.
	currentScope := make(map[string][]string)
	seen := make(map[string]map[string]bool) // book -> ref -> already added
	storyNums := make([]string, 0, len(obsStoryScopes))
	for num := range obsStoryScopes {
		storyNums = append(storyNums, num)
	}
	sort.Strings(storyNums)
	for _, num := range storyNums {
		for book, refs := range obsStoryScopes[num] {
			if seen[book] == nil {
				seen[book] = make(map[string]bool)
			}
			for _, ref := range refs {
				if seen[book][ref] {
					continue
				}
				seen[book][ref] = true
				currentScope[book] = append(currentScope[book], ref)
			}
		}
	}

	// Set type - OBS uses gloss/textStories
	m.Type = sb.Type{
		FlavorType: sb.FlavorType{
			Name: "gloss",
			Flavor: sb.Flavor{
				Name: "textStories",
			},
			CurrentScope: currentScope,
		},
	}

	// OBS uses a different copyright format
	m.Copyright = BuildCopyright(manifest, true)

	// Copy common root files (README.md, .gitignore, .gitea, .github)
	if err := CopyCommonRootFiles(inDir, outDir, m); err != nil {
		return nil, err
	}

	// Copy LICENSE.md to root (uses embedded default if RC doesn't have one).
	if err := CopyLicenseToRoot(inDir, outDir); err != nil {
		return nil, fmt.Errorf("copying root LICENSE.md: %w", err)
	}

	// Determine the content directory from the manifest project path.
	// OBS has a single project whose path is typically "./content" but may be "."
	// when the markdown files live in the repository root.
	contentPath := "content"
	if len(manifest.Projects) > 0 {
		p := strings.TrimPrefix(manifest.Projects[0].Path, "./")
		if p != "" {
			contentPath = p
		}
	}
	contentDir := inDir
	if contentPath != "." {
		contentDir = filepath.Join(inDir, contentPath)
	}

	// Set name and description from front/title.md when present. The RC
	// language gets the title text; English is always present, falling back
	// to "Open Bible Stories" when the title doesn't provide it.
	title, err := readOBSTitle(contentDir)
	if err != nil {
		return nil, err
	}
	nameMap := make(map[string]string)
	if title != "" {
		nameMap[manifest.DublinCore.Language.Identifier] = title
	}
	if _, ok := nameMap["en"]; !ok {
		nameMap["en"] = "Open Bible Stories"
	}
	descMap := make(map[string]string, len(nameMap))
	for k, v := range nameMap {
		descMap[k] = v
	}
	m.Identification.Name = nameMap
	m.Identification.Description = descMap

	if err := copyOBSContent(contentDir, outDir, m); err != nil {
		return nil, err
	}

	// Copy LICENSE.md to ingredients/LICENSE.md (uses embedded default if RC doesn't have one).
	licIng, err := CopyLicenseIngredient(inDir, outDir)
	if err != nil {
		return nil, fmt.Errorf("copying ingredients/LICENSE.md: %w", err)
	}
	m.Ingredients["ingredients/LICENSE.md"] = licIng

	return m, nil
}

// copyOBSContent copies OBS content into ingredients/content/. Only the story
// files 01.md-50.md plus front.md and back.md are copied; everything else in
// the content directory (front/title.md, config files, etc.) is skipped.
// front.md is sourced from front/intro.md (or a flat front.md), and back.md
// from back/intro.md (or a flat back.md) — the front/ and back/ directories
// themselves are never recreated in the output.
func copyOBSContent(contentDir, outDir string, m *sb.Metadata) error {
	// Story files 01.md-50.md, each with its fixed Bible reference scope.
	storyNums := make([]string, 0, len(obsStoryScopes))
	for num := range obsStoryScopes {
		storyNums = append(storyNums, num)
	}
	sort.Strings(storyNums)
	for _, num := range storyNums {
		name := num + ".md"
		src := filepath.Join(contentDir, name)
		if _, err := os.Stat(src); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("checking content file %s: %w", name, err)
		}
		ingredientKey := "ingredients/content/" + name
		ing, err := CopyFileWithScope(src, outDir, ingredientKey, obsStoryScopes[num])
		if err != nil {
			return fmt.Errorf("copying content file %s: %w", name, err)
		}
		m.Ingredients[ingredientKey] = ing
	}

	// front.md and back.md, flattened from front/intro.md and back/intro.md.
	for _, section := range []string{"front", "back"} {
		candidates := []string{
			filepath.Join(contentDir, section, "intro.md"),
			filepath.Join(contentDir, section+".md"),
		}
		for _, src := range candidates {
			info, err := os.Stat(src)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return fmt.Errorf("checking %s content: %w", section, err)
			}
			if info.IsDir() {
				continue
			}
			ingredientKey := "ingredients/content/" + section + ".md"
			ing, err := CopyFileAndComputeIngredient(src, outDir, ingredientKey)
			if err != nil {
				return fmt.Errorf("copying %s content: %w", section, err)
			}
			m.Ingredients[ingredientKey] = ing
			break
		}
	}

	return nil
}

// readOBSTitle returns the trimmed contents of front/title.md in the content
// directory, or "" if the file does not exist.
func readOBSTitle(contentDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(contentDir, "front", "title.md"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("reading front/title.md: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}
