package util

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	. "github.com/smartystreets/goconvey/convey"
)

type testDecodeMapToStructStruct struct {
	Name     string    `mapstructure:"name"`
	Age      int       `mapstructure:"age"`
	Money    float64   `mapstructure:"money"`
	BirthDay time.Time `mapstructure:"birth_day"`
	JoinTime time.Time `mapstructure:"join_time"`
}

func TestDecodeMapToStructP(t *testing.T) {
	Convey("TestDecodeMapToStructP", t, func() {
		name := gofakeit.Name()
		birth := time.Now().Add(-10 * 24 * 365 * time.Hour)
		joinTime := time.Now().Add(-1 * 24 * 265 * time.Hour)
		age := 24
		money := 108.24
		mapData := map[string]interface{}{
			"name":      name,
			"age":       age,
			"money":     money,
			"birth_day": birth,
			"join_time": joinTime.Format(time.RFC3339),
		}

		var structData testDecodeMapToStructStruct
		err := DecodeMapToStructP(mapData, &structData)
		So(err, ShouldBeNil)
		So(structData.Name, ShouldEqual, name)
		So(structData.BirthDay, ShouldEqual, birth)
		So(structData.Age, ShouldEqual, age)
		So(structData.Money, ShouldEqual, money)
		So(structData.JoinTime.Unix(), ShouldEqual, joinTime.Unix())
	})
}
