package gitlab

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/helpers"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *gitlab) AddUsersFromApi(link, typeTo string, info data.Permission) (*data.Permission, error) {
	jsonBody, err := createAddUserRequestBody(info.GitlabId, info.AccessLevel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create body")
	}

	params := data.RequestParams{
		Method: http.MethodPost,
		Link:   fmt.Sprintf("https://gitlab.com/api/v4/%ss/%s/members", typeTo, regexp.MustCompile("/").ReplaceAllString(link, "%2F")),
		Body:   jsonBody,
		Query:  nil,
		Header: map[string]string{
			"Content-Type":  "application/json",
			"PRIVATE-TOKEN": g.superUserToken,
		},
		Timeout: time.Second * 30,
	}

	res, err := helpers.MakeHttpRequest(params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make http request")
	}

	res, err = helpers.HandleHttpResponseStatusCode(res, params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check response status code")
	}
	if res == nil {
		return nil, nil
	}

	var response data.Permission
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal body")
	}

	return &response, nil
}

func createAddUserRequestBody(gitlabId, accessLevel int64) ([]byte, error) {
	body := struct {
		UserId      int64 `json:"user_id"`
		AccessLevel int64 `json:"access_level"`
	}{
		UserId:      gitlabId,
		AccessLevel: accessLevel,
	}

	return json.Marshal(body)
}
