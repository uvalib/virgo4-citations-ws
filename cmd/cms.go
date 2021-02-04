package main

import (
	"strings"
)

type cmsEncoder struct {
	cfg          serviceConfigFormat
	url          string
	preferCiteAs bool
	data         *genericCitation
}

func newCmsEncoder(cfg serviceConfigFormat, preferCiteAs bool) *cmsEncoder {
	e := cmsEncoder{}

	e.cfg = cfg
	e.preferCiteAs = preferCiteAs

	return &e
}

func (e *cmsEncoder) Init(url string) {
	e.url = url
}

func (e *cmsEncoder) Populate(parts citationParts) error {
	/*
	   opt[:strip_protocol] = true

	   volume    = setup_volume(true)
	   issue     = setup_issue(true)
	   pages     = setup_pages(true)
	   publisher = setup_pub_info
	*/

	var err error

	opts := genericCitationOpts{
		stripProtocol:  true,
		volumePrefix:   true,
		issuePrefix:    true,
		pagesPrefix:    true,
		publisherPlace: true,
	}

	if e.data, err = newGenericCitation(e.url, parts, opts); err != nil {
		return err
	}

	return nil
}

func (e *cmsEncoder) Label() string {
	return e.cfg.Label
}

func (e *cmsEncoder) ContentType() string {
	return e.cfg.ContentType
}

func (e *cmsEncoder) FileName() string {
	return ""
}

func (e *cmsEncoder) Contents() (string, error) {
	if e.preferCiteAs == true && len(e.data.citeAs) > 0 {
		return strings.Join(e.data.citeAs, "\n"), nil
	}

	res := ""

	/*
	   # === Author(s)
	   # First author in "Last, First" form; second author in "First Last" form;
	   # three or more authors shown as the first author followed by ", et al.".
	   # If the author is the same as the publisher, skip the author.
	   eds   = []
	   comps = []
	   trans = []
	   get_related_names(false).each do |name_and_role|
	     name_and_role = name_and_role.to_s
	     if (name = name_and_role.sub!(/\W+Editor\W*$/i, ''))
	       eds   << name
	     elsif (name = name_and_role.sub!(/\W+Compiler\W*$/i, ''))
	       comps << name
	     elsif (name = name_and_role.sub!(/\W+Translator\W*$/i, ''))
	       trans << name
	     end
	   end
	   authors.delete_if do |v|
	     (v == publisher) || comps.include?(v) || trans.include?(v)
	   end
	   if authors.present?
	     result << cmos_names(authors, true)
	     if (authors - eds).empty?
	       eds = []
	       result << ', ed'
	       result << 's' if authors.size > 1
	     elsif (authors - comps).empty?
	       comps = []
	       result << ', comp'
	       result << 's' if authors.size > 1
	     elsif (authors - trans).empty?
	       trans = []
	       result << ', trans'
	     end
	     result << '.'
	   end
	*/

	pub := []string{e.data.publisher}
	pubCompTrans := pub
	pubCompTrans = append(pubCompTrans, e.data.compilers...)
	pubCompTrans = append(pubCompTrans, e.data.translators...)

	authors := removeEntries(e.data.authors, pubCompTrans)
	editors := removeEntries(e.data.editors, pubCompTrans)
	advisors := removeEntries(e.data.advisors, pubCompTrans)
	compilers := removeEntries(e.data.compilers, pub)
	translators := removeEntries(e.data.translators, pub)

	var creators []string
	creators = append(creators, authors...)
	creators = append(creators, editors...)
	creators = append(creators, advisors...)

	numCreators := len(creators)
	if numCreators > 0 {
		res += cmsNames(creators)

		nonEditors := removeEntries(creators, editors)
		if len(nonEditors) == 0 {
			editors = []string{}
			res += ", ed"
			if numCreators > 1 {
				res += "s"
			}
		}

		nonCompilers := removeEntries(creators, compilers)
		if len(nonCompilers) == 0 {
			compilers = []string{}
			res += ", comp"
			if numCreators > 1 {
				res += "s"
			}
		}

		nonTranslators := removeEntries(creators, translators)
		if len(nonTranslators) == 0 {
			translators = []string{}
			res += ", trans"
		}

		res += "."
	}

	/*
	   # === Item Title
	   # Titles of larger works (books, journals, etc) are italicized; title of
	   # shorter works (poems, articles, etc) are in quotes.  If the article
	   # title contains double quotes, convert them to single quotes before
	   # wrapping the title in double quotes.
	   if title.present?
	     result << SPACE unless result.blank?
	     title = mla_citation_title(title)
	     if is_article
	       title.gsub!(/[#{DQUOTE}\p{Pi}\p{Pf}]/u, SQUOTE)
	       result << %Q("#{title}.")
	     else
	       result << "<em>#{title}</em>."
	     end
	   end
	*/

	if e.data.title != "" {
		res = appendUnlessEndsWith(res, " ", []string{" "})

		title := mlaTitle(e.data.title)

		if e.data.isArticle == true {
			res += `"` + doubleToSingleQuotes(title) + `."`
		} else {
			res += italics(title) + "."
		}
	}

	/*
	   # === Editors, Compilers, or Translators
	   actors = { 'Edited' => eds, 'Compiled' => comps, 'Translated' => trans }
	   actors.each_pair do |action, names|
	     next unless names.present?
	     result << SPACE unless result.blank? || result.end_with?(SPACE)
	     result << action << ' by ' << cmos_names(names) << '.'
	   end
	*/

	if len(editors) > 0 {
		res = appendUnlessEndsWith(res, " ", []string{" "})
		res += "Edited by " + cmsNames(editors) + "."
	}

	if len(compilers) > 0 {
		res = appendUnlessEndsWith(res, " ", []string{" "})
		res += "Compiled by " + cmsNames(compilers) + "."
	}

	if len(translators) > 0 {
		res = appendUnlessEndsWith(res, " ", []string{" "})
		res += "Translated by " + cmsNames(translators) + "."
	}

	/*
	   # === Container Title
	   if journal.present?
	     result << SPACE unless result.blank? || result.end_with?(SPACE)
	     journal = mla_citation_title(journal)
	     result << "<em>#{journal}</em>"
	   end
	*/

	if e.data.journal != "" {
		res = appendUnlessEndsWith(res, " ", []string{" "})
		res += italics(mlaTitle(e.data.journal))
	}

	/*
	   # === Version/Edition
	   if edition.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << clean_end_punctuation(edition)
	   end

	*/

	res = appendWithComma(res, cleanEndPunctuation(e.data.edition))

	/*
	   # === Container Editors
	   if editors.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     editors = cmos_names(editors)
	     result << clean_end_punctuation(editors)
	   end
	*/

	// NOTE: these are journal editors (as opposed to book editors); not yet implemented

	/*
	   # === Accession Number (for archival collections)
	   if an.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << an
	   end
	*/

	// NOTE: archival items should all have a cite_as entry, obviating the need to handle accession number

	/*
	   # === Volume
	   if volume.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << volume
	   end
	*/

	res = appendWithComma(res, e.data.volume)

	/*
	   # === Issue
	   if issue.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << issue
	   end
	*/

	res = appendWithComma(res, e.data.issue)

	/*
	   # === Publisher
	   if publisher.present?
	     unless result.blank?
	       result << '.'   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << publisher
	   end
	*/

	res = appendWithComma(res, e.data.publisher)

	/*
	   # === Date of publication
	   # Should be "YYYY" for a book; "[Day] Mon. YYYY" for an article.
	   if date.present?
	     date_string = export_date(date, month_names: true)
	     year, month, day = (date_string || date).split('/')
	     month = "#{month[0,3]}." if month && (month.size > 3)
	     if year && month && day && is_article
	       date = "#{day} #{month} #{year}"
	     elsif year && month && is_article
	       date = "#{year}, #{month}"
	     elsif year
	       date = year.sub(/^(\d{4}).*$/, '\1')
	     end
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << date
	   end
	*/

	if e.data.date != "" {
		res = appendWithComma(res, mlaDate(e.data.year, e.data.month, e.data.day, e.data.isArticle))
	}

	/*
	   # === Pages
	   if pages.present?
	     unless result.blank?
	       result << ','   if result =~ /\d$/
	       result << ':'   unless result.end_with?(SPACE, '.', ',', ':')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << pages
	   end
	*/

	res = appendWithComma(res, e.data.pages)

	/*
	   # === URL/DOI
	   if link.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << link
	   end
	*/

	res = appendWithComma(res, e.data.link)

	/*
	   # The end of the citation should be a period.
	   result << '.' unless result.end_with?('.')

	   result
	*/

	res = appendUnlessEndsWith(res, ".", []string{"."})

	return res, nil
}

