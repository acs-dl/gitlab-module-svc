package models

import (
	"github.com/acs-dl/gitlab-module-svc/resources"
)

func NewInputsModel() resources.Inputs {
	result := resources.Inputs{
		Key: resources.Key{
			ID:   "inputs",
			Type: resources.INPUTS,
		},
		Attributes: resources.InputsAttributes{
			Username:    "string",
			Link:        "string",
			AccessLevel: "number",
		},
	}

	return result
}

func NewInputsResponse() resources.InputsResponse {
	return resources.InputsResponse{
		Data: NewInputsModel(),
	}
}
