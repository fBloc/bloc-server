package flow

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"

	"github.com/julienschmidt/httprouter"
)

func GetDraftByOriginID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "get draft flow by origin_id"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(logTags, "failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil, "get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	originID := ps.ByName("origin_id")
	if originID == "" {
		fService.Logger.Warningf(logTags, "lack query param origin_id")
		web.WriteBadRequestDataResp(&w, r, "origin_id cannot be blank")
		return
	}
	originUUID, err := value_object.ParseToUUID(originID)
	if err != nil {
		fService.Logger.Errorf(logTags, "parse query param origin_id: %s to uuid error: %v", originID, err)
		web.WriteBadRequestDataResp(&w, r, "cannot parse origin_id to uuid")
		return
	}
	logTags["origin_id"] = originID

	flowIns, err := fService.Flow.GetDraftByOriginID(originUUID)
	if err != nil {
		fService.Logger.Errorf(logTags, "get latest draft_flow by origin_id error: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "get latest draft_flow by origin_id failed")
		return
	}
	if flowIns.IsZero() {
		fService.Logger.Warningf(logTags, "get latest draft_flow by origin_id match no record")
		web.WriteSucResp(&w, r, nil)
		return
	}

	retFlow := fromAggWithLatestRunFlowView(flowIns, reqUser)
	if !retFlow.Read {
		fService.Logger.Warningf(logTags, "user has no read permission of this draft flow")
		web.WritePermissionNotEnough(&w, r, "user have no read permission on this flow")
		return
	}
	fService.Logger.Infof(logTags, "finished")

	web.WriteSucResp(&w, r, retFlow)
}

func FilterDraftByName(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "filter draft flow by name"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(
			logTags,
			"failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil,
			"get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	nameContains := r.URL.Query().Get("name__contains")

	aggSlice, err := fService.Flow.FilterDraft(reqUser.ID, nameContains)
	if err != nil {
		fService.Logger.Errorf(
			logTags,
			"filter draft flow by name__contains:%s error: %v",
			nameContains, err)
		web.WriteInternalServerErrorResp(&w, r, err, "visit repository failed")
		return
	}
	ret := fromAggSlice(aggSlice, reqUser)

	fService.Logger.Infof(
		logTags,
		"filter draft flow by name__contains:%s getted amount: %d",
		nameContains, len(ret))
	web.WriteSucResp(&w, r, ret)
}

// CreateDraft
func CreateDraft(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "create draft flow"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(
			logTags,
			"failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil,
			"get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	var reqFlow Flow
	err := json.NewDecoder(r.Body).Decode(&reqFlow)
	if err != nil {
		fService.Logger.Errorf(
			logTags, "json unmarshal req body to flow failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, err.Error())
		return
	}

	reqFlow.CreateUserID = reqUser.ID

	// 保存, 特别说明：这里没有检测OriginID是否为空是因为
	// 1. 若为空：表示新建
	// 2. 不为空：表示创建的已有flow的修改
	if !reqFlow.OriginID.IsNil() {
		arr, err := fService.Flow.GetDraftByOriginID(reqFlow.OriginID)
		if err != nil {
			fService.Logger.Errorf(
				logTags, "get draft flow by origin_id:%s failed: %v",
				reqFlow.OriginID.String(), err)
			web.WriteInternalServerErrorResp(&w, r, err, "GetDraftByOriginID error")
			return
		}
		// 相同origin_id的draft只能创建一个，不能在已有的情况下再次创建
		if !arr.IsZero() {
			fService.Logger.Errorf(
				logTags,
				"the same origin_id:%s draft flow already exist! cannot create another one.",
				reqFlow.OriginID.String())
			web.WriteBadRequestDataResp(&w, r, "")
			return
		}

		flowIns, err := fService.Flow.CreateDraftFromExistFlow(
			reqFlow.Name,
			reqFlow.CreateUserID, reqFlow.OriginID,
			reqFlow.Position, reqFlow.getAggregateFlowFunctionIDMapFlowFunction())

		if err != nil {
			fService.Logger.Errorf(logTags, "create draft flow failed: %v", err)
			web.WriteInternalServerErrorResp(&w, r, err, "create draft flow error")
			return
		}

		web.WriteSucResp(&w, r, fromAggWithoutUserPermission(flowIns))
		return
	}

	// > 新建
	flowIns, err := fService.Flow.CreateDraftFromScratch(
		reqFlow.Name, reqFlow.CreateUserID,
		reqFlow.Position, reqFlow.getAggregateFlowFunctionIDMapFlowFunction())
	if err != nil {
		fService.Logger.Errorf(logTags, "create draft flow error: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "create flow error")
		return
	}

	fService.Logger.Infof(logTags, "finished create draft flow. id:%s", reqUser.Name, flowIns.ID.String())
	web.WriteSucResp(&w, r, fromAggWithoutUserPermission(flowIns))
}

// PubDraft 提交草稿上线
func PubDraft(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "publish draft flow"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(
			logTags, "failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil, "get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	// 从传入获取id
	draftFlowID := ps.ByName("id")
	draftFlowUUID, err := web.ParseStrValueToUUID("id", draftFlowID)
	if err != nil {
		fService.Logger.Errorf(
			logTags,
			"parse draft_flow_id in query param:%s to uuid failed: %v", draftFlowID, err)
		web.WriteBadRequestDataResp(&w, r, err.Error())
		return
	}
	logTags["draft_flow_id"] = draftFlowID

	// 从id获取到flowIns
	draftFlowIns, err := fService.Flow.GetByID(draftFlowUUID)
	if err != nil {
		fService.Logger.Errorf(logTags, "find draft flow by id failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "find draft flow by id failed")
		return
	}
	if draftFlowIns.IsZero() {
		fService.Logger.Warningf(logTags, "find draft flow by id find no record")
		web.WriteBadRequestDataResp(&w, r, "id find no draft flow")
		return
	}
	if !draftFlowIns.IsDraft {
		fService.Logger.Errorf(logTags, "find flow but not draft!")
		web.WriteBadRequestDataResp(&w, r, "id find no draft flow")
		return
	}
	logTags["draft_flow_name"] = draftFlowIns.Name

	// 权限检查
	if !draftFlowIns.UserCanWrite(reqUser) {
		fService.Logger.Errorf(logTags, "need write permission")
		web.WritePermissionNotEnough(&w, r, "need write permission to pub draft flow")
	}
	draftFlowIns.CreateUserID = reqUser.ID

	// 正式提交的需要做有效性检测
	startFlowBloc, ok := draftFlowIns.FlowFunctionIDMapFlowFunction[config.FlowFunctionStartID]
	if !ok {
		msg := "failed flow valid check: miss start function"
		fService.Logger.Warningf(logTags, msg)
		web.WriteBadRequestDataResp(&w, r, msg)
		return
	}
	if len(startFlowBloc.DownstreamFlowFunctionIDs) <= 0 {
		msg := "failed flow valid check: not allowed create flow without function"
		fService.Logger.Warningf(logTags, msg)
		web.WriteBadRequestDataResp(&w, r, msg)
		return
	}

	funcIDMapFunction, err := fService.Function.IDMapFunctionAll()
	if err != nil {
		msg := "failed flow valid check: visit function failed(used to complete reference)"
		fService.Logger.Errorf(logTags, msg)
		web.WriteBadRequestDataResp(&w, r, msg)
		return
	}
	// 检查FlowFunctionIDMapFlowFunction下面配置的每个function_id是否正确
	// 同时也将查到的function赋予到对应的值，方便后续的连接类型有效性检查
	for flowFuncID, flowFunc := range draftFlowIns.FlowFunctionIDMapFlowFunction {
		if flowFuncID == config.FlowFunctionStartID {
			continue
		}
		function, ok := funcIDMapFunction[flowFunc.FunctionID]
		if !ok {
			msg := fmt.Sprintf(
				"function:%s's function_id: %s cannot find corresponding function",
				flowFunc.Note, flowFunc.FunctionID)
			fService.Logger.Errorf(logTags, msg)
			web.WriteBadRequestDataResp(&w, r, msg)
			return
		}
		flowFunc.Function = function
	}

	// 具体开始检查每个节点的各项配置是否正确/有效
	for flowFuncID, flowFunc := range draftFlowIns.FlowFunctionIDMapFlowFunction {
		valid, err := flowFunc.CheckValid(
			flowFuncID,
			draftFlowIns.FlowFunctionIDMapFlowFunction,
		)
		if !valid {
			msg := fmt.Sprintf("function:%s failed valid check: %v", flowFunc.Note, err)
			fService.Logger.Errorf(logTags, msg)
			web.WriteBadRequestDataResp(&w, r, msg)
			return
		}
	}

	// 通过有效性测试，开始创建
	aggF, err := fService.Flow.CreateOnlineFromDraft(draftFlowIns)
	if err != nil {
		fService.Logger.Errorf(logTags, "publish failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "publish failed")
	}
	fService.Logger.Infof(logTags, "suc published draft")

	deleteAmount, err := fService.Flow.DeleteDraftByOriginID(draftFlowIns.OriginID)

	fService.Logger.Infof(logTags, "finished delete draft amount: %d, error: %v", deleteAmount, err)
	web.WriteSucResp(&w, r, fromAggWithoutUserPermission(aggF))
}

// Update 用户前端更新的draft_flow,此时id不是origin_id
func UpdateDraft(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "update draft flow"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(
			logTags,
			"failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil,
			"get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	var reqFlow Flow
	err := json.NewDecoder(r.Body).Decode(&reqFlow)
	if err != nil {
		fService.Logger.Errorf(
			logTags, "json unmarshal to flow failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, err.Error())
		return
	}

	// > id不能为空
	if reqFlow.ID.IsNil() {
		fService.Logger.Errorf(logTags, "draft flow's id cannot be null")
		web.WriteBadRequestDataResp(&w, r, "must have id field")
		return
	}

	flowIns, err := fService.Flow.GetByID(reqFlow.ID)
	logTags["draft_flow_id"] = reqFlow.ID.String()
	if err != nil {
		fService.Logger.Errorf(
			logTags, "get draft flow by id failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "get flow by id failed")
		return
	}
	if flowIns.IsZero() {
		fService.Logger.Warningf(
			logTags, "get draft flow by id match no record")
		web.WriteBadRequestDataResp(&w, r, "find no flow by this id")
		return
	}
	if !flowIns.IsDraft {
		fService.Logger.Warningf(
			logTags, "the getted flow is not draft at all")
		web.WriteBadRequestDataResp(&w, r, "such id flow is not draft!")
		return
	}

	// > 检测当前用户是否有update此flow的权限
	if !flowIns.UserCanWrite(reqUser) {
		fService.Logger.Infof(
			logTags, "user have no write permission to update this draft flow")
		web.WritePermissionNotEnough(&w, r, "need write permission to update")
	}

	// > 开始更新相应字段
	if reqFlow.Name != "" && reqFlow.Name != flowIns.Name {
		err := fService.Flow.PatchName(reqFlow.ID, reqFlow.Name)
		baseLogMsg := fmt.Sprintf(
			"change draft flow's name from:%s to:%s",
			flowIns.Name, reqFlow.Name)

		if err != nil {
			fService.Logger.Errorf(logTags, "%s. error: %v", baseLogMsg, err)
			web.WriteInternalServerErrorResp(&w, r, err, "update name failed")
			return
		}
		fService.Logger.Infof(logTags, baseLogMsg)
	}

	if reqFlow.Position != nil {
		err := fService.Flow.PatchPosition(reqFlow.ID, reqFlow.Position)
		baseLogMsg := "change draft flow's position"
		if err != nil {
			fService.Logger.Errorf(logTags, "%s. error: %v", baseLogMsg, err)
			web.WriteInternalServerErrorResp(&w, r, err, "update position failed")
			return
		}
		fService.Logger.Infof(logTags, baseLogMsg)
	}

	// 若是修改节点构成，需要再次进行**粗略的**有效性检测（草稿不需要太严格的检验）
	if len(reqFlow.FlowFunctionIDMapFlowFunction) != 0 {
		_, ok := reqFlow.FlowFunctionIDMapFlowFunction[config.FlowFunctionStartID]
		if !ok {
			web.WriteBadRequestDataResp(&w, r, "Lack start bloc")
			return
		}
		err := fService.Flow.PatchFlowFunctionIDMapFlowFunction(
			reqFlow.ID, reqFlow.getAggregateFlowFunctionIDMapFlowFunction())
		baseLogMsg := fmt.Sprintf(
			"change draft flow's FlowFunctionIDMapFlowFunction: %v",
			reqFlow.FlowFunctionIDMapFlowFunction)
		if err != nil {
			fService.Logger.Errorf(logTags, "%s. error: %v", baseLogMsg, err)
			web.WriteInternalServerErrorResp(&w, r, err, "update function composition failed")
			return
		}
		fService.Logger.Infof(logTags, baseLogMsg)
	}

	fService.Logger.Infof(logTags, "finished")
	web.WritePlainSucOkResp(&w, r)
}

// DeleteDraftByOriginID 只有对此draft_flow有delete权限的能够删除flow，通过originID全部删除
func DeleteDraftByOriginID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "delete draft flow"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(
			&w, r, nil,
			"get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	originID := ps.ByName("origin_id")
	if originID == "" {
		fService.Logger.Infof(
			logTags, "miss draft flow's origin_id param in get path")
		web.WriteBadRequestDataResp(&w, r, "origin_id param must exist")
	}
	logTags["origin_id"] = originID

	uuOriginID, err := value_object.ParseToUUID(originID)
	if err != nil {
		fService.Logger.Errorf(logTags, "pass origin_id to uuid failed")
		web.WriteBadRequestDataResp(&w, r, "parse origin_id to uuid failed: %v", err)
		return
	}

	flowIns, err := fService.Flow.GetDraftByOriginID(uuOriginID)
	if err != nil {
		fService.Logger.Errorf(
			logTags, "get draft flow by origin_id failed: %v", err)
		web.WriteInternalServerErrorResp(
			&w, r, nil, "visit flow repository failed")
		return
	}
	if flowIns.IsZero() {
		fService.Logger.Errorf(
			logTags, "get draft flow by origin_id match no record")
		web.WriteBadRequestDataResp(&w, r, "origin_id find no flow")
		return
	}

	if !flowIns.UserCanDelete(reqUser) {
		fService.Logger.Warningf(logTags, "user have no delete permission to delete it")
		web.WritePermissionNotEnough(
			&w, r, "only user with delete permission to this draft flow can delete")
		return
	}

	deleteCount, err := fService.Flow.DeleteByOriginID(uuOriginID)
	if err != nil {
		fService.Logger.Errorf(logTags, "delete draft flow failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "delete failed")
		return
	}

	fService.Logger.Infof(logTags, "finished with delete amount: %d", deleteCount)
	web.WriteDeleteSucResp(&w, r, deleteCount)
}
