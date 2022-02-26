package flow

import (
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"

	"github.com/julienschmidt/httprouter"
)

func GetPermission(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "get permission of flow"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(
			logTags,
			"failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil,
			"get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	flowID := r.URL.Query().Get("flow_id")
	if flowID == "" {
		fService.Logger.Warningf(logTags, "lack get param flow_id")
		web.WriteBadRequestDataResp(&w, r, "get param must contain flow_id")
		return
	}
	logTags["flow_id"] = flowID
	flowUUID, err := value_object.ParseToUUID(flowID)
	if err != nil {
		fService.Logger.Warningf(logTags, "parse flow_id to uuid failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, "parse flow_id to uuid failed")
		return
	}

	aggF, err := fService.Flow.GetByID(flowUUID)
	if err != nil {
		fService.Logger.Errorf(logTags, "get flow by id failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "get flow by flow_id error")
		return
	}
	if aggF.IsZero() {
		fService.Logger.Warningf(logTags, "get flow by id match no record")
		web.WriteBadRequestDataResp(&w, r, "flow_id find no flow")
		return
	}

	permsResp := PermissionResp{
		Read:             aggF.UserCanRead(reqUser),
		Write:            aggF.UserCanWrite(reqUser),
		Execute:          aggF.UserCanExecute(reqUser),
		Delete:           aggF.UserCanDelete(reqUser),
		AssignPermission: aggF.UserCanAssignPermission(reqUser),
	}

	fService.Logger.Infof(logTags, "finished")
	web.WriteSucResp(&w, r, permsResp)
}

func AddUserPermission(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "get permission of flow"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(
			logTags,
			"failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil,
			"get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	req := BuildPermissionReqAndCheck(&w, r, r.Body)
	if req == nil {
		fService.Logger.Warningf(logTags, "build req failed")
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
		fService.Logger.Errorf(logTags, "add permission failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "add permission failed")
		return
	}
	fService.Logger.Infof(
		logTags, "suc add permission: %s", req.PermissionType.String())
	web.WritePlainSucOkResp(&w, r)
}

func DeleteUserPermission(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "delete permission of flow"

	req := BuildPermissionReqAndCheck(&w, r, r.Body)
	if req == nil {
		fService.Logger.Warningf(
			logTags, "build req from body failed")
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
		fService.Logger.Errorf(
			logTags,
			"remove user permission failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "remove user permission failed")
		return
	}

	fService.Logger.Infof(
		logTags, "suc remove permission: %s", req.PermissionType.String())
	web.WritePlainSucOkResp(&w, r)
}
