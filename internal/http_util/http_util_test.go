package http_util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/fBloc/bloc-server/interfaces/web"
	. "github.com/smartystreets/goconvey/convey"
)

type testData struct {
	ID   int
	Name string
}

func TestHttpUtil(t *testing.T) {
	var testDataSlice []testData
	idMapName := make(map[int]string)
	for i := 0; i < 10; i++ {
		name := gofakeit.Name()
		testDataSlice = append(testDataSlice, testData{
			ID: i, Name: name,
		})
		idMapName[i] = name
	}

	Convey("test Get", t, func(c C) {
		ts := httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					idInQuery := r.URL.Query().Get("id")
					idInInt, _ := strconv.Atoi(idInQuery)
					for _, i := range testDataSlice {
						if i.ID == idInInt {
							byteData, err := json.Marshal(i)
							c.So(err, ShouldBeNil)
							_, err = w.Write(byteData)
							return
						}
					}
				},
			),
		)
		defer ts.Close()

		searchID := gofakeit.RandomInt([]int{1, 2, 3, 4, 5, 6, 7, 8, 9})
		var resp testData
		_, err := Get(fmt.Sprintf("%s/?id=%d", ts.URL, searchID), BlankHeader, &resp)
		So(err, ShouldBeNil)
		So(resp.Name, ShouldEqual, idMapName[searchID])
		ts.Close()
	})

	Convey("test Post", t, func(c C) {
		ts := httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					var req testData
					err := json.NewDecoder(r.Body).Decode(&req)
					if err != nil {
						web.WriteBadRequestDataResp(&w, err.Error())
						return
					}
					for _, i := range testDataSlice {
						if i.ID == req.ID {
							byteData, err := json.Marshal(i)
							c.So(err, ShouldBeNil)
							_, err = w.Write(byteData)
							return
						}
					}
				},
			),
		)
		defer ts.Close()

		searchID := gofakeit.RandomInt([]int{1, 2, 3, 4, 5, 6, 7, 8, 9})
		req := testData{ID: searchID}
		reqBytes, err := json.Marshal(req)
		So(err, ShouldBeNil)
		var resp testData
		_, err = Post(
			fmt.Sprintf("%s", ts.URL),
			BlankHeader, reqBytes, &resp)
		So(err, ShouldBeNil)
		So(resp.Name, ShouldEqual, idMapName[searchID])
	})

	Convey("test Set/Get header", t, func(c C) {
		headerKey := "header"
		headerValue := "value"
		ts := httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					val := r.Header.Get(headerKey)
					c.So(val, ShouldEqual, headerValue)

					idInQuery := r.URL.Query().Get("id")
					idInInt, _ := strconv.Atoi(idInQuery)
					for _, i := range testDataSlice {
						if i.ID == idInInt {
							byteData, err := json.Marshal(i)
							c.So(err, ShouldBeNil)
							_, err = w.Write(byteData)
							return
						}
					}
				},
			),
		)
		defer ts.Close()

		searchID := gofakeit.RandomInt([]int{1, 2, 3, 4, 5, 6, 7, 8, 9})
		var resp testData
		_, err := Get(
			fmt.Sprintf("%s/?id=%d", ts.URL, searchID),
			map[string]string{headerKey: headerValue},
			&resp)
		So(err, ShouldBeNil)
		So(resp.Name, ShouldEqual, idMapName[searchID])
		ts.Close()
	})
}
