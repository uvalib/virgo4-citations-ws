package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// git commit used for this build; supplied at compile time
var gitCommit string

type serviceVersion struct {
	BuildVersion string `json:"build,omitempty"`
	GoVersion    string `json:"go_version,omitempty"`
	GitCommit    string `json:"git_commit,omitempty"`
}

type servicePools struct {
	client *http.Client
}

type serviceContext struct {
	randomSource *rand.Rand
	config       *serviceConfig
	version      serviceVersion
	pools        servicePools
}

func (p *serviceContext) initVersion() {
	buildVersion := "unknown"
	files, _ := filepath.Glob("buildtag.*")
	if len(files) == 1 {
		buildVersion = strings.Replace(files[0], "buildtag.", "", 1)
	}

	p.version = serviceVersion{
		BuildVersion: buildVersion,
		GoVersion:    fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
		GitCommit:    gitCommit,
	}

	log.Printf("[SERVICE] version.BuildVersion = [%s]", p.version.BuildVersion)
	log.Printf("[SERVICE] version.GoVersion    = [%s]", p.version.GoVersion)
	log.Printf("[SERVICE] version.GitCommit    = [%s]", p.version.GitCommit)
}

func (p *serviceContext) initPools() {
	// client setup

	connTimeout := integerWithMinimum(p.config.Pools.ConnTimeout, 1)
	readTimeout := integerWithMinimum(p.config.Pools.ReadTimeout, 1)

	poolsClient := &http.Client{
		Timeout: time.Duration(readTimeout) * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   time.Duration(connTimeout) * time.Second,
				KeepAlive: 60 * time.Second,
			}).DialContext,
			MaxIdleConns:        100, // we could be hitting one of a dozen or so pools,
			MaxIdleConnsPerHost: 10,  // so spread out the idle connections among them
			IdleConnTimeout:     90 * time.Second,
		},
	}

	p.pools = servicePools{
		client: poolsClient,
	}
}

func initializeService(cfg *serviceConfig) *serviceContext {
	p := serviceContext{}

	p.config = cfg
	p.randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))

	p.initVersion()
	p.initPools()

	return &p
}
