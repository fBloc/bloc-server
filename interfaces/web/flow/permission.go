package flow

import (
	"github.com/google/uuid"
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
	FlowID         uuid.UUID          `json:"flow_id"`
}

type PermissionResp struct {
	Read             bool `json:"read"`
	Write            bool `json:"write"`
	Execute          bool `json:"execute"`
	Delete           bool `json:"delete"`
	AssignPermission bool `json:"assign_permission"`
}
