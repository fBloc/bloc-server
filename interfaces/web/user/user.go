package user

import (
	"errors"
	"net/http"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/internal/json_date"
	"github.com/fBloc/bloc-server/services/user"
	"github.com/fBloc/bloc-server/value_object"
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
	Token       value_object.UUID  `json:"token,omitempty"`
	ID          value_object.UUID  `json:"id,omitempty"`
	CreateTime  json_date.JsonDate `json:"create_time"`
	IsSuper     bool               `json:"super"`
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
		CreateTime: json_date.New(aggU.CreateTime),
		IsSuper:    aggU.IsSuper,
	}
}

func LoginRespFromAgg(w *http.ResponseWriter, aggU *aggregate.User) {
	tmp := FromAgg(aggU)
	tmp.Token = aggU.ID // 只有login的才返回token
	web.WriteSucResp(w, tmp)
}

func FilterRespFromAggs(w *http.ResponseWriter, aggUs []aggregate.User) {
	us := make([]*User, len(aggUs))
	for i, j := range aggUs {
		tmp := FromAgg(&j)
		tmp.ID = j.ID
		us[i] = tmp
	}
	web.WriteSucResp(w, us)
}
