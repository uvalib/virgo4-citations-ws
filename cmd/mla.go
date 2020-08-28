package main

import (
	"errors"
	"strings"
)

type mlaEncoder struct {
	url         string
	extension   string
	contentType string
	data        *genericCitation
}

func newMlaEncoder(cfg serviceConfigFormat) *mlaEncoder {
	e := mlaEncoder{}

	e.extension = cfg.Extension
	e.contentType = cfg.ContentType

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

	if e.data, err = newGenericCitation(parts, opts); err != nil {
		return err
	}

	return nil
}

func (e *mlaEncoder) ContentType() string {
	return e.contentType
}

func (e *mlaEncoder) FileName() string {
	return ""
}

func (e *mlaEncoder) Contents() (string, error) {
	if len(e.data.citeAs) > 0 {
		return strings.Join(e.data.citeAs, "\n"), nil
	}

	res := ""

	var authors []string

	for _, author := range e.data.authors {
		if author == e.data.publisher {
			continue
		}

		authors = append(authors, author)
	}

	numAuthors := len(authors)
	if numAuthors > 0 {
		list := capitalize(authors[0])
		switch {
		case numAuthors > 2:
			list += ", et al"

		case numAuthors == 2:
			list += ", and " + nameReverse(authors[1])
		}

		res += cleanEndPunctuation(list)
	}

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

	   # === Container Title
	   if journal.present?
	     result << SPACE unless result.blank? || result.end_with?(SPACE)
	     journal = mla_citation_title(journal)
	     result << "<em>#{journal}</em>"
	   end

	   # === Version/Edition
	   if edition.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << clean_end_punctuation(edition)
	   end

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

	   # === Accession Number (for archival collections)
	   if an.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << an
	   end

	   # === Publisher
	   if publisher.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << publisher
	   end

	   # === Volume
	   if volume.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << volume
	   end

	   # === Issue
	   if issue.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << issue
	   end

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

	   # === Pages
	   if pages.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << pages
	   end

	   # === URL/DOI
	   if link.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << link
	   end

	   # The end of the citation should be a period.
	   result << '.' unless result.end_with?('.')

	   result
	*/

	return "", errors.New("non-explicit MLA citations not yet implemented")
	//return res, nil
}
