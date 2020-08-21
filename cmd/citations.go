package main

import (
	"net/http"
	"path"
)

type citationType interface {
	Init(string)
	Populate(citationParts) error
	ContentType() string
	FileName() string
	Contents() (string, error)
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

	// v4 uses "item", but unAPI uses "id".  these are equivalent identifiers
	s.url = c.ginCtx.Query("item")
	if s.url == "" {
		s.url = c.ginCtx.Query("id")
	}

	if id := path.Base(s.url); id != "" && s.svc.config.URLPrefix != "" {
		s.v4url = s.svc.config.URLPrefix + id
	}
}

func (s *citationsContext) log(format string, args ...interface{}) {
	s.client.log(format, args...)
}

func (s *citationsContext) warn(format string, args ...interface{}) {
	s.client.warn(format, args...)
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

func (s *citationsContext) handleCitationRequest(fmt citationType) serviceResponse {
	resp := s.collectCitationParts()

	if resp.err != nil {
		return resp
	}

	fmt.Init(s.v4url)

	if err := fmt.Populate(s.parts); err != nil {
		return serviceResponse{status: http.StatusInternalServerError, err: err}
	}

	return serviceResponse{status: http.StatusOK}
}
