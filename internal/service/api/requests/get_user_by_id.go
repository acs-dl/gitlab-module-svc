package requests

import "net/http"

func NewGetUserByIdRequest(r *http.Request) (int64, error) {
	return RetrieveId(r)
}
