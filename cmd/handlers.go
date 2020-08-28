package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (p *serviceContext) citationHandler(c *gin.Context, fmt citationType) {
	cl := clientContext{}
	cl.init(p, c)

	s := citationsContext{}
	s.init(p, &cl)

	cl.logRequest()
	resp := s.handleCitationRequest(fmt)
	cl.logResponse(resp)

	if resp.err != nil {
		c.String(resp.status, resp.err.Error())
		return
	}

	s.serveCitation(fmt)
}

func (p *serviceContext) apaHandler(c *gin.Context) {
	p.citationHandler(c, newApaEncoder(p.config.Formats.APA))
}

func (p *serviceContext) cmsHandler(c *gin.Context) {
	p.citationHandler(c, newCmsEncoder(p.config.Formats.CMS))
}

func (p *serviceContext) mlaHandler(c *gin.Context) {
	p.citationHandler(c, newMlaEncoder(p.config.Formats.MLA))
}

func (p *serviceContext) risHandler(c *gin.Context) {
	p.citationHandler(c, newRisEncoder(p.config.Formats.RIS))
}

func (p *serviceContext) unapiHandler(c *gin.Context) {
	id := c.Query("id")
	format := c.Query("format")

	// no params: formats for any objects this endpoint will provide (ris only)
	// id param only: formats for this object (ris only)
	// in these cases, response will be the same (modulo an id attribute, and http status)
	if format == "" {
		idAttr := ""
		status := http.StatusOK

		if id != "" {
			idAttr = fmt.Sprintf(` id="%s"`, id)
			status = http.StatusMultipleChoices
		}

		formatsXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?><formats%s><format name="ris" type="%s" /></formats>`, idAttr, p.config.Formats.RIS.ContentType)

		c.Header("Content-Type", "application/xml")
		c.String(status, formatsXML)

		return
	}

	// id and format params: the citation itself
	p.citationHandler(c, newRisEncoder(p.config.Formats.RIS))
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

	data, err := citation.Contents()

	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	fileName := citation.FileName()
	contentType := citation.ContentType()

	if s.client.opts.inline == true || fileName == "" {
		c.Header("Content-Type", contentType)
		c.String(http.StatusOK, data)
		return
	}

	reader := strings.NewReader(data)
	contentLength := int64(len(data))

	extraHeaders := map[string]string{
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, fileName),
	}

	c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
}
