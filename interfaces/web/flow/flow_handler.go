package flow

import (
	"encoding/json"
	"net/http"

	"github.com/fBloc/bloc-backend-go/interfaces/web"
	"github.com/fBloc/bloc-backend-go/interfaces/web/req_context"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// GetFlowByID 需要特别注意的是，通过id获取的话，就可能是获取到的老版本的flow
func GetFlowByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	flowID := ps.ByName("id")
	if flowID == "" {
		web.WriteBadRequestDataResp(&w, "id cannot be nil")
		return
	}

	flowIns, err := fService.Flow.GetByIDStr(flowID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "")
		return
	}
	if flowIns.IsZero() {
		web.WriteSucResp(&w, nil)
		return
	}

	var retFlow *Flow
	// 因为用户可能是新用户，没有加到老版本flow权限中
	// 但是能发起此请求的、一定是有新版本权限的，老版本依然有对应权限
	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil,
			"get requser from context failed")
		return
	}
	retFlow = fromAggWithLatestRunFunctionView(flowIns, reqUser)

	if !flowIns.Newest { // 老版本的话，只返回read权限(别的操作都不该有了)
		retFlow = fromAggWithoutUserPermission(flowIns)
		retFlow.Write = false
		retFlow.Execute = false
		retFlow.Delete = false
		retFlow.AssignPermission = false
	}
	web.WriteSucResp(&w, retFlow)
}

// GetFlowByOriginID 通过origin_id精确查询
func GetFlowByOriginID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	originID := ps.ByName("origin_id")
	arrFlowID := r.URL.Query().Get("arrangement_flow_id") // 获取特定arrangement下的flow最近一次运行记录

	flowIns, err := fService.Flow.GetOnlineByOriginIDStr(originID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "get latest flow by origin_id failed")
		return
	}
	if flowIns.IsZero() {
		web.WriteSucResp(&w, nil)
		return
	}
	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil,
			"get requser from context failed")
		return
	}

	var retFlow *Flow
	if arrFlowID == "" {
		retFlow = fromAggWithLatestRunFunctionView(flowIns, reqUser)
	} else {
		retFlow = fromAggWithLatestRunOfCertainArrangement(flowIns, reqUser, arrFlowID)
	}

	if !retFlow.Read {
		web.WritePermissionNotEnough(&w, "user have no read permission on this flow")
		return
	}

	web.WriteSucResp(&w, retFlow)
}

// FilterFlow 获取flow
func FilterFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil,
			"get requser from context failed")
		return
	}

	flowID := r.URL.Query().Get("id")
	// 若flowID存在的话表示是精确查找
	if flowID != "" {
		// 通过ID查询并返回flow
		agg, err := fService.Flow.GetByIDStr(flowID)
		if err != nil {
			web.WriteInternalServerErrorResp(&w, err, "")
			return
		}
		ret := fromAggWithLatestRunFlowView(agg, reqUser)
		web.WriteSucResp(&w, ret)
		return
	}

	// 否则是过滤查找
	aggSlice, err := fService.Flow.FilterOnline(reqUser.ID, r.URL.Query().Get("name__contains"))
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit repository failed")
		return
	}
	ret := fromAggSliceWithLatestRun(aggSlice, reqUser)

	web.WriteSucResp(&w, ret)
}

// Patch 用户前端保存创建/更新的flow的运行配置
func SetExecuteControlAttributes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var reqFlow Flow
	err := json.NewDecoder(r.Body).Decode(&reqFlow)
	if err != nil {
		web.WriteBadRequestDataResp(&w, err.Error())
		return
	}

	// > id不能为空
	if reqFlow.ID == uuid.Nil {
		web.WriteBadRequestDataResp(&w, "must have id field")
		return
	}

	flowIns, err := fService.Flow.GetByID(reqFlow.ID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "get flow by id failed")
		return
	}
	if flowIns.IsZero() {
		web.WriteBadRequestDataResp(&w, "find no flow by this id")
		return
	}

	// > 检测当前用户是否有update此flow的权限
	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil,
			"get requser from context failed")
		return
	}

	if !flowIns.UserCanExecute(reqUser) {
		web.WritePermissionNotEnough(&w, "need execute permission to update")
	}

	// > 开始更新相应字段
	var changedFieldAmount uint8
	// >> 更新crontab
	if !reqFlow.Crontab.Equal(&flowIns.Crontab) {
		changedFieldAmount++
		flowIns.Crontab = *reqFlow.Crontab
	}

	// >> 更新trigger_key
	if reqFlow.TriggerKey != flowIns.TriggerKey {
		changedFieldAmount++
		flowIns.TriggerKey = reqFlow.TriggerKey
	}

	// >> 更新超时设置
	if reqFlow.TimeoutInSeconds != flowIns.TimeoutInSeconds {
		changedFieldAmount++
		flowIns.TimeoutInSeconds = reqFlow.TimeoutInSeconds
	}
	// >> 更新重试策略
	if reqFlow.RetryAmount != flowIns.RetryAmount ||
		reqFlow.RetryIntervalInSecond != flowIns.RetryIntervalInSecond {
		changedFieldAmount++
		if reqFlow.RetryAmount > 0 && reqFlow.RetryIntervalInSecond <= 0 {
			// 对于设置了重试次数的，重试间隔不能设置为0，暂时默认给个1秒
			reqFlow.RetryIntervalInSecond = 1
		}
		flowIns.RetryAmount = reqFlow.RetryAmount
		flowIns.RetryIntervalInSecond = reqFlow.RetryIntervalInSecond
	}

	// >> 更新是否支持在运行的时候也发布
	if reqFlow.AllowParallelRun != flowIns.AllowParallelRun {
		changedFieldAmount++
		flowIns.AllowParallelRun = reqFlow.AllowParallelRun
	}

	if changedFieldAmount > 0 {
		err = fService.Flow.ReplaceByID(reqFlow.ID, flowIns)
		if err != nil {
			web.WriteInternalServerErrorResp(&w, err, "update failed")
			return
		}
	}

	web.WritePlainSucOkResp(&w)
}

// DeleteFlowByOriginID 只有delete user能够删除flow，通过originID全部删除
func DeleteFlowByOriginID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	originID := ps.ByName("origin_id")
	if originID == "" {
		web.WriteBadRequestDataResp(&w, "origin_id param must exist")
	}
	uuOriginID, err := uuid.Parse(originID)
	if err != nil {
		web.WriteBadRequestDataResp(&w,
			"parse origin_id to uuid failed:", err.Error())
		return
	}

	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil,
			"get requser from context failed")
		return
	}

	aggFlow, err := fService.Flow.GetOnlineByOriginID(uuOriginID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, nil,
			"get flow from origin_id failed")
		return
	}
	if !aggFlow.UserCanDelete(reqUser) {
		web.WritePermissionNotEnough(&w, "need delete permission")
		return
	}

	// TODO：应该检查此originID的在线flow是不是还在被arrangement引用，如果是的话不能删除
	deleteCount, err := fService.Flow.DeleteByOriginID(uuOriginID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "delete failed")
		return
	}
	// TODO：不应该只删除flow就完事了，对应的flowRunRecord、fucntionRunRecord ... 也应该删除
	web.WriteDeleteSucResp(&w, deleteCount)
}
