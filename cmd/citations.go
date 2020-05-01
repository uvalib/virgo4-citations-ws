package main

import (
	"fmt"
	"net/http"
)

type citationsContext struct {
	svc    *serviceContext
	client *clientContext
}

type serviceResponse struct {
	status int         // http status code
	data   interface{} // data to return as JSON
	err    error       // error, if any
}

func (s *citationsContext) init(p *serviceContext, c *clientContext) {
	s.svc = p
	s.client = c
}

func (s *citationsContext) log(format string, args ...interface{}) {
	s.client.log(format, args...)
}

func (s *citationsContext) err(format string, args ...interface{}) {
	s.client.err(format, args...)
}

func (s *citationsContext) handleRISRequest() serviceResponse {
	rec, resp := s.queryPoolRecord()

	if resp.err != nil {
		return resp
	}

	for _, field := range rec.Fields {
		if field.RISCode != "" {
			s.log("%s  %s", field.RISCode, field.Value)
		}
	}

	return serviceResponse{status: http.StatusNotImplemented, err: fmt.Errorf("handleRISRequest() not yet implemented")}
}