func cmsNames(authors []string) string {
	/*
	   # Format a list of names for Chicago Manual of Style citations.
	   #
	   # @param [Array<String>] names    One or more personal or corporate names.
	   # @param [Boolean]       authors  Treat as author names if *true*.
	   #
	   # @return [String]
	   #
	   # === Usage Notes
	   # For author citations use `cmos_names(src, true)` to emit the first listed
	   # name in bibliographic order (with the surname first).  Otherwise all
	   # names are emitted in reading order (with the surname last).
	   #
	   def cmos_names(names, authors = false)
	     total  = names.size
	     et_al  = (total > 10)
	     names  = et_al ? names.take(7) : names.dup
	     first  = names.shift
	     result = authors ? capitalize(first.dup) : name_reverse(first)
	     if names.present?
	       names.map! { |n| name_reverse(n) }
	       final = et_al ? 'et al' : "and #{names.pop}"
	       result << ', ' << names.join(', ') unless names.blank?
	       result << ' ' << final
	     end
	     clean_end_punctuation(result)
	   end
	*/
	res := ""

	total := len(authors)

	etAl := total > 10

	names := authors
	if etAl == true {
		names = names[:7]
	}

	var first string

	first, names = names[0], names[1:]

	res += capitalize(first)

	if len(names) > 0 {
		var readingNames []string
		for _, name := range names {
			readingNames = append(readingNames, readingOrder(name))
		}

		var last string

		last, readingNames = readingNames[len(readingNames)-1], readingNames[:len(readingNames)-1]

		if len(readingNames) > 0 {
			res += ", " + strings.Join(readingNames, ", ")
		}

		if etAl == true {
			res += " et al"
		} else {
			res += " and " + last
		}
	}

	res = cleanEndPunctuation(res)

	return res
}
