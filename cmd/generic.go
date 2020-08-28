package main

import (
	"log"
	"regexp"
	"strings"
)

type citationREs struct {
	volume             *regexp.Regexp
	issue              *regexp.Regexp
	pages              *regexp.Regexp
	yearLeading        *regexp.Regexp
	yearTrailing       *regexp.Regexp
	editionFirst       *regexp.Regexp
	editionCorrect     *regexp.Regexp
	editionCorrectable *regexp.Regexp
	fieldEnd           *regexp.Regexp
	capitalizeable     *regexp.Regexp
}

var re citationREs

// data common among CMS/APA/MLA citations
type genericCitation struct {
	opts      genericCitationOpts
	isArticle bool
	citeAs    []string
	authors   []string
	editors   []string
	advisors  []string
	title     string
	journal   string
	volume    string
	issue     string
	pages     string
	edition   string
	publisher string
	date      string
	link      string
}

// options to control the slight differences in data population
type genericCitationOpts struct {
	stripProtocol  bool
	volumePrefix   bool
	issuePrefix    bool
	pagesPrefix    bool
	publisherPlace bool
}

func newGenericCitation(parts citationParts, opts genericCitationOpts) (*genericCitation, error) {
	/*
	   # If a citation has been specified, use that rather than constructing a
	   # citation from metadata.
	   result = cite_as
	   return result if result
	   result = ''

	   # Determine type of formatting required.
	   is_article = (doc_type == :article)

	   authors   = get_author_list
	   title     = setup_title_info
	   editors   = nil # TO-DO: Journal editors as opposed to book editors
	   journal   = export_journal
	   volume    = setup_volume(xxx)
	   issue     = setup_issue(xxx)
	   pages     = setup_pages(xxx)
	   edition   = setup_edition
	   publisher = setup_pub_info
	   date      = setup_pub_date
	   link      = setup_link(url, xxx)
	   an        = accession_number # jlj5aj: unused because it comes from cite_as, which overrides citation building?

	   # Make adjustments if necessary # jlj5aj: unnecessary because of above?
	   an = nil if is_article
	   date = nil if an
	*/

	c := genericCitation{opts: opts}

	// check for explicit citation
	c.citeAs = parts["explicit"]
	if len(c.citeAs) > 0 {
		return &c, nil
	}

	// set options
	c.isArticle = firstElementOf(parts["format"]) == "article"

	// set values
	c.setupAuthors(parts["author"])
	c.setupEditors(parts["editor"])
	c.setupAdvisors(parts["advisor"])
	c.setupTitle(firstElementOf(parts["title"]), firstElementOf(parts["subtitle"]))
	c.setupJournal(firstElementOf(parts["journal"]))
	c.setupVolume(firstElementOf(parts["volume"]))
	c.setupIssue(firstElementOf(parts["issue"]))
	c.setupPages(firstElementOf(parts["pages"]))
	c.setupEdition(firstElementOf(parts["edition"]))
	c.setupPublisher(firstElementOf(parts["publisher"]), firstElementOf(parts["published_location"]))
	c.setupDate(firstElementOf(parts["published_date"]))
	c.setupLink(firstElementOf(parts["url"]))

	c.log()

	return &c, nil
}

func (c *genericCitation) log() {
	log.Printf("generic citation:")

	for _, author := range c.authors {
		log.Printf("  author    : [%s]", author)
	}

	for _, editor := range c.editors {
		log.Printf("  editor    : [%s]", editor)
	}

	for _, advisor := range c.advisors {
		log.Printf("  advisor   : [%s]", advisor)
	}

	log.Printf("  title     : [%s]", c.title)

	log.Printf("  journal   : [%s]", c.journal)
	log.Printf("  volume    : [%s]", c.volume)
	log.Printf("  issue     : [%s]", c.issue)
	log.Printf("  pages     : [%s]", c.pages)
	log.Printf("  edition   : [%s]", c.edition)
	log.Printf("  publisher : [%s]", c.publisher)
	log.Printf("  date      : [%s]", c.date)
	log.Printf("  link      : [%s]", c.link)
}

func (c *genericCitation) setupAuthors(authors []string) {
	c.authors = authors
}

func (c *genericCitation) setupEditors(editors []string) {
	c.editors = editors
}

func (c *genericCitation) setupAdvisors(advisors []string) {
	c.advisors = advisors
}

func (c *genericCitation) setupTitle(title, subtitle string) {
	fullTitle := title

	if subtitle != "" {
		fullTitle = fullTitle + ": " + subtitle
	}

	c.title = fullTitle
}

func (c *genericCitation) setupJournal(journal string) {
	c.journal = journal
}

