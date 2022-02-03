package http_server

import (
	"encoding/json"
	"testing"

	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/interfaces/web/user"
	"github.com/fBloc/bloc-server/internal/http_util"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLogin(t *testing.T) {
	Convey("login should fail", t, func() {
		loginUser := user.User{
			Name:        config.DefaultUserName + "aha",
			RaWPassword: config.DefaultUserPassword + "wuhu"}
		postBody, err := json.Marshal(loginUser)
		So(err, ShouldBeNil)

		resp := struct {
			web.RespMsg
			Data user.User `json:"data"`
		}{}
		_, err = http_util.Post(
			serverAddress+"/api/v1/login",
			http_util.BlankHeader, postBody, &resp)
		So(err, ShouldBeNil)
		So(resp.Data.Token.IsNil(), ShouldBeTrue)
	})

	Convey("login should suc", t, func() {
		loginUser := user.User{
			Name:        config.DefaultUserName,
			RaWPassword: config.DefaultUserPassword}
		postBody, err := json.Marshal(loginUser)
		So(err, ShouldBeNil)

		resp := struct {
			Code int       `json:"status_code"`
			Msg  string    `json:"status_msg"`
			Data user.User `json:"data"`
		}{}
		_, err = http_util.Post(
			serverAddress+"/api/v1/login",
			http_util.BlankHeader, postBody, &resp)
		So(err, ShouldBeNil)
		So(resp.Data.Token.IsNil(), ShouldBeFalse)
	})
}
