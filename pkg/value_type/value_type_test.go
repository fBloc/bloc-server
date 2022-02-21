package value_type

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCheckValueTypeValueValid(t *testing.T) {
	Convey("int check", t, func() {
		So(CheckValueTypeValueValid(IntValueType, 1), ShouldBeTrue)
		So(CheckValueTypeValueValid(IntValueType, "1"), ShouldBeFalse)
		So(CheckValueTypeValueValid(IntValueType, true), ShouldBeFalse)
		So(CheckValueTypeValueValid(IntValueType, []int{1, 2}), ShouldBeTrue)
		So(CheckValueTypeValueValid(IntValueType, []string{"1", "2"}), ShouldBeTrue)
		So(CheckValueTypeValueValid(IntValueType, []string{"1", "2", "xxx"}), ShouldBeFalse)
	})

	Convey("float check", t, func() {
		So(CheckValueTypeValueValid(FloatValueType, 1), ShouldBeTrue)
		So(CheckValueTypeValueValid(FloatValueType, []int{1, 2}), ShouldBeTrue)
		So(CheckValueTypeValueValid(FloatValueType, []float64{1.231, 2}), ShouldBeTrue)
		So(CheckValueTypeValueValid(FloatValueType, []string{"1", "2"}), ShouldBeTrue)
		So(CheckValueTypeValueValid(FloatValueType, []string{"1.2453", "2"}), ShouldBeTrue)
		So(CheckValueTypeValueValid(FloatValueType, []string{"1", "2", "xxx"}), ShouldBeFalse)
	})

	Convey("string check", t, func() {
		So(CheckValueTypeValueValid(StringValueType, 1), ShouldBeFalse)
		So(CheckValueTypeValueValid(StringValueType, 1.234), ShouldBeFalse)
		So(CheckValueTypeValueValid(StringValueType, []int{1, 2}), ShouldBeTrue)
		So(CheckValueTypeValueValid(StringValueType, []float64{1.231, 2}), ShouldBeTrue)
		So(CheckValueTypeValueValid(StringValueType, []string{"1", "2"}), ShouldBeTrue)
		So(CheckValueTypeValueValid(StringValueType, []string{"1.2453", "2"}), ShouldBeTrue)
		So(CheckValueTypeValueValid(StringValueType, []string{"1", "2", "xxx"}), ShouldBeTrue)
	})

	Convey("bool check", t, func() {
		So(CheckValueTypeValueValid(BoolValueType, 1), ShouldBeFalse)
		So(CheckValueTypeValueValid(BoolValueType, 1.234), ShouldBeFalse)
		So(CheckValueTypeValueValid(BoolValueType, "xxx"), ShouldBeFalse)
		So(CheckValueTypeValueValid(BoolValueType, true), ShouldBeTrue)
		So(CheckValueTypeValueValid(BoolValueType, false), ShouldBeTrue)
	})
}
