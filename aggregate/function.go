package aggregate

import (
	"github.com/fBloc/bloc-backend-go/pkg/function_developer_implement"
	"github.com/fBloc/bloc-backend-go/pkg/ipt"
	"github.com/fBloc/bloc-backend-go/pkg/opt"

	"github.com/google/uuid"
)

type Function struct {
	ID            uuid.UUID
	Name          string
	GroupName     string
	Description   string
	Ipts          ipt.IptSlice
	Opts          []*opt.Opt
	IptDigest     string
	OptDigest     string
	ProcessStages []string
	ExeFunc       function_developer_implement.FunctionDeveloperImplementInterface
	// 用于权限
	ReadUserIDs             []uuid.UUID
	ExecuteUserIDs          []uuid.UUID
	AssignPermissionUserIDs []uuid.UUID
}

func (f *Function) CoreString() string {
	return ipt.IptString(f.Ipts) + opt.OptString(f.Opts)
}

func (f *Function) IsZero() bool {
	if f == nil {
		return true
	}
	return f.ID == uuid.Nil
}

func (f *Function) String() string {
	return f.Name + f.GroupName + f.Description + f.CoreString()
}

func (f *Function) UserCanRead(user *User) bool {
	if user.IsSuper {
		return true
	}
	userID := user.ID
	for _, uID := range f.ReadUserIDs {
		if uID == userID {
			return true
		}
	}
	return false
}

func (f *Function) UserCanExecute(user *User) bool {
	if user.IsSuper {
		return true
	}
	userID := user.ID
	for _, uID := range f.ExecuteUserIDs {
		if uID == userID {
			return true
		}
	}
	return false
}

func (f *Function) UserCanAssignPermission(user *User) bool {
	if user.IsSuper {
		return true
	}
	userID := user.ID
	for _, uID := range f.AssignPermissionUserIDs {
		if uID == userID {
			return true
		}
	}
	return false
}
