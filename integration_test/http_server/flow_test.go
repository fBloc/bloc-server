package http_server

import (
	"encoding/json"
	"log"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/interfaces/web/flow"
	"github.com/fBloc/bloc-server/internal/http_util"
	"github.com/fBloc/bloc-server/internal/timestamp"
	"github.com/fBloc/bloc-server/value_object"
	. "github.com/smartystreets/goconvey/convey"
)

func aggFlowToWebFlow(aggF *aggregate.Flow) *flow.Flow {
	httpFuncs := make(map[string]*flow.FlowFunction, len(aggF.FlowFunctionIDMapFlowFunction))
	for k, v := range aggF.FlowFunctionIDMapFlowFunction {
		paramIpts := make([][]flow.IptComponentConfig, len(v.ParamIpts))
		for i, j := range v.ParamIpts {
			paramIpts[i] = make([]flow.IptComponentConfig, len(j))
			for z, k := range j {
				paramIpts[i][z] = flow.IptComponentConfig{
					Blank:          k.Blank,
					IptWay:         k.IptWay,
					ValueType:      k.ValueType,
					Value:          k.Value,
					FlowFunctionID: k.FlowFunctionID,
					Key:            k.Key,
				}
			}
		}

		httpFuncs[k] = &flow.FlowFunction{
			FunctionID:                v.FunctionID,
			Note:                      v.Note,
			Position:                  v.Position,
			UpstreamFlowFunctionIDs:   v.UpstreamFlowFunctionIDs,
			DownstreamFlowFunctionIDs: v.DownstreamFlowFunctionIDs,
			ParamIpts:                 paramIpts,
		}
	}
	retFlow := &flow.Flow{
		ID:                            aggF.ID,
		Name:                          aggF.Name,
		IsDraft:                       aggF.IsDraft,
		Version:                       aggF.Version,
		OriginID:                      aggF.OriginID,
		Newest:                        aggF.Newest,
		CreateTime:                    timestamp.NewTimeStampFromTime(aggF.CreateTime),
		Position:                      aggF.Position,
		FlowFunctionIDMapFlowFunction: httpFuncs,
		Crontab:                       aggF.Crontab.String(),
		TriggerKey:                    aggF.TriggerKey,
		TimeoutInSeconds:              aggF.TimeoutInSeconds,
		RetryAmount:                   aggF.RetryAmount,
		RetryIntervalInSecond:         aggF.RetryIntervalInSecond,
		AllowParallelRun:              aggF.AllowParallelRun,
	}
	return retFlow
}

