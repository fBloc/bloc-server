package flow

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fBloc/bloc-backend-go/config"
	"github.com/fBloc/bloc-backend-go/interfaces/web"
	"github.com/fBloc/bloc-backend-go/interfaces/web/req_context"
	"github.com/fBloc/bloc-backend-go/value_object"

	"github.com/julienschmidt/httprouter"
)

func GetDraftByOriginID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	originID := ps.ByName("origin_id")
	if originID == "" {
		web.WriteBadRequestDataResp(&w, "origin_id cannot be blank")
		return
	}
	originUUID, err := value_object.ParseToUUID(originID)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "cannot parse origin_id to uuid")
		return
	}

	flowIns, err := fService.Flow.GetDraftByOriginID(originUUID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "get latest draft_flow by origin_id failed")
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

	retFlow := fromAggWithLatestRunFlowView(flowIns, reqUser)
	if !retFlow.Read {
		web.WritePermissionNotEnough(&w, "user have no read permission on this flow")
		return
	}

	web.WriteSucResp(&w, retFlow)
}

func FilterDraftByName(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil,
			"get requser from context failed")
		return
	}

	aggSlice, err := fService.Flow.FilterDraft(
		reqUser.ID, r.URL.Query().Get("name__contains"))
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit repository failed")
		return
	}
	ret := fromAggSlice(aggSlice, reqUser)

	web.WriteSucResp(&w, ret)
}

// CreateDraft
func CreateDraft(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var reqFlow Flow
	err := json.NewDecoder(r.Body).Decode(&reqFlow)
	if err != nil {
		web.WriteBadRequestDataResp(&w, err.Error())
		return
	}

	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil,
			"get requser from context failed")
		return
	}

	reqFlow.CreateUserID = reqUser.ID

	// 保存, 特别说明：这里没有检测OriginID是否为空是因为
	// 1. 若为空：表示新建
	// 2. 不为空：表示创建的已有flow的修改
	if !reqFlow.OriginID.IsNil() {
		arr, err := fService.Flow.GetDraftByOriginID(reqFlow.OriginID)
		if err != nil {
			web.WriteInternalServerErrorResp(&w, err, "GetDraftByOriginID error")
			return
		}
		// 相同origin_id的draft只能创建一个，不能在已有的情况下再次创建
		if !arr.IsZero() {
			web.WriteBadRequestDataResp(&w, "draft只能创建一个，不能在已有的情况下再次创建")
			return
		}

		flowIns, err := fService.Flow.CreateDraftFromExistFlow(
			reqFlow.Name,
			reqFlow.CreateUserID, reqFlow.OriginID,
			reqFlow.Position, reqFlow.getAggregateFlowFunctionIDMapFlowFunction())

		if err != nil {
			web.WriteInternalServerErrorResp(&w, err, "create flow error")
			return
		}

		web.WriteSucResp(&w, fromAggWithoutUserPermission(flowIns))
		return
	}

	// > 新建
	flowIns, err := fService.Flow.CreateDraftFromScratch(
		reqFlow.Name, reqFlow.CreateUserID,
		reqFlow.Position, reqFlow.getAggregateFlowFunctionIDMapFlowFunction())
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "create flow error")
		return
	}
	web.WriteSucResp(&w, fromAggWithoutUserPermission(flowIns))
}

