package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
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
	trailingPeriods    *regexp.Regexp
	capitalizeable     *regexp.Regexp
	doubleQuoted       *regexp.Regexp
	lowerLastNamePart  *regexp.Regexp
	doiPrefix          *regexp.Regexp
	doiURL             *regexp.Regexp
	urlProtocol        *regexp.Regexp
	year               *regexp.Regexp
	romanNumeral       *regexp.Regexp
	leadingYear        *regexp.Regexp
	trailingYear       *regexp.Regexp
}

var re citationREs
var nameSuffixMap map[string]bool
var lowerCaseWordMap map[string]bool

// data common among CMS/APA/MLA citations
type genericCitation struct {
	v4url           string
	opts            genericCitationOpts
	isArticle       bool
	citeAs          []string
	authors         []string
	editors         []string
	advisors        []string
	compilers       []string
	translators     []string
	title           string
	format          string
	journal         string
	volume          string
	issue           string
	pages           string
	pageFrom        string
	pageTo          string
	edition         string
	publisher       string
	publicationType string
	dataSource      string
	date            string
	link            string
	year            int
	month           int
	day             int
}

// options to control the slight differences in data population
type genericCitationOpts struct {
	verbose        bool
	stripProtocol  bool
	volumePrefix   bool
	issuePrefix    bool
	pagesPrefix    bool
	publisherPlace bool
}

func newGenericCitation(v4url string, parts citationParts, opts genericCitationOpts) (*genericCitation, error) {
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
	c.v4url = v4url

	// set values
	citeAs := parts["explicit"]
	authors := parts["author"]
	editors := parts["editor"]
	advisors := parts["advisor"]
	compilers := parts["compiler"]
	translators := parts["translator"]
	title := firstElementOf(parts["title"])
	format := firstElementOf(parts["format"])
	subtitle := firstElementOf(parts["subtitle"])
	journal := firstElementOf(parts["journal"])
	volume := firstElementOf(parts["volume"])
	issue := firstElementOf(parts["issue"])
	pages := firstElementOf(parts["pages"])
	edition := firstElementOf(parts["edition"])
	publisher := firstElementOf(parts["publisher"])
	publishedLocation := firstElementOf(parts["published_location"])
	publicationType := firstElementOf(parts["publication_type"])
	dataSource := firstElementOf(parts["data_source"])
	date := firstElementOf(parts["published_date"])
	url := firstElementOf(parts["url"])
	doi := firstElementOf(parts["doi"])
	serialNumbers := parts["serial_number"]
	isOnlineOnly := firstElementOf(parts["is_online_only"])
	isVirgoURL := firstElementOf(parts["is_virgo_url"])

	// set options
	c.isArticle = format == "article"

	c.setupCiteAs(citeAs)
	c.setupAuthors(authors)
	c.setupEditors(editors)
	c.setupAdvisors(advisors)
	c.setupCompilers(compilers)
	c.setupTranslators(translators)
	c.setupTitle(title, subtitle)
	c.setupFormat(format)
	c.setupJournal(journal)
	c.setupVolume(volume)
	c.setupIssue(issue)
	c.setupPages(pages)
	c.setupEdition(edition)
	c.setupPublisher(publisher, publishedLocation)
	c.setupPublicationType(publicationType)
	c.setupDataSource(dataSource)
	c.setupDate(date)
	c.setupLink(url, doi, isOnlineOnly, isVirgoURL, serialNumbers)

	c.log(parts)

	return &c, nil
}

func (c *genericCitation) log(parts citationParts) {
	if c.opts.verbose == false {
		return
	}

	log.Printf("collected parts:")
	for k, v := range parts {
		log.Printf("    %-20s : %#v", k, v)
	}

	log.Printf("generic citation:")

	for _, author := range c.authors {
		log.Printf("    author          : [%s]", author)
	}

	for _, editor := range c.editors {
		log.Printf("    editor          : [%s]", editor)
	}

	for _, advisor := range c.advisors {
		log.Printf("    advisor         : [%s]", advisor)
	}

	for _, compiler := range c.compilers {
		log.Printf("    compiler        : [%s]", compiler)
	}

	for _, translator := range c.translators {
		log.Printf("    translator      : [%s]", translator)
	}

	log.Printf("    title           : [%s]", c.title)
	log.Printf("    format          : [%s]", c.format)
	log.Printf("    journal         : [%s]", c.journal)
	log.Printf("    volume          : [%s]", c.volume)
	log.Printf("    issue           : [%s]", c.issue)
	log.Printf("    pages           : [%s]", c.pages)
	log.Printf("    pageFrom        : [%s]", c.pageFrom)
	log.Printf("    pageTo          : [%s]", c.pageTo)
	log.Printf("    edition         : [%s]", c.edition)
	log.Printf("    publisher       : [%s]", c.publisher)
	log.Printf("    publicationType : [%s]", c.publicationType)
	log.Printf("    dataSource      : [%s]", c.dataSource)
	log.Printf("    date            : [%s]  (%d) (%d) (%d)", c.date, c.year, c.month, c.day)
	log.Printf("    link            : [%s]", c.link)
}

