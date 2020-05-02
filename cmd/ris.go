package main

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
)

const RISTypeTag = "TY"
const RISEndTag = "ER"
const RISLineFormat = "%s  - %s\r\n"

type risEncoder struct {
	fileName  string
	tagValues map[string][]string
	re        *regexp.Regexp
}

func newRisEncoder(id string) *risEncoder {
	r := risEncoder{}

	r.fileName = fmt.Sprintf("%s.ris", id)
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
	return "application/x-research-info-systems"
}

func (r *risEncoder) FileName() string {
	return r.fileName
}

func (r *risEncoder) FileContents() (string, error) {
	switch l := len(r.tagValues[RISTypeTag]); {
	case l == 0:
		err := fmt.Errorf("missing RIS Type value")
		log.Printf("%s", err.Error())
		return "", err

		/*
			case l > 1:
				err := fmt.Errorf("too many RIS Type values")
				log.Printf(err.Error())
				return "", err
		*/
	}

	tags := []string{}
	for tag := range r.tagValues {
		if tag != RISTypeTag {
			tags = append(tags, tag)
		}
	}

	sort.Strings(tags)

	data := r.recordByJoiningTypes(tags)

	return data, nil
}

func (r *risEncoder) recordByJoiningTypes(tags []string) string {
	var b strings.Builder

	types := strings.Join(r.tagValues[RISTypeTag], "; ")
	fmt.Fprintf(&b, RISLineFormat, RISTypeTag, types)

	for _, tag := range tags {
		for _, value := range r.tagValues[tag] {
			fmt.Fprintf(&b, RISLineFormat, tag, value)
		}
	}

	fmt.Fprintf(&b, RISLineFormat, RISEndTag, "")

	return b.String()
}

/*
func (r *risEncoder) recordsByType(tags []string) string {
	var b strings.Builder

	for _, typ := range r.tagValues[RISTypeTag] {
		fmt.Fprintf(&b, RISLineFormat, RISTypeTag, typ)

		for _, tag := range tags {
			for _, value := range r.tagValues[tag] {
				fmt.Fprintf(&b, RISLineFormat, tag, value)
			}
		}

		fmt.Fprintf(&b, RISLineFormat, RISEndTag, "")
	}

	return b.String()
}
*/
