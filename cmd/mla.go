package main

import (
	"fmt"
	"strings"
)

type mlaEncoder struct {
	cfg          serviceConfigFormat
	url          string
	preferCiteAs bool
	data         *genericCitation
}

func newMlaEncoder(cfg serviceConfigFormat, preferCiteAs bool) *mlaEncoder {
	e := mlaEncoder{}

	e.cfg = cfg
	e.preferCiteAs = preferCiteAs

	return &e
}

func (e *mlaEncoder) Init(url string) {
	e.url = url
}

func (e *mlaEncoder) Populate(parts citationParts) error {
	/*
	   opt[:strip_protocol] = true

	   volume    = setup_volume(true)
	   issue     = setup_issue(true)
	   pages     = setup_pages(true)
	   publisher = clean_end_punctuation(export_publisher)
	*/

	var err error

	opts := genericCitationOpts{
		stripProtocol:  true,
		volumePrefix:   true,
		issuePrefix:    true,
		pagesPrefix:    true,
		publisherPlace: false,
	}

	if e.data, err = newGenericCitation(e.url, parts, opts); err != nil {
		return err
	}

	return nil
}

func (e *mlaEncoder) Label() string {
	return e.cfg.Label
}

func (e *mlaEncoder) ContentType() string {
	return e.cfg.ContentType
}

func (e *mlaEncoder) FileName() string {
	return ""
}

