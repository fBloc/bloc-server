package function

import (
	"net/http"

	"github.com/fBloc/bloc-backend-go/interfaces/web"
	"github.com/fBloc/bloc-backend-go/interfaces/web/req_context"
	"github.com/fBloc/bloc-backend-go/value_object"

	"github.com/julienschmidt/httprouter"
)

func GetPermissionByFunctionID(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	functionID := r.URL.Query().Get("function_id")
	if functionID == "" {
		web.WriteBadRequestDataResp(&w, "get param must contain function_id")
		return
	}
	functionUUID, err := value_object.ParseToUUID(functionID)
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
	req := BuildPermissionReqAndCheck(&w, r, r.Body)
	if req == nil {
		return
	}

	// 开始实际更新数据
	var err error
	if req.PermissionType == Read {
		err = fService.Function.AddReader(req.FunctionID, req.UserID)
	} else if req.PermissionType == Execute {
		err = fService.Function.AddExecuter(req.FunctionID, req.UserID)
	} else if req.PermissionType == AssignPermission {
		err = fService.Function.AddAssigner(req.FunctionID, req.UserID)
	}
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "add user permission failed")
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
		err = fService.Function.RemoveReader(req.FunctionID, req.UserID)
	} else if req.PermissionType == Execute {
		err = fService.Function.RemoveExecuter(req.FunctionID, req.UserID)
	} else if req.PermissionType == AssignPermission {
		err = fService.Function.RemoveAssigner(req.FunctionID, req.UserID)
	}
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "remove user permission failed")
		return
	}
	web.WritePlainSucOkResp(&w)
}
