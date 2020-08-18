package main

import (
	"errors"
)

type apaEncoder struct {
	url         string
	extension   string
	contentType string
}

func newApaEncoder(cfg serviceConfigFormat) *apaEncoder {
	e := apaEncoder{}

	e.extension = cfg.Extension
	e.contentType = cfg.ContentType

	return &e
}

func (e *apaEncoder) Init(url string) {
	e.url = url
}

func (e *apaEncoder) Populate(parts citationParts) error {
	return errors.New("apa format not yet implemented")
}

func (e *apaEncoder) ContentType() string {
	return e.contentType
}

func (e *apaEncoder) FileName() string {
	return ""
}

func (e *apaEncoder) FileContents() (string, error) {
	return "fixme", nil
}
