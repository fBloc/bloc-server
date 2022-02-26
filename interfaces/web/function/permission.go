package function

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"
)

type FunctionPermissionType int

const (
	UnknownPermission FunctionPermissionType = iota
	Read
	Execute
	AssignPermission
	maxPermissionType
)

func (fP *FunctionPermissionType) IsValid() bool {
	intVal := int(*fP)
	return intVal > int(UnknownPermission) && intVal < int(maxPermissionType)
}

type PermissionReq struct {
	PermissionType FunctionPermissionType `json:"permission_type"`
	FunctionID     value_object.UUID      `json:"function_id"`
	UserID         value_object.UUID      `json:"user_id"`
}

func buildPermissionReqAndCheck(
	w *http.ResponseWriter, r *http.Request, body io.ReadCloser,
) *PermissionReq {
	var req PermissionReq
	err := json.NewDecoder(body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(w, r, "not valid json data: %v", err)
		return nil
	}
	if req.FunctionID.IsNil() {
		web.WriteBadRequestDataResp(w, r, "must have function_id")
		return nil
	}
	if !req.PermissionType.IsValid() {
		web.WriteBadRequestDataResp(w, r, "permission_type not valid")
		return nil
	}

	aggF, err := fService.Function.GetByID(req.FunctionID)
	if err != nil {
		web.WriteInternalServerErrorResp(w, r, err, "get function by id error")
		return nil
	}
	if aggF.IsZero() {
		web.WriteBadRequestDataResp(w, r, "function_id find no function")
		return nil
	}

	// 检查当前用户是否对此function有操作添加用户的权限
	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(w, r, nil,
			"get requser from context failed")
		return nil
	}
	if !aggF.UserCanAssignPermission(reqUser) && !reqUser.IsSuper {
		web.WritePermissionNotEnough(w, r, "need assign_permission permission")
		return nil
	}
	return &req
}

type PermissionResp struct {
	Read             bool `json:"read"`
	Execute          bool `json:"execute"`
	AssignPermission bool `json:"assign_permission"`
}
