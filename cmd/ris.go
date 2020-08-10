package main

import (
	"fmt"
	"html"
	"path"
	"sort"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// subset of tags needed in code below
const risTagAuthor = "AU"
const risTagAuthorPrimary = "A1"
const risTagAuthorSecondary = "A2"
const risTagAuthorTertiary = "A3"
const risTagAuthorSubsidiary = "A4"
const risTagEnd = "ER"
const risTagKeyword = "KW"
const risTagLibrary = "DP"
const risTagNote = "N1"
const risTagPeriodicalName = "J2"
const risTagType = "TY"
const risTagURL = "UR"

// misc definitions
const risTypeGeneric = "GEN"
const risTruncateLength = 255
const risLineEnding = "\r\n"
const risLineFormat = "%s  - %s" + risLineEnding

type tagValueMap map[string][]string

type risEncoder struct {
	url         string
	extension   string
	contentType string
	tagValues   tagValueMap
	policy      *bluemonday.Policy
}

var risPartsMap map[string][]string
var risTypesMap map[string]string

func newRisEncoder(cfg serviceConfigFormat, url string) *risEncoder {
	r := risEncoder{}

	r.url = url
	r.extension = cfg.Extension
	r.contentType = cfg.ContentType
	r.tagValues = make(tagValueMap)
	r.policy = bluemonday.StrictPolicy()

	return &r
}

func (r *risEncoder) populateCitation(parts citationParts) {
	for part, values := range parts {
		risCodes := risPartsMap[part]

		if len(risCodes) == 0 {
			continue
		}

		for _, risCode := range risCodes {
			for _, value := range values {
				risValue := value

				if risCode == risTagType {
					risValue = risTypesMap[value]
					if risValue == "" {
						risValue = risTypeGeneric
					}
				}

				r.addTagValue(risCode, risValue)
			}
		}
	}
}

func (r *risEncoder) addTagValue(risTag, value string) {
	tag := strings.ToUpper(risTag)

	r.tagValues[tag] = append(r.tagValues[tag], value)
}

func (r *risEncoder) ContentType() string {
	return r.contentType
}

func (r *risEncoder) FileName() string {
	filename := path.Base(r.url)

	if r.extension != "" {
		filename += "." + r.extension
	}

	return filename
}

func (r *risEncoder) isAuthorOrKeywordTag(tag string) bool {
	switch tag {
	case risTagAuthor:
	case risTagAuthorPrimary:
	case risTagAuthorSecondary:
	case risTagAuthorTertiary:
	case risTagAuthorSubsidiary:
	case risTagKeyword:

	default:
		return false
	}

	return true
}

func (r *risEncoder) isRepeatableTag(tag string) bool {
	switch {
	case r.isAuthorOrKeywordTag(tag):
	case tag == risTagNote:

	default:
		return false
	}

	return true
}

func (r *risEncoder) isNonAsteriskTag(tag string) bool {
	switch {
	case r.isAuthorOrKeywordTag(tag):
	case tag == risTagPeriodicalName:

	default:
		return false
	}

	return true
}

func (r *risEncoder) isLimitedLengthTag(tag string) bool {
	return r.isAuthorOrKeywordTag(tag)
}

func (r *risEncoder) cleanString(val string) string {
	cleaned := val

	cleaned = strings.ReplaceAll(cleaned, `►`, `>`)
	cleaned = strings.ReplaceAll(cleaned, `▶`, `>`)
	cleaned = strings.ReplaceAll(cleaned, `•`, `*`)
	cleaned = strings.ReplaceAll(cleaned, `·`, `*`)
	cleaned = strings.ReplaceAll(cleaned, `–`, `-`)
	cleaned = strings.ReplaceAll(cleaned, `—`, `--`)
	cleaned = strings.ReplaceAll(cleaned, `→`, `->`)
	cleaned = strings.ReplaceAll(cleaned, `←`, `<-`)
	cleaned = strings.ReplaceAll(cleaned, `↔`, `<->`)
	cleaned = strings.ReplaceAll(cleaned, `⇒`, `=>`)
	cleaned = strings.ReplaceAll(cleaned, `⇐`, `<=`)
	cleaned = strings.ReplaceAll(cleaned, `⇔`, `<=>`)
	cleaned = strings.ReplaceAll(cleaned, `≤`, `<=`)
	cleaned = strings.ReplaceAll(cleaned, `≦`, `<=`)
	cleaned = strings.ReplaceAll(cleaned, `≥`, `>=`)
	cleaned = strings.ReplaceAll(cleaned, `≧`, `>=`)
	cleaned = strings.ReplaceAll(cleaned, `©`, `(c)`)
	cleaned = strings.ReplaceAll(cleaned, `®`, `(R)`)
	cleaned = strings.ReplaceAll(cleaned, `’`, `'`)
	cleaned = strings.ReplaceAll(cleaned, `‹`, `'`)
	cleaned = strings.ReplaceAll(cleaned, `›`, `'`)
	cleaned = strings.ReplaceAll(cleaned, `«`, `"`)
	cleaned = strings.ReplaceAll(cleaned, `»`, `"`)

	cleaned = r.policy.Sanitize(cleaned)
	cleaned = html.UnescapeString(cleaned)
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

func (r *risEncoder) getTagValue(tag, val string) string {
	value := strings.ReplaceAll(val, `\n`, "\n")

	// clean up each line of data
	var lines []string
	for _, line := range strings.Split(value, "\n") {
		cleaned := r.cleanString(line)

		if r.isNonAsteriskTag(tag) == true {
			cleaned = strings.ReplaceAll(cleaned, `*`, `#`)
		}

		if r.isLimitedLengthTag(tag) == true {
			if len(cleaned) > risTruncateLength {
				cleaned = cleaned[0:risTruncateLength]
			}
		}

		if cleaned != "" {
			lines = append(lines, cleaned)
		}
	}

	return strings.Join(lines, risLineEnding)
}

func (r *risEncoder) FileContents() (string, error) {
	if len(r.tagValues[risTagType]) == 0 {
		r.addTagValue(risTagType, risTypeGeneric)
	}

	r.addTagValue(risTagNote, r.url)

	tags := []string{}
	for tag := range r.tagValues {
		if tag != risTagType {
			tags = append(tags, tag)
		}
	}

	sort.Strings(tags)

	var b strings.Builder

	// use first type as type
	fmt.Fprintf(&b, risLineFormat, risTagType, r.tagValues[risTagType][0])

	r.recordBody(&b, tags)

	fmt.Fprintf(&b, risLineFormat, risTagEnd, "")

	return b.String(), nil
}

func (r *risEncoder) recordBody(b *strings.Builder, tags []string) {
	merged := make(tagValueMap)

	// first pass: merge tags that are not repeatable
	for _, tag := range tags {
		switch {
		// preserve individual tag values for repeatable tags
		case r.isRepeatableTag(tag):
			merged[tag] = r.tagValues[tag]

		// special-case URL joining (to prevent the separator from being treated as part of the URL)
		case tag == risTagURL:
			merged[tag] = []string{strings.Join(r.tagValues[tag], " ; ")}

		// tags for which we want to use the first occurrence only
		case tag == risTagLibrary:
			merged[tag] = []string{r.tagValues[tag][0]}

		// straightforward joining of separate values
		default:
			merged[tag] = []string{strings.Join(r.tagValues[tag], ", ")}
		}
	}

	// second pass: write cleaned tag values one by one
	for _, tag := range tags {
		for _, val := range merged[tag] {
			fmt.Fprintf(b, risLineFormat, tag, r.getTagValue(tag, val))
		}
	}
}

func init() {
	// mapping of citation parts to RIS code(s)
	risPartsMap = make(map[string][]string)

	risPartsMap["abstract"] = []string{"AB"}
	risPartsMap["author"] = []string{"AU"}
	risPartsMap["call_number"] = []string{"CN"}
	risPartsMap["content_provider"] = []string{"DB"}
	risPartsMap["description"] = []string{"N1"}
	risPartsMap["doi"] = []string{"DO"}
	risPartsMap["format"] = []string{"TY"}
	risPartsMap["full_text_url"] = []string{"L2"}
	risPartsMap["genre"] = []string{"M3"}
	risPartsMap["id"] = []string{"ID"}
	risPartsMap["issue"] = []string{"IS"}
	risPartsMap["journal"] = []string{"T2"}
	risPartsMap["language"] = []string{"LA"}
	risPartsMap["library"] = []string{"DP"}
	risPartsMap["location"] = []string{"AN"}
	risPartsMap["published_location"] = []string{"CY"}
	risPartsMap["published_date"] = []string{"DA", "PY"}
	risPartsMap["publisher"] = []string{"PB"}
	risPartsMap["rights"] = []string{"C4"}
	risPartsMap["serial_number"] = []string{"SN"}
	risPartsMap["series"] = []string{"T2"}
	risPartsMap["subject"] = []string{"KW"}
	risPartsMap["subtitle"] = []string{"T2"}
	risPartsMap["title"] = []string{"TI"}
	risPartsMap["url"] = []string{"UR"}
	risPartsMap["volume"] = []string{"VL"}

	// mapping of citation formats (citation part "format") to RIS type
	risTypesMap = make(map[string]string)

	risTypesMap["art"] = "ART"
	risTypesMap["article"] = "JOUR"
	risTypesMap["book"] = "BOOK"
	risTypesMap["generic"] = "GEN"
	risTypesMap["government_document"] = "GOVDOC"
	risTypesMap["journal"] = "JOUR"
	risTypesMap["manuscript"] = "MANSCPT"
	risTypesMap["map"] = "MAP"
	risTypesMap["music"] = "MUSIC"
	risTypesMap["news"] = "NEWS"
	risTypesMap["sound"] = "SOUND"
	risTypesMap["thesis"] = "THES"
	risTypesMap["video"] = "VIDEO"
}
