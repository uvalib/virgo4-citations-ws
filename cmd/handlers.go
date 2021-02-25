package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (p *serviceContext) citationHandler(c *clientContext, json bool, citations []citationType) {
	s := citationsContext{}
	s.init(p, c)

	c.logRequest()
	resp := s.handleCitationRequest(citations)
	c.logResponse(resp)

	if resp.err != nil {
		c.ginCtx.String(resp.status, resp.err.Error())
		return
	}

	// single non-json or inline citation uses configured values
	if (json == false || s.client.opts.inline == true) && len(citations) == 1 {
		s.serveSingleCitation(citations[0])
		return
	}

	s.serveMultipleCitations(citations)
}

func (p *serviceContext) allHandler(c *gin.Context) {
	cl := clientContext{}
	cl.init(p, c)

	p.citationHandler(&cl, true, []citationType{
		newMlaEncoder(p.config.Formats.MLA, true),
		newApaEncoder(p.config.Formats.APA, true),
		newCmsEncoder(p.config.Formats.CMS, true),
		newLbbEncoder(p.config.Formats.LBB, true),
	})
}

func (p *serviceContext) apaHandler(c *gin.Context) {
	cl := clientContext{}
	cl.init(p, c)

	p.citationHandler(&cl, true, []citationType{newApaEncoder(p.config.Formats.APA, true)})
}

func (p *serviceContext) citeAsHandler(c *gin.Context) {
	cl := clientContext{}
	cl.init(p, c)

	p.citationHandler(&cl, true, []citationType{newCiteAsEncoder(p.config.Formats.CiteAs)})
}

func (p *serviceContext) cmsHandler(c *gin.Context) {
	cl := clientContext{}
	cl.init(p, c)

	p.citationHandler(&cl, true, []citationType{newCmsEncoder(p.config.Formats.CMS, true)})
}

func (p *serviceContext) lbbHandler(c *gin.Context) {
	cl := clientContext{}
	cl.init(p, c)

	p.citationHandler(&cl, true, []citationType{newLbbEncoder(p.config.Formats.LBB, true)})
}

func (p *serviceContext) mlaHandler(c *gin.Context) {
	cl := clientContext{}
	cl.init(p, c)

	p.citationHandler(&cl, true, []citationType{newMlaEncoder(p.config.Formats.MLA, true)})
}

func (p *serviceContext) risHandler(c *gin.Context) {
	cl := clientContext{}
	cl.init(p, c)

	p.citationHandler(&cl, false, []citationType{newRisEncoder(p.config.Formats.RIS)})
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
	cl := clientContext{}
	cl.init(p, c)

	p.citationHandler(&cl, false, []citationType{newRisEncoder(p.config.Formats.RIS)})
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

func (s *citationsContext) getContents(citation citationType) (string, error) {
	data, err := citation.Contents()

	if err != nil {
		return "", err
	}

	// extra formatting/sanitizing could go here

	return data, nil
}

func (s *citationsContext) serveSingleCitation(citation citationType) {
	c := s.client.ginCtx

	data, err := s.getContents(citation)

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

	resp := []citationResp{}

	for _, citation := range citations {
		data, err := s.getContents(citation)

		if err != nil {
			s.log("WARNING: failed to generate %s citation: %s", citation.Label(), err.Error())
			continue
		}

		resp = append(resp, citationResp{Label: citation.Label(), Value: data})
	}

	c.JSON(http.StatusOK, resp)
}
