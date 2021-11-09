package flow

import "github.com/google/uuid"

type PermissionReq struct {
	FlowID uuid.UUID `json:"flow_id"`
	UserID uuid.UUID `json:"user_token"`
}

type PermissionResp struct {
	Read    bool `json:"read"`
	Write   bool `json:"write"`
	Execute bool `json:"execute"`
	Super   bool `json:"super"`
}
