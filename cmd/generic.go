package main

import (
	"errors"
	"log"
)

// data common among CMS/APA/MLA citations
type genericCitation struct {
	isArticle       bool
	citeAs          []string
	authors         []string
	title           string
	editors         []string
	journal         string
	volume          string
	issue           string
	pages           string
	edition         string
	publisher       string
	date            string
	link            string
	accessionNumber string
}

// options to control the slight differences in data population
type genericCitationOpts struct {
	stripProtocol  bool
	volumePrefix   bool
	issuePrefix    bool
	pagesPrefix    bool
	publisherPlace bool
}

func firstElementOf(s []string) string {
	// return first element of slice, or blank string if empty
	val := ""

	if len(s) > 0 {
		val = s[0]
	}

	return val
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
	   editors   = nil # TODO: Journal editors as opposed to book editors
	   journal   = export_journal
	   volume    = setup_volume(xxx)
	   issue     = setup_issue(xxx)
	   pages     = setup_pages(xxx)
	   edition   = setup_edition
	   publisher = setup_pub_info
	   date      = setup_pub_date
	   link      = setup_link(url, xxx)
	   an        = accession_number

	   # Make adjustments if necessary
	   an = nil if is_article
	   date = nil if an
	*/

	c := genericCitation{}

	// check for explicit citation
	c.citeAs = parts["explicit"]
	if len(c.citeAs) > 0 {
		return &c, nil
	}

	// set options
	c.isArticle = firstElementOf(parts["format"]) == "article"

	// set values
	// TODO: implement
	for part, values := range parts {
		log.Printf("part [%s] => %v", part, values)
	}

	c.authors = []string{"author 1", "author 2"}
	c.title = "title"
	c.editors = []string{}
	c.journal = "journal"
	c.volume = "volume"
	c.issue = "issue"
	c.pages = "pages"
	c.edition = "edition"
	c.publisher = "publisher"
	c.date = "date"
	c.link = "link"
	c.accessionNumber = "accession_number"

	// adjust values
	if c.isArticle == true {
		c.accessionNumber = ""
	}

	if c.accessionNumber != "" {
		c.date = ""
	}

	log.Printf("generic citation: %#v", c)

	//	return &c, nil
	return nil, errors.New("full generic citation parsing not yet implemented")
}
