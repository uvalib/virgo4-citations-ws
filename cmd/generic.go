package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

type citationREs struct {
	volume             *regexp.Regexp
	issue              *regexp.Regexp
	pages              *regexp.Regexp
	editionFirst       *regexp.Regexp
	editionCorrect     *regexp.Regexp
	editionCorrectable *regexp.Regexp
	fieldEnd           *regexp.Regexp
	capitalizeable     *regexp.Regexp
	doubleQuoted       *regexp.Regexp
	lowerLastNamePart  *regexp.Regexp
}

var re citationREs
var nameSuffixes map[string]bool

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
	year      int
	month     int
	day       int
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
		log.Printf("    author    : [%s]", author)
	}

	for _, editor := range c.editors {
		log.Printf("    editor    : [%s]", editor)
	}

	for _, advisor := range c.advisors {
		log.Printf("    advisor   : [%s]", advisor)
	}

	log.Printf("    title     : [%s]", c.title)
	log.Printf("    journal   : [%s]", c.journal)
	log.Printf("    volume    : [%s]", c.volume)
	log.Printf("    issue     : [%s]", c.issue)
	log.Printf("    pages     : [%s]", c.pages)
	log.Printf("    edition   : [%s]", c.edition)
	log.Printf("    publisher : [%s]", c.publisher)
	log.Printf("    date      : [%s]  (%d) (%d) (%d)", c.date, c.year, c.month, c.day)
	log.Printf("    link      : [%s]", c.link)
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
	fullVolume := cleanEndPunctuation(volume)

	if fullVolume != "" && c.opts.volumePrefix == true && re.volume.MatchString(fullVolume) == false {
		fullVolume = "vol. " + fullVolume
	}

	c.volume = fullVolume
}

func (c *genericCitation) setupIssue(issue string) {
	fullIssue := cleanEndPunctuation(issue)

	if fullIssue != "" && c.opts.issuePrefix == true && re.issue.MatchString(fullIssue) == false {
		fullIssue = "no. " + fullIssue
	}

	c.issue = fullIssue
}

func (c *genericCitation) setupPages(pages string) {
	fullPages := re.pages.ReplaceAllString(pages, "")

	if fullPages != "" && c.opts.pagesPrefix == true {
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

	default:
		fullEdition = fullEdition + " ed."
	}

	c.edition = fullEdition
}

