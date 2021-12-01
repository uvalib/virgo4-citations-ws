package main

import (
	"fmt"
	"html"
	"path"
	"sort"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// subset of ris tags needed in code below
const risTagAbstract = "AB"
const risTagAccessionNumber = "AN"
const risTagAuthor = "AU"
const risTagAuthorPrimary = "A1"
const risTagAuthorSecondary = "A2"
const risTagAuthorSubsidiary = "A4"
const risTagAuthorTertiary = "A3"
const risTagCallNumber = "CN"
const risTagDOI = "DO"
const risTagDatabase = "DB"
const risTagDate = "DA"
const risTagEnd = "ER"
const risTagFullTextLink = "L2"
const risTagIssueNumber = "IS"
const risTagJournalTitle = "T2"
const risTagKeyword = "KW"
const risTagLanguage = "LA"
const risTagLibrary = "DP"
const risTagNote = "N1"
const risTagPeriodicalName = "J2"
const risTagPlacePublished = "CY"
const risTagPublicationYear = "PY"
const risTagPublisher = "PB"
const risTagReferenceID = "ID"
const risTagRights = "C4"
const risTagSerialNumber = "SN"
const risTagSubtitle = risTagJournalTitle
const risTagTitle = "TI"
const risTagType = "TY"
const risTagTypeOfWork = "M3"
const risTagURL = "UR"
const risTagVolumeNumber = "VL"

// ris types
const risTypeArt = "ART"
const risTypeBook = "BOOK"
const risTypeGeneric = "GEN"
const risTypeGovernmentDocument = "GOVDOC"
const risTypeJournal = "JOUR"
const risTypeManuscript = "MANSCPT"
const risTypeMap = "MAP"
const risTypeMusic = "MUSIC"
const risTypeNews = "NEWS"
const risTypeSound = "SOUND"
const risTypeThesis = "THES"
const risTypeVideo = "VIDEO"

// misc definitions
const risTruncateLength = 255
const risLineEnding = "\r\n"
const risLineFormat = "%s  - %s" + risLineEnding

type tagValueMap map[string][]string

type risEncoder struct {
	cfg       serviceConfigFormat
	url       string
	tagValues tagValueMap
	policy    *bluemonday.Policy
}

var risPartsMap map[string][]string
var risTypesMap map[string]string
var risAppendMap map[string]string

func newRisEncoder(cfg serviceConfigFormat) *risEncoder {
	e := risEncoder{}

	e.cfg = cfg
	e.tagValues = make(tagValueMap)
	e.policy = bluemonday.UGCPolicy()

	return &e
}

func (e *risEncoder) Init(c *clientContext, url string) {
	e.url = url
}

func (e *risEncoder) Populate(parts citationParts) error {
	for part, values := range parts {
		risCodes := risPartsMap[part]
		postfix := risAppendMap[part]

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
				} else {
					if postfix != "" {
						risValue = risValue + " " + postfix
					}
				}

				e.addTagValue(risCode, risValue)
			}
		}
	}

	if len(e.tagValues[risTagType]) == 0 {
		e.addTagValue(risTagType, risTypeGeneric)
	}

	// type-specific cleanups

	risType := e.tagValues[risTagType][0]

	// merge title/subtitle for book type citations, using first instance of each

	if risType == risTypeBook {
		if len(e.tagValues[risTagSubtitle]) > 0 {
			title := firstElementOf(e.tagValues[risTagTitle])
			subtitle := firstElementOf(e.tagValues[risTagSubtitle])

			fullTitle := title
			if subtitle != "" {
				fullTitle = fullTitle + ": " + subtitle
			}

			fullTitle = removeTrailingPeriods(fullTitle)

			e.tagValues[risTagTitle] = []string{fullTitle}
			delete(e.tagValues, risTagSubtitle)
		}
	}

	return nil
}

func (e *risEncoder) addTagValue(risTag, value string) {
	tag := strings.ToUpper(risTag)

	e.tagValues[tag] = append(e.tagValues[tag], value)
}

func (e *risEncoder) Label() string {
	return e.cfg.Label
}

func (e *risEncoder) ContentType() string {
	return e.cfg.ContentType
}

func (e *risEncoder) FileName() string {
	filename := path.Base(e.url)

	if e.cfg.Extension != "" {
		filename += "." + e.cfg.Extension
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
	url := fmt.Sprintf(`<a href="%s">%s</a>`, e.url, e.url)
	e.addTagValue(risTagNote, url)

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

	risPartsMap["abstract"] = []string{risTagAbstract}
	risPartsMap["advisor"] = []string{risTagAuthor}
	risPartsMap["author"] = []string{risTagAuthor}
	risPartsMap["call_number"] = []string{risTagCallNumber}
	risPartsMap["content_provider"] = []string{risTagDatabase}
	risPartsMap["description"] = []string{risTagNote}
	risPartsMap["doi"] = []string{risTagDOI}
	risPartsMap["editor"] = []string{risTagAuthor}
	risPartsMap["format"] = []string{risTagType}
	risPartsMap["full_text_url"] = []string{risTagFullTextLink}
	risPartsMap["genre"] = []string{risTagTypeOfWork}
	risPartsMap["id"] = []string{risTagReferenceID}
	risPartsMap["issue"] = []string{risTagIssueNumber}
	risPartsMap["journal"] = []string{risTagJournalTitle}
	risPartsMap["language"] = []string{risTagLanguage}
	risPartsMap["library"] = []string{risTagLibrary}
	risPartsMap["location"] = []string{risTagAccessionNumber}
	risPartsMap["published_location"] = []string{risTagPlacePublished}
	risPartsMap["published_date"] = []string{risTagDate, risTagPublicationYear}
	risPartsMap["publisher"] = []string{risTagPublisher}
	risPartsMap["rights"] = []string{risTagRights}
	risPartsMap["serial_number"] = []string{risTagSerialNumber}
	risPartsMap["series"] = []string{risTagJournalTitle}
	risPartsMap["subject"] = []string{risTagKeyword}
	risPartsMap["subtitle"] = []string{risTagJournalTitle}
	risPartsMap["title"] = []string{risTagTitle}
	risPartsMap["url"] = []string{risTagURL}
	risPartsMap["volume"] = []string{risTagVolumeNumber}

	// mapping of citation formats (citation part "format") to RIS type
	risTypesMap = make(map[string]string)

	risTypesMap["art"] = risTypeArt
	risTypesMap["article"] = risTypeJournal
	risTypesMap["book"] = risTypeBook
	risTypesMap["generic"] = risTypeGeneric
	risTypesMap["government_document"] = risTypeGovernmentDocument
	risTypesMap["journal"] = risTypeJournal
	risTypesMap["manuscript"] = risTypeManuscript
	risTypesMap["map"] = risTypeMap
	risTypesMap["music"] = risTypeMusic
	risTypesMap["news"] = risTypeNews
	risTypesMap["sound"] = risTypeSound
	risTypesMap["thesis"] = risTypeThesis
	risTypesMap["video"] = risTypeVideo

	// mapping of strings to append to authors, depending on type
	risAppendMap = make(map[string]string)

	risAppendMap["advisor"] = "(advisor)"
	risAppendMap["editor"] = "(editor)"
}
