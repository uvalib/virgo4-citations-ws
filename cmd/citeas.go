package main

import (
	"errors"
	"strings"
)

type citeAsEncoder struct {
	cfg  serviceConfigFormat
	url  string
	data *genericCitation
}

func newCiteAsEncoder(cfg serviceConfigFormat) *citeAsEncoder {
	e := citeAsEncoder{}

	e.cfg = cfg

	return &e
}

func (e *citeAsEncoder) Init(url string) {
	e.url = url
}

func (e *citeAsEncoder) Populate(parts citationParts) error {
	var err error

	// these options will effectively not be used
	opts := genericCitationOpts{}

	if e.data, err = newGenericCitation(e.url, parts, opts); err != nil {
		return err
	}

	return nil
}

func (e *citeAsEncoder) Label() string {
	return e.cfg.Label
}

func (e *citeAsEncoder) ContentType() string {
	return e.cfg.ContentType
}

func (e *citeAsEncoder) FileName() string {
	return ""
}

func (e *citeAsEncoder) Contents() (string, error) {
	if len(e.data.citeAs) > 0 {
		return strings.Join(e.data.citeAs, "\n"), nil
	}

	return "", errors.New("no explicit citation for this item")
}
