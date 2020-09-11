package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"sort"
	"strings"
)

const envPrefix = "VIRGO4_CITATIONS_WS"

type serviceConfigPools struct {
	ConnTimeout string `json:"conn_timeout,omitempty"`
	ReadTimeout string `json:"read_timeout,omitempty"`
}

type serviceConfigJWT struct {
	Key        string `json:"key,omitempty"`
	Expiration int    `json:"expiration,omitempty"`
}

type serviceConfigFormat struct {
	Label       string `json:"label,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	Extension   string `json:"extension,omitempty"`
}

type serviceConfigFormats struct {
	APA    serviceConfigFormat `json:"apa,omitempty"`
	CiteAs serviceConfigFormat `json:"cite_as,omitempty"`
	CMS    serviceConfigFormat `json:"cms,omitempty"`
	MLA    serviceConfigFormat `json:"mla,omitempty"`
	RIS    serviceConfigFormat `json:"ris,omitempty"`
}

type serviceConfig struct {
	Port      string               `json:"port,omitempty"`
	URLPrefix string               `json:"url_prefix,omitempty"`
	JWT       serviceConfigJWT     `json:"jwt,omitempty"`
	Pools     serviceConfigPools   `json:"pools,omitempty"`
	Formats   serviceConfigFormats `json:"formats,omitempty"`
}

func getSortedJSONEnvVars() []string {
	var keys []string

	for _, keyval := range os.Environ() {
		key := strings.Split(keyval, "=")[0]
		if strings.HasPrefix(key, envPrefix+"_JSON_") {
			keys = append(keys, key)
		}
	}

	sort.Strings(keys)

	return keys
}

func loadConfig() *serviceConfig {
	cfg := serviceConfig{}

	// json configs

	envs := getSortedJSONEnvVars()

	valid := true

	for _, env := range envs {
		log.Printf("[CONFIG] loading %s ...", env)
		if val := os.Getenv(env); val != "" {
			dec := json.NewDecoder(bytes.NewReader([]byte(val)))
			dec.DisallowUnknownFields()

			if err := dec.Decode(&cfg); err != nil {
				log.Printf("error decoding %s: %s", env, err.Error())
				valid = false
			}
		}
	}

	if valid == false {
		log.Printf("exiting due to json decode error(s) above")
		os.Exit(1)
	}

	//bytes, err := json.MarshalIndent(cfg, "", "  ")
	bytes, err := json.Marshal(cfg)
	if err != nil {
		log.Printf("error encoding config json: %s", err.Error())
		os.Exit(1)
	}

	log.Printf("[CONFIG] composite json:\n%s", string(bytes))

	return &cfg
}
