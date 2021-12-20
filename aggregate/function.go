package aggregate

import (
	"time"

	"github.com/fBloc/bloc-server/pkg/function_developer_implement"
	"github.com/fBloc/bloc-server/pkg/ipt"
	"github.com/fBloc/bloc-server/pkg/opt"
	"github.com/fBloc/bloc-server/pkg/value_type"
	"github.com/fBloc/bloc-server/value_object"
)

type Function struct {
	ID            value_object.UUID
	Name          string
	GroupName     string
	ProviderName  string
	Description   string
	Ipts          ipt.IptSlice
	Opts          []*opt.Opt
	IptDigest     string
	OptDigest     string
	ProcessStages []string
	ExeFunc       function_developer_implement.FunctionDeveloperImplementInterface
	// heartbeat
	LastAliveTime time.Time
	// 用于权限
	ReadUserIDs             []value_object.UUID
	ExecuteUserIDs          []value_object.UUID
	AssignPermissionUserIDs []value_object.UUID
}

func (f *Function) CoreString() string {
	return ipt.IptString(f.Ipts) + opt.OptString(f.Opts)
}

func (f *Function) IsZero() bool {
	if f == nil {
		return true
	}
	return f.ID.IsNil()
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

func (f *Function) OptKeyMapValueType() map[string]value_type.ValueType {
	keyMapValueType := make(map[string]value_type.ValueType, len(f.Opts))
	for _, i := range f.Opts {
		keyMapValueType[i.Key] = i.ValueType
	}
	return keyMapValueType
}

func (f *Function) OptKeyMapIsArray() map[string]bool {
	keyMapIsArray := make(map[string]bool, len(f.Opts))
	for _, i := range f.Opts {
		keyMapIsArray[i.Key] = i.IsArray
	}
	return keyMapIsArray
}
