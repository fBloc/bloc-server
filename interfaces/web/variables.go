package web

type RequestContextKey string

const (
	RequestContextUserKey RequestContextKey = "req_user"
	RequestContextUUIDKey RequestContextKey = "req_uuid"
)
