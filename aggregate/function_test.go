package aggregate

import (
	"testing"

	"github.com/fBloc/bloc-server/pkg/ipt"
	"github.com/fBloc/bloc-server/pkg/opt"
	"github.com/fBloc/bloc-server/pkg/value_type"
	"github.com/fBloc/bloc-server/value_object"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	funcID       = value_object.NewUUID()
	fakeFunction = Function{
		ID:           funcID,
		Name:         "two add",
		GroupName:    "math operation",
		ProviderName: "test",
		Description:  "test function",
		Ipts: ipt.IptSlice{
			{
				Key:     "to_add_ints",
				Display: "to_add_ints",
				Must:    true,
				Components: []*ipt.IptComponent{
					{
						ValueType:       value_type.IntValueType,
						FormControlType: value_object.InputFormControl,
						Hint:            "加数",
						DefaultValue:    0,
						AllowMulti:      true,
					},
				},
			},
		},
		Opts: []*opt.Opt{
			{
				Key:         "sum",
				Description: "sum of your inputs",
				ValueType:   value_type.IntValueType,
				IsArray:     false,
			},
		},
		ProcessStages:           []string{"parsing ipt", "finished parse ipt & start do the math", "finished"},
		ReadUserIDs:             []value_object.UUID{readeUser.ID, allPermissionUser.ID},
		ExecuteUserIDs:          []value_object.UUID{executeUser.ID, allPermissionUser.ID},
		AssignPermissionUserIDs: []value_object.UUID{allPermissionUser.ID},
	}
)

func TestFunctionString(t *testing.T) {
	Convey("string", t, func() {
		funcStr := fakeFunction.String()
		So(funcStr, ShouldNotEqual, "")
	})
}

func TestFunctionIsZero(t *testing.T) {
	Convey("zero", t, func() {
		var nilFunc *Function
		So(nilFunc.IsZero(), ShouldBeTrue)
		nilFunc = &Function{}
		So(nilFunc.IsZero(), ShouldBeTrue)
	})

	Convey("not zero", t, func() {
		So(fakeFunction.IsZero(), ShouldBeFalse)
	})
}

func TestFunctionUserCanRead(t *testing.T) {
	Convey("User cannot read", t, func() {
		So(fakeFunction.UserCanRead(&executeUser), ShouldBeFalse)
	})

	Convey("User can read", t, func() {
		So(fakeFunction.UserCanRead(&readeUser), ShouldBeTrue)
		So(fakeFunction.UserCanRead(&allPermissionUser), ShouldBeTrue)
		So(fakeFunction.UserCanRead(&superUser), ShouldBeTrue)
	})
}

func TestFunctionUserCanExecute(t *testing.T) {
	Convey("User cannot execute", t, func() {
		So(fakeFunction.UserCanExecute(&readeUser), ShouldBeFalse)
	})

	Convey("User can execute", t, func() {
		So(fakeFunction.UserCanExecute(&executeUser), ShouldBeTrue)
		So(fakeFunction.UserCanExecute(&allPermissionUser), ShouldBeTrue)
		So(fakeFunction.UserCanExecute(&superUser), ShouldBeTrue)
	})
}

func TestFunctionUserCanAssignPermission(t *testing.T) {
	Convey("User cannot assign permission", t, func() {
		So(fakeFunction.UserCanAssignPermission(&readeUser), ShouldBeFalse)
		So(fakeFunction.UserCanAssignPermission(&executeUser), ShouldBeFalse)
	})

	Convey("User can assign permission", t, func() {
		So(fakeFunction.UserCanAssignPermission(&allPermissionUser), ShouldBeTrue)
		So(fakeFunction.UserCanAssignPermission(&superUser), ShouldBeTrue)
	})
}

func TestFunctionOptKeyMapValueType(t *testing.T) {
	Convey("OptKeyMapValueType", t, func() {
		optKeyMapValueType := fakeFunction.OptKeyMapValueType()
		So(optKeyMapValueType["sum"], ShouldEqual, value_type.IntValueType)
	})
}

func TestFunctionOptKeyMapIsArray(t *testing.T) {
	Convey("OptKeyMapIsArray", t, func() {
		optKeyMapIsArray := fakeFunction.OptKeyMapIsArray()
		So(optKeyMapIsArray["sum"], ShouldEqual, false)
	})
}
