package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (p *serviceContext) citationHandler(c *gin.Context, citations []citationType) {
	cl := clientContext{}
	cl.init(p, c)

	s := citationsContext{}
	s.init(p, &cl)

	cl.logRequest()
	resp := s.handleCitationRequest(citations)
	cl.logResponse(resp)

	if resp.err != nil {
		c.String(resp.status, resp.err.Error())
		return
	}

	// single citation uses configured values
	if len(citations) == 1 {
		s.serveSingleCitation(citations[0])
		return
	}

	s.serveMultipleCitations(citations)
}

func (p *serviceContext) allHandler(c *gin.Context) {
	p.citationHandler(c, []citationType{
		newCiteAsEncoder(p.config.Formats.CiteAs),
		newMlaEncoder(p.config.Formats.MLA, false),
		newApaEncoder(p.config.Formats.APA, false),
		newCmsEncoder(p.config.Formats.CMS, false),
	})
}

func (p *serviceContext) apaHandler(c *gin.Context) {
	p.citationHandler(c, []citationType{newApaEncoder(p.config.Formats.APA, true)})
}

func (p *serviceContext) citeAsHandler(c *gin.Context) {
	p.citationHandler(c, []citationType{newCiteAsEncoder(p.config.Formats.CiteAs)})
}

func (p *serviceContext) cmsHandler(c *gin.Context) {
	p.citationHandler(c, []citationType{newCmsEncoder(p.config.Formats.CMS, true)})
}

func (p *serviceContext) mlaHandler(c *gin.Context) {
	p.citationHandler(c, []citationType{newMlaEncoder(p.config.Formats.MLA, true)})
}

func (p *serviceContext) risHandler(c *gin.Context) {
	p.citationHandler(c, []citationType{newRisEncoder(p.config.Formats.RIS)})
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
	p.citationHandler(c, []citationType{newRisEncoder(p.config.Formats.RIS)})
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

func (s *citationsContext) serveSingleCitation(citation citationType) {
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

func (s *citationsContext) serveMultipleCitations(citations []citationType) {
	c := s.client.ginCtx

	// build json of multi-formats

	type citationResp struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}

	var resp []citationResp

	for _, citation := range citations {
		data, err := citation.Contents()

		if err != nil {
			s.log("WARNING: failed to generate %s citation: %s", citation.Label(), err.Error())
			continue
		}

		resp = append(resp, citationResp{Label: citation.Label(), Value: data})
	}

	c.JSON(http.StatusOK, resp)
}
