package requests

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func RetrieveId(r *http.Request) (int64, error) {
	id := chi.URLParam(r, "id")

	if id == "" {
		return 0, errors.New("`id` param is not specified")
	}

	return strconv.ParseInt(id, 10, 64)
}
