package main

import (
	"net/http"
	"path"
)

type citationType interface {
	ContentType() string
	FileName() string
	FileContents() (string, error)
}

type citationParts map[string][]string

type citationsContext struct {
	svc    *serviceContext
	client *clientContext
	url    string
	v4url  string
	parts  citationParts
}

type serviceResponse struct {
	status int   // http status code
	err    error // error, if any
}

func (s *citationsContext) init(p *serviceContext, c *clientContext) {
	s.svc = p
	s.client = c

	s.url = c.ginCtx.Query("item")

	if id := path.Base(s.url); id != "" && s.svc.config.URLPrefix != "" {
		s.v4url = s.svc.config.URLPrefix + id
	}
}

func (s *citationsContext) log(format string, args ...interface{}) {
	s.client.log(format, args...)
}

func (s *citationsContext) err(format string, args ...interface{}) {
	s.client.err(format, args...)
}

func (s *citationsContext) collectCitationParts() serviceResponse {
	rec, resp := s.queryPoolRecord()

	if resp.err != nil {
		return resp
	}

	s.parts = make(citationParts)

	for _, field := range rec.Fields {
		if field.CitationPart != "" && field.Value != "" {
			s.parts[field.CitationPart] = append(s.parts[field.CitationPart], field.Value)
		}
	}

	return serviceResponse{status: http.StatusOK}
}

func (s *citationsContext) handleRISRequest() (citationType, serviceResponse) {
	resp := s.collectCitationParts()

	if resp.err != nil {
		return nil, resp
	}

	ris := newRisEncoder(s.svc.config.Formats.RIS, s.v4url)

	ris.populateCitation(s.parts)

	return ris, serviceResponse{status: http.StatusOK}
}
