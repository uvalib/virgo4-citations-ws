package main

import (
	"fmt"
	"net/http"
)

type citationsContext struct {
	svc    *serviceContext
	client *clientContext
	url    string
}

type searchResponse struct {
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

func (s *citationsContext) handleRISRequest() searchResponse {
	if s.url != "" {
		s.log("url = [%s]", s.url)
	} else {
		s.err("empty url")
	}

	return searchResponse{status: http.StatusNotImplemented, err: fmt.Errorf("not yet implemented")}
}
