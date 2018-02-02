package model

import "github.com/satori/go.uuid"

type TokenResponse struct {
	*Response
	AccessToken uuid.UUID `json:"access_token,omitempty"`
}
