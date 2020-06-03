package main

import (
	"fmt"
	"html"
	"log"
	"path"
	"regexp"
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
const risTagType = "TY"
const risTagURL = "UR"

// misc definitions
const risTypeGeneric = "GEN"
const risLineEnding = "\r\n"
const risLineFormat = "%s  - %s" + risLineEnding

type tagValueMap map[string][]string

type risEncoder struct {
	url         string
	extension   string
	contentType string
	tagValues   tagValueMap
	re          *regexp.Regexp
	policy      *bluemonday.Policy
}

func newRisEncoder(cfg serviceConfigFormat, url string) *risEncoder {
	r := risEncoder{}

	r.url = url
	r.extension = cfg.Extension
	r.contentType = cfg.ContentType
	r.tagValues = make(tagValueMap)
	r.re = regexp.MustCompile(`^([[:upper:]]|[[:digit:]]){2}$`)
	r.policy = bluemonday.StrictPolicy()

	return &r
}

func (r *risEncoder) addTagValue(risTag, value string) {
	tag := strings.ToUpper(risTag)

	if r.re.MatchString(tag) == false {
		log.Printf("skipping invalid RIS tag: [%s]", tag)
		return
	}

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

	data := r.singleRecordByJoiningTypes(tags)

	return data, nil
}

func (r *risEncoder) isRepeatableTag(tag string) bool {
	switch tag {
	case risTagAuthor:
	case risTagAuthorPrimary:
	case risTagAuthorSecondary:
	case risTagAuthorTertiary:
	case risTagAuthorSubsidiary:
	case risTagNote:
	case risTagKeyword:

	default:
		return false
	}

	return true
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

func (r *risEncoder) getTagValue(val string) string {
	value := strings.ReplaceAll(val, `\n`, "\n")

	// clean up each line of data
	var lines []string
	for _, line := range strings.Split(value, "\n") {
		cleaned := r.cleanString(line)
		if cleaned != "" {
			lines = append(lines, cleaned)
		}
	}

	return strings.Join(lines, risLineEnding)
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

		default:
			merged[tag] = []string{strings.Join(r.tagValues[tag], ",")}
		}
	}

	// second pass: write cleaned tag values one by one
	for _, tag := range tags {
		for _, val := range merged[tag] {
			fmt.Fprintf(b, risLineFormat, tag, r.getTagValue(val))
		}
	}
}

func (r *risEncoder) singleRecordByJoiningTypes(tags []string) string {
	var b strings.Builder

	// use first type as type
	fmt.Fprintf(&b, risLineFormat, risTagType, r.tagValues[risTagType][0])

	r.recordBody(&b, tags)

	fmt.Fprintf(&b, risLineFormat, risTagEnd, "")

	return b.String()
}
