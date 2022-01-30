package aggregate

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewUser(t *testing.T) {
	Convey("new user should fail", t, func() {
		_, err := NewUser(
			"", gofakeit.Password(false, false, false, false, false, 16), false)
		So(err, ShouldNotBeNil)

		_, err = NewUser(gofakeit.Name(), "", false)
		So(err, ShouldNotBeNil)
	})

	Convey("new user", t, func() {
		u, err := NewUser(
			gofakeit.Name(),
			gofakeit.Password(false, false, false, false, false, 16),
			false)
		So(err, ShouldBeNil)
		So(u, ShouldNotBeNil)
	})
}

func TestChangeSalt(t *testing.T) {
	Convey("ChangeSalt", t, func() {
		sameRawPasswd := gofakeit.Password(false, false, false, false, false, 16)

		uA, _ := NewUser(gofakeit.Name(), sameRawPasswd, false)
		uB, _ := NewUser(gofakeit.Name(), sameRawPasswd, false)
		ChangeSalt("pdtonyshaweb")
		uC, _ := NewUser(gofakeit.Name(), sameRawPasswd, false)
		So(uA.RawPassword, ShouldEqual, uB.RawPassword)
		So(uA.RawPassword, ShouldEqual, uC.RawPassword)
		So(uB.RawPassword, ShouldEqual, uB.RawPassword)

		So(uA.Password, ShouldEqual, uB.Password)
		So(uA.Password, ShouldNotEqual, uC.Password)
	})
}

func TestUserIsZero(t *testing.T) {
	Convey("zero check", t, func() {
		u, _ := NewUser(
			"", gofakeit.Password(false, false, false, false, false, 16), false)
		So(u.IsZero(), ShouldBeTrue)

		u = &User{}
		So(u.IsZero(), ShouldBeTrue)

		u, _ = NewUser(
			gofakeit.Name(),
			gofakeit.Password(false, false, false, false, false, 16),
			false)
		So(u.IsZero(), ShouldBeFalse)
	})
}

func TestUserIsRawPasswordMatch(t *testing.T) {
	Convey("user password match check of nil user", t, func() {
		var user *User
		passwdMatch, err := user.IsRawPasswordMatch("whatever")
		So(err, ShouldNotBeNil)
		So(passwdMatch, ShouldBeFalse)
	})

	Convey("user password match check", t, func() {
		u, _ := NewUser(
			gofakeit.Name(),
			gofakeit.Password(false, false, false, false, false, 16),
			false)
		passwdMatch, err := u.IsRawPasswordMatch(u.RawPassword)
		So(err, ShouldBeNil)
		So(passwdMatch, ShouldBeTrue)
	})

	Convey("user password match check missmatch", t, func() {
		u, _ := NewUser(
			gofakeit.Name(),
			gofakeit.Password(false, false, false, false, false, 16),
			false)
		passwdMatch, err := u.IsRawPasswordMatch(u.RawPassword + "miss")
		So(err, ShouldBeNil)
		So(passwdMatch, ShouldBeFalse)
	})
}
