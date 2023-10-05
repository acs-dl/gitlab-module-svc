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

func (g *gitlab) UpdateUserFromApi(info data.Permission) (*data.Permission, error) {
	params := data.RequestParams{
		Method: http.MethodPut,
		Link:   fmt.Sprintf("https://gitlab.com/api/v4/%ss/%s/members/%d", info.Type, regexp.MustCompile("/").ReplaceAllString(info.Link, "%2F"), info.GitlabId),
		Body:   nil,
		Query: map[string]string{
			"access_level": fmt.Sprintf("%v", info.AccessLevel),
		},
		Header: map[string]string{
			"PRIVATE-TOKEN": g.superUserToken,
		},
		Timeout: time.Second * 30,
	}

	res, err := helpers.MakeHttpRequest(params)
	if err != nil {
		return nil, err
	}

	res, err = helpers.HandleHttpResponseStatusCode(res, params)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errors.Errorf("No user with `%s` username was found in gitlab API", info.Username)
	}

	var response data.Permission
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal response body")
	}

	return &response, nil
}
