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

type citationsContext struct {
	svc    *serviceContext
	client *clientContext
	url    string
	v4url  string
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

func (s *citationsContext) handleRISRequest() (citationType, serviceResponse) {
	rec, resp := s.queryPoolRecord()

	if resp.err != nil {
		return nil, resp
	}

	ris := newRisEncoder(s.svc.config.Formats.RIS, s.v4url)

	for _, field := range rec.Fields {
		if field.RISCode != "" {
			ris.addTagValue(field.RISCode, field.Value)
		}
	}

	return ris, serviceResponse{status: http.StatusOK}
}
