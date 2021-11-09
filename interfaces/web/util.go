package web

import (
	"errors"

	"github.com/google/uuid"
)

func ParseStrValueToGoogleUUID(key, value string) (uuid.UUID, error) {
	if value == "" {
		return uuid.Nil, errors.New(key + " cannot be blank")
	}
	return uuid.Parse(value)
}
