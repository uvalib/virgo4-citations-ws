package main

import (
	"fmt"
	"log"
	"path"
	"regexp"
	"sort"
	"strings"
)

const RISTypeTag = "TY"
const RISURLTag = "UR"
const RISSerialNumberTag = "SN"
const RISEndTag = "ER"
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
	if len(r.tagValues[RISTypeTag]) == 0 {
		r.addTagValue(RISTypeTag, RISTypeGeneric)
	}

	if len(r.tagValues[RISURLTag]) == 0 {
		r.addTagValue(RISURLTag, r.url)
	}

	tags := []string{}
	for tag := range r.tagValues {
		if tag != RISTypeTag {
			tags = append(tags, tag)
		}
	}

	sort.Strings(tags)

	data := r.singleRecordByJoiningTypes(tags)

	return data, nil
}

func (r *risEncoder) singleRecordByJoiningTypes(tags []string) string {
	var b strings.Builder

	types := strings.Join(r.tagValues[RISTypeTag], "; ")
	fmt.Fprintf(&b, RISLineFormat, RISTypeTag, types)

	r.recordBody(&b, tags)

	fmt.Fprintf(&b, RISLineFormat, RISEndTag, "")

	return b.String()
}

/*
func (r *risEncoder) multipleRecordsByType(tags []string) string {
	var b strings.Builder

	for _, typ := range r.tagValues[RISTypeTag] {
		fmt.Fprintf(&b, RISLineFormat, RISTypeTag, typ)

		r.recordBody(&b, tags)

		fmt.Fprintf(&b, RISLineFormat, RISEndTag, "")
	}

	return b.String()
}
*/

func (r *risEncoder) recordBody(b *strings.Builder, tags []string) {
	for _, tag := range tags {
		switch tag {
		case RISSerialNumberTag:
			fmt.Fprintf(b, RISLineFormat, tag, strings.Join(r.tagValues[tag], ", "))

		default:
			for _, value := range r.tagValues[tag] {
				fmt.Fprintf(b, RISLineFormat, tag, value)
			}
		}
	}
}
