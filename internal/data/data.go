package data

import (
	"io"
	"net/http"
	"time"
)

const (
	ModuleName = "gitlab"
	Project    = "project"
	Group      = "group"
)

type ModuleRequest struct {
	ID            string    `db:"id" structs:"id"`
	UserID        int64     `db:"user_id" structs:"user_id"`
	Module        string    `db:"module" structs:"module"`
	Payload       string    `db:"payload" structs:"payload"`
	CreatedAt     time.Time `db:"created_at" structs:"created_at"`
	RequestStatus string    `db:"request_status" structs:"request_status"`
	Error         string    `db:"error" structs:"error"`
}

type ModulePayload struct {
	RequestId string `json:"request_id"`
	UserId    string `json:"user_id"`
	Action    string `json:"action"`

	//other fields that are required for module
	Type        string   `json:"type"`
	Link        string   `json:"link"`
	Links       []string `json:"links"`
	Username    string   `json:"username"`
	GitlabId    string   `json:"gitlab_id"`
	AccessLevel int64    `json:"access_level"`
}

type UnverifiedPayload struct {
	Action string           `json:"action"`
	Users  []UnverifiedUser `json:"users"`
}

var Roles = map[int64]string{
	0:  "No access",
	10: "Guest",
	20: "Reporter",
	30: "Developer",
	40: "Maintainer",
	50: "Owner",
}

var RolesArray = []int64{0, 10, 20, 30, 40, 50}

type RequestParams struct {
	Method  string
	Link    string
	Body    []byte
	Query   map[string]string
	Header  map[string]string
	Timeout time.Duration
}

type ResponseParams struct {
	Body       io.ReadCloser
	Header     http.Header
	StatusCode int
}
