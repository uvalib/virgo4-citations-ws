package main

import (
	"errors"
)

type mlaEncoder struct {
	url         string
	extension   string
	contentType string
}

func newMlaEncoder(cfg serviceConfigFormat) *mlaEncoder {
	e := mlaEncoder{}

	e.extension = cfg.Extension
	e.contentType = cfg.ContentType

	return &e
}

func (e *mlaEncoder) Init(url string) {
	e.url = url
}

func (e *mlaEncoder) Populate(parts citationParts) error {
	return errors.New("mla format not yet implemented")
}

func (e *mlaEncoder) ContentType() string {
	return e.contentType
}

func (e *mlaEncoder) FileName() string {
	return ""
}

func (e *mlaEncoder) FileContents() (string, error) {
	return "fixme", nil
}