// PubDraft 提交草稿上线
func PubDraft(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 从传入获取id
	draftFlowID := ps.ByName("id")
	draftFlowUUID, err := web.ParseStrValueToUUID("id", draftFlowID)
	if err != nil {
		web.WriteBadRequestDataResp(&w, err.Error())
		return
	}

	// 从id获取到flowIns
	draftFlowIns, err := fService.Flow.GetByID(draftFlowUUID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "find flow by id failed")
		return
	}
	if draftFlowIns.IsZero() {
		web.WriteBadRequestDataResp(&w, "id find no draft flow")
		return
	}
	if !draftFlowIns.IsDraft {
		web.WriteBadRequestDataResp(&w, "id find no draft flow")
		return
	}

	// 权限检查
	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil,
			"get requser from context failed")
		return
	}
	if !draftFlowIns.UserCanWrite(reqUser) {
		web.WritePermissionNotEnough(&w, "need write permission to pub draft flow")
	}
	draftFlowIns.CreateUserID = reqUser.ID

	// 正式提交的需要做有效性检测
	startFlowBloc, ok := draftFlowIns.FlowFunctionIDMapFlowFunction[config.FlowFunctionStartID]
	if !ok {
		web.WriteBadRequestDataResp(&w, "有效性检测失败，缺少开始bloc")
		return
	}
	if len(startFlowBloc.DownstreamFlowFunctionIDs) <= 0 {
		web.WriteBadRequestDataResp(&w, "有效性检测失败，一个节点都没有，不允许创建")
		return
	}

	funcIDMapFunction, err := fService.Function.IDMapFunctionAll()
	if err != nil {
		web.WriteBadRequestDataResp(&w, "访问全部function信息（用于补全引用）时失败")
		return
	}
	for flowFuncID, flowFunc := range draftFlowIns.FlowFunctionIDMapFlowFunction {
		if flowFuncID == config.FlowFunctionStartID {
			continue
		}
		function, ok := funcIDMapFunction[flowFunc.FunctionID]
		if !ok {
			web.WriteBadRequestDataResp(&w, fmt.Sprintf(
				"ID为%s节点的function_id不对，找不到对应的function",
				flowFunc.FunctionID))
			return
		}
		flowFunc.Function = function
		valid, errorStr := flowFunc.CheckValid(
			draftFlowIns.FlowFunctionIDMapFlowFunction,
		)
		if !valid {
			web.WriteBadRequestDataResp(&w, errorStr)
			return
		}
	}

	// 通过有效性测试，开始创建
	aggF, err := fService.Flow.CreateOnlineFromDraft(draftFlowIns)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "")
	}
	fService.Flow.DeleteDraftByOriginID(draftFlowIns.OriginID)

	web.WriteSucResp(&w, fromAggWithoutUserPermission(aggF))
}

// Update 用户前端更新的draft_flow,此时id不是origin_id
func UpdateDraft(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var reqFlow Flow
	err := json.NewDecoder(r.Body).Decode(&reqFlow)
	if err != nil {
		web.WriteBadRequestDataResp(&w, err.Error())
		return
	}

	// > id不能为空
	if reqFlow.ID.IsNil() {
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
	if !flowIns.IsDraft {
		web.WriteBadRequestDataResp(&w, "such id flow is not draft!")
		return
	}

	// > 检测当前用户是否有update此flow的权限
	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil,
			"get requser from context failed")
		return
	}
	if !flowIns.UserCanWrite(reqUser) {
		web.WritePermissionNotEnough(&w, "need write permission to update")
	}

	// > 开始更新相应字段
	if reqFlow.Name != "" {
		flowIns.Name = reqFlow.Name
	}

	if reqFlow.Position != nil {
		flowIns.Position = reqFlow.Position
	}

	// 若是修改节点构成，需要再次进行**粗略的**有效性检测（草稿不需要太严格的检验）
	if len(reqFlow.FlowFunctionIDMapFlowFunction) != 0 {
		_, ok := reqFlow.FlowFunctionIDMapFlowFunction[config.FlowFunctionStartID]
		if !ok {
			web.WriteBadRequestDataResp(&w, "Lack start bloc")
			return
		}
		flowIns.FlowFunctionIDMapFlowFunction = reqFlow.getAggregateFlowFunctionIDMapFlowFunction()
	}

	// > 更新
	err = fService.Flow.ReplaceByID(reqFlow.ID, flowIns)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "update failed")
		return
	}

	web.WritePlainSucOkResp(&w)
}

// DeleteDraftByOriginID 只有对此draft_flow有delete权限的能够删除flow，通过originID全部删除
func DeleteDraftByOriginID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	originID := ps.ByName("origin_id")
	if originID == "" {
		web.WriteBadRequestDataResp(&w, "origin_id param must exist")
	}
	uuOriginID, err := value_object.ParseToUUID(originID)
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

	flowIns, err := fService.Flow.GetDraftByOriginID(uuOriginID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, nil,
			"visit flow repository failed")
		return
	}
	if flowIns.IsZero() {
		web.WriteBadRequestDataResp(&w, "origin_id find no flow")
		return
	}

	if !flowIns.UserCanDelete(reqUser) {
		web.WritePermissionNotEnough(
			&w,
			"only user with delete permission to this draft flow can delete")
		return
	}

	deleteCount, err := fService.Flow.DeleteByOriginID(uuOriginID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "delete failed")
		return
	}
	web.WriteDeleteSucResp(&w, deleteCount)
}