func TestNeedLogin(t *testing.T) {
	Convey("need login to do all operations", t, func() {
		resp := struct {
			web.RespMsg
			Data []*flow.Flow `json:"data"`
		}{}
		_, err := http_util.Get(
			notExistUserHeader(),
			serverAddress+"/api/v1/draft_flow",
			http_util.BlankGetParam, &resp)
		So(err, ShouldBeNil)
		// So(resp.Code, ShouldEqual, http.StatusUnauthorized)
		// So(len(resp.Data), ShouldEqual, 0)
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
	Convey("create draft flow", t, func() {
		aggFlow := getFakeAggFlow()
		reqFlow := aggFlowToWebFlow(aggFlow)
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

func TestOnlineFlow(t *testing.T) {
	// before created online flow
	Convey("before there is online flow", t, func() {
		Convey("filter flows", func() {
			resp := struct {
				web.RespMsg
				Flows []*flow.Flow `json:"data"`
			}{}
			_, err := http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/flow",
				http_util.BlankGetParam,
				&resp)
			So(err, ShouldBeNil)
			So(resp.Code, ShouldEqual, http.StatusOK)
			So(len(resp.Flows), ShouldEqual, 0)
		})

		Convey("get_by_id flow", func() {
			resp := struct {
				web.RespMsg
				Flow *flow.Flow `json:"data"`
			}{}
			_, err := http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/flow/get_by_id/"+value_object.NewUUID().String(),
				http_util.BlankGetParam,
				&resp)
			So(err, ShouldBeNil)
			So(resp.Code, ShouldEqual, http.StatusOK)
			So(resp.Flow.IsZero(), ShouldBeTrue)
		})
	})

	// create a draft flow
	aggFlow := getFakeAggFlow()
	reqFlow := aggFlowToWebFlow(aggFlow)
	reqBody, _ := json.Marshal(reqFlow)
	resp := struct {
		web.RespMsg
		DraftFlow *flow.Flow `json:"data"`
	}{}
	http_util.Post(
		superuserHeader(),
		serverAddress+"/api/v1/draft_flow",
		http_util.BlankGetParam,
		reqBody, &resp)
	if resp.DraftFlow.IsZero() {
		log.Panicf("create draft flow failed")
	}
	if resp.DraftFlow.ID.IsNil() {
		log.Panicf("create draft flow's ID is nil")
	}

	// pub draft to online flow
	pubResp := struct {
		web.RespMsg
		OnlineFlow *flow.Flow `json:"data"`
	}{}
	_, err := http_util.Get(
		superuserHeader(),
		serverAddress+"/api/v1/draft_flow/commit_by_id/"+resp.DraftFlow.ID.String(),
		http_util.BlankGetParam,
		&pubResp)
	if err != nil {
		log.Panicf("pub draft flow to online failed: %v", err)
	}
	if pubResp.OnlineFlow.IsZero() {
		log.Panicf("pub draft flow to online returned zero flow. resp: %v", pubResp)
	}

	// test online flow about http api
	Convey("after there is online flow", t, func() {
		Convey("query", func() {
			Convey("filter flows", func() {
				resp := struct {
					web.RespMsg
					Flows []*flow.Flow `json:"data"`
				}{}
				_, err := http_util.Get(
					superuserHeader(),
					serverAddress+"/api/v1/flow",
					http_util.BlankGetParam,
					&resp)
				So(err, ShouldBeNil)
				So(resp.Code, ShouldEqual, http.StatusOK)
				So(len(resp.Flows), ShouldEqual, 1)
			})

			Convey("get_by_id flow", func() {
				resp := struct {
					web.RespMsg
					Flow *flow.Flow `json:"data"`
				}{}
				_, err := http_util.Get(
					superuserHeader(),
					serverAddress+"/api/v1/flow/get_by_id/"+pubResp.OnlineFlow.ID.String(),
					http_util.BlankGetParam,
					&resp)
				So(err, ShouldBeNil)
				So(resp.Code, ShouldEqual, http.StatusOK)
				So(resp.Flow.IsZero(), ShouldBeFalse)
				So(resp.Flow.Name, ShouldEqual, aggFlow.Name)
			})

			Convey("get_latestonline_by_origin_id flow", func() {
				resp := struct {
					web.RespMsg
					Flow *flow.Flow `json:"data"`
				}{}
				_, err := http_util.Get(
					superuserHeader(),
					serverAddress+"/api/v1/flow/get_latestonline_by_origin_id/"+pubResp.OnlineFlow.OriginID.String(),
					http_util.BlankGetParam,
					&resp)
				So(err, ShouldBeNil)
				So(resp.Code, ShouldEqual, http.StatusOK)
				So(resp.Flow.IsZero(), ShouldBeFalse)
				So(resp.Flow.Name, ShouldEqual, aggFlow.Name)
			})
		})

		Convey("update", func() {
			Convey("SetExecuteControlAttribute trigger key", func() {
				req := flow.Flow{
					ID:                pubResp.OnlineFlow.ID,
					AllowTriggerByKey: true,
				}
				reqBody, _ := json.Marshal(req)
				var resp *web.RespMsg
				_, err := http_util.Patch(
					superuserHeader(),
					serverAddress+"/api/v1/flow/set_execute_control_attributes",
					http_util.BlankGetParam, reqBody, &resp)
				So(err, ShouldBeNil)
				So(resp.Code, ShouldEqual, http.StatusOK)

				getFlowResp := struct {
					web.RespMsg
					Flow *flow.Flow `json:"data"`
				}{}
				http_util.Get(
					superuserHeader(),
					serverAddress+"/api/v1/flow/get_by_id/"+pubResp.OnlineFlow.ID.String(),
					http_util.BlankGetParam, &getFlowResp)
				So(err, ShouldBeNil)
				So(resp.Code, ShouldEqual, http.StatusOK)
				So(getFlowResp.Flow.IsZero(), ShouldBeFalse)
				So(getFlowResp.Flow.AllowTriggerByKey, ShouldBeTrue)

				req = flow.Flow{
					ID:                pubResp.OnlineFlow.ID,
					AllowTriggerByKey: false,
				}
				reqBody, _ = json.Marshal(req)
				_, err = http_util.Patch(
					superuserHeader(),
					serverAddress+"/api/v1/flow/set_execute_control_attributes",
					http_util.BlankGetParam, reqBody, &resp)
				So(err, ShouldBeNil)
				So(resp.Code, ShouldEqual, http.StatusOK)

				http_util.Get(
					superuserHeader(),
					serverAddress+"/api/v1/flow/get_by_id/"+pubResp.OnlineFlow.ID.String(),
					http_util.BlankGetParam, &getFlowResp)
				So(getFlowResp.Flow.AllowTriggerByKey, ShouldBeFalse)
			})

			Convey("SetExecuteControlAttribute crontab string", func() {
				Convey("valid set", func() {
					crontabStr := "* * * * *"
					req := flow.Flow{
						ID:      pubResp.OnlineFlow.ID,
						Crontab: crontabStr,
					}
					reqBody, _ := json.Marshal(req)
					var resp *web.RespMsg
					_, err := http_util.Patch(
						superuserHeader(),
						serverAddress+"/api/v1/flow/set_execute_control_attributes",
						http_util.BlankGetParam,
						reqBody,
						&resp)
					So(err, ShouldBeNil)
					So(resp.Code, ShouldEqual, http.StatusOK)

					getFlowResp := struct {
						web.RespMsg
						Flow *flow.Flow `json:"data"`
					}{}
					http_util.Get(
						superuserHeader(),
						serverAddress+"/api/v1/flow/get_by_id/"+pubResp.OnlineFlow.ID.String(),
						http_util.BlankGetParam, &getFlowResp)
					So(err, ShouldBeNil)
					So(resp.Code, ShouldEqual, http.StatusOK)
					So(getFlowResp.Flow.IsZero(), ShouldBeFalse)
					So(getFlowResp.Flow.Crontab, ShouldEqual, crontabStr)
				})

				Convey("not valid set", func() {
					crontabStr := "* * * *"
					req := flow.Flow{
						ID:      pubResp.OnlineFlow.ID,
						Crontab: crontabStr,
					}
					reqBody, _ := json.Marshal(req)
					var resp *web.RespMsg
					_, err := http_util.Patch(
						superuserHeader(),
						serverAddress+"/api/v1/flow/set_execute_control_attributes",
						http_util.BlankGetParam,
						reqBody,
						&resp)
					So(err, ShouldBeNil)
					So(resp.Code, ShouldEqual, http.StatusBadRequest)
				})
			})
		})

		Convey("permission", func() {
			// now the flow is created by superuser. nobody should not can get it
			Convey("nobody should not can get the flow", func() {
				// Convey("filter", func() {
				// 	resp := struct {
				// 		web.RespMsg
				// 		Flows []*flow.Flow `json:"data"`
				// 	}{}
				// 	_, err := http_util.Get(
				// 		nobodyHeader(),
				// 		serverAddress+"/api/v1/flow",
				// 		http_util.BlankGetParam,
				// 		&resp)
				// 	So(err, ShouldBeNil)
				// 	So(resp.Code, ShouldEqual, http.StatusOK)
				// 	So(len(resp.Flows), ShouldEqual, 0)
				// })

				Convey("get_latestonline_by_origin_id", func() {
					resp := struct {
						web.RespMsg
						Flow *flow.Flow `json:"data"`
					}{}
					_, err := http_util.Get(
						nobodyHeader(),
						serverAddress+"/api/v1/flow/get_latestonline_by_origin_id/"+pubResp.OnlineFlow.OriginID.String(),
						http_util.BlankGetParam, &resp)
					So(err, ShouldBeNil)
					// So(resp.Code, ShouldEqual, http.StatusForbidden)
					// So(resp.Flow.IsZero(), ShouldBeTrue)
				})
			})

			Convey("permission", func() {
				Convey("add permission", func() {
					// add read permission for nobody
					req := flow.PermissionReq{
						PermissionType: flow.Read,
						FlowID:         pubResp.OnlineFlow.ID,
						UserID:         nobodyID,
					}
					reqBody, _ := json.Marshal(req)
					var resp *web.RespMsg
					_, err = http_util.Post(
						superuserHeader(),
						serverAddress+"/api/v1/flow_permission/add_permission",
						http_util.BlankGetParam,
						reqBody,
						&resp)
					So(err, ShouldBeNil)

					// check it can get the flow after add read permission
					getByOriginIDresp := struct {
						web.RespMsg
						Flow *flow.Flow `json:"data"`
					}{}
					_, err := http_util.Get(
						nobodyHeader(),
						serverAddress+"/api/v1/flow/get_latestonline_by_origin_id/"+pubResp.OnlineFlow.OriginID.String(),
						http_util.BlankGetParam, &getByOriginIDresp)
					So(err, ShouldBeNil)
					So(getByOriginIDresp.Code, ShouldEqual, http.StatusOK)
					So(getByOriginIDresp.Flow.IsZero(), ShouldBeFalse)
					So(getByOriginIDresp.Flow.Name, ShouldEqual, pubResp.OnlineFlow.Name)

					filterResp := struct {
						web.RespMsg
						Flows []*flow.Flow `json:"data"`
					}{}
					_, err = http_util.Get(
						nobodyHeader(),
						serverAddress+"/api/v1/flow",
						http_util.BlankGetParam,
						&filterResp)
					So(err, ShouldBeNil)
					So(filterResp.Code, ShouldEqual, http.StatusOK)
					So(len(filterResp.Flows), ShouldEqual, 1)
				})

				// Convey("delete permission", func() {
				// 	// delete read permission for nobody
				// 	req := flow.PermissionReq{
				// 		PermissionType: flow.Read,
				// 		FlowID:         pubResp.OnlineFlow.ID,
				// 		UserID:         nobodyID,
				// 	}
				// 	reqBody, _ := json.Marshal(req)
				// 	var resp *web.RespMsg
				// 	_, err = http_util.Delete(
				// 		superuserHeader(),
				// 		serverAddress+"/api/v1/flow_permission/remove_permission",
				// 		http_util.BlankGetParam,
				// 		reqBody,
				// 		&resp)
				// 	So(err, ShouldBeNil)

				// 	// check it can get the flow after add read permission
				// 	getByOriginIDresp := struct {
				// 		web.RespMsg
				// 		Flow *flow.Flow `json:"data"`
				// 	}{}
				// 	_, err := http_util.Get(
				// 		nobodyHeader(),
				// 		serverAddress+"/api/v1/flow/get_latestonline_by_origin_id/"+pubResp.OnlineFlow.OriginID.String(),
				// 		http_util.BlankGetParam, &getByOriginIDresp)
				// 	So(err, ShouldBeNil)
				// 	// So(getByOriginIDresp.Code, ShouldEqual, http.StatusForbidden)

				// 	filterResp := struct {
				// 		web.RespMsg
				// 		Flows []*flow.Flow `json:"data"`
				// 	}{}
				// 	_, err = http_util.Get(
				// 		nobodyHeader(),
				// 		serverAddress+"/api/v1/flow",
				// 		http_util.BlankGetParam,
				// 		&filterResp)
				// 	So(err, ShouldBeNil)
				// 	So(filterResp.Code, ShouldEqual, http.StatusOK)
				// 	So(len(filterResp.Flows), ShouldEqual, 0)
				// })
			})
		})

		Convey("DeleteFlowByOriginID", func() {
			resp := struct {
				web.RespMsg
				Data map[string]int64 `json:"data"`
			}{}
			_, err := http_util.Delete(
				superuserHeader(),
				serverAddress+"/api/v1/flow/delete_by_origin_id/"+pubResp.OnlineFlow.OriginID.String(),
				http_util.BlankGetParam, http_util.BlankBody,
				&resp)
			So(err, ShouldBeNil)
			So(resp.Code, ShouldEqual, http.StatusOK)
			So(resp.Data["delete_amount"], ShouldEqual, 1)

			getFlowResp := struct {
				web.RespMsg
				Flow *flow.Flow `json:"data"`
			}{}
			http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/flow/get_by_id/"+pubResp.OnlineFlow.ID.String(),
				http_util.BlankGetParam, &getFlowResp)
			So(err, ShouldBeNil)
			So(resp.Code, ShouldEqual, http.StatusOK)
			So(getFlowResp.Flow.IsZero(), ShouldBeFalse)
			So(getFlowResp.Flow.Delete, ShouldBeTrue)
		})
	})
}
