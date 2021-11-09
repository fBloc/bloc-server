package user

import (
	"errors"
	"net/http"

	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/interfaces/web"
	"github.com/fBloc/bloc-backend-go/internal/json_date"
	"github.com/fBloc/bloc-backend-go/services/user"

	"github.com/google/uuid"
)

var uService *user.UserService

func InjectUserService(uS *user.UserService) {
	uService = uS
}

func InitialUserExistOrCreate(
	name, raWPassword string,
) (exist bool, err error) {
	if uService == nil {
		return false, errors.New("have to inject user service first")
	}
	exist, _, err = uService.Login(name, raWPassword)
	if err != nil {
		return
	}
	if exist { // 能登录成功，肯定就是存在了
		return
	}
	// 不存在默认初始化用户，创建之
	err = uService.AddUser(name, raWPassword, true)
	return
}

type User struct {
	Name        string             `json:"name"`
	RaWPassword string             `json:"password"`
	Token       uuid.UUID          `json:"token,omitempty"`
	CreateTime  json_date.JsonDate `json:"create_time"`
	IsSuper     bool               `json:"is_superuser"`
}

func (u *User) IsZero() bool {
	return u == nil
}

func FromAgg(aggU *aggregate.User) *User {
	if aggU.IsZero() {
		return nil
	}
	return &User{
		Name:       aggU.Name,
		Token:      aggU.ID,
		CreateTime: json_date.New(aggU.CreateTime),
		IsSuper:    aggU.IsSuper,
	}
}

func HttpSucRespFromAgg(w *http.ResponseWriter, aggU *aggregate.User) {
	web.WriteSucResp(w, FromAgg(aggU))
}

func HttpSucRespFromAggs(w *http.ResponseWriter, aggUs []aggregate.User) {
	us := make([]*User, len(aggUs))
	for i, j := range aggUs {
		us[i] = FromAgg(&j)
	}
	web.WriteSucResp(w, us)
}
