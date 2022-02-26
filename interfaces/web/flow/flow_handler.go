package flow

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"

	"github.com/julienschmidt/httprouter"
)

// GetFlowByID 需要特别注意的是，通过id获取的话，就可能是获取到的老版本的flow
func GetFlowByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "get flow by id"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(logTags, "failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil, "get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	flowID := ps.ByName("id")
	if flowID == "" {
		fService.Logger.Infof(logTags, "lack id in url path")
		web.WriteBadRequestDataResp(&w, r, "id cannot be nil")
		return
	}
	logTags["flow_id"] = flowID

	flowIns, err := fService.Flow.GetByIDStr(flowID)
	if err != nil {
		fService.Logger.Errorf(logTags, "get flow by id error: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "")
		return
	}
	if flowIns.IsZero() {
		fService.Logger.Warningf(logTags, "get flow by id match no record")
		web.WriteSucResp(&w, r, nil)
		return
	}

	var retFlow *Flow
	// 因为用户可能是新用户，没有加到老版本flow权限中
	// 但是能发起此请求的、一定是有新版本权限的，老版本依然有对应权限
	retFlow = fromAggWithLatestRunFunctionView(flowIns, reqUser)

	if !flowIns.Newest { // 老版本的话，只返回read权限(别的操作都不该有了)
		fService.Logger.Infof(logTags, "old version flow")
		retFlow = fromAggWithoutUserPermission(flowIns)
		retFlow.Write = false
		retFlow.Execute = false
		retFlow.Delete = false
		retFlow.AssignPermission = false
	}

	fService.Logger.Infof(logTags, "finished")
	web.WriteSucResp(&w, r, retFlow)
}

func GetFlowByCertainFlowRunRecord(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "get flow by flow_run_record id"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(
			logTags, "failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil, "get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	flowRunRecordIDStr := ps.ByName("flow_run_record_id")
	if flowRunRecordIDStr == "" {
		fService.Logger.Warningf(logTags, "lack flow_run_record_id in url path")
		web.WriteBadRequestDataResp(&w, r, "flow_run_record_id cannot be nil")
		return
	}
	logTags["flow_run_record_id"] = flowRunRecordIDStr

	flowRunRecordUUID, err := value_object.ParseToUUID(flowRunRecordIDStr)
	if err != nil {
		fService.Logger.Warningf(
			logTags, "parse flow_run_record_id to uuid failed: %v", err)
		web.WriteBadRequestDataResp(
			&w, r, "parse flowRunRecordIDStr(%s) to uuid failed",
			flowRunRecordIDStr)
		return
	}

	aggFlowRunRecord, err := fService.FlowRunRecord.GetByID(flowRunRecordUUID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, r, err, "fetch FlowRunRecord ins failed")
		return
	}

	flowIns, err := fService.Flow.GetByID(aggFlowRunRecord.FlowID)
	if err != nil {
		fService.Logger.Errorf(logTags, "get by id failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "get flow by id failed")
		return
	}
	if flowIns.IsZero() {
		fService.Logger.Warningf(logTags, "get by id match no record")
		web.WriteInternalServerErrorResp(
			&w, r, err,
			"get no flow by this flow_id(%s)", aggFlowRunRecord.FlowID.String())
		return
	}

	retFlow := fromAggWithCertainRunFunctionView(flowIns, aggFlowRunRecord, reqUser)
	if !flowIns.Newest { // 老版本的话，只返回read权限(别的操作都不该有了)
		fService.Logger.Infof(logTags, "old version flow")
		retFlow = fromAggWithoutUserPermission(flowIns)
		retFlow.Write = false
		retFlow.Execute = false
		retFlow.Delete = false
		retFlow.AssignPermission = false
	}

	fService.Logger.Infof(logTags, "finished")
	web.WriteSucResp(&w, r, retFlow)
}

// GetFlowByOriginID 通过origin_id精确查询
func GetFlowByOriginID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "get flow by origin_id"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(
			logTags, "failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil, "get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	originID := ps.ByName("origin_id")
	arrFlowID := r.URL.Query().Get("arrangement_flow_id") // 获取特定arrangement下的flow最近一次运行记录

	flowIns, err := fService.Flow.GetOnlineByOriginIDStr(originID)
	if err != nil {
		fService.Logger.Errorf(logTags, "get flow by origin_id error: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "get latest flow by origin_id failed")
		return
	}
	if flowIns.IsZero() {
		fService.Logger.Infof(logTags, "match no record")
		web.WriteSucResp(&w, r, nil)
		return
	}

	var retFlow *Flow
	if arrFlowID == "" {
		retFlow = fromAggWithLatestRunFunctionView(flowIns, reqUser)
	} else {
		retFlow = fromAggWithLatestRunOfCertainArrangement(flowIns, reqUser, arrFlowID)
	}

	if !retFlow.Read {
		fService.Logger.Infof(logTags, "have no read permission")
		web.WritePermissionNotEnough(&w, r, "user have no read permission on this flow")
		return
	}

	fService.Logger.Infof(logTags, "finished")
	web.WriteSucResp(&w, r, retFlow)
}

// FilterFlow 获取flow
func FilterFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "filter flow"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(
			logTags, "failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil, "get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	flowID := r.URL.Query().Get("id")
	// 若flowID存在的话表示是精确查找
	if flowID != "" {
		// 通过ID查询并返回flow
		agg, err := fService.Flow.GetByIDStr(flowID)
		if err != nil {
			fService.Logger.Errorf(logTags, "get by id failed: %v", err)
			web.WriteInternalServerErrorResp(&w, r, err, "")
			return
		}
		ret := fromAggWithLatestRunFlowView(agg, reqUser)

		fService.Logger.Infof(logTags, "finished")
		web.WriteSucResp(&w, r, ret)
		return
	}

	// 否则是过滤查找
	aggSlice, err := fService.Flow.FilterOnline(reqUser, r.URL.Query().Get("name__contains"))
	if err != nil {
		fService.Logger.Errorf(logTags, "filter failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "visit repository failed")
		return
	}
	ret := fromAggSliceWithLatestRun(aggSlice, reqUser)

	fService.Logger.Infof(logTags, "finished")
	web.WriteSucResp(&w, r, ret)
}

// Patch 用户前端保存创建/更新的flow的运行配置
func SetExecuteControlAttributes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "set flow's execute control attribute"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(logTags, "failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil, "get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	var reqFlow Flow
	err := json.NewDecoder(r.Body).Decode(&reqFlow)
	if err != nil {
		fService.Logger.Warningf(
			logTags, "json unmarshal req body to flow failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, err.Error())
		return
	}

	// > id不能为空
	if reqFlow.ID.IsNil() {
		fService.Logger.Warningf(logTags, "lack flow_id")
		web.WriteBadRequestDataResp(&w, r, "must have id field")
		return
	}
	logTags["flow_id"] = reqFlow.ID.String()

	flowIns, err := fService.Flow.GetByID(reqFlow.ID)
	if err != nil {
		fService.Logger.Errorf(logTags, "get flow by id failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "get flow by id failed")
		return
	}
	if flowIns.IsZero() {
		fService.Logger.Warningf(logTags, "get flow by id match no record")
		web.WriteBadRequestDataResp(&w, r, "find no flow by this id")
		return
	}

	// > 检测当前用户是否有update此flow的权限
	if !flowIns.UserCanExecute(reqUser) {
		fService.Logger.Warningf(logTags, "user lack execute permission")
		web.WritePermissionNotEnough(&w, r, "need execute permission to update")
	}

	// >> 更新crontab
	if !reqFlow.Crontab.Equal(flowIns.Crontab) {
		err := fService.Flow.PatchCrontab(
			reqFlow.ID, *reqFlow.Crontab)
		baseLogMsg := fmt.Sprintf(
			"change flow's crontab from:%s to:%s",
			flowIns.Crontab.String(), reqFlow.Crontab.String())

		if err != nil {
			fService.Logger.Errorf(logTags, "%s. error: %v", baseLogMsg, err)
			web.WriteInternalServerErrorResp(&w, r, err, "update crontab failed")
			return
		}
		fService.Logger.Infof(logTags, baseLogMsg)
	}

	// >> 更新trigger_key
	if reqFlow.TriggerKey != flowIns.TriggerKey {
		err := fService.Flow.PatchTriggerKey(
			reqFlow.ID, reqFlow.TriggerKey)
		baseLogMsg := fmt.Sprintf(
			"change flow's trigger_key from:%s to:%s",
			flowIns.TriggerKey, reqFlow.TriggerKey)

		if err != nil {
			fService.Logger.Errorf(logTags, "%s. error: %v", baseLogMsg, err)
			web.WriteInternalServerErrorResp(&w, r, err, "update trigger_key failed")
			return
		}
		fService.Logger.Infof(logTags, baseLogMsg)
	}

	// >> 更新超时设置
	if reqFlow.TimeoutInSeconds != flowIns.TimeoutInSeconds {
		err := fService.Flow.PatchTimeout(
			reqFlow.ID, reqFlow.TimeoutInSeconds)
		baseLogMsg := fmt.Sprintf(
			"change flow's timeout from:%d to:%d",
			flowIns.TimeoutInSeconds, reqFlow.TimeoutInSeconds)

		if err != nil {
			fService.Logger.Errorf(logTags, "%s. error: %v", baseLogMsg, err)
			web.WriteInternalServerErrorResp(&w, r, err, "update timeout failed")
			return
		}
		fService.Logger.Infof(logTags, baseLogMsg)
	}

	// >> 更新重试策略
	if reqFlow.RetryAmount != flowIns.RetryAmount ||
		reqFlow.RetryIntervalInSecond != flowIns.RetryIntervalInSecond {
		if reqFlow.RetryAmount > 0 && reqFlow.RetryIntervalInSecond <= 0 {
			// 对于设置了重试次数的，重试间隔不能设置为0，暂时默认给个1秒
			reqFlow.RetryIntervalInSecond = 1
		}

		err := fService.Flow.PatchRetryStrategy(
			reqFlow.ID, reqFlow.RetryAmount, reqFlow.RetryIntervalInSecond)
		baseLogMsg := fmt.Sprintf(
			"change flow's retry_strategy from:%d-%ds to:%d-%ds",
			flowIns.RetryAmount, flowIns.RetryIntervalInSecond,
			reqFlow.RetryAmount, reqFlow.RetryIntervalInSecond)
		if err != nil {
			fService.Logger.Errorf(logTags, "%s. error: %v", baseLogMsg, err)
			web.WriteInternalServerErrorResp(&w, r, err, "update retry_strategy failed")
			return
		}
		fService.Logger.Infof(logTags, baseLogMsg)
	}

	// >> 更新是否支持在运行的时候也发布
	if reqFlow.AllowParallelRun != flowIns.AllowParallelRun {
		err := fService.Flow.PatchAllowParallelRun(
			reqFlow.ID, reqFlow.AllowParallelRun)
		baseLogMsg := fmt.Sprintf(
			"change flow's allow_parallel_run from:%t to:%t",
			flowIns.AllowParallelRun, reqFlow.AllowParallelRun)
		if err != nil {
			fService.Logger.Errorf(logTags, "%s. error: %v", baseLogMsg, err)
			web.WriteInternalServerErrorResp(&w, r, err, "update allow_parallel_run failed")
			return
		}
		fService.Logger.Infof(logTags, baseLogMsg)
	}

	fService.Logger.Infof(logTags, "finished")
	web.WritePlainSucOkResp(&w, r)
}

// DeleteFlowByOriginID 只有delete user能够删除flow，通过originID全部删除
func DeleteFlowByOriginID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "delete flow by origin_id"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(logTags, "failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil,
			"get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	originID := ps.ByName("origin_id")
	if originID == "" {
		fService.Logger.Warningf(logTags, "lack origin_id url path")
		web.WriteBadRequestDataResp(&w, r, "origin_id param must exist")
	}
	uuOriginID, err := value_object.ParseToUUID(originID)
	if err != nil {
		fService.Logger.Warningf(logTags, "parse origin_id to uuid failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, "parse origin_id to uuid failed: %v", err)
		return
	}
	logTags["origin_id"] = originID

	aggFlow, err := fService.Flow.GetOnlineByOriginID(uuOriginID)
	if err != nil {
		fService.Logger.Errorf(logTags, "get flow from origin_id failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, nil,
			"get flow from origin_id failed")
		return
	}

	if !aggFlow.UserCanDelete(reqUser) {
		fService.Logger.Errorf(logTags, "user donnot have delete permission")
		web.WritePermissionNotEnough(&w, r, "need delete permission")
		return
	}

	// this is fake delete.
	// it will only mark the flow as deleted and no longer can be seen in frontend
	// actually delete nothing of it
	deleteCount, err := fService.Flow.DeleteByOriginID(uuOriginID)
	if err != nil {
		fService.Logger.Errorf(logTags, "delete failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "delete failed")
		return
	}

	fService.Logger.Infof(logTags, "finished with delete amount: %d", deleteCount)
	web.WriteDeleteSucResp(&w, r, deleteCount)
}
