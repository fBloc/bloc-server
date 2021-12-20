package web

import (
	"errors"

	"github.com/fBloc/bloc-server/value_object"
)

func ParseStrValueToUUID(key, value string) (value_object.UUID, error) {
	if value == "" {
		return value_object.NillUUID, errors.New(key + " cannot be blank")
	}
	return value_object.ParseToUUID(value)
}
