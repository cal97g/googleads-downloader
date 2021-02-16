package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
)

// profile contains a list of search queries.
type profile struct {
	Queries []*query `yaml:"queries" validate:"required"`
}

// validate ensures that the queries key exists.
func (p *profile) Validate() error {
	validate := *validator.New()
	return validate.Struct(p)
}

// interpolate injects template vars into the queries.
func (p *profile) interpolate(vars map[string]string) error {
	for _, q := range p.Queries {
		if err := q.interpolate(vars); err != nil {
			return err
		}
	}
	return nil
}

// query represents a Google Ads API Search query.
type query struct {
	FilePrefix string `yaml:"filename_prefix" validate:"required"`
	Query      string `yaml:"query" validate:"required"`
}

// interpolate injects template vars into the query, in place.
func (q *query) interpolate(vars map[string]string) error {
	for k, v := range vars {
		re, err := regexp.Compile(fmt.Sprintf("{{%s}}", k))
		if err != nil {
			return err
		}
		q.Query = re.ReplaceAllString(q.Query, v)
		q.FilePrefix = re.ReplaceAllString(q.FilePrefix, v)
	}

	// check if all template tags gave been replaced
	re := regexp.MustCompile("{{.*}}")
	matches := re.FindAllString(q.Query, -1)
	if len(matches) > 0 {
		return fmt.Errorf("unset template vars: %v", matches)
	}

	return nil
}

// execute makes a single call to Google Ads API.
func (q *query) execute(apiVersion string, accessToken, developerToken, mccID, accID string, pageSize *int, pageToken *string, backoffIntervals []int, queryLog logrus.FieldLogger) ([]byte, error) {

	if len(backoffIntervals) == 0 {
		backoffIntervals = []int{30, 60}
	}

	url := fmt.Sprintf("https://googleads.googleapis.com/%s/customers/%s/googleAds:search", apiVersion, accID)

	reqBody := struct {
		Query     string  `json:"query"`
		PageSize  *int    `json:"page_size,omitempty"`
		PageToken *string `json:"page_token,omitempty"`
	}{
		q.Query,
		pageSize,
		pageToken,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal body")
	}

	attempt := 1

	for {
		req, err := http.NewRequest("POST", url, bytes.NewReader(reqBodyBytes))
		if err != nil {
			return nil, fmt.Errorf("create POST request")
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		req.Header.Set("developer-token", developerToken)
		if mccID != "" {
			req.Header.Set("login-customer-id", mccID)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			// no retry here
			return nil, fmt.Errorf("make request %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// try to read body, ignore errors
			body, _ := ioutil.ReadAll(resp.Body)

			queryLog.Infof("Error Status %d, body: %s", resp.StatusCode, string(body))

			if resp.StatusCode >= 500 {
				// 502: temp network error, retry with backoff
				if attempt-1 == len(backoffIntervals) {
					queryLog.Infof("Exhausted %d retries", len(backoffIntervals))

					return nil, fmt.Errorf("response status: %d", resp.StatusCode)
				}

				waitTime := backoffIntervals[attempt-1]

				queryLog.Infof(
					"Waiting %d seconds before retrying (%d/%d)",
					waitTime, attempt, len(backoffIntervals))

				time.Sleep(time.Second * time.Duration(waitTime))
				attempt++
				continue
			}

			// not retryable
			return nil, fmt.Errorf("non retryable status code %d", resp.StatusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read response body: %w", err)
		}

		return body, nil
	}

}

// profileFromPath is a helper func that returns a valid Profile from a file path.
func profileFromPath(path string) (*profile, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var p profile
	if err := yaml.Unmarshal(bytes, &p); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	return &p, nil
}
