package flow

import (
	"encoding/json"
	"net/http"

	"github.com/fBloc/bloc-backend-go/interfaces/web"
	"github.com/fBloc/bloc-backend-go/interfaces/web/req_context"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

func GetPermission(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID := r.URL.Query().Get("user_token")
	flowID := r.URL.Query().Get("flow_id")

	if userID == "" || flowID == "" {
		web.WriteBadRequestDataResp(&w, "get param must contain both user_token & flow_id")
		return
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "parse user_token to uuid failed")
		return
	}
	flowUUID, err := uuid.Parse(flowID)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "parse flow_id to uuid failed")
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

	aggF, err := fService.Flow.GetByID(flowUUID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "get flow by flow_id error")
		return
	}
	if aggF.IsZero() {
		web.WriteBadRequestDataResp(&w, "flow_id find no flow")
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
	if req.FlowID == uuid.Nil || req.UserID == uuid.Nil {
		web.WriteBadRequestDataResp(&w, "must have both flow_id & user_token")
		return
	}

	aggF, err := fService.Flow.GetByID(req.FlowID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "get flow by id error")
		return
	}
	if aggF.IsZero() {
		web.WriteBadRequestDataResp(&w, "flow_id find no flow")
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
		err = fService.Flow.AddReader(req.FlowID, req.UserID)
	} else if who == "write" {
		err = fService.Flow.AddWriter(req.FlowID, req.UserID)
	} else if who == "execute" {
		err = fService.Flow.AddExecuter(req.FlowID, req.UserID)
	} else if who == "super" {
		err = fService.Flow.AddSuper(req.FlowID, req.UserID)
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
	if req.FlowID == uuid.Nil || req.UserID == uuid.Nil {
		web.WriteBadRequestDataResp(&w, "must have both flow_id & user_token")
		return
	}

	aggF, err := fService.Flow.GetByID(req.FlowID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "get flow by id error")
		return
	}
	if aggF.IsZero() {
		web.WriteBadRequestDataResp(&w, "flow_id find no flow")
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
		err = fService.Flow.DeleteReader(req.FlowID, req.UserID)
	} else if who == "write" {
		err = fService.Flow.DeleteWriter(req.FlowID, req.UserID)
	} else if who == "execute" {
		err = fService.Flow.DeleteExecuter(req.FlowID, req.UserID)
	} else if who == "super" {
		err = fService.Flow.DeleteSuper(req.FlowID, req.UserID)
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
