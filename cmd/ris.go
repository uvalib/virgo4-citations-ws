package main

import (
	"fmt"
	"log"
	"path"
	"regexp"
	"sort"
	"strings"
)

// required tags
const risTagType = "TY"
const risTagEnd = "ER"

// special handling tags
const risTagURL = "UR"
const risTagLibrary = "DB"

// repeatable tags
const risTagAuthor = "AU"
const risTagAuthorPrimary = "A1"
const risTagAuthorSecondary = "A2"
const risTagAuthorTertiary = "A3"
const risTagAuthorSubsidiary = "A4"
const risTagKeyword = "KW"

// allowed multi-line tags
const risTagAbstract = "AB"
const risTagNote1 = "N1"
const risTagNote2 = "N2"

// misc definitions
const risTypeGeneric = "GEN"
const risLineEnding = "\r\n"
const risLineFormat = "%s  - %s" + risLineEnding

type risEncoder struct {
	url         string
	extension   string
	contentType string
	tagValues   map[string][]string
	re          *regexp.Regexp
}

func newRisEncoder(cfg serviceConfigFormat, url string) *risEncoder {
	r := risEncoder{}

	r.url = url
	r.extension = cfg.Extension
	r.contentType = cfg.ContentType
	r.tagValues = make(map[string][]string)
	r.re = regexp.MustCompile(`^([[:upper:]]|[[:digit:]]){2}$`)

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

	r.addTagValue(risTagNote1, r.url)

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
	case risTagNote1:
	case risTagKeyword:

	default:
		return false
	}

	return true
}

func (r *risEncoder) isAllowedMultilineTag(tag string) bool {
	switch tag {
	case risTagAbstract:
	case risTagNote2:

	default:
		return false
	}

	return true
}

func (r *risEncoder) singleRecordByJoiningTypes(tags []string) string {
	var b strings.Builder

	types := strings.Join(r.tagValues[risTagType], "; ")
	fmt.Fprintf(&b, risLineFormat, risTagType, types)

	r.recordBody(&b, tags)

	fmt.Fprintf(&b, risLineFormat, risTagEnd, "")

	return b.String()
}

/*
func (r *risEncoder) multipleRecordsByType(tags []string) string {
	var b strings.Builder

	for _, typ := range r.tagValues[risTagType] {
		fmt.Fprintf(&b, risLineFormat, risTagType, typ)

		r.recordBody(&b, tags)

		fmt.Fprintf(&b, risLineFormat, risTagEnd, "")
	}

	return b.String()
}
*/

func (r *risEncoder) recordBody(b *strings.Builder, tags []string) {
	for _, tag := range tags {
		switch {
		// multiple URL special handling
		case tag == risTagURL:
			fmt.Fprintf(b, risLineFormat, tag, strings.Join(r.tagValues[tag], " ; "))

		// first-entry-only tags
		case tag == risTagLibrary:
			fmt.Fprintf(b, risLineFormat, tag, r.tagValues[tag][0])

		// allowed multi-line tags
		case r.isAllowedMultilineTag(tag):
			// join any duplicates with this tag, in case it was repeated
			data := strings.Join(r.tagValues[tag], "\n")

			// clean up each line of data
			var lines []string
			for _, line := range strings.Split(data, "\n") {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" {
					lines = append(lines, trimmed)
				}
			}

			fmt.Fprintf(b, risLineFormat, tag, strings.Join(lines, risLineEnding))

		// repeatable tags
		case r.isRepeatableTag(tag):
			for _, value := range r.tagValues[tag] {
				fmt.Fprintf(b, risLineFormat, tag, value)
			}

		// simple concatenation of any other possibly duplicated tags
		default:
			fmt.Fprintf(b, risLineFormat, tag, strings.Join(r.tagValues[tag], ", "))
		}
	}
}
