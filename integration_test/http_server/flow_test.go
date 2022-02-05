package http_server

import (
	"encoding/json"
	"log"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/interfaces/web/client"
	"github.com/fBloc/bloc-server/interfaces/web/flow"
	"github.com/fBloc/bloc-server/internal/http_util"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNeedLogin(t *testing.T) {
	SkipConvey("need login to do all operations", t, func() {
		// TODO
	})
}

func TestFilterBeforeCreateFlow(t *testing.T) {
	Convey("before create filter should return blank", t, func() {
		resp := struct {
			web.RespMsg
			Data []*flow.Flow `json:"data"`
		}{}
		Convey("filter draft flow", func() {
			_, err := http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/draft_flow",
				http_util.BlankGetParam,
				&resp)
			So(err, ShouldBeNil)
			So(len(resp.Data), ShouldEqual, 0)
		})

		Convey("filter online flow", func() {
			_, err := http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/flow",
				http_util.BlankGetParam,
				&resp)
			So(err, ShouldBeNil)
			So(len(resp.Data), ShouldEqual, 0)
		})
	})
}

func TestDraftFlow(t *testing.T) {
	// register the two functions
	registerFunction := client.RegisterFuncReq{
		Who: fakeAggFunction.ProviderName,
		GroupNameMapFunctions: map[string][]*client.HttpFunction{
			fakeAggFunction.GroupName: []*client.HttpFunction{
				{
					Name:          aggFuncAdd.Name,
					GroupName:     aggFuncAdd.GroupName,
					Description:   aggFuncAdd.Description,
					Ipts:          aggFuncAdd.Ipts,
					Opts:          aggFuncAdd.Opts,
					ProcessStages: aggFuncAdd.ProcessStages,
				},
				{
					Name:          aggFuncMultiply.Name,
					GroupName:     aggFuncMultiply.GroupName,
					Description:   aggFuncMultiply.Description,
					Ipts:          aggFuncMultiply.Ipts,
					Opts:          aggFuncMultiply.Opts,
					ProcessStages: aggFuncMultiply.ProcessStages,
				},
			},
		},
	}
	registerFunctionBody, _ := json.Marshal(registerFunction)
	registerFunctionResp := struct {
		web.RespMsg
		Data client.RegisterFuncReq `json:"data"`
	}{}
	_, err := http_util.Post(
		http_util.BlankHeader,
		serverAddress+"/api/v1/client/register_functions",
		http_util.BlankGetParam, registerFunctionBody, &registerFunctionResp)
	if err != nil {
		log.Fatalf("register function error: %v", err)
	}
	if registerFunctionResp.Code != http.StatusOK {
		log.Fatalf("register function failed: %v", registerFunctionResp)
	}
	for _, function := range registerFunctionResp.Data.GroupNameMapFunctions[fakeAggFunction.GroupName] {
		if function.Name == aggFuncAdd.Name {
			aggFuncAdd.ID = function.ID
		} else if function.Name == aggFuncMultiply.Name {
			aggFuncMultiply.ID = function.ID
		}
	}

	Convey("create draft flow", t, func() {
		aggFlow := getFakeAggFlow()
		reqFlow := flow.FromAggWithoutUserPermission(aggFlow)
		reqBody, _ := json.Marshal(reqFlow)
		resp := struct {
			web.RespMsg
			Data *flow.Flow `json:"data"`
		}{}
		_, err := http_util.Post(
			superuserHeader(),
			serverAddress+"/api/v1/draft_flow",
			http_util.BlankGetParam,
			reqBody, &resp)
		So(err, ShouldBeNil)
		So(resp.Data.IsZero(), ShouldBeFalse)
		So(resp.Data.ID.IsNil(), ShouldBeFalse)
		So(resp.Data.OriginID.IsNil(), ShouldBeFalse)
		createdFlow := resp.Data
		createdFlowOringinID := resp.Data.OriginID

		Convey("get by origin_id draft flow", func() {
			Convey("get should miss", func() {
				resp := struct {
					web.RespMsg
					Data *flow.Flow `json:"data"`
				}{}
				_, err := http_util.Get(
					superuserHeader(),
					serverAddress+"/api/v1/draft_flow/get_by_origin_id/"+createdFlowOringinID.String()+"miss",
					http_util.BlankGetParam, &resp)
				So(err, ShouldBeNil)
				So(resp.Data.IsZero(), ShouldBeTrue)
			})

			Convey("get should hit", func() {
				resp := struct {
					web.RespMsg
					Data *flow.Flow `json:"data"`
				}{}
				_, err := http_util.Get(
					superuserHeader(),
					serverAddress+"/api/v1/draft_flow/get_by_origin_id/"+createdFlowOringinID.String(),
					http_util.BlankGetParam, &resp)
				So(err, ShouldBeNil)
				So(resp.Data.IsZero(), ShouldBeFalse)
			})
		})

		Convey("patch draft flow", func() {
			Convey("patch name", func() {
				var resp web.RespMsg
				patchFlow := createdFlow
				patchFlow.Name = gofakeit.Name()
				patchBody, _ := json.Marshal(patchFlow)
				_, err := http_util.Patch(
					superuserHeader(), serverAddress+"/api/v1/draft_flow",
					http_util.BlankGetParam, patchBody, &resp)
				So(err, ShouldBeNil)
				So(resp.Code, ShouldEqual, http.StatusOK)

				regetResp := struct {
					web.RespMsg
					Data *flow.Flow `json:"data"`
				}{}
				http_util.Get(
					superuserHeader(),
					serverAddress+"/api/v1/draft_flow/get_by_origin_id/"+createdFlowOringinID.String(),
					http_util.BlankGetParam, &regetResp)
				So(regetResp.Data.Name, ShouldEqual, patchFlow.Name)
			})
		})

		Convey("delete draft flow", func() {
			// before delete should have 1 flow
			filterResp := struct {
				web.RespMsg
				Data []*flow.Flow `json:"data"`
			}{}
			_, err := http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/draft_flow",
				http_util.BlankGetParam,
				&filterResp)
			So(err, ShouldBeNil)
			So(len(filterResp.Data), ShouldEqual, 1)

			// delete the flow
			resp := struct {
				web.RespMsg
				Data map[string]int64 `json:"data"`
			}{}
			_, err = http_util.Delete(
				superuserHeader(),
				serverAddress+"/api/v1/draft_flow/delete_by_origin_id/"+createdFlowOringinID.String(),
				http_util.BlankGetParam, http_util.BlankBody, &resp)
			So(err, ShouldBeNil)
			So(resp.Code, ShouldEqual, http.StatusOK)
			So(resp.Data["delete_amount"], ShouldEqual, 1)

			// after delete should have 0 flow
			filterResp = struct {
				web.RespMsg
				Data []*flow.Flow `json:"data"`
			}{}
			_, err = http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/draft_flow",
				http_util.BlankGetParam,
				&filterResp)
			So(err, ShouldBeNil)
			So(len(filterResp.Data), ShouldEqual, 0)
		})

		Reset(func() {
			// delete the created draft flow
			var resp interface{}
			http_util.Delete(
				superuserHeader(),
				serverAddress+"/api/v1/draft_flow/delete_by_origin_id/"+createdFlowOringinID.String(),
				http_util.BlankGetParam, http_util.BlankBody, &resp)
		})
	})
}
