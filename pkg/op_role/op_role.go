package op_role

type OpRole string

const (
	Reader   OpRole = "reader"
	Writer   OpRole = "writer"
	Executer OpRole = "executer"
	Super    OpRole = "super"
)
