package main

import (
	"errors"
)

type cmsEncoder struct {
	url         string
	extension   string
	contentType string
}

func newCmsEncoder(cfg serviceConfigFormat) *cmsEncoder {
	e := cmsEncoder{}

	e.extension = cfg.Extension
	e.contentType = cfg.ContentType

	return &e
}

func (e *cmsEncoder) Init(url string) {
	e.url = url
}

func (e *cmsEncoder) Populate(parts citationParts) error {
	return errors.New("cms format not yet implemented")
}

func (e *cmsEncoder) ContentType() string {
	return e.contentType
}

func (e *cmsEncoder) FileName() string {
	return ""
}

func (e *cmsEncoder) FileContents() (string, error) {
	return "fixme", nil
}
