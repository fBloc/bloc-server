package user

import (
	"encoding/json"
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"

	"github.com/julienschmidt/httprouter"
)

// LoginHandler 处理用户登录请求
func LoginHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "not valid json request data: %s", err.Error())
		return
	}

	isNameMatchPwd, sameNameUser, err := uService.Login(u.Name, u.RaWPassword)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit user repository failed")
		return
	}
	if !isNameMatchPwd {
		resp := &web.RespMsg{
			Code: http.StatusUnauthorized,
			Msg:  "name - password not match",
		}
		w.Write(resp.JSONBytes())
		return
	}

	LoginRespFromAgg(&w, sameNameUser)
}

// FilterByName
func FilterByName(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	nameContains := r.URL.Query().Get("name__contains")
	users, err := uService.FilterByNameContains(nameContains)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit user repository failed")
		return
	}

	FilterRespFromAggs(&w, users)
}

// AddUser POST添加用户 - 只有superuser才能够添加用户
func AddUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "not valid json request data: %s", err.Error())
		return
	}

	if u.Name == "" || u.RaWPassword == "" {
		web.WriteBadRequestDataResp(&w, "param must contains name & password & is_superuser")
		return
	}

	err = uService.AddUser(u.Name, u.RaWPassword, u.IsSuper)
	if err != nil {
		web.WriteBadRequestDataResp(&w, err.Error())
		return
	}
	web.WritePlainSucOkResp(&w)
}

// DeleteUser Delete删除用户 - 只有superuser才能够删除用户
func DeleteUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	_, err := uService.DeleteUserByIDString(id)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "delete user failed")
		return
	}
	web.WritePlainSucOkResp(&w)
}