func (e *mlaEncoder) Contents() (string, error) {
	if e.preferCiteAs == true && len(e.data.citeAs) > 0 {
		return strings.Join(e.data.citeAs, "\n"), nil
	}

	res := ""

	/*
	   # === Author(s)
	   # First author in "Last, First" form; second author in "First Last" form;
	   # three or more authors shown as the first author followed by ", et al.".
	   # If the author is the same as the publisher, skip the author.
	   authors.delete_if { |v| v == publisher }
	   if authors.present?
	     total = authors.size
	     list  = capitalize(authors.first.dup)
	     if total > 2
	       list << ', et al'
	     elsif total > 1
	       list << ', and ' << name_reverse(authors.last)
	     end
	     result << clean_end_punctuation(list)
	     # Indicate if the "authors" are actually editors of the work.
	     eds =
	       get_related_names(false).map { |name_and_role|
	         name_and_role.to_s.sub!(/\W+Editor\W*$/i, '')
	       }.reject(&:blank?)
	     if (authors - eds).empty?
	       result << '.' if total > 2
	       result << ', editor'
	       result << 's' if total > 1
	     end
	     result << '.'
	   end
	*/

	remove := []string{e.data.publisher}

	authors := removeEntries(e.data.authors, remove)
	editors := removeEntries(e.data.editors, remove)
	advisors := removeEntries(e.data.advisors, remove)

	var creators []string
	creators = append(creators, authors...)
	creators = append(creators, editors...)
	creators = append(creators, advisors...)

	numCreators := len(creators)
	if numCreators > 0 {
		list := capitalize(creators[0])
		switch {
		case numCreators > 2:
			list += ", et al"

		case numCreators == 2:
			list += ", and " + readingOrder(creators[1])
		}

		res += cleanEndPunctuation(list)

		nonEditors := removeEntries(creators, editors)

		if len(nonEditors) == 0 {
			if numCreators > 2 {
				res += "."
			}

			res += ", editor"

			if numCreators > 1 {
				res += "s"
			}
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
	       result << "\"#{title}.\""
	     else
	       result << "<em>#{title}</em>."
	     end
	   end
	*/

	if e.data.title != "" {
		if res != "" {
			res += " "
		}

		title := mlaTitle(e.data.title)

		if e.data.isArticle == true {
			res += `"` + doubleToSingleQuotes(title) + `."`
		} else {
			res += italics(title) + "."
		}
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
	     total = editors.size
	     list  = editors.first.dup
	     if total > 2
	       list << ', et al.'
	     elsif total > 1
	       list << ', and ' << name_reverse(editors.last)
	     end
	     result << clean_end_punctuation(list)
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
	   # === Publisher
	   if publisher.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << publisher
	   end
	*/

	res = appendWithComma(res, e.data.publisher)

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

		    # Format a date string for use with export formats.
		    #
		    # @param [String] value
		    # @param [Hash]   opt
		    #
		    # @option opt [String]  :separator
		    # @option opt [Boolean] :allow_extra_text
		    # @option opt [Boolean] :default_20th_century
		    # @option opt [Boolean] :month_names
		    #
		    # @return [String]
		    # @return [nil]                 If the string did not have a date value.
		    #
		    def export_date(value, opt = {})
		      yy = mm = dd = xx = nil
		      case value
		        when DATE_MM_DD_YY then mm, dd, yy, xx = $LAST_MATCH_INFO[1,4]
		        when DATE_YY_MM_DD then yy, mm, dd, xx = $LAST_MATCH_INFO[1,4]
		        when DATE_MM_YY    then mm, yy, xx = $LAST_MATCH_INFO[1,3]
		        when DATE_YY_MM    then yy, mm, xx = $LAST_MATCH_INFO[1,3]
		        when DATE_YY       then yy, xx = $LAST_MATCH_INFO[1,2]
		      end
		      return if yy.blank?
		      opt = {
		        separator:            '/',
		        allow_extra_text:     false,
		        default_20th_century: true,
		        month_names:          false,
		      }.merge(opt)
		      if mm.blank?
		        mm = nil
		      elsif opt[:month_names]
		        mm = Date.const_get(:MONTHNAMES)[mm.to_i]
		      end
		      dd = nil if dd.blank?
		      xx = xx.delete(opt[:separator]) if xx.present?
		      xx = nil if xx.blank? || !opt[:allow_extra_text]

		      # Adjust year, with the heuristic that two-digit years are actually
		      # years from the 20th century.
		      result =
		        case yy.length
		          when 3 then "0#{yy}"
		          when 2 then opt[:default_20th_century] ? "19#{yy}" : "00#{yy}"
		          when 1 then "000#{yy}"
		          else        yy
		        end

		      # Zero-fill the month; if there is extra text a slash is needed even if
		      # the value is missing.
		      result << opt[:separator] if mm || xx
		      result << ((mm.length == 1) ? "0#{mm}" : mm) if mm

		      # Zero-fill the day; if there is extra text a slash is needed even if the
		      # value is missing.
		      result << opt[:separator] if dd || xx
		      result << ((dd.length == 1) ? "0#{dd}" : dd) if dd

		      # Append the extra text if present.
		      result << "#{opt[:separator]}#{xx}" if xx
		      result
		    end
	*/

	if e.data.date != "" {
		res = appendWithComma(res, mlaDate(e.data.year, e.data.month, e.data.day, e.data.isArticle))
	}

	/*
	   # === Pages
	   if pages.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
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
	*/

	res = appendUnlessEndsWith(res, ".", []string{"."})

	return res, nil
}

func mlaTitle(s string) string {
	/*
	   # Format a title for use in an MLA citation.
	   #
	   # All words other than connector words are capitalized.  If the title ends
	   # with a period, the period is removed so that the caller as control over
	   # where the period is added back in.  Other terminal punctuation
	   # (including "...") is left untouched.
	   #
	   # @param [String] title_text
	   #
	   # @return [String]
	   #
	   # Replaces:
	   # @see Blacklight::Solr::Document::MarcExport#mla_citation_title
	   #
	   # TO-DO: Implement using UVA::Utils::StringMethods#titleize
	   #
	   def mla_citation_title(title_text, *)
	     no_upcase = %w(a an and but by for it of the to with)
	     words = title_text.to_s.strip.split(SPACE)
	     words.map { |w|
	       no_upcase.include?(w) ? w : capitalize(w)
	     }.join(SPACE).sub(/(?<!\.\.)\.$/, '')
	   end
	*/

	noCapitalize := []string{"a", "an", "and", "but", "by", "for", "it", "of", "the", "to", "with"}

	oldWords := wordsBySeparator(s, " ")
	var newWords []string

	for _, word := range oldWords {
		if sliceContainsString(noCapitalize, strings.ToLower(word)) == true {
			newWords = append(newWords, word)
		} else {
			newWords = append(newWords, capitalize(word))
		}
	}

	title := strings.Join(newWords, " ")

	switch {
	case strings.HasSuffix(title, "..."):
		return title
	case strings.HasSuffix(title, "."):
		return title[:len(title)-1]
	}

	return title
}

func mlaDate(y, m, d int, isArticle bool) string {
	res := ""

	month := monthName(m)
	if len(month) > 3 {
		month = month[:3] + "."
	}

	switch {
	case isArticle == true && y != 0 && month != "" && d != 0:
		res = fmt.Sprintf("%d %s %d", d, month, y)

	case isArticle == true && y != 0 && month != "":
		res = fmt.Sprintf("%d, %s", y, month)

	case y != 0:
		res = fmt.Sprintf("%d", y)
	}

	return res
}
