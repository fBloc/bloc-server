package aggregate

import (
	"time"

	"github.com/fBloc/bloc-backend-go/internal/util"

	"github.com/google/uuid"
)

var salt = "may the force be with you"

func ChangeSalt(userSalt string) {
	if userSalt == "" {
		return
	}
	salt = userSalt
}

type User struct {
	ID          uuid.UUID
	Name        string
	RaWPassword string
	Password    string
	CreateTime  time.Time
	IsSuper     bool
}

func NewUser(name, rawPassword string, isSuper bool) User {
	return User{
		ID:          uuid.New(),
		Name:        name,
		RaWPassword: rawPassword,
		Password:    encodePassword(rawPassword),
		CreateTime:  time.Now(),
		IsSuper:     isSuper,
	}
}

func (u *User) IsZero() bool {
	if u == nil {
		return true
	}
	return u.ID == uuid.Nil
}

func encodePassword(rawPassword string) string {
	return util.Sha1([]byte(rawPassword + salt))
}

func (u *User) IsRawPasswordMatch(rawPassword string) (bool, error) {
	if u.IsZero() {
		return false, nil
	}
	if u.Password == encodePassword(rawPassword) {
		return true, nil
	}
	return false, nil
}