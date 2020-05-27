package main

import (
	"fmt"
	"log"
	"path"
	"regexp"
	"sort"
	"strings"
)

const RISTagType = "TY"
const RISTagAuthor = "AU"
const RISTagNote = "N1"
const RISTagKeyword = "KW"
const RISTagURL = "UR"
const RISTagEnd = "ER"

const RISTypeGeneric = "GEN"
const RISLineFormat = "%s  - %s\r\n"

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
	if len(r.tagValues[RISTagType]) == 0 {
		r.addTagValue(RISTagType, RISTypeGeneric)
	}

	r.addTagValue(RISTagNote, r.url)

	tags := []string{}
	for tag := range r.tagValues {
		if tag != RISTagType {
			tags = append(tags, tag)
		}
	}

	sort.Strings(tags)

	data := r.singleRecordByJoiningTypes(tags)

	return data, nil
}

func (r *risEncoder) singleRecordByJoiningTypes(tags []string) string {
	var b strings.Builder

	types := strings.Join(r.tagValues[RISTagType], "; ")
	fmt.Fprintf(&b, RISLineFormat, RISTagType, types)

	r.recordBody(&b, tags)

	fmt.Fprintf(&b, RISLineFormat, RISTagEnd, "")

	return b.String()
}

/*
func (r *risEncoder) multipleRecordsByType(tags []string) string {
	var b strings.Builder

	for _, typ := range r.tagValues[RISTagType] {
		fmt.Fprintf(&b, RISLineFormat, RISTagType, typ)

		r.recordBody(&b, tags)

		fmt.Fprintf(&b, RISLineFormat, RISTagEnd, "")
	}

	return b.String()
}
*/

func (r *risEncoder) recordBody(b *strings.Builder, tags []string) {
	for _, tag := range tags {
		switch {
		// multiple URL special handling
		case tag == RISTagURL:
			fmt.Fprintf(b, RISLineFormat, tag, strings.Join(r.tagValues[tag], " ; "))

		// repeatable fields
		case tag == RISTagAuthor || tag == RISTagNote || tag == RISTagKeyword:
			for _, value := range r.tagValues[tag] {
				fmt.Fprintf(b, RISLineFormat, tag, value)
			}

		// simple concatenation of any other possibly duplicated fields
		default:
			fmt.Fprintf(b, RISLineFormat, tag, strings.Join(r.tagValues[tag], ", "))
		}
	}
}
