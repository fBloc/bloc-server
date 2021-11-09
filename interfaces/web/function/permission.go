package function

import "github.com/google/uuid"

type PermissionReq struct {
	FunctionID uuid.UUID `json:"function_id"`
	UserID     uuid.UUID `json:"user_token"`
}

type PermissionResp struct {
	Read    bool `json:"read"`
	Write   bool `json:"write"`
	Execute bool `json:"execute"`
	Super   bool `json:"super"`
}