func (c *genericCitation) setupVolume(volume string) {
	if volume == "" {
		return
	}

	fullVolume := cleanEndPunctuation(volume)

	if c.opts.volumePrefix == true && re.volume.MatchString(fullVolume) == false {
		fullVolume = "vol. " + fullVolume
	}

	c.volume = fullVolume
}

func (c *genericCitation) setupIssue(issue string) {
	if issue == "" {
		return
	}

	fullIssue := cleanEndPunctuation(issue)

	if c.opts.issuePrefix == true && re.issue.MatchString(fullIssue) == false {
		fullIssue = "no. " + fullIssue
	}

	c.issue = fullIssue
}

func (c *genericCitation) setupPages(pages string) {
	if pages == "" {
		return
	}

	fullPages := re.pages.ReplaceAllString(pages, "")

	if c.opts.pagesPrefix == true {
		prefix := "p."
		if strings.Contains(fullPages, "-") {
			prefix = "pp."
		}

		fullPages = prefix + " " + fullPages
	}

	c.pages = fullPages
}

func (c *genericCitation) setupEdition(edition string) {
	fullEdition := cleanEndPunctuation(edition)

	if fullEdition == "" {
		return
	}

	switch {
	case re.editionFirst.MatchString(fullEdition):
		// first editions do not need to be specified
		fullEdition = ""

	case re.editionCorrect.MatchString(fullEdition):
		// looks good

	case re.editionCorrectable.MatchString(fullEdition):
		// close enough; we can correct these
		fullEdition = re.editionCorrectable.ReplaceAllString(fullEdition, " ed.")

	case re.editionCorrect.MatchString(fullEdition):
		fullEdition = fullEdition + " ed."
	}

	c.edition = fullEdition
}

func (c *genericCitation) setupPublisher(publisher, publishedPlace string) {
	if publisher == "" {
		return
	}

	name := cleanEndPunctuation(publisher)
	fullPublisher := name

	if c.opts.publisherPlace == true {
		place := cleanEndPunctuation(publishedPlace)

		if name != "" && place != "" {
			switch {
			case strings.Contains(name, place):
				fullPublisher = name
			case strings.Contains(place, name):
				fullPublisher = place
			default:
				fullPublisher = name + ": " + place
			}
		}
	}

	c.publisher = fullPublisher
}

func (c *genericCitation) setupDate(date string) {
	if date == "" {
		return
	}

	c.date = date

	if groups := re.yearLeading.FindStringSubmatch(date); len(groups) > 0 {
		c.date = groups[2]
		return
	}

	if groups := re.yearTrailing.FindStringSubmatch(date); len(groups) > 0 {
		c.date = groups[2]
		return
	}
}

func (c *genericCitation) setupLink(link string) {
	// TODO: implement me
	c.link = ""
}

func cleanEndPunctuation(s string) string {
	cleaned := re.fieldEnd.ReplaceAllString(s, "")

	cleaned = stripEnclosers(cleaned, '(', ')')
	cleaned = stripEnclosers(cleaned, '[', ']')

	return cleaned
}

func stripEnclosers(outer string, opener, closer rune) string {
	if strings.HasPrefix(outer, string(opener)) == false {
		return outer
	}

	if strings.HasSuffix(outer, string(closer)) == false {
		return outer
	}

	inner := outer[1 : len(outer)-1]

	diff := 0

	for _, c := range inner {
		switch {
		case c == opener:
			diff++
		case c == closer:
			diff--
		}

		if diff < 0 {
			return outer
		}
	}

	return strings.TrimSpace(inner)
}

func capitalize(s string) string {
	if re.capitalizeable.MatchString(s) == false {
		return s
	}

	return strings.ToUpper(string(s[0])) + s[1:]
}

func nameReverse(s string) string {
	// TODO: implement me
	return strings.ToUpper(s)
}

func init() {
	re.volume = regexp.MustCompile(`(?i)^vol`)
	re.issue = regexp.MustCompile(`(?i)^(n[ou]|iss)`)
	re.pages = regexp.MustCompile(`p{1,2}\.\s+`)
	re.yearLeading = regexp.MustCompile(`^(\s*)(\d{4})(.*)$`)
	re.yearTrailing = regexp.MustCompile(`^(.*)(\d{4})(\s*)$`)
	re.editionFirst = regexp.MustCompile(`(?i)^(1st|first)`)
	re.editionCorrect = regexp.MustCompile(`(?i) eds?\.( |$)`)
	re.editionCorrectable = regexp.MustCompile(`(?i) ed(|ition)[[:punct:]]*$`)
	re.fieldEnd = regexp.MustCompile(`[,;:\/\s]+$`)
	re.capitalizeable = regexp.MustCompile(`^[a-z][a-z\s]`)
}
