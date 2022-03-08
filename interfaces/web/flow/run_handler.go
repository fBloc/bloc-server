package flow

import (
	"net/http"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/event"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"

	"github.com/julienschmidt/httprouter"
)

func RunByTriggerKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "trigger_key flow to run"
	fService.Logger.Infof(logTags, "start")

	triggerKey := ps.ByName("trigger_key")
	if triggerKey == "" {
		fService.Logger.Warningf(logTags, "lack trigger_key in url path")
		web.WriteBadRequestDataResp(&w, r, "trigger_key cannot be nil")
		return
	}
	logTags["trigger_key"] = triggerKey

	flows, err := fService.Flow.Filter(
		value_object.NewRepositoryFilter().AddEqual("trigger_key", triggerKey))
	if err != nil {
		fService.Logger.Errorf(logTags, "filter flows by trigger_key failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "filter flows by trigger_key failed")
		return
	}
	if len(flows) == 0 {
		web.WriteBadRequestDataResp(&w, r,
			"trigger_key match no flow")
		return
	}
	if len(flows) > 1 {
		fService.Logger.Errorf(logTags,
			"matched more than one flow!")
	}

	for _, flowIns := range flows {
		thisLogTags := logTags
		thisLogTags["origin_id"] = flowIns.OriginID.String()

		// create new run record
		aggFlowRunRecord, err := aggregate.NewKeyTriggeredFlowRunRecord(
			r.Context(), &flowIns, triggerKey)
		if err != nil {
			fService.Logger.Errorf(
				logTags, "create aggregate flow_run_record failed: %v", err)
			web.WriteInternalServerErrorResp(&w, r, err, "build aggregate flow failed")
			return
		}

		err = fService.FlowRunRecord.Create(aggFlowRunRecord)
		if err != nil {
			fService.Logger.Errorf(
				logTags, "persist flow_run_record failed: %v", err)
			web.WriteInternalServerErrorResp(
				&w, r, err, "create flow run record to repository failed")
			return
		}
		logTags["flow_run_record_id"] = aggFlowRunRecord.ID.String()

		err = event.PubEvent(&event.FlowToRun{FlowRunRecordID: aggFlowRunRecord.ID})
		if err != nil {
			fService.Logger.Errorf(logTags, "pub event failed: %v", err)
			web.WriteInternalServerErrorResp(&w, r, err, "pub event failed")
			return
		}
	}

	fService.Logger.Infof(logTags, "finished")
	web.WritePlainSucOkResp(&w, r)
}

func Run(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "frontend trigger flow to run"
	fService.Logger.Infof(logTags, "start")

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(logTags, "failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil, "get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	flowOriginID := ps.ByName("origin_id")
	if flowOriginID == "" {
		fService.Logger.Infof(logTags, "lack origin_id in url path")
		web.WriteBadRequestDataResp(&w, r, "origin_id cannot be nil")
		return
	}

	// 通过获取对应的flow检测flow_origin_id是否有效
	flowIns, err := fService.Flow.GetOnlineByOriginIDStr(flowOriginID)
	if err != nil {
		fService.Logger.Errorf(logTags, "get flow by origin_id error: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "visit flow by origin_id failed")
		return
	}
	if flowIns.IsZero() {
		fService.Logger.Warningf(logTags, "get flow by origin_id match no record")
		web.WriteBadRequestDataResp(&w, r, "origin_id find no flow")
		return
	}
	logTags["origin_id"] = flowOriginID

	// 检查用户是否有执行权限
	if !flowIns.UserCanExecute(reqUser) {
		fService.Logger.Warningf(logTags, "user lack execute permission")
		web.WritePermissionNotEnough(&w, r, "need execute permission")
		return
	}

	// create new run record
	aggFlowRunRecord, err := aggregate.NewUserTriggeredFlowRunRecord(r.Context(), flowIns, reqUser)
	if err != nil {
		fService.Logger.Errorf(
			logTags, "create aggregate flow_run_record failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "build aggregate flow failed")
		return
	}

	err = fService.FlowRunRecord.Create(aggFlowRunRecord)
	if err != nil {
		fService.Logger.Errorf(
			logTags, "persist flow_run_record failed: %v", err)
		web.WriteInternalServerErrorResp(
			&w, r, err, "create flow run record to repository failed")
		return
	}
	logTags["flow_run_record_id"] = aggFlowRunRecord.ID.String()

	err = event.PubEvent(&event.FlowToRun{FlowRunRecordID: aggFlowRunRecord.ID})
	if err != nil {
		fService.Logger.Errorf(logTags, "pub event failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "pub event failed")
		return
	}

	fService.Logger.Infof(logTags, "finished")
	web.WritePlainSucOkResp(&w, r)
}

func CancelRun(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "cancel flow run"
	fService.Logger.Infof(logTags, "start")

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(logTags, "failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil, "get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	flowOriginID := ps.ByName("origin_id")
	if flowOriginID == "" {
		fService.Logger.Infof(logTags, "lack origin_id in url path")
		web.WriteBadRequestDataResp(&w, r, "origin_id cannot be nil")
		return
	}
	logTags["origin_id"] = flowOriginID

	// 通过获取对应的flow检测flow_origin_id是否有效
	flowIns, err := fService.Flow.GetOnlineByOriginIDStr(flowOriginID)
	if err != nil {
		fService.Logger.Errorf(logTags, "get flow by origin_id error: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "visit flow by origin_id failed")
		return
	}
	if flowIns.IsZero() {
		fService.Logger.Warningf(logTags, "get flow by origin_id match no record")
		web.WriteBadRequestDataResp(&w, r, "origin_id find no flow")
		return
	}

	// 检查用户是否有执行权限
	if !flowIns.UserCanExecute(reqUser) {
		fService.Logger.Warningf(logTags, "user lack execute permission")
		web.WritePermissionNotEnough(&w, r, "need execute permission")
		return
	}

	// 获取全部的运行中任务
	aggFRRs, err := fService.FlowRunRecord.AllRunRecordOfFlowTriggeredByFlowID(flowIns.ID)
	if err != nil {
		fService.Logger.Errorf(logTags, "visit run record of this flow failed")
		web.WriteInternalServerErrorResp(&w, r, err, "visit run record of this flow failed")
		return
	}

	// 取消全部运行中任务
	for _, i := range aggFRRs {
		// TODO 需要并行吗？
		err := fService.FlowRunRecord.UserCancel(i.ID, reqUser.ID)
		if err != nil {
			fService.Logger.Errorf(logTags,
				"cancel flow_run_record(id:%s) failed: %v", i.ID.String(), err)
			web.WriteInternalServerErrorResp(&w, r, err, "cancel record failed")
			return
		}
	}

	fService.Logger.Infof(logTags, "finished cancel %d task", len(aggFRRs))
	web.WritePlainSucOkResp(&w, r)
}
