package model

import (
	"errors"
)

type Translation struct {
	AddedOn string `json:"added_on"`
}

func (t Translation) Validate() error {
	if t.AddedOn == "" {
		return errors.New("required: AddedOn")
	}

	return nil
}
