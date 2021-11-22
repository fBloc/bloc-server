package value_object

import (
	"fmt"

	"github.com/google/uuid"
)

type UUID uuid.UUID

var NillUUID = UUID(uuid.Nil)

func NewUUID() UUID {
	return UUID(uuid.New())
}

func ParseToUUID(s string) (UUID, error) {
	gUUID, err := uuid.Parse(s)
	tUUID := UUID(gUUID)
	return tUUID, err
}

func (u *UUID) IsNil() bool {
	return uuid.UUID(*u) == uuid.Nil
}

func (u *UUID) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		*u = UUID(uuid.Nil)
		return nil

	}
	var err error
	id, err := uuid.ParseBytes(data)
	if err != nil {
		return fmt.Errorf("%x: invalid format", data)
	}
	*u = UUID(id)
	return nil
}

func (u UUID) MarshalText() ([]byte, error) {
	if u == UUID(uuid.Nil) {
		return []byte{}, nil
	}
	return uuid.UUID(u).MarshalText()
}

func (u UUID) String() string {
	return uuid.UUID(u).String()
}
