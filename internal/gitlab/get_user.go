package gitlab

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/helpers"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *gitlab) GetUserFromApi(username string) (*data.User, error) {
	params := data.RequestParams{
		Method: http.MethodGet,
		Link:   "https://gitlab.com/api/v4/users",
		Body:   nil,
		Query: map[string]string{
			"username": username,
		},
		Header: map[string]string{
			"PRIVATE-TOKEN": g.userToken,
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
		return nil, ErrNoSuchUser
	}

	return retrieveUserFromResponse(res.Body, username)
}

func retrieveUserFromResponse(body io.ReadCloser, username string) (*data.User, error) {
	var response []data.User
	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal body")
	}

	if len(response) == 0 {
		return nil, ErrNoSuchUser
	}

	for i := range response {
		if response[i].GitlabUsername == username {
			return &response[i], nil
		}

	}

	return nil, ErrNoSuchUser
}
