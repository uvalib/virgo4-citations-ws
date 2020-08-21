package main

import (
	"errors"
	"strings"
)

type cmsEncoder struct {
	url         string
	extension   string
	contentType string
	data        *genericCitation
}

func newCmsEncoder(cfg serviceConfigFormat) *cmsEncoder {
	e := cmsEncoder{}

	e.extension = cfg.Extension
	e.contentType = cfg.ContentType

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

	if e.data, err = newGenericCitation(parts, opts); err != nil {
		return err
	}

	return nil
}

func (e *cmsEncoder) ContentType() string {
	return e.contentType
}

func (e *cmsEncoder) FileName() string {
	return ""
}

func (e *cmsEncoder) Contents() (string, error) {
	if len(e.data.citeAs) > 0 {
		return strings.Join(e.data.citeAs, "\n"), nil
	}

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

	     # === Editors, Compilers, or Translators
	     actors = { 'Edited' => eds, 'Compiled' => comps, 'Translated' => trans }
	     actors.each_pair do |action, names|
	       next unless names.present?
	       result << SPACE unless result.blank? || result.end_with?(SPACE)
	       result << action << ' by ' << cmos_names(names) << '.'
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
	       editors = cmos_names(editors)
	       result << clean_end_punctuation(editors)
	     end

	     # === Accession Number (for archival collections)
	     if an.present?
	       unless result.blank?
	         result << ','   unless result.end_with?(SPACE, '.', ',')
	         result << SPACE unless result.end_with?(SPACE)
	       end
	       result << an
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

	     # === Publisher
	     if publisher.present?
	       unless result.blank?
	         result << '.'   unless result.end_with?(SPACE, '.', ',')
	         result << SPACE unless result.end_with?(SPACE)
	       end
	       result << publisher
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
	         result << ','   if result =~ /\d$/
	         result << ':'   unless result.end_with?(SPACE, '.', ',', ':')
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

	return "", errors.New("non-explicit CMS citations not yet implemented")
}
