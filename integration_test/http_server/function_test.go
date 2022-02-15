package http_server

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/interfaces/web/function"
	"github.com/fBloc/bloc-server/interfaces/web/user"
	"github.com/fBloc/bloc-server/internal/http_util"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFilterFunction(t *testing.T) {
	Convey("filter functions", t, func() {
		functionResp := struct {
			web.RespMsg
			Data []function.GroupFunctions `json:"data"`
		}{}
		_, err := http_util.Get(
			superuserHeader(),
			serverAddress+"/api/v1/function",
			http_util.BlankGetParam,
			&functionResp)
		So(err, ShouldBeNil)
		So(functionResp.Code, ShouldEqual, http.StatusOK)
		So(len(functionResp.Data), ShouldEqual, 1)
		So(functionResp.Data[0].GroupName, ShouldEqual, fakeAggFunction.GroupName)
		So(functionResp.Data[0].Functions[0].ID, ShouldEqual, fakeAggFunction.ID)
		So(functionResp.Data[0].Functions[0].Name, ShouldEqual, fakeAggFunction.Name)
	})
}

func TestFunctionPermission(t *testing.T) {
	basePath := "/api/v1/function_permission"

	Convey("get permission", t, func() {
		Convey("superuser should have all permission", func() {
			permissionResp := struct {
				web.RespMsg
				Data function.PermissionResp `json:"data"`
			}{}
			_, err := http_util.Get(
				superuserHeader(),
				serverAddress+basePath,
				map[string]string{"function_id": fakeAggFunction.ID.String()},
				&permissionResp)
			So(err, ShouldBeNil)
			So(permissionResp.Code, ShouldEqual, http.StatusOK)
			So(permissionResp.Data.Read, ShouldBeTrue)
			So(permissionResp.Data.Execute, ShouldBeTrue)
			So(permissionResp.Data.AssignPermission, ShouldBeTrue)
		})

		Convey("nobody should login failed", func() {
			permissionResp := struct {
				web.RespMsg
				Data function.PermissionResp `json:"data"`
			}{}
			_, err := http_util.Get(
				notExistUserHeader(),
				serverAddress+basePath,
				map[string]string{"function_id": fakeAggFunction.ID.String()},
				&permissionResp)
			So(err, ShouldBeNil)
			So(permissionResp.Code, ShouldEqual, http.StatusUnauthorized)
		})
	})

	Convey("permission add/delete", t, func() {
		Convey("add a user for read", func() {
			// add a user
			addUser := user.User{
				Name:        gofakeit.Name(),
				RaWPassword: gofakeit.Password(false, false, false, false, false, 16)}
			addPostBody, _ := json.Marshal(addUser)
			var addUserResp web.RespMsg
			_, err := http_util.Post(
				superuserHeader(),
				serverAddress+"/api/v1/user",
				http_util.BlankGetParam, addPostBody, &addUserResp)
			So(err, ShouldBeNil)
			So(addUserResp.Code, ShouldEqual, http.StatusOK)

			// login to get the token
			loginUser := user.User{
				Name:        addUser.Name,
				RaWPassword: addUser.RaWPassword}
			loginBody, err := json.Marshal(loginUser)
			So(err, ShouldBeNil)
			resp := struct {
				web.RespMsg
				Data user.User `json:"data"`
			}{}
			_, err = http_util.Post(
				http_util.BlankHeader,
				serverAddress+"/api/v1/login",
				http_util.BlankGetParam, loginBody, &resp)
			So(err, ShouldBeNil)
			So(resp.Data.Token.IsNil(), ShouldBeFalse)
			So(resp.Data.ID.IsNil(), ShouldBeFalse)

			thisUserID := resp.Data.ID
			thisUserLoginedHeader := map[string]string{"token": resp.Data.Token.String()}

			// check the user have no read permission of the function before add
			permissionResp := struct {
				web.RespMsg
				Data function.PermissionResp `json:"data"`
			}{}
			_, err = http_util.Get(
				thisUserLoginedHeader, serverAddress+basePath,
				map[string]string{"function_id": fakeAggFunction.ID.String()},
				&permissionResp)
			So(err, ShouldBeNil)
			So(permissionResp.Code, ShouldEqual, http.StatusOK)
			So(permissionResp.Data.Read, ShouldBeFalse)

			// check the user can really not get the function
			functionResp := struct {
				web.RespMsg
				Data []function.GroupFunctions `json:"data"`
			}{}
			_, err = http_util.Get(
				thisUserLoginedHeader,
				serverAddress+"/api/v1/function",
				http_util.BlankGetParam, &functionResp)
			So(err, ShouldBeNil)
			So(functionResp.Code, ShouldEqual, http.StatusOK)
			So(len(functionResp.Data), ShouldEqual, 0)

			// add read permission for that user
			addRead := function.PermissionReq{
				PermissionType: function.Read,
				FunctionID:     fakeAggFunction.ID,
				UserID:         thisUserID}
			addReadPostBody, _ := json.Marshal(addRead)
			var addPermissionResp web.RespMsg
			_, err = http_util.Post(
				superuserHeader(),
				serverAddress+"/api/v1/function_permission/add_permission",
				http_util.BlankGetParam, addReadPostBody, &addPermissionResp)
			So(err, ShouldBeNil)

			// check the permission is suc added
			permissionResp = struct {
				web.RespMsg
				Data function.PermissionResp `json:"data"`
			}{}
			_, err = http_util.Get(
				thisUserLoginedHeader,
				serverAddress+basePath,
				map[string]string{"function_id": fakeAggFunction.ID.String()},
				&permissionResp)
			So(err, ShouldBeNil)
			So(permissionResp.Code, ShouldEqual, http.StatusOK)
			So(permissionResp.Data.Read, ShouldBeTrue)
			So(permissionResp.Data.Execute, ShouldBeFalse)
			So(permissionResp.Data.AssignPermission, ShouldBeFalse)

			_, err = http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/function",
				http_util.BlankGetParam,
				&functionResp)
			So(err, ShouldBeNil)
			So(functionResp.Code, ShouldEqual, http.StatusOK)
			So(len(functionResp.Data), ShouldEqual, 1)
			So(functionResp.Data[0].GroupName, ShouldEqual, fakeAggFunction.GroupName)
			So(functionResp.Data[0].Functions[0].ID, ShouldEqual, fakeAggFunction.ID)
			So(functionResp.Data[0].Functions[0].Name, ShouldEqual, fakeAggFunction.Name)

			Convey("delete a user for read", func() {
				// delete that user of read permission
				removeRead := function.PermissionReq{
					PermissionType: function.Read,
					FunctionID:     fakeAggFunction.ID,
					UserID:         thisUserID}
				removeReadPostBody, _ := json.Marshal(removeRead)
				var removePermissionResp web.RespMsg
				_, err = http_util.Delete(
					superuserHeader(),
					serverAddress+"/api/v1/function_permission/remove_permission",
					http_util.BlankGetParam, removeReadPostBody, &removePermissionResp)
				So(err, ShouldBeNil)

				// check delete suc
				permissionResp = struct {
					web.RespMsg
					Data function.PermissionResp `json:"data"`
				}{}
				_, err = http_util.Get(
					thisUserLoginedHeader,
					serverAddress+basePath,
					map[string]string{"function_id": fakeAggFunction.ID.String()},
					&permissionResp)
				So(err, ShouldBeNil)
				So(permissionResp.Code, ShouldEqual, http.StatusOK)
				So(permissionResp.Data.Read, ShouldBeFalse)
			})
		})
	})
}
