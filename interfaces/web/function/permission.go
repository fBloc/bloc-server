package function

import "github.com/google/uuid"

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
	FunctionID     uuid.UUID              `json:"function_id"`
}

type PermissionResp struct {
	Read             bool `json:"read"`
	Execute          bool `json:"execute"`
	AssignPermission bool `json:"assign_permission"`
}
