package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func MakeHttpRequest(params data.RequestParams) (*data.ResponseParams, error) {
	req, err := http.NewRequest(params.Method, params.Link, bytes.NewReader(params.Body))
	if err != nil {
		return nil, errors.Wrap(err, "couldn't create request")
	}

	ctx, cancel := context.WithTimeout(context.Background(), params.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	if params.Header != nil {
		for key, value := range params.Header {
			req.Header.Set(key, value)
		}
	}

	if params.Query != nil {
		q := req.URL.Query()
		for key, value := range params.Query {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error making http request")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}
	clearBody := io.NopCloser(bytes.NewReader(body))

	return &data.ResponseParams{
		Body:       clearBody,
		Header:     response.Header,
		StatusCode: response.StatusCode,
	}, nil
}

func HandleHttpResponseStatusCode(response *data.ResponseParams, params data.RequestParams) (*data.ResponseParams, error) {
	switch status := response.StatusCode; {
	case status >= http.StatusOK && status < http.StatusMultipleChoices:
		return response, nil
	case status == http.StatusNotFound:
		return nil, nil
	case status == http.StatusTooManyRequests:
		return HandleTooManyRequests(response.Header, params)
	case status < http.StatusOK || status >= http.StatusMultipleChoices:
		return nil, errors.New(fmt.Sprintf("error in response `%s`", http.StatusText(response.StatusCode)))
	default:
		return nil, errors.New(fmt.Sprintf("unexpected response status `%s`", http.StatusText(response.StatusCode)))
	}
}

func HandleTooManyRequests(header http.Header, params data.RequestParams) (*data.ResponseParams, error) {
	durationInSeconds, err := strconv.ParseInt(header.Get("Retry-After"), 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse `retry-after `header")
	}

	timeoutDuration := time.Second * time.Duration(durationInSeconds)
	log.Printf("we need to wait `%s`", timeoutDuration.String())
	time.Sleep(timeoutDuration)

	return MakeHttpRequest(params)
}

func MakeRequestWithPagination(params data.RequestParams) ([]byte, error) {
	result := make([]interface{}, 0)

	for page := 1; ; page++ {
		params.Query["page"] = strconv.Itoa(page)

		res, err := MakeHttpRequest(params)
		if err != nil {
			return nil, errors.Wrap(err, "failed to make http request")
		}

		res, err = HandleHttpResponseStatusCode(res, params)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check response status code")
		}
		if res == nil {
			return nil, errors.Errorf("error in response, status %v", res.StatusCode)
		}

		var response []interface{}
		if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal body")
		}

		if len(response) == 0 {
			break
		}

		result = append(result, response...)
	}

	return json.Marshal(result)
}
