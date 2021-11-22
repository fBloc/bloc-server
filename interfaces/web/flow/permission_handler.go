package flow

import (
	"net/http"

	"github.com/fBloc/bloc-backend-go/interfaces/web"
	"github.com/fBloc/bloc-backend-go/interfaces/web/req_context"
	"github.com/fBloc/bloc-backend-go/value_object"

	"github.com/julienschmidt/httprouter"
)

func GetPermission(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	flowID := r.URL.Query().Get("flow_id")

	if flowID == "" {
		web.WriteBadRequestDataResp(&w, "get param must contain flow_id")
		return
	}
	flowUUID, err := value_object.ParseToUUID(flowID)
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
	req := BuildPermissionReqAndCheck(&w, r, r.Body)
	if req == nil {
		return
	}

	// 开始实际更新数据
	var err error
	if req.PermissionType == Read {
		err = fService.Flow.AddReader(req.FlowID, req.UserID)
	} else if req.PermissionType == Write {
		err = fService.Flow.AddWriter(req.FlowID, req.UserID)
	} else if req.PermissionType == Execute {
		err = fService.Flow.AddExecuter(req.FlowID, req.UserID)
	} else if req.PermissionType == Delete {
		err = fService.Flow.AddDeleter(req.FlowID, req.UserID)
	} else if req.PermissionType == AssignPermission {
		err = fService.Flow.AddAssigner(req.FlowID, req.UserID)
	}
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "add permission failed")
		return
	}
	web.WritePlainSucOkResp(&w)
}

func DeleteUserPermission(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	req := BuildPermissionReqAndCheck(&w, r, r.Body)
	if req == nil {
		return
	}

	// 开始实际更新数据
	var err error
	if req.PermissionType == Read {
		err = fService.Flow.RemoveReader(req.FlowID, req.UserID)
	} else if req.PermissionType == Write {
		err = fService.Flow.RemoveWriter(req.FlowID, req.UserID)
	} else if req.PermissionType == Execute {
		err = fService.Flow.RemoveExecuter(req.FlowID, req.UserID)
	} else if req.PermissionType == Delete {
		err = fService.Flow.RemoveDeleter(req.FlowID, req.UserID)
	} else if req.PermissionType == AssignPermission {
		err = fService.Flow.RemoveAssigner(req.FlowID, req.UserID)
	}
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "remove user permission failed")
		return
	}
	web.WritePlainSucOkResp(&w)
}
