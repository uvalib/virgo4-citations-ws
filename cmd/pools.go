package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/uvalib/virgo4-api/v4api"
)

func (s *citationsContext) queryPoolRecord() (*v4api.Record, serviceResponse) {
	var err error

	url := s.client.ginCtx.Query("url")

	if url == "" {
		err = fmt.Errorf("missing url")
		s.err(err.Error())
		return nil, serviceResponse{status: http.StatusBadRequest, err: err}
	}

	s.log("url = [%s]", s.url)

	req, reqErr := http.NewRequest("GET", url, nil)
	if reqErr != nil {
		s.log("[POOL] NewRequest() failed: %s", reqErr.Error())
		err = fmt.Errorf("failed to create pool record request")
		return nil, serviceResponse{status: http.StatusInternalServerError, err: err}
	}

	req.Header.Set("Authorization", s.client.ginCtx.GetHeader("Authorization"))

	start := time.Now()
	res, resErr := s.svc.pools.client.Do(req)
	elapsedMS := int64(time.Since(start) / time.Millisecond)

	// external service failure logging

	if resErr != nil {
		status := http.StatusBadRequest
		errMsg := resErr.Error()
		if strings.Contains(errMsg, "Timeout") {
			status = http.StatusRequestTimeout
			errMsg = fmt.Sprintf("%s timed out", url)
		} else if strings.Contains(errMsg, "connection refused") {
			status = http.StatusServiceUnavailable
			errMsg = fmt.Sprintf("%s refused connection", url)
		}

		s.log("[POOL] client.Do() failed: %s", resErr.Error())
		s.log("ERROR: Failed response from %s %s - %d:%s. Elapsed Time: %d (ms)", req.Method, url, status, errMsg, elapsedMS)
		err = fmt.Errorf("failed to receive pool record response")
		return nil, serviceResponse{status: http.StatusInternalServerError, err: err}
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errMsg := fmt.Errorf("unexpected status code %d", res.StatusCode)
		s.log("[POOL] unexpected status code %d", res.StatusCode)
		s.log("ERROR: Failed response from %s %s - %d:%s. Elapsed Time: %d (ms)", req.Method, url, res.StatusCode, errMsg, elapsedMS)
		err = fmt.Errorf("received pool record response code %d", res.StatusCode)
		return nil, serviceResponse{status: http.StatusInternalServerError, err: err}
	}

	var rec v4api.Record

	decoder := json.NewDecoder(res.Body)

	// external service failure logging (scenario 2)

	if decErr := decoder.Decode(&rec); decErr != nil {
		s.log("[POOL] Decode() failed: %s", decErr.Error())
		s.log("ERROR: Failed response from %s %s - %d:%s. Elapsed Time: %d (ms)", req.Method, url, http.StatusInternalServerError, decErr.Error(), elapsedMS)
		err = fmt.Errorf("failed to decode pool record response")
		return nil, serviceResponse{status: http.StatusInternalServerError, err: err}
	}

	// external service success logging

	s.log("Successful pool record response from %s %s. Elapsed Time: %d (ms)", req.Method, url, elapsedMS)

	return &rec, serviceResponse{status: http.StatusOK}
}
