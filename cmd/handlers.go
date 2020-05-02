package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (p *serviceContext) risHandler(c *gin.Context) {
	cl := clientContext{}
	cl.init(p, c)

	s := citationsContext{}
	s.init(p, &cl)

	cl.logRequest()
	ris, resp := s.handleRISRequest()
	cl.logResponse(resp)

	if resp.err != nil {
		c.String(resp.status, resp.err.Error())
		return
	}

	s.serveCitation(ris)
}

func (p *serviceContext) ignoreHandler(c *gin.Context) {
}

func (p *serviceContext) versionHandler(c *gin.Context) {
	cl := clientContext{}
	cl.init(p, c)

	c.JSON(http.StatusOK, p.version)
}

func (p *serviceContext) healthCheckHandler(c *gin.Context) {
	cl := clientContext{}
	cl.init(p, c)

	s := citationsContext{}
	s.init(p, &cl)

	if s.client.opts.verbose == false {
		s.client.nolog = true
	}

	// build response

	type hcResp struct {
		Healthy bool   `json:"healthy"`
		Message string `json:"message,omitempty"`
	}

	hcMap := make(map[string]hcResp)

	hcMap["self"] = hcResp{Healthy: true}

	c.JSON(http.StatusOK, hcMap)
}

func (s *citationsContext) serveCitation(citation citationType) {
	c := s.client.ginCtx

	data, err := citation.FileContents()

	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if s.client.opts.inline == true {
		c.String(http.StatusOK, data)
		return
	}

	reader := strings.NewReader(data)
	contentLength := int64(len(data))
	contentType := citation.ContentType()
	fileName := citation.FileName()

	extraHeaders := map[string]string{
		"Content-Disposition:": fmt.Sprintf(`attachment; filename="%s"`, fileName),
	}

	c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
}
