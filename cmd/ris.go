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

func newRisEncoder(cfg serviceConfigFormat) *risEncoder {
	e := risEncoder{}

	e.extension = cfg.Extension
	e.contentType = cfg.ContentType
	e.tagValues = make(tagValueMap)
	e.policy = bluemonday.StrictPolicy()

	return &e
}

func (e *risEncoder) Init(url string) {
	e.url = url
}

func (e *risEncoder) Populate(parts citationParts) error {
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

				e.addTagValue(risCode, risValue)
			}
		}
	}

	return nil
}

func (e *risEncoder) addTagValue(risTag, value string) {
	tag := strings.ToUpper(risTag)

	e.tagValues[tag] = append(e.tagValues[tag], value)
}

func (e *risEncoder) ContentType() string {
	return e.contentType
}

func (e *risEncoder) FileName() string {
	filename := path.Base(e.url)

	if e.extension != "" {
		filename += "." + e.extension
	}

	return filename
}

func (e *risEncoder) isAuthorOrKeywordTag(tag string) bool {
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

func (e *risEncoder) isRepeatableTag(tag string) bool {
	switch {
	case e.isAuthorOrKeywordTag(tag):
	case tag == risTagNote:

	default:
		return false
	}

	return true
}

func (e *risEncoder) isNonAsteriskTag(tag string) bool {
	switch {
	case e.isAuthorOrKeywordTag(tag):
	case tag == risTagPeriodicalName:

	default:
		return false
	}

	return true
}

func (e *risEncoder) isLimitedLengthTag(tag string) bool {
	return e.isAuthorOrKeywordTag(tag)
}

func (e *risEncoder) cleanString(val string) string {
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

	cleaned = e.policy.Sanitize(cleaned)
	cleaned = html.UnescapeString(cleaned)
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

func (e *risEncoder) getTagValue(tag, val string) string {
	value := strings.ReplaceAll(val, `\n`, "\n")

	// clean up each line of data
	var lines []string
	for _, line := range strings.Split(value, "\n") {
		cleaned := e.cleanString(line)

		if e.isNonAsteriskTag(tag) == true {
			cleaned = strings.ReplaceAll(cleaned, `*`, `#`)
		}

		if e.isLimitedLengthTag(tag) == true {
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

func (e *risEncoder) Contents() (string, error) {
	if len(e.tagValues[risTagType]) == 0 {
		e.addTagValue(risTagType, risTypeGeneric)
	}

	e.addTagValue(risTagNote, e.url)

	tags := []string{}
	for tag := range e.tagValues {
		if tag != risTagType {
			tags = append(tags, tag)
		}
	}

	sort.Strings(tags)

	var b strings.Builder

	// use first type as type
	fmt.Fprintf(&b, risLineFormat, risTagType, e.tagValues[risTagType][0])

	e.recordBody(&b, tags)

	fmt.Fprintf(&b, risLineFormat, risTagEnd, "")

	return b.String(), nil
}

func (e *risEncoder) recordBody(b *strings.Builder, tags []string) {
	merged := make(tagValueMap)

	// first pass: merge tags that are not repeatable
	for _, tag := range tags {
		switch {
		// preserve individual tag values for repeatable tags
		case e.isRepeatableTag(tag):
			merged[tag] = e.tagValues[tag]

		// special-case URL joining (to prevent the separator from being treated as part of the URL)
		case tag == risTagURL:
			merged[tag] = []string{strings.Join(e.tagValues[tag], " ; ")}

		// tags for which we want to use the first occurrence only
		case tag == risTagLibrary:
			merged[tag] = []string{e.tagValues[tag][0]}

		// straightforward joining of separate values
		default:
			merged[tag] = []string{strings.Join(e.tagValues[tag], ", ")}
		}
	}

	// second pass: write cleaned tag values one by one
	for _, tag := range tags {
		for _, val := range merged[tag] {
			fmt.Fprintf(b, risLineFormat, tag, e.getTagValue(tag, val))
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
