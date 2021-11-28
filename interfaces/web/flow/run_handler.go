package flow

import (
	"net/http"

	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/event"
	"github.com/fBloc/bloc-backend-go/interfaces/web"
	"github.com/fBloc/bloc-backend-go/interfaces/web/req_context"

	"github.com/julienschmidt/httprouter"
)

func Run(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	flowOriginID := ps.ByName("origin_id")
	if flowOriginID == "" {
		web.WriteBadRequestDataResp(&w, "origin_id cannot be nil")
		return
	}

	// 通过获取对应的flow检测flow_origin_id是否有效
	flowIns, err := fService.Flow.GetOnlineByOriginIDStr(flowOriginID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit flow by origin_id failed")
		return
	}
	if flowIns.IsZero() {
		web.WriteBadRequestDataResp(&w, "origin_id find no flow")
		return
	}

	// 检查用户是否有执行权限
	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil, "get requser from context failed")
		return
	}
	if !flowIns.UserCanExecute(reqUser) {
		web.WritePermissionNotEnough(&w, "need execute permission")
		return
	}

	// create new run record
	aggFlowRunRecord := aggregate.NewUserTriggeredRunRecord(*flowIns, reqUser.ID)
	err = fService.FlowRunRecord.Create(aggFlowRunRecord)
	if err != nil {
		fService.Logger.Errorf("user %s trigger flow run failed. origin_id %s, error: %s",
			reqUser.Name, flowOriginID, err.Error())
		web.WriteInternalServerErrorResp(
			&w, err,
			"create flow run record to repository failed")
	}
	fService.Logger.Infof("user %s triggered flow run. origin_id %s",
		reqUser.Name, flowOriginID)
	event.PubEvent(&event.FlowToRun{FlowRunRecordID: aggFlowRunRecord.ID})

	web.WritePlainSucOkResp(&w)
}

func CancelRun(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	flowOriginID := ps.ByName("origin_id")
	if flowOriginID == "" {
		web.WriteBadRequestDataResp(&w, "origin_id cannot be nil")
		return
	}

	// 通过获取对应的flow检测flow_origin_id是否有效
	flowIns, err := fService.Flow.GetOnlineByOriginIDStr(flowOriginID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit flow by origin_id failed")
		return
	}
	if flowIns.IsZero() {
		web.WriteBadRequestDataResp(&w, "origin_id find no flow")
		return
	}

	// 检查用户是否有执行权限
	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil, "get requser from context failed")
		return
	}
	if !flowIns.UserCanExecute(reqUser) {
		web.WritePermissionNotEnough(&w, "need execute permission")
		return
	}

	// 获取全部的运行中任务
	aggFRRs, err := fService.FlowRunRecord.AllRunRecordOfFlowTriggeredByFlowID(flowIns.ID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit run record of this flow failed")
		return
	}

	// 取消全部运行中任务
	for _, i := range aggFRRs {
		// TODO 需要并行吗？
		err := fService.FlowRunRecord.UserCancel(i.ID, reqUser.ID)
		if err != nil {
			web.WriteInternalServerErrorResp(&w, err, "cancel record failed")
			return
		}
	}
	fService.Logger.Infof("user %s canceld flow run. origin_id %s",
		reqUser.Name, flowOriginID)

	web.WritePlainSucOkResp(&w)
}
