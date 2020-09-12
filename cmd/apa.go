package main

import (
	"fmt"
	"strings"
)

type apaEncoder struct {
	cfg          serviceConfigFormat
	url          string
	preferCiteAs bool
	data         *genericCitation
}

func newApaEncoder(cfg serviceConfigFormat, preferCiteAs bool) *apaEncoder {
	e := apaEncoder{}

	e.cfg = cfg
	e.preferCiteAs = preferCiteAs

	return &e
}

func (e *apaEncoder) Init(url string) {
	e.url = url
}

func (e *apaEncoder) Populate(parts citationParts) error {
	/*
		add_pp    = (doc_sub_type == :newspaper_article)

		volume    = setup_volume
		issue     = setup_issue
		pages     = setup_pages(add_pp)
		publisher = setup_pub_info
	*/

	addPP := firstElementOf(parts["format"]) == "news"

	var err error

	opts := genericCitationOpts{
		stripProtocol:  false,
		volumePrefix:   false,
		issuePrefix:    false,
		pagesPrefix:    addPP,
		publisherPlace: true,
	}

	if e.data, err = newGenericCitation(e.url, parts, opts); err != nil {
		return err
	}

	return nil
}

func (e *apaEncoder) Label() string {
	return e.cfg.Label
}

func (e *apaEncoder) ContentType() string {
	return e.cfg.ContentType
}

func (e *apaEncoder) FileName() string {
	return ""
}

