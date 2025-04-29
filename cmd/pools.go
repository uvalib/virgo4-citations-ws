package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/uvalib/virgo4-api/v4api"
	"github.com/uvalib/virgo4-jwt/v4jwt"
)

func (s *citationsContext) queryPoolRecord() (*v4api.Record, serviceResponse) {
	var err error

	if s.url == "" {
		err = fmt.Errorf("missing or invalid url")
		s.warn(err.Error())
		return nil, serviceResponse{status: http.StatusBadRequest, err: err}
	}

	// create a short-lived single-use token, (ab)using IsUVA claim to make
	// sure we get all possible info from the pool.

	// there is a risk that anyone can access protected resources through
	// this service, but that is limited to the citation info exposed by each pool.

	claims := v4jwt.V4Claims{IsUVA: true}

	token, jwtErr := v4jwt.Mint(claims, time.Duration(s.svc.config.JWT.Expiration)*time.Minute, s.svc.config.JWT.Key)
	if jwtErr != nil {
		err = fmt.Errorf("failed to mint JWT: %s", jwtErr.Error())
		s.err(err.Error())
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