func (c *genericCitation) setupPublisher(publisher, publishedPlace string) {
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
	c.date = ""
	c.year = 0
	c.month = 0
	c.day = 0

	// solr: YYYY
	if t, err := time.Parse("2006", date); err == nil {
		c.date = date
		c.year = t.Year()
		return
	}

	// eds: YYYY-MM-DD (but allow for M or D as well)
	if t, err := time.Parse("2006-1-2", date); err == nil {
		c.date = date
		c.year = t.Year()
		// TODO: check if eds pool is effectively sending 01 placeholders for unknown MM and DD.
		// have seen records citations in V3 with just a year, but in V4 they would be 1/1 of
		// that year, since the eds pool is sending YYYY-01-01.  until known, omit month/day:
		//c.month = int(t.Month())
		//c.day = t.Day()
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

func readingOrder(name string) string {
	/*
	   # Transforms the form "last_name, first_name [middle_names][, suffix]"
	   # into the form "first_name [middle_names] last_name[, suffix]".  If the
	   # name could not be reversed, the original name is returned.
	   #
	   # The method is somewhat poorly-named because the intent is really to
	   # restore the reading order of a name that is in bibliographic order.
	   #
	   # @param [String] name
	   #
	   # @return [String]
	   #
	   # Replaces:
	   # @see Blacklight::Solr::Document::MarcExport#name_reverse
	   #
	   # === Usage Notes
	   # If *name* is in bibliographic order, make sure that the surname is
	   # capitalized appropriately.  E.g., if *name* is "De la Croix, Jean" then
	   # the result will be "Jean De la Croix", whereas "de la Croix, Jean" will
	   # result in "Jean de la Croix".
	   #
	   def name_reverse(name, *)
	     name = name.to_s.squish
	     return name if name.blank?

	     # Assuming this is a corporate name that should not be reversed.
	     return name if name.include?('(') || name.include?(')')

	     # Remove suffix(es) from the end; these will stay at the end of the name
	     # even after the other parts are reordered.
	     comma_parts  = name.split(',').map(&:strip).reject(&:blank?)
	     suffix_parts = []
	     while NAME_SUFFIX_WORDS.include?(comma_parts.last)
	       suffix_parts.unshift(comma_parts.pop)
	     end
	     suffixes = (', ' << suffix_parts.join(', ') unless suffix_parts.empty?)

	     # If the remaining name has two or more comma-separated parts then assume
	     # it was in bibliographic order.  Otherwise, extract the last name parts,
	     # leaving behind the given name(s).
	     if comma_parts.size > 1
	       last_name   = comma_parts.shift
	       other_names = comma_parts.join(', ').presence
	     else
	       name_parts  = comma_parts.first.split(' ')
	       last_name   = extract_last_name!(name_parts)
	       other_names = name_parts.join(' ').presence
	     end
	     other_names << ' ' if other_names

	     # Return with the names in reading order.
	     "#{other_names}#{last_name}#{suffixes}"
	   end
	*/

	if name == "" || strings.Contains(name, ")") == true || strings.Contains(name, "(") == true {
		return name
	}

	var commaParts []string
	var suffixParts []string

	var lastName string
	var otherNames string
	var suffixes string

	for _, p := range strings.Split(name, ",") {
		part := strings.TrimSpace(p)
		if part == "" {
			continue
		}

		commaParts = append(commaParts, part)
	}

	for {
		if len(commaParts) == 0 {
			break
		}

		last := commaParts[len(commaParts)-1]

		if nameSuffixes[last] == false {
			break
		}

		suffixParts = append([]string{last}, suffixParts...)
		commaParts = commaParts[:len(commaParts)-1]
	}

	if len(suffixParts) > 0 {
		suffixes = fmt.Sprintf(", %s", strings.Join(suffixParts, ", "))
	}

	if len(commaParts) > 1 {
		lastName, commaParts = commaParts[0], commaParts[1:]
		otherNames = strings.Join(commaParts, ", ")
	} else {
		// FIXME: potential out-of-bounds for certain input?  V3 did not seem to check for this, so we won't either
		nameParts := strings.Split(commaParts[0], " ")

		/*
		   # Remove the elements from the end of *name_parts* which appear to be a
		   # last name, accounting for multi-part names like "de la Croix" or
		   # "v. Ribbentrop".
		   #
		   # @param [Array<String>] name_parts   Array to be modified.
		   #
		   # @return [String]
		   #
		   # === Implementation Notes
		   # It is assumed that the full name represented by *name_parts* is comprised
		   # of zero or more "given" names and a surname which may begin with zero or
		   # more lowercase words (like "de" or "la") followed by one or more surnames
		   # (or ordinal designations like "VIII") which each begin with a capital
		   # (although the surnames themselves may contain spaces).
		   #
		   def extract_last_name!(name_parts)
		     surname   = name_parts.pop
		     lowercase = /^\p{Lower}+([\s.-]\p{Lower})*$/u
		     if name_parts.any? { |part| part =~ lowercase }
		       result = [surname]
		       result.unshift(name_parts.pop) while name_parts.last !~ lowercase
		       result.unshift(name_parts.pop) while name_parts.last =~ lowercase
		       result.join(' ')
		     else
		       surname
		     end
		   end
		*/

		lastName, nameParts = nameParts[len(nameParts)-1], nameParts[:len(nameParts)-1]

		hasLowerLast := false
		for _, part := range nameParts {
			if re.lowerLastNamePart.MatchString(part) == true {
				hasLowerLast = true
				break
			}
		}

		if hasLowerLast == true {
			lastParts := []string{lastName}

			for {
				if len(nameParts) == 0 {
					break
				}

				lastPart := nameParts[len(nameParts)-1]

				if re.lowerLastNamePart.MatchString(lastPart) == true {
					break
				}

				lastParts = append([]string{lastPart}, lastParts...)
			}

			for {
				if len(nameParts) == 0 {
					break
				}

				lastPart := nameParts[len(nameParts)-1]

				if re.lowerLastNamePart.MatchString(lastPart) == false {
					break
				}

				lastParts = append([]string{lastPart}, lastParts...)
			}

			lastName = strings.Join(lastParts, " ")
		}

		otherNames = strings.Join(nameParts, " ")
	}

	if otherNames != "" {
		otherNames += " "
	}

	return fmt.Sprintf("%s%s%s", otherNames, lastName, suffixes)
}

func doubleToSingleQuotes(s string) string {
	return re.doubleQuoted.ReplaceAllString(s, `'`)
}

func monthName(m int) string {
	if m < 1 || m > 12 {
		return ""
	}

	t := time.Month(m)
	return t.String()[:3]
}

func init() {
	re.volume = regexp.MustCompile(`(?i)^vol`)
	re.issue = regexp.MustCompile(`(?i)^(n[ou]|iss)`)
	re.pages = regexp.MustCompile(`p{1,2}\.\s+`)
	re.editionFirst = regexp.MustCompile(`(?i)^(1st|first)`)
	re.editionCorrect = regexp.MustCompile(`(?i) eds?\.( |$)`)
	re.editionCorrectable = regexp.MustCompile(`(?i) ed(|ition)[[:punct:]]*$`)
	re.fieldEnd = regexp.MustCompile(`[,;:\/\s]+$`)
	re.capitalizeable = regexp.MustCompile(`^[a-z][a-z\s]`)
	re.doubleQuoted = regexp.MustCompile(`(?U)[#{"}\p{Pi}\p{Pf}]`)
	re.lowerLastNamePart = regexp.MustCompile(`(?U)^([[:lower:]])+([[[:lower:]]\s.-])*$`)

	nameSuffixes = make(map[string]bool)

	suffixes := []string{
		"B.A.",
		"BA",
		"B.S.",
		"BS",
		"D.D.",
		"DD",
		"D.Phil.",
		"DPhil",
		"D.D.S.",
		"DDS",
		"Ed.D.",
		"EdD",
		"Esquire",
		"Esq.",
		"J.D.",
		"JD",
		"Junior",
		"Jnr",
		"Jr.",
		"Jr",
		"LL.D.",
		"LLD",
		"M.B.A.",
		"MBA",
		"M.D.",
		"MD",
		"M.A.",
		"MA",
		"Ph.D.",
		"PhD",
		"R.N.",
		"RN",
		"Senior",
		"Snr",
		"Sr.",
		"Sr",
	}

	for _, suffix := range suffixes {
		nameSuffixes[suffix] = true
	}
}
