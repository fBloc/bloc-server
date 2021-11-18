package value_object

type PermissionType int

const (
	Read PermissionType = iota + 1
	Write
	Execute
	Delete
	AssignPermission
	Super
)
