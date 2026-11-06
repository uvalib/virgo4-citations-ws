package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/uvalib/virgo4-api/v4api"
	"github.com/uvalib/virgo4-jwt/v4jwt"
)

func (s *citationsContext) urlParseStrict(rawURL string) (*url.URL, error) {
	// meticulously validate item urls passed to this service

	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil, err
	}

	// ensure the url points to a webpage
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme: [%s]", u.Scheme)
	}

	// ensure a host is supplied in the url
	if u.Host == "" {
		return nil, fmt.Errorf("missing host")
	}

	// if the host contains a port, validate it with net.SplitHostPort, which will
	// catch things like "hostname:123:456" that slip past url.Parse{,RequestURI}.

	// check whether host contains a port delimiter, ignoring hosts that end in the
	// IPv6 bracket delimiter (because they will contain colons, but not a port):
	if strings.Contains(u.Host, ":") && !strings.HasSuffix(u.Host, "]") {
		_, _, err := net.SplitHostPort(u.Host)
		if err != nil {
			return nil, fmt.Errorf("invalid host/port [%s]: %w", u.Host, err)
		}
	}

	return u, nil
}

func (s *citationsContext) itemURLisFromAllowedHost() bool {
	// check host from provided url against configured whitelist

	u, err := s.urlParseStrict(s.url)
	if err != nil {
		s.warn("item url parsing failed: %s", err.Error())
		return false
	}

	s.log("item url     = [%s]", s.url)
	s.log("u.Host       = [%s]", u.Host)
	s.log("u.Hostname() = [%s]", u.Hostname())

	for _, pattern := range s.svc.config.Pools.AllowedHosts {
		re, err := regexp.Compile(pattern)
		if err != nil {
			s.warn("skipping invalid item url allowed host pattern [%s]: %s", pattern, err.Error())
			continue
		}
		if re.MatchString(u.Host) {
			s.log("host [%s] matched pattern [%s]", u.Host, pattern)
			return true
		}
		s.log("host [%s] did NOT match pattern [%s]", u.Host, pattern)
	}

	return false
}

func (s *citationsContext) queryPoolRecord() (*v4api.Record, serviceResponse) {
	var err error

	// validate that the user-provided item url is from an allowed host.
	// otherwise an attacker can abuse this to get a token with IsUVA privileges

	if s.itemURLisFromAllowedHost() == false {
		err = fmt.Errorf("item url is invalid or unsupported")
		s.warn("%s", err.Error())
		return nil, serviceResponse{status: http.StatusBadRequest, err: err}
	}

	// create a short-lived single-use token, (ab)using IsUVA claim to make
	// sure we get all possible info from the pool.

	claims := v4jwt.V4Claims{IsUVA: true}

	token, jwtErr := v4jwt.Mint(claims, time.Duration(s.svc.config.JWT.Expiration)*time.Minute, s.svc.config.JWT.Key)
	if jwtErr != nil {
		err = fmt.Errorf("failed to mint JWT: %s", jwtErr.Error())
		s.err("%s", err.Error())
		return nil, serviceResponse{status: http.StatusBadRequest, err: err}
	}

	// the citation query parameter is only used by the solr pool, and is not relevant to other pools
	req, reqErr := http.NewRequest("GET", s.url+"?citation=1", nil)
	if reqErr != nil {
		s.log("[POOL] NewRequest() failed: %s", reqErr.Error())
		err = fmt.Errorf("failed to create pool record request")
		return nil, serviceResponse{status: http.StatusInternalServerError, err: err}
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	start := time.Now()
	res, resErr := s.svc.pools.client.Do(req)
	elapsedMS := int64(time.Since(start) / time.Millisecond)

	// external service failure logging

	if resErr != nil {
		status := http.StatusBadRequest
		errMsg := resErr.Error()
		if strings.Contains(errMsg, "Timeout") {
			status = http.StatusRequestTimeout
			errMsg = fmt.Sprintf("%s timed out", s.url)
		} else if strings.Contains(errMsg, "connection refused") {
			status = http.StatusServiceUnavailable
			errMsg = fmt.Sprintf("%s refused connection", s.url)
		}

		s.log("[POOL] client.Do() failed: %s", resErr.Error())
		s.err("Failed response from %s %s - %d:%s. Elapsed Time: %d (ms)", req.Method, s.url, status, errMsg, elapsedMS)
		err = fmt.Errorf("failed to receive pool record response")
		return nil, serviceResponse{status: http.StatusInternalServerError, err: err}
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errMsg := fmt.Errorf("unexpected status code %d", res.StatusCode)
		s.log("[POOL] unexpected status code %d", res.StatusCode)

		// 404 errors from the pool are likely old bookmarked items that have since become shadowed, or no longer exist
		if res.StatusCode == http.StatusNotFound {
			s.warn("Failed response from %s %s - %d:%s. Elapsed Time: %d (ms)", req.Method, s.url, res.StatusCode, errMsg, elapsedMS)
		} else {
			s.err("Failed response from %s %s - %d:%s. Elapsed Time: %d (ms)", req.Method, s.url, res.StatusCode, errMsg, elapsedMS)
		}

		err = fmt.Errorf("received pool record response code %d", res.StatusCode)
		return nil, serviceResponse{status: http.StatusInternalServerError, err: err}
	}

	var rec v4api.Record

	decoder := json.NewDecoder(res.Body)

	// external service failure logging (scenario 2)

	if decErr := decoder.Decode(&rec); decErr != nil {
		s.log("[POOL] Decode() failed: %s", decErr.Error())
		s.err("Failed response from %s %s - %d:%s. Elapsed Time: %d (ms)", req.Method, s.url, http.StatusInternalServerError, decErr.Error(), elapsedMS)
		err = fmt.Errorf("failed to decode pool record response")
		return nil, serviceResponse{status: http.StatusInternalServerError, err: err}
	}

	// external service success logging

	s.log("Successful pool record response from %s %s. Elapsed Time: %d (ms)", req.Method, s.url, elapsedMS)

	return &rec, serviceResponse{status: http.StatusOK}
}
