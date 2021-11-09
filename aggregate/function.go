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
	ReadUserIDs    []uuid.UUID
	WriteUserIDs   []uuid.UUID
	ExecuteUserIDs []uuid.UUID
	SuperUserIDs   []uuid.UUID
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
	for _, readAbleUserID := range f.ReadUserIDs {
		if readAbleUserID == userID {
			return true
		}
	}
	return false
}

func (f *Function) UserCanWrite(user *User) bool {
	if user.IsSuper {
		return true
	}
	userID := user.ID
	for _, writeAbleUserID := range f.WriteUserIDs {
		if writeAbleUserID == userID {
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
	for _, exeAbleUserID := range f.ExecuteUserIDs {
		if exeAbleUserID == userID {
			return true
		}
	}
	return false
}

func (f *Function) UserIsSuper(user *User) bool {
	if user.IsSuper {
		return true
	}
	userID := user.ID
	for _, superUserID := range f.SuperUserIDs {
		if superUserID == userID {
			return true
		}
	}
	return false
}