func (c *genericCitation) setupCiteAs(citeAs []string) {
	c.citeAs = citeAs
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

func (c *genericCitation) setupCompilers(compilers []string) {
	c.compilers = compilers
}

func (c *genericCitation) setupTranslators(translators []string) {
	c.translators = translators
}

func (c *genericCitation) setupTitle(title, subtitle string) {
	fullTitle := title

	if subtitle != "" {
		fullTitle = fullTitle + ": " + subtitle
	}

	fullTitle = removeTrailingPeriods(fullTitle)

	c.title = fullTitle
}

func (c *genericCitation) setupFormat(format string) {
	c.format = format
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
	ends := wordsBySeparator(re.pages.ReplaceAllString(pages, ""), "-")

	pageRange := ""
	prefix := ""

	pageFrom := ""
	pageTo := ""

	switch {
	case len(ends) == 2:
		prefix = "pp."
		pageFrom = strings.TrimSpace(ends[0])
		pageTo = strings.TrimSpace(ends[1])
		pageRange = pageFrom + " - " + pageTo

	case len(ends) == 1:
		prefix = "p."
		pageFrom = strings.TrimSpace(ends[0])
		pageRange = pageFrom
	}

	fullPages := ""

	if pageRange != "" {
		if c.opts.pagesPrefix == true {
			fullPages = prefix + " " + pageRange
		} else {
			fullPages = pageRange
		}
	}

	c.pages = fullPages
	c.pageFrom = pageFrom
	c.pageTo = pageTo
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

func (c *genericCitation) setupPublicationType(publicationType string) {
	c.publicationType = strings.ToLower(publicationType)
}

func (c *genericCitation) setupDataSource(dataSource string) {
	c.dataSource = strings.ToLower(dataSource)
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
				fullPublisher = place + ": " + name
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
		// that year, since the eds pool is sending YYYY-01-01.  until known, just trust month/day:
		c.month = int(t.Month())
		c.day = t.Day()
		return
	}

	// fallback: attempt to extract leading/trailing year

	if re.leadingYear.MatchString(date) == true {
		if groups := re.leadingYear.FindStringSubmatch(date); len(groups) > 0 {
			yearStr := groups[1]
			if year, err := strconv.Atoi(yearStr); err == nil {
				c.date = yearStr
				c.year = year
				return
			}
		}
	}

	if re.trailingYear.MatchString(date) == true {
		if groups := re.trailingYear.FindStringSubmatch(date); len(groups) > 0 {
			yearStr := groups[1]
			if year, err := strconv.Atoi(yearStr); err == nil {
				c.date = yearStr
				c.year = year
				return
			}
		}
	}
}

func (c *genericCitation) setupLink(url, doi, isOnlineOnly, isVirgoURL string, serialNumbers []string) {
	c.link = ""

	/*
	   # Get the link (DOI or URL) for the item for use in citations.
	   #
	   # DOI is preferred; if present it is converted into URL form.
	   #
	   # NOTE: Because the OpenURL link for articles is sort of ugly, this method
	   # will return *nil* for articles unless they have a DOI.
	   #
	   # @param [String,Boolean] url     If set (not *nil* or *false*) then return
	   #                                   a link even if the item is not an
	   #                                   exclusively online item.  If a String,
	   #                                   then also use that value in place of
	   #                                   the result from #export_url.
	   # @param [Array]  args
	   #
	   # @return [String]
	   # @return [nil]
	   #
	   def setup_link(url = nil, *args)
	     # Supply/override *url* as needed.
	     if (doi = dois.first.presence)
	       # Prefer DOI if it's directly available.
	       url = 'https://doi.org/' + doi.sub(/^doi:/, '')
	     elsif (u = get_url.first) && (u =~ %r{^https?://(\w+\.)?doi\.org/})
	       # Prefer DOI if it's present indirectly -- even if a *url* parameter
	       # was provided.
	       url = u
	     elsif url.blank? && !born_digital?
	       # Only show a link if this is an electronic item.
	       return
	     end

	     # Only show a URL if this is an electronic item.
	     return if url.blank? || !born_digital?

	     # Create the URL as a link unless the context requires simple text.
	     opt = args.last.is_a?(Hash) ? args.last.dup : {}
	     result  = opt.delete(:strip_protocol) ? url.sub(%r{^\w+://}, '') : url
	     context = opt.delete(:context)
	     if context && ![:email, :export].include?(context)
	       link_to(result, url, opt)
	     else
	       result
	     end
	   end

	   # Indicate whether this document has online content which did not have an
	   # original form that was published through print media.
	   #
	   # For citation purposes, a born-digital item includes a "Retrieved from"
	   # notation.
	   #
	   def born_digital?(*)
	     # TO-DO: Verify this definition...
	     online_only? && isbns.blank? && issns.blank?
	   end
	*/

	isBornDigital := isOnlineOnly == "true" && len(serialNumbers) == 0

	link := ""
	if isVirgoURL == "true" {
		link = c.v4url
	}

	switch {
	case doi != "":
		link = "https://doi.org/" + re.doiPrefix.ReplaceAllString(doi, "")

	case url != "" && re.doiURL.MatchString(url):
		link = url

	case url == "" && isBornDigital == false:
		return
	}

	if link == "" || isBornDigital == false {
		return
	}

	fullLink := link
	if c.opts.stripProtocol == true {
		link = re.urlProtocol.ReplaceAllString(fullLink, "")
	}

	c.link = fmt.Sprintf(`<a href="%s">%s</a>`, fullLink, link)
}

func cleanEndPunctuation(s string) string {
	cleaned := re.fieldEnd.ReplaceAllString(s, "")

	cleaned = stripEnclosers(cleaned, '(', ')')
	cleaned = stripEnclosers(cleaned, '[', ']')

	return cleaned
}

func removeTrailingPeriods(s string) string {
	cleaned := re.trailingPeriods.ReplaceAllString(s, "")

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

	if strings.TrimSpace(name) == "" || strings.Contains(name, ")") == true || strings.Contains(name, "(") == true {
		return name
	}

	commaParts := wordsBySeparator(name, ",")
	var suffixParts []string

	var lastName string
	var otherNames string
	var suffixes string

	for {
		if len(commaParts) == 0 {
			break
		}

		last := commaParts[len(commaParts)-1]

		if nameSuffixMap[last] == false {
			break
		}

		suffixParts = append([]string{last}, suffixParts...)
		commaParts = commaParts[:len(commaParts)-1]
	}

	if len(suffixParts) > 0 {
		suffixes = strings.Join(suffixParts, ", ")
		if len(commaParts) > 0 {
			suffixes = ", " + suffixes
		}
	}

	switch {
	case len(commaParts) > 1:
		lastName, commaParts = commaParts[0], commaParts[1:]
		otherNames = strings.Join(commaParts, ", ")

	case len(commaParts) == 1:
		nameParts := wordsBySeparator(commaParts[0], " ")

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
				nameParts = nameParts[:len(nameParts)-1]
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
				nameParts = nameParts[:len(nameParts)-1]
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

func abbreviateName(name string) string {
	/*
	   # Transforms the form "last_name[, first_name [middle_names][, suffix]"
	   # into the form "last_name[, initials][, suffix]".
	   #
	   # For now it's assumed that if a name has parentheses then it's probably a
	   # corporate name that can't be "abbreviated" in this way.  Names like this
	   # are returned unmodified.
	   #
	   # @param [String] name
	   #
	   # @return [String]
	   #
	   # Replaces:
	   # @see Blacklight::Solr::Document::MarcExport#abbreviate_name
	   #
	   def abbreviate_name(name, *)
	     name = name.to_s.squish
	     return name if name.blank?

	     # Assuming this is a corporate name.
	     return name if name.include?('(') || name.include?(')')

	     # If there are no commas then assume this a corporate name or a person
	     # with only one name (e.g. "Christo", "Michelangelo").  Remove date(s)
	     # from the end if the name includes them.
	     parts  = name.split(',').map(&:strip).reject(&:blank?)
	     result = parts.shift.to_s
	     parts.pop if parts.last =~ /\d{4}/

	     # The rest of the parts (first name and/or middle name(s)) are turned
	     # into initials.  However, suffixes remain untouched:
	     # (1) Roman numerals (e.g. "Henry VIII")
	     # (2) Suffix         (e.g. "Martin Luther King, Jr.")
	     if parts.present?
	       result << ', '
	       result <<
	         parts.map { |part|
	           part.split(SPACE).map { |s|
	             next if LOWERCASE_WORDS.include?(s)
	             keep = s.include?('.')
	             keep ||= (s =~ /^([IX][IVX]*|VI*)$/i)   # (1)
	             keep ||= NAME_SUFFIX_WORDS.include?(s)  # (2)
	             keep ? s : "#{s[0]}."
	           }.compact.join(SPACE)
	         }.join(', ')
	     end

	     capitalize(result)
	   end
	*/

	if strings.TrimSpace(name) == "" || strings.Contains(name, ")") == true || strings.Contains(name, "(") == true {
		return name
	}

	parts := wordsBySeparator(name, ",")

	res := ""

	res, parts = parts[0], parts[1:]

	if len(parts) > 0 && re.year.MatchString(parts[len(parts)-1]) == true {
		parts = parts[:len(parts)-1]
	}

	if len(parts) > 0 {
		res += ", "

		var resParts []string

		for _, part := range parts {
			var resPieces []string

			for _, piece := range wordsBySeparator(part, " ") {
				if lowerCaseWordMap[strings.ToLower(piece)] == true {
					continue
				}

				keep := strings.Contains(piece, ".")
				keep = keep || re.romanNumeral.MatchString(piece)
				keep = keep || nameSuffixMap[piece]

				if keep == false {
					piece = string(piece[0]) + "."
				}

				resPieces = append(resPieces, piece)
			}

			resParts = append(resParts, strings.Join(resPieces, " "))
		}

		res += strings.Join(resParts, ", ")
	}

	res = capitalize(res)

	return res
}

func doubleToSingleQuotes(s string) string {
	return re.doubleQuoted.ReplaceAllString(s, `'`)
}

func monthName(m int) string {
	if m < 1 || m > 12 {
		return ""
	}

	t := time.Month(m)

	return t.String()
}

func appendWithComma(str, part string) string {
	if part == "" {
		return str
	}

	res := str
	res = appendUnlessEndsWith(res, ",", []string{" ", ".", ","})
	res = appendUnlessEndsWith(res, " ", []string{" "})
	res += part

	return res
}

func appendUnlessEndsWith(str, part string, ends []string) string {
	if str == "" {
		return str
	}

	for _, end := range ends {
		if strings.HasSuffix(str, end) == true {
			return str
		}
	}

	res := str + part

	return res
}

func wordsBySeparator(word, separator string) []string {
	var words []string

	for _, p := range strings.Split(word, separator) {
		part := strings.TrimSpace(p)
		if part == "" {
			continue
		}

		words = append(words, part)
	}

	return words
}

func smallCaps(s string) string {
	return fmt.Sprintf(`<span style="font-variant: small-caps;">%s</span>`, s)
}

func italics(s string) string {
	return fmt.Sprintf(`<em>%s</em>`, s)
}

func init() {
	re.volume = regexp.MustCompile(`(?i)^vol`)
	re.issue = regexp.MustCompile(`(?i)^(n[ou]|iss)`)
	re.pages = regexp.MustCompile(`p{1,2}\.\s+`)
	re.editionFirst = regexp.MustCompile(`(?i)^(1st|first)`)
	re.editionCorrect = regexp.MustCompile(`(?i) eds?\.( |$)`)
	re.editionCorrectable = regexp.MustCompile(`(?i) ed(|ition)[[:punct:]]*$`)
	re.fieldEnd = regexp.MustCompile(`[,;:\/\s]+$`)
	re.trailingPeriods = regexp.MustCompile(`\.+$`)
	re.capitalizeable = regexp.MustCompile(`^[a-z][a-z\s]`)
	re.doubleQuoted = regexp.MustCompile(`(?U)[#{"}\p{Pi}\p{Pf}]`)
	re.lowerLastNamePart = regexp.MustCompile(`(?U)^([[:lower:]])+([[[:lower:]]\s.-])*$`)
	re.doiPrefix = regexp.MustCompile(`^doi:`)
	re.doiURL = regexp.MustCompile(`^https?://(\w+\.)?doi\.org/`)
	re.urlProtocol = regexp.MustCompile(`^\w+://`)
	re.year = regexp.MustCompile(`\d{4}`)
	re.romanNumeral = regexp.MustCompile(`(?i)^([IX][IVX]*|VI*)$`)
	re.leadingYear = regexp.MustCompile(`^\s*(\d{1,4}).*?$`)
	re.trailingYear = regexp.MustCompile(`^.*?(\d{1,4})\s*$`)

	var list []string

	nameSuffixMap = make(map[string]bool)

	list = []string{
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

	for _, s := range list {
		nameSuffixMap[s] = true
	}

	lowerCaseWordMap = make(map[string]bool)

	list = []string{
		"van der",
		"van den",
		"and",
		"ca.",
		"del",
		"des",
		"etc",
		"for",
		"les",
		"los",
		"nor",
		"pp.",
		"the",
		"und",
		"van",
		"von",
		"vs.",
		"an",
		"as",
		"at",
		"by",
		"de",
		"di",
		"du",
		"el",
		"et",
		"in",
		"la",
		"le",
		"of",
		"on",
		"or",
		"to",
		"v.",
		"a",
		"à",
		"e",
		"o",
		"y",
	}

	for _, s := range list {
		lowerCaseWordMap[s] = true
	}
}
