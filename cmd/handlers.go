package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *serviceContext) risHandler(c *gin.Context) {
	cl := clientContext{}
	cl.init(p, c)

	s := citationsContext{}
	s.init(p, &cl)

	cl.logRequest()
	resp := s.handleRISRequest()
	cl.logResponse(resp)

	if resp.err != nil {
		c.String(resp.status, resp.err.Error())
		return
	}

	c.JSON(resp.status, resp.data)
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
