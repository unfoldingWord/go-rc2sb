package rc2sb

import (
	"strings"

	"github.com/unfoldingWord/go-rc2sb/handler"
	"github.com/unfoldingWord/go-rc2sb/rc"
)

// subjectAliases are subjects RC manifests use for content of another subject.
var subjectAliases = map[string]string{
	"Bible translation comprehension questions": "Translation Questions",
}

// identifierSubjects are the subjects of the resource identifiers RC manifests use,
// for an older manifest whose dublin_core names no subject (e.g. Translation Academy
// tags with only "type: man" and "identifier: ta").
var identifierSubjects = map[string]string{
	"obs":     "Open Bible Stories",
	"obs-sn":  "OBS Study Notes",
	"obs-sq":  "OBS Study Questions",
	"obs-tn":  "OBS Translation Notes",
	"obs-tq":  "OBS Translation Questions",
	"obs-twl": "TSV OBS Translation Words Links",
	"sn":      "Study Notes",
	"sq":      "Study Questions",
	"ta":      "Translation Academy",
	"tn":      "Translation Notes",
	"tq":      "Translation Questions",
	"tw":      "Translation Words",
	"twl":     "TSV Translation Words Links",
	"ugnt":    "Greek New Testament",
	"uhb":     "Hebrew Old Testament",
	"ulb":     "Bible",
	"udb":     "Bible",
	"ult":     "Aligned Bible",
	"ust":     "Aligned Bible",
	"glt":     "Aligned Bible",
	"gst":     "Aligned Bible",
}

// obsSubjects are the OBS forms of the helps subjects, which an RC takes when
// its only project is "obs" (e.g. a "Study Notes" RC of OBS study notes).
var obsSubjects = map[string]string{
	"Translation Notes":           "OBS Translation Notes",
	"Translation Questions":       "OBS Translation Questions",
	"Study Notes":                 "OBS Study Notes",
	"Study Questions":             "OBS Study Questions",
	"TSV Translation Notes":       "TSV OBS Translation Notes",
	"TSV Translation Questions":   "TSV OBS Translation Questions",
	"TSV Study Notes":             "TSV OBS Study Notes",
	"TSV Study Questions":         "TSV OBS Study Questions",
	"TSV Translation Words Links": "TSV OBS Translation Words Links",
}

// ResolveSubject returns the subject an RC converts as, from its manifest's
// dublin_core.subject and its projects' identifiers. It applies the spellings
// RC manifests use for a known subject ("Bible Stories", "Translator Notes",
// aliases) and gives a helps subject its OBS form when the RC is of OBS.
// Convert uses it to pick the handler; DCS uses it so a catalog entry carries
// the subject its burrito converts as.
func ResolveSubject(subject string, projectIDs []string) string {
	subject = strings.TrimSpace(subject)
	if strings.EqualFold(subject, "bible stories") {
		subject = "Open Bible Stories"
	}
	const translator = "Translator "
	if len(subject) > len(translator) && strings.EqualFold(subject[:len(translator)], translator) {
		subject = "Translation " + subject[len(translator):]
	}
	subject = registeredSpelling(subject)
	if alias, ok := subjectAliases[subject]; ok {
		subject = alias
	}
	if obs, ok := obsSubjects[subject]; ok && onlyOBS(projectIDs) {
		subject = obs
	}
	return subject
}

// ResolveRCSubject is ResolveSubject for an RC manifest: when its dublin_core names
// no subject, the subject comes from its dublin_core.identifier (e.g. "ta").
func ResolveRCSubject(subject, identifier string, projectIDs []string) string {
	if strings.TrimSpace(subject) == "" {
		subject = identifierSubjects[strings.ToLower(strings.TrimSpace(identifier))]
	}
	return ResolveSubject(subject, projectIDs)
}

// registeredSpelling returns the registered subject or alias the subject names,
// ignoring case (e.g. "Open Bible stories"), or the subject unchanged.
func registeredSpelling(subject string) string {
	for _, s := range handler.SupportedSubjects() {
		if strings.EqualFold(s, subject) {
			return s
		}
	}
	for alias := range subjectAliases {
		if strings.EqualFold(alias, subject) {
			return alias
		}
	}
	return subject
}

// Supports reports whether Convert converts an RC of the given resolved subject.
func Supports(subject string) bool {
	_, err := handler.Lookup(subject)
	return err == nil
}

// SupportedSubjects returns the resolved subjects Convert converts, sorted.
func SupportedSubjects() []string {
	return handler.SupportedSubjects()
}

func onlyOBS(projectIDs []string) bool {
	if len(projectIDs) == 0 {
		return false
	}
	for _, id := range projectIDs {
		if !strings.EqualFold(strings.TrimSpace(id), "obs") {
			return false
		}
	}
	return true
}

func manifestSubject(manifest *rc.Manifest) string {
	ids := make([]string, 0, len(manifest.Projects))
	for _, p := range manifest.Projects {
		ids = append(ids, p.Identifier)
	}
	return ResolveRCSubject(manifest.DublinCore.Subject, manifest.DublinCore.Identifier, ids)
}
