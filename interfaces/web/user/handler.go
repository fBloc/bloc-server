package user

import (
	"encoding/json"
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"

	"github.com/julienschmidt/httprouter"
)

// LoginHandler 处理用户登录请求
func LoginHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "login"

	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		uService.Logger.Warningf(
			logTags,
			"json unmarshal to user failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, "not valid json request data: %s", err.Error())
		return
	}

	isNameMatchPwd, sameNameUser, err := uService.Login(u.Name, u.RaWPassword)
	if err != nil {
		uService.Logger.Errorf(
			logTags, "visit user repository failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "visit user repository failed")
		return
	}
	if !isNameMatchPwd {
		uService.Logger.Warningf(logTags,
			"name - password not match. name: %s", u.Name)
		web.WriteBadRequestDataResp(&w, r, "name - password not match")
		return
	}

	uService.Logger.Infof(logTags, "finished")
	LoginRespFromAgg(&w, r, sameNameUser)
}

// Info Loginned user to get its info
func Info(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "get loginned user info"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		uService.Logger.Errorf(logTags, "failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil, "get requser from context failed")
		return
	}

	uService.Logger.Infof(logTags, "finished")
	web.WriteSucResp(&w, r, FromAggToInfo(reqUser))
}

// FilterByName
func FilterByName(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "filter user"

	nameContains := r.URL.Query().Get("name__contains")
	users, err := uService.FilterByNameContains(nameContains)
	if err != nil {
		uService.Logger.Errorf(logTags,
			"filter user failed: %v. name__contains: %s", err, nameContains)
		web.WriteInternalServerErrorResp(&w, r, err, "visit user repository failed")
		return
	}

	uService.Logger.Infof(logTags, "finished")
	FilterRespFromAggs(&w, r, users)
}

// AddUser POST添加用户 - 只有superuser才能够添加用户
func AddUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "add user"

	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		uService.Logger.Warningf(logTags, "json unmarshal to user failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, "not valid json request data: %s", err.Error())
		return
	}
	logTags["user name"] = u.Name

	if u.Name == "" || u.RaWPassword == "" {
		uService.Logger.Warningf(logTags, "name & password both blank")
		web.WriteBadRequestDataResp(&w, r, "param must contains name & password & is_superuser")
		return
	}

	err = uService.AddUser(u.Name, u.RaWPassword, u.IsSuper)
	if err != nil {
		uService.Logger.Errorf(
			logTags, "add user failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, err.Error())
		return
	}

	uService.Logger.Infof(logTags, "finished")
	web.WritePlainSucOkResp(&w, r)
}

// DeleteUser Delete删除用户 - 只有superuser才能够删除用户
func DeleteUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "delete user"

	id := ps.ByName("id")
	logTags["user_id"] = id

	deleteAmount, err := uService.DeleteUserByIDString(id)
	if err != nil {
		uService.Logger.Errorf(logTags, "delete user failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "delete user failed")
		return
	}

	uService.Logger.Infof(logTags,
		"finished with deleted amount: %d", deleteAmount)
	web.WritePlainSucOkResp(&w, r)
}
