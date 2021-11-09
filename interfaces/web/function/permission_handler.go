package function

import (
	"encoding/json"
	"net/http"

	"github.com/fBloc/bloc-backend-go/interfaces/web"
	"github.com/fBloc/bloc-backend-go/interfaces/web/req_context"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

func GetPermissionByFunctionID(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID := r.URL.Query().Get("user_token")
	functionID := r.URL.Query().Get("function_id")

	if userID == "" || functionID == "" {
		web.WriteBadRequestDataResp(&w, "get param must contain both user_token & function_id")
		return
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "parse user_token to uuid failed")
		return
	}
	functionUUID, err := uuid.Parse(functionID)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "parse function_id to uuid failed")
		return
	}

	aggUser, err := fService.UserCacheService.GetUserByID(userUUID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "get user by user_token error")
		return
	}
	if aggUser.IsZero() {
		web.WriteBadRequestDataResp(&w, "user_token find no user")
		return
	}

	aggF, err := fService.Function.GetByID(functionUUID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "get function by function_id error")
		return
	}
	if aggF.IsZero() {
		web.WriteBadRequestDataResp(&w, "function_id find no function")
		return
	}

	permsResp := PermissionResp{
		Read:    aggF.UserCanRead(&aggUser),
		Write:   aggF.UserCanWrite(&aggUser),
		Execute: aggF.UserCanExecute(&aggUser),
		Super:   aggF.UserIsSuper(&aggUser),
	}
	web.WriteSucResp(&w, permsResp)
}

func AddUserPermission(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	who := ps.ByName("who")

	var req PermissionReq
	var err error
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "not valid json data："+err.Error())
		return
	}
	if req.FunctionID == uuid.Nil || req.UserID == uuid.Nil {
		web.WriteBadRequestDataResp(&w, "must have both function_id & user_token")
		return
	}

	aggF, err := fService.Function.GetByID(req.FunctionID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "get function by id error")
		return
	}
	if aggF.IsZero() {
		web.WriteBadRequestDataResp(&w, "function_id find no function")
		return
	}

	// 检查当前用户是否对此bloc有操作添加用户的权限
	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil,
			"get requser from context failed")
		return
	}
	if !aggF.UserIsSuper(reqUser) {
		web.WritePermissionNotEnough(&w, "need super permission")
		return
	}

	// 开始实际更新数据
	if who == "read" {
		err = fService.Function.AddReader(req.FunctionID, req.UserID)
	} else if who == "write" {
		err = fService.Function.AddWriter(req.FunctionID, req.UserID)
	} else if who == "execute" {
		err = fService.Function.AddExecuter(req.FunctionID, req.UserID)
	} else if who == "super" {
		err = fService.Function.AddSuper(req.FunctionID, req.UserID)
	} else {
		web.WriteBadRequestDataResp(&w, "last get path should be in read/write/execute/super")
		return
	}
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "update user permission failed")
		return
	}
	web.WritePlainSucOkResp(&w)
}

func DeleteUserPermission(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	who := ps.ByName("who")

	var req PermissionReq
	var err error
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "not valid json data："+err.Error())
		return
	}
	if req.FunctionID == uuid.Nil || req.UserID == uuid.Nil {
		web.WriteBadRequestDataResp(&w, "must have both function_id & user_token")
		return
	}

	aggF, err := fService.Function.GetByID(req.FunctionID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "get function by id error")
		return
	}
	if aggF.IsZero() {
		web.WriteBadRequestDataResp(&w, "function_id find no function")
		return
	}

	// 检查当前用户是否对此bloc有操作添加用户的权限
	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil,
			"get requser from context failed")
		return
	}
	if !aggF.UserIsSuper(reqUser) {
		web.WritePermissionNotEnough(&w, "need super permission")
		return
	}

	// 开始实际更新数据
	if who == "read" {
		err = fService.Function.DeleteReader(req.FunctionID, req.UserID)
	} else if who == "write" {
		err = fService.Function.DeleteWriter(req.FunctionID, req.UserID)
	} else if who == "execute" {
		err = fService.Function.DeleteExecuter(req.FunctionID, req.UserID)
	} else if who == "super" {
		err = fService.Function.DeleteSuper(req.FunctionID, req.UserID)
	} else {
		web.WriteBadRequestDataResp(&w, "last path should be in read/write/execute/super")
		return
	}
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "update user permission failed")
		return
	}
	web.WritePlainSucOkResp(&w)
}
