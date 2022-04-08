package user

import (
	"errors"
	"net/http"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/internal/timestamp"
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
	ID          value_object.UUID    `json:"id,omitempty"`
	Token       value_object.UUID    `json:"token,omitempty"` // only return when login in
	Name        string               `json:"name"`
	RaWPassword string               `json:"password"`
	CreateTime  *timestamp.Timestamp `json:"create_time"`
	IsSuper     bool                 `json:"super"`
}

func (u *User) IsZero() bool {
	if u == nil {
		return true
	}
	return u.ID.IsNil()
}

func FromAgg(aggU *aggregate.User) *User {
	if aggU.IsZero() {
		return nil
	}
	return &User{
		ID:         aggU.ID,
		Name:       aggU.Name,
		CreateTime: timestamp.NewTimeStampFromTime(aggU.CreateTime),
		IsSuper:    aggU.IsSuper,
	}
}

func LoginRespFromAgg(w *http.ResponseWriter, r *http.Request, aggU *aggregate.User) {
	tmp := FromAgg(aggU)
	tmp.Token = aggU.Token // only login should return token!
	web.WriteSucResp(w, r, tmp)
}

func FilterRespFromAggs(w *http.ResponseWriter, r *http.Request, aggUs []aggregate.User) {
	us := make([]*User, len(aggUs))
	for i, j := range aggUs {
		tmp := FromAgg(&j)
		us[i] = tmp
	}
	web.WriteSucResp(w, r, us)
}
