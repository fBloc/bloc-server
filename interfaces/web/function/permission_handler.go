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
	functionID := r.URL.Query().Get("function_id")
	if functionID == "" {
		web.WriteBadRequestDataResp(&w, "get param must contain function_id")
		return
	}
	functionUUID, err := uuid.Parse(functionID)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "parse function_id to uuid failed")
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

	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil,
			"get requser from context failed")
		return
	}

	permsResp := PermissionResp{
		Read:             aggF.UserCanRead(reqUser),
		Execute:          aggF.UserCanExecute(reqUser),
		AssignPermission: aggF.UserCanAssignPermission(reqUser),
	}
	web.WriteSucResp(&w, permsResp)
}

func AddUserPermission(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req PermissionReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "not valid json data："+err.Error())
		return
	}
	if req.FunctionID == uuid.Nil {
		web.WriteBadRequestDataResp(&w, "function_id cannot be blank")
		return
	}
	if !req.PermissionType.IsValid() {
		web.WriteBadRequestDataResp(&w, "permission_type not valid")
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

	// 检查当前用户是否对此function有操作添加用户的权限
	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil,
			"get requser from context failed")
		return
	}
	if !aggF.UserCanAssignPermission(reqUser) {
		web.WritePermissionNotEnough(&w, "need assign_permission permission")
		return
	}

	// 开始实际更新数据
	if req.PermissionType == Read {
		err = fService.Function.AddReader(req.FunctionID, reqUser.ID)
	} else if req.PermissionType == Execute {
		err = fService.Function.AddExecuter(req.FunctionID, reqUser.ID)
	} else if req.PermissionType == AssignPermission {
		err = fService.Function.AddAssigner(req.FunctionID, reqUser.ID)
	}
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "add user permission failed")
		return
	}
	web.WritePlainSucOkResp(&w)
}

func DeleteUserPermission(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req PermissionReq
	var err error
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "not valid json data："+err.Error())
		return
	}
	if req.FunctionID == uuid.Nil {
		web.WriteBadRequestDataResp(&w, "function_id cannot be blank")
		return
	}
	if !req.PermissionType.IsValid() {
		web.WriteBadRequestDataResp(&w, "permission_type not valid")
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

	// 检查当前用户是否对此function有操作添加用户的权限
	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil,
			"get requser from context failed")
		return
	}
	if !aggF.UserCanAssignPermission(reqUser) {
		web.WritePermissionNotEnough(&w, "need assign_permission permission")
		return
	}

	// 开始实际更新数据
	if req.PermissionType == Read {
		err = fService.Function.RemoveReader(req.FunctionID, reqUser.ID)
	} else if req.PermissionType == Execute {
		err = fService.Function.RemoveExecuter(req.FunctionID, reqUser.ID)
	} else if req.PermissionType == AssignPermission {
		err = fService.Function.RemoveAssigner(req.FunctionID, reqUser.ID)
	}
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "remove user permission failed")
		return
	}
	web.WritePlainSucOkResp(&w)
}
