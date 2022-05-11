package http_server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/interfaces/web/user"
	"github.com/fBloc/bloc-server/internal/http_util"
	"github.com/fBloc/bloc-server/internal/timestamp"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	toAddUserName   = gofakeit.Name()
	toAddUserPasswd = gofakeit.Password(false, false, false, false, false, 16)
)

func TestUserFilterByName(t *testing.T) {
	Convey("filter by name hit", t, func() {
		resp := struct {
			web.RespMsg
			Data []aggregate.User `json:"data"`
		}{}
		name := config.DefaultUserName[1 : len(config.DefaultUserName)-1]
		_, err := http_util.Get(
			superuserHeader(),
			fmt.Sprintf("%s%s", serverAddress, "/api/v1/user"),
			map[string]string{"name__contains": name}, &resp)
		So(err, ShouldBeNil)
		So(len(resp.Data), ShouldBeGreaterThan, 0)
		So(resp.Data[0].ID.IsNil(), ShouldBeFalse)
		So(resp.Data[0].Token.IsNil(), ShouldBeTrue) // filter should not return token
	})

	Convey("filter by name miss", t, func() {
		resp := struct {
			web.RespMsg
			Data []user.User `json:"data"`
		}{}
		name := config.DefaultUserName + "miss"
		_, err := http_util.Get(
			superuserHeader(),
			fmt.Sprintf("%s%s", serverAddress, "/api/v1/user"),
			map[string]string{"name__contains": name}, &resp)
		So(err, ShouldBeNil)
		So(len(resp.Data), ShouldEqual, 0)
	})
}

func TestUserInfo(t *testing.T) {
	Convey("get logginned user info", t, func() {
		resp := struct {
			web.RespMsg
			User *struct {
				Name       string               `json:"name"`
				CreateTime *timestamp.Timestamp `json:"create_time"`
				IsSuper    bool                 `json:"super"`
				IsAdmin    bool                 `json:"is_admin"`
			} `json:"data"`
		}{}

		Convey("not exist user should require token", func() {
			_, err := http_util.Get(
				notExistUserHeader(),
				fmt.Sprintf("%s%s", serverAddress, "/api/v1/user/info"),
				nil, &resp)
			So(err, ShouldBeNil)
			So(resp.User, ShouldBeNil)
			So(resp.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("admin user hit", func() {
			_, err := http_util.Get(
				superuserHeader(),
				fmt.Sprintf("%s%s", serverAddress, "/api/v1/user/info"),
				nil, &resp)
			So(err, ShouldBeNil)
			So(resp.Code, ShouldEqual, http.StatusOK)
			So(resp.User.IsAdmin, ShouldBeTrue)
		})

		Convey("nobody user hit", func() {
			_, err := http_util.Get(
				nobodyHeader(),
				fmt.Sprintf("%s%s", serverAddress, "/api/v1/user/info"),
				nil, &resp)
			So(err, ShouldBeNil)
			So(resp.Code, ShouldEqual, http.StatusOK)
			So(resp.User.IsAdmin, ShouldBeFalse)
		})
	})
}

func TestAddDeleteUser(t *testing.T) {
	Convey("AddUser & DeleteUser", t, func() {
		addUser := user.User{
			Name:        toAddUserName,
			RaWPassword: toAddUserPasswd}
		addPostBody, _ := json.Marshal(addUser)
		var addResp web.RespMsg
		_, err := http_util.Post(
			superuserHeader(),
			serverAddress+"/api/v1/user",
			http_util.BlankGetParam, addPostBody, &addResp)
		So(err, ShouldBeNil)
		So(addResp.Code, ShouldEqual, http.StatusOK)

		resp := struct {
			web.RespMsg
			Data []user.User `json:"data"`
		}{}

		_, err = http_util.Get(
			superuserHeader(),
			fmt.Sprintf("%s%s", serverAddress, "/api/v1/user"),
			map[string]string{"name__contains": toAddUserName}, &resp)
		So(err, ShouldBeNil)
		So(len(resp.Data), ShouldBeGreaterThan, 0)
		So(resp.Data[0].ID.IsNil(), ShouldBeFalse)
		So(resp.Data[0].Token.IsNil(), ShouldBeTrue)
		theUserID := resp.Data[0].ID

		Convey("DeleteUser", func() {
			var resp web.RespMsg
			_, err := http_util.Delete(
				superuserHeader(),
				fmt.Sprintf(
					"%s%s",
					serverAddress, "/api/v1/user/delete_by_id/"+theUserID.String()),
				http_util.BlankGetParam, []byte{}, &resp)
			So(err, ShouldBeNil)
			So(resp.Code, ShouldEqual, http.StatusOK)

			deletedResp := struct {
				web.RespMsg
				Data []user.User `json:"data"`
			}{}

			_, err = http_util.Get(
				superuserHeader(),
				fmt.Sprintf("%s%s", serverAddress, "/api/v1/user"),
				map[string]string{"name__contains": toAddUserName}, &resp)
			So(err, ShouldBeNil)
			So(len(deletedResp.Data), ShouldEqual, 0)
		})
	})
}
