package function

import (
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"

	"github.com/julienschmidt/httprouter"
)

func GetPermissionByFunctionID(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "get permission of function"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(logTags, "failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil,
			"get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	functionID := r.URL.Query().Get("function_id")
	if functionID == "" {
		fService.Logger.Warningf(
			logTags, "get param miss function_id")
		web.WriteBadRequestDataResp(&w, r, "get param must contain function_id")
		return
	}
	logTags["function_id"] = functionID

	functionUUID, err := value_object.ParseToUUID(functionID)
	if err != nil {
		fService.Logger.Warningf(
			logTags, "parse function_id to uuid failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, "parse function_id to uuid failed")
		return
	}

	aggF, err := fService.Function.GetByID(functionUUID)
	if err != nil {
		fService.Logger.Errorf(
			logTags, "visit function by id failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "get function by function_id error")
		return
	}
	if aggF.IsZero() {
		fService.Logger.Warningf(
			logTags, "visit function by id match no record")
		web.WriteBadRequestDataResp(&w, r, "function_id find no function")
		return
	}

	permsResp := PermissionResp{
		Read:             aggF.UserCanRead(reqUser),
		Execute:          aggF.UserCanExecute(reqUser),
		AssignPermission: aggF.UserCanAssignPermission(reqUser),
	}

	fService.Logger.Infof(logTags, "finished")
	web.WriteSucResp(&w, r, permsResp)
}

func AddUserPermission(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "add permission of function"

	req := buildPermissionReqAndCheck(&w, r, r.Body)
	if req == nil {
		fService.Logger.Warningf(
			logTags, "build permission from body failed")
		return
	}
	logTags["user_id"] = req.UserID.String()
	logTags["function_id"] = req.FunctionID.String()

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
		fService.Logger.Errorf(
			logTags, "add permission failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "add user permission failed")
		return
	}

	fService.Logger.Infof(
		logTags, "finished add permission: %v", req.PermissionType)
	web.WritePlainSucOkResp(&w, r)
}

func DeleteUserPermission(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "delete permission of function"

	req := buildPermissionReqAndCheck(&w, r, r.Body)
	if req == nil {
		fService.Logger.Warningf(
			logTags, "build permission from body failed")
		return
	}
	logTags["user_id"] = req.UserID.String()
	logTags["function_id"] = req.FunctionID.String()

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
		fService.Logger.Errorf(
			logTags, "remove permission failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "remove user permission failed")
		return
	}

	fService.Logger.Infof(
		logTags, "finished delete permission: %v", req.PermissionType)
	web.WritePlainSucOkResp(&w, r)
}
