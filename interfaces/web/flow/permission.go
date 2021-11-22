package flow

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/fBloc/bloc-backend-go/interfaces/web"
	"github.com/fBloc/bloc-backend-go/interfaces/web/req_context"
	"github.com/fBloc/bloc-backend-go/value_object"
)

type FlowPermissionType int

const (
	UnknownPermission FlowPermissionType = iota
	Read
	Write
	Execute
	Delete
	AssignPermission
	maxPermissionType
)

func (fP *FlowPermissionType) IsValid() bool {
	intVal := int(*fP)
	return intVal > int(UnknownPermission) && intVal < int(maxPermissionType)
}

type PermissionReq struct {
	PermissionType FlowPermissionType `json:"permission_type"`
	FlowID         value_object.UUID  `json:"flow_id"`
	UserID         value_object.UUID  `json:"user_id"`
}

func BuildPermissionReqAndCheck(w *http.ResponseWriter, r *http.Request, body io.ReadCloser) *PermissionReq {
	var req PermissionReq
	err := json.NewDecoder(body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(w, "not valid json data："+err.Error())
		return nil
	}
	if req.FlowID.IsNil() {
		web.WriteBadRequestDataResp(w, "must have flow_id")
		return nil
	}
	if !req.PermissionType.IsValid() {
		web.WriteBadRequestDataResp(w, "permission_type not valid")
		return nil
	}

	aggF, err := fService.Flow.GetByID(req.FlowID)
	if err != nil {
		web.WriteInternalServerErrorResp(w, err, "get flow by id error")
		return nil
	}
	if aggF.IsZero() {
		web.WriteBadRequestDataResp(w, "flow_id find no flow")
		return nil
	}

	// 检查当前用户是否对此function有操作添加用户的权限
	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(w, nil,
			"get requser from context failed")
		return nil
	}
	if !aggF.UserCanAssignPermission(reqUser) && !reqUser.IsSuper {
		web.WritePermissionNotEnough(w, "need assign_permission permission")
		return nil
	}
	return &req
}

type PermissionResp struct {
	Read             bool `json:"read"`
	Write            bool `json:"write"`
	Execute          bool `json:"execute"`
	Delete           bool `json:"delete"`
	AssignPermission bool `json:"assign_permission"`
}
