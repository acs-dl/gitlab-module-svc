package gitlab

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/helpers"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *gitlab) RemoveUserFromApi(link, typeTo string, gitlabId int64) error {
	params := data.RequestParams{
		Method: http.MethodDelete,
		Link:   fmt.Sprintf("https://gitlab.com/api/v4/%ss/%s/members/%d", typeTo, regexp.MustCompile("/").ReplaceAllString(link, "%2F"), gitlabId),
		Body:   nil,
		Query:  nil,
		Header: map[string]string{
			"PRIVATE-TOKEN": g.superUserToken,
		},
		Timeout: time.Second * 30,
	}

	res, err := helpers.MakeHttpRequest(params)
	if err != nil {
		return err
	}

	res, err = helpers.HandleHttpResponseStatusCode(res, params)
	if err != nil {
		return err
	}
	if res == nil {
		return errors.Errorf("No user with `%d` Gitlab ID was found in `%s`", gitlabId, link)
	}

	return nil
}
