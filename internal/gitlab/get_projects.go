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

func (g *gitlab) GetProjectsFomApi(link string) ([]data.Sub, error) {
	response, err := helpers.MakeRequestWithPagination(data.RequestParams{
		Method: http.MethodGet,
		Link:   fmt.Sprintf("https://gitlab.com/api/v4/groups/%s/projects", regexp.MustCompile("/").ReplaceAllString(link, "%2F")),
		Body:   nil,
		Query: map[string]string{
			"per_page": "100",
		},
		Header: map[string]string{
			"PRIVATE-TOKEN": g.superUserToken,
		},
		Timeout: time.Second * 30,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request with pagination")
	}

	var result []data.Sub
	if err = json.Unmarshal(response, &result); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal response body")
	}

	return result, nil
}
