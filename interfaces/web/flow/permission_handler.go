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
	flowID := r.URL.Query().Get("flow_id")

	if flowID == "" {
		web.WriteBadRequestDataResp(&w, "get param must contain flow_id")
		return
	}
	flowUUID, err := uuid.Parse(flowID)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "parse flow_id to uuid failed")
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

	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil,
			"get requser from context failed")
		return
	}

	permsResp := PermissionResp{
		Read:             aggF.UserCanRead(reqUser),
		Write:            aggF.UserCanWrite(reqUser),
		Execute:          aggF.UserCanExecute(reqUser),
		Delete:           aggF.UserCanDelete(reqUser),
		AssignPermission: aggF.UserCanAssignPermission(reqUser),
	}
	web.WriteSucResp(&w, permsResp)
}

func AddUserPermission(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req PermissionReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "not valid json data："+err.Error())
		return
	}
	if req.FlowID == uuid.Nil {
		web.WriteBadRequestDataResp(&w, "must have flow_id")
		return
	}
	if !req.PermissionType.IsValid() {
		web.WriteBadRequestDataResp(&w, "permission_type not valid")
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
		err = fService.Flow.AddReader(req.FlowID, reqUser.ID)
	} else if req.PermissionType == Write {
		err = fService.Flow.AddWriter(req.FlowID, reqUser.ID)
	} else if req.PermissionType == Execute {
		err = fService.Flow.AddExecuter(req.FlowID, reqUser.ID)
	} else if req.PermissionType == Delete {
		err = fService.Flow.AddDeleter(req.FlowID, reqUser.ID)
	} else if req.PermissionType == AssignPermission {
		err = fService.Flow.AddAssigner(req.FlowID, reqUser.ID)
	}
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "add permission failed")
		return
	}
	web.WritePlainSucOkResp(&w)
}

func DeleteUserPermission(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req PermissionReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "not valid json data："+err.Error())
		return
	}
	if req.FlowID == uuid.Nil {
		web.WriteBadRequestDataResp(&w, "must have flow_id")
		return
	}
	if !req.PermissionType.IsValid() {
		web.WriteBadRequestDataResp(&w, "permission_type not valid")
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
		err = fService.Flow.RemoveReader(req.FlowID, reqUser.ID)
	} else if req.PermissionType == Write {
		err = fService.Flow.RemoveWriter(req.FlowID, reqUser.ID)
	} else if req.PermissionType == Execute {
		err = fService.Flow.RemoveExecuter(req.FlowID, reqUser.ID)
	} else if req.PermissionType == Delete {
		err = fService.Flow.RemoveDeleter(req.FlowID, reqUser.ID)
	} else if req.PermissionType == AssignPermission {
		err = fService.Flow.RemoveAssigner(req.FlowID, reqUser.ID)
	}
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "remove user permission failed")
		return
	}
	web.WritePlainSucOkResp(&w)
}
