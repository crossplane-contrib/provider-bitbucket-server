/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-bitbucket-server/internal/clients/bitbucket"
)

// Client defines the API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

type errorResponse struct {
	Errors []struct {
		Context       *string `json:"context"`
		Message       string  `json:"message"`
		ExceptionName *string `json:"exceptionName"`
	} `json:"errors"`

	code int
}

func (e errorResponse) Error() string {
	if len(e.Errors) > 0 {
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(e)
		if err != nil {
			return fmt.Sprintf("printing as json failed: %v", err)
		}
		return fmt.Sprintf("%v %v", e.code, buf.String())
	}
	return fmt.Sprintf("HTTP status %v", e.code)
}

// IsNotFound is a 404 error
func IsNotFound(err error) bool {
	var errResp errorResponse
	if errors.As(err, &errResp) {
		log.Printf("IsNotFound %+v", errResp)
		return errResp.code == http.StatusNotFound
	}
	return false
}

// NotFoundError is 404
func NotFoundError() error {
	return errorResponse{code: http.StatusNotFound}
}

func (c *Client) sendRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() // nolint

	// fmt.Printf("%v %v -> %v\n", req.Method, req.URL.String(), res.StatusCode)
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes errorResponse
		errRes.code = res.StatusCode
		if err = json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			fmt.Println(err.Error())
		}

		if res.StatusCode == http.StatusNotFound {
			return bitbucket.ErrNotFound
		}

		return errRes
	}

	if v != nil {
		if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
			return err
		}
	} /* else {
		fmt.Printf("Body not decoded: %v\n", req.URL)
		io.Copy(os.Stdout, res.Body)
		fmt.Println()
	}*/

	return nil
}

// Pagination defines response pagination
type Pagination struct {
	Size       int  `json:"size"`
	Limit      int  `json:"limit"`
	IsLastPage bool `json:"isLastPage"`
}
