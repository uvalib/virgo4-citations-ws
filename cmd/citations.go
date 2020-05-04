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
	svc     *serviceContext
	client  *clientContext
	itemURL string
	itemID  string
}

type serviceResponse struct {
	status int   // http status code
	err    error // error, if any
}

func (s *citationsContext) init(p *serviceContext, c *clientContext) {
	s.svc = p
	s.client = c

	s.itemURL = c.ginCtx.Query("item")
	s.itemID = path.Base(s.itemURL)
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

	ris := newRisEncoder(s.svc.config.Formats.RIS, s.itemID)

	for _, field := range rec.Fields {
		if field.RISCode != "" {
			ris.addTagValue(field.RISCode, field.Value)
		}
	}

	return ris, serviceResponse{status: http.StatusOK}
}