func (e *apaEncoder) Contents() (string, error) {
	if e.preferCiteAs == true && len(e.data.citeAs) > 0 {
		return strings.Join(e.data.citeAs, "\n"), nil
	}

	res := ""

	/*
	   # === Author(s)
	   # No more than seven names are listed in "Last, F. M." form.  If there
	   # are between 2 and 7 total authors they are listed separated by commas
	   # with '& ' before the last one.  If there are more than 7 authors, only
	   # the first 6 are listed separated by commas, then an ellipsis (...)
	   # followed by the final author.
	   if authors.present?
	     list = authors.map { |name| abbreviate_name(name) }
	     total_authors = list.size
	     final_author  = list.pop
	     result <<
	       case total_authors
	         when 1    then final_author
	         when 2..7 then list.join(', ') + ', &amp; ' + final_author
	         else           list[0,6].join(', ') + ', ... ' + final_author
	       end
	     # Indicate if the "authors" are actually editors of the work.
	     eds =
	       get_related_names(false).map { |name_and_role|
	         name_and_role.to_s.sub!(/\W+Editor\W*$/i, '')
	       }.reject(&:blank?)
	     if (authors - eds).empty?
	       s = ('s' if total_authors > 1)
	       result << " (Ed#{s}.)."
	     end
	   end
	*/

	if len(e.data.authors) > 0 {
		var authors []string
		for _, author := range e.data.authors {
			// FIXME
			authors = append(authors, abbreviateName(author))
		}

		var last string

		total := len(authors)

		last, authors = authors[len(authors)-1], authors[:len(authors)-1]

		switch {
		case total == 1:
			res += last

		case (total >= 2) && (total <= 7):
			res += strings.Join(authors, ", ") + ", &amp; " + last

		default:
			res += strings.Join(authors[0:6], ", ") + ", ... " + last
		}

		nonEditors := removeEntries(e.data.authors, e.data.editors)

		if len(nonEditors) == 0 {
			s := ""
			if total > 1 {
				s = "s"
			}

			res += "(Ed" + s + ".)."
		}
	}

	/*
	   # === Date of publication
	   # Should be "(YYYY)" for a book; "(YYYY, Month [Day])" for an article.
	   if date.blank?
	     result << '.'   unless result.blank? || result.end_with?('.')
	   else
	     result << SPACE unless result.blank? || result.end_with?(SPACE)
	     date_string = export_date(date, month_names: true)
	     year, month, day = (date_string || date).split('/')
	     if year && month && day && is_article
	       date = "#{year}, #{month} #{day}"
	     elsif year && month && is_article
	       date = "#{year}, #{month}"
	     elsif year
	       date = year.sub(/^(\d{4}).*$/, '\1')
	     end
	     result << "(#{date})."
	   end
	*/

	if e.data.date != "" {
		if res != "" && strings.HasSuffix(res, " ") == false {
			res += " "
		}

		date := ""

		month := monthName(e.data.month)

		switch {
		case e.data.isArticle && e.data.year != 0 && month != "" && e.data.day != 0:
			date = fmt.Sprintf("%d, %s %d", e.data.year, month, e.data.day)

		case e.data.isArticle && e.data.year != 0 && month != "":
			date = fmt.Sprintf("%d, %s", e.data.year, month)

		case e.data.year != 0:
			date = fmt.Sprintf("%d", e.data.year)
		}

		res += "(" + date + ")."
	} else {
		if res != "" && strings.HasSuffix(res, ".") == false {
			res += "."
		}
	}

	/*
	   # === Item Title
	   # The title is in sentence-case (only the first word and proper nouns
	   # are capitalized); if there is a sub-title, it also has the first word
	   # capitalized followed by lower-case words.
	   if title.present?
	     result << SPACE unless result.blank? || result.end_with?(SPACE)
	     title = clean_end_punctuation(title)
	     result << (is_article ? title : "<em>#{title}</em>")
	   end
	*/

	if e.data.title != "" {
		if res != "" && strings.HasSuffix(res, " ") == false {
			res += " "
		}

		title := cleanEndPunctuation(e.data.title)

		if e.data.isArticle == true {
			res += title
		} else {
			res += "<em>" + title + "</em>"
		}
	}

	/*
	   # === Container Editors
	   if editors.present?
	     unless result.blank?
	       result << '.'   unless result.end_with?('.')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     list = editors.map { |name| abbreviate_name(name) }
	     total_editors = list.size
	     final_editor  = list.pop
	     result <<
	       case total_editors
	         when 1    then final_editor
	         when 2..7 then list.join(', ') + ', &amp; ' + final_editor
	         else           list[0,6].join(', ') + ', ... ' + final_editor
	       end
	   end
	*/

	// NOTE: these are journal editors (as opposed to book editors); not yet implemented

	/*
	   # === Container Title
	   # Journal titles are capitalized like MLA titles.
	   if journal.present?
	     unless result.blank?
	       result << '.'   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     journal = mla_citation_title(journal)
	     result << "<em>#{journal}</em>"
	   end
	*/

	if e.data.journal != "" {
		if res != "" && strings.HasSuffix(res, " ") == false && strings.HasSuffix(res, ".") == false && strings.HasSuffix(res, ",") == false {
			res += "."
		}
		if res != "" && strings.HasSuffix(res, " ") == false {
			res += " "
		}

		res += "<em>" + mlaTitle(e.data.journal) + "</em>"
	}

	/*
	   # === Version/Edition
	   if edition.present?
	     result << SPACE unless result.blank? || result.end_with?(SPACE)
	     edition = clean_end_punctuation(edition)
	     result << "(#{edition})."
	   elsif journal.blank?
	     result << '.' unless result.end_with?('.')
	   end
	*/

	if e.data.edition != "" {
		if res != "" && strings.HasSuffix(res, " ") == false {
			res += " "
		}

		res += "(" + cleanEndPunctuation(e.data.edition) + ")."
	} else if e.data.journal == "" {
		if res != "" && strings.HasSuffix(res, ".") == false {
			res += "."
		}
	}

	/*
	   # === Volume
	   if volume.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << clean_end_punctuation(volume)
	   end
	*/

	if e.data.volume != "" {
		if res != "" && strings.HasSuffix(res, " ") == false && strings.HasSuffix(res, ".") == false && strings.HasSuffix(res, ",") == false {
			res += ","
		}
		if res != "" && strings.HasSuffix(res, " ") == false {
			res += " "
		}

		res += cleanEndPunctuation(e.data.volume)
	}

	/*
	   # === Issue
	   if issue.present?
	     if volume.blank?
	       result << SPACE unless result.blank? || result.end_with?(SPACE)
	     end
	     issue = clean_end_punctuation(issue)
	     result << "(#{issue})"
	   end
	*/

	if e.data.issue != "" {
		if e.data.volume == "" {
			if res != "" && strings.HasSuffix(res, " ") == false {
				res += " "
			}
		}

		res += "(" + cleanEndPunctuation(e.data.issue) + ")"
	}

	/*
	   # === Pages
	   # For articles, pages do not include "p." or "pp." *except* for articles
	   # in a newspaper.
	   if pages.present?
	     unless result.blank?
	       result << ','   unless result.end_with?(SPACE, '.', ',')
	       result << SPACE unless result.end_with?(SPACE)
	     end
	     result << pages
	   end
	*/

	if e.data.pages != "" {
		if res != "" && strings.HasSuffix(res, " ") == false && strings.HasSuffix(res, ".") == false && strings.HasSuffix(res, ",") == false {
			res += ","
		}
		if res != "" && strings.HasSuffix(res, " ") == false {
			res += " "
		}

		res += e.data.pages
	}

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
	     result << '.'
	   end
	*/

	if e.data.publisher != "" {
		if res != "" && strings.HasSuffix(res, " ") == false && strings.HasSuffix(res, ".") == false && strings.HasSuffix(res, ",") == false {
			res += ","
		}
		if res != "" && strings.HasSuffix(res, " ") == false {
			res += " "
		}

		res += e.data.publisher + "."
	}

	/*
	   # The end of the citation proper should be a period.
	   result << '.' unless result.blank? || result.end_with?('.')
	*/

	if res != "" && strings.HasSuffix(res, ".") == false {
		res += "."
	}

	/*
	   # === URL/DOI
	   if link.present?
	     result << SPACE unless result.blank? || result.end_with?(SPACE)
	     result << 'Retrieved from '
	     result << link
	   end

	   result
	*/

	if e.data.link != "" {
		if res != "" && strings.HasSuffix(res, " ") == false {
			res += " "
		}

		res += "Retrieved from " + e.data.link
	}

	return res, nil
}
