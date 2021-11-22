package flow_run_record

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fBloc/bloc-backend-go/interfaces/web"
	"github.com/fBloc/bloc-backend-go/internal/util"
	"github.com/fBloc/bloc-backend-go/value_object"

	"github.com/pkg/errors"
)

/*
普遍使用参数：
1. limit
2. offset
3. 时间范围

常见使用场景：
1. flow_origin_id: 以flow的origin_id作为查找依据，返回的就是此flow的[无论此记录是直接运行/flow配置的运行/arrangement的运行都会出现]（不过可能包含老版本flow的运行历史）
2. arrangement_flow_id： 此返回的是特定arrangement下的此flow的运行记录，应该只在arrangement作为入口查看历史的时候才会用得到
3. flow_id： 以flow的id作为参数，表示获取特定版本的flow的运行历史
*/

type FlowFunctionRecordFilter struct {
	FlowIDStr       string                `mapstructure:"flow_id"`
	FlowOriginIDStr string                `mapstructure:"flow_origin_id"`
	TriggerTime     time.Time             `mapstructure:"trigger_time"`
	StartTime       time.Time             `mapstructure:"start_time"`
	EndTime         time.Time             `mapstructure:"end_time"`
	State           value_object.RunState `mapstructure:"status"`
	TimeoutCanceled bool                  `mapstructure:"timeout_canceled"`
	Canceled        bool                  `mapstructure:"canceled"`
}

func BuildFromWebRequestParams(
	query url.Values,
) (*value_object.RepositoryFilter, *value_object.RepositoryFilterOption, error) {
	filterInGetPathMapReqKeys, err := web.ParseReqQueryToGroupedFilters(query)
	if err != nil {
		return nil, nil, err
	}

	filter := value_object.NewRepositoryFilter()
	filterOp := value_object.NewRepositoryFilterOption()
	filterKeyMapVal := make(map[string]interface{})
	filterKeyMapFilterInGetPath := make(map[string]web.FilterInGetPath)
	for filterInGetPath, reqKeys := range filterInGetPathMapReqKeys {
		for _, key := range reqKeys {
			var completeKey string
			if strings.HasPrefix(filterInGetPath.String(), "__") {
				completeKey = fmt.Sprintf("%s%s", key, filterInGetPath.String())
			} else {
				completeKey = key
			}
			val := query.Get(completeKey)
			if filterInGetPath == web.FilterInGetPathLimit {
				intVal, err := strconv.Atoi(val)
				if err != nil {
					return nil, nil, errors.New(
						web.FilterInGetPathLimit.String() + " field value must be int")
				}
				filterOp.SetLimit(intVal)
			} else if filterInGetPath == web.FilterInGetPathOffset {
				intVal, err := strconv.Atoi(val)
				if err != nil {
					return nil, nil, errors.New(
						web.FilterInGetPathOffset.String() + " field value must be int")
				}
				filterOp.SetOffset(intVal)
			} else if filterInGetPath == web.FilterInGetPathSort {
				// TODO 这里不应该写死 asc/desc
				if strings.EqualFold(val, "asc") {
					filterOp.SetAsc()
				} else if strings.EqualFold(val, "desc") {
					filterOp.SetDesc()
				}
			} else {
				filterKeyMapVal[key] = val
				filterKeyMapFilterInGetPath[key] = filterInGetPath
			}
		}
	}

	var fFRR FlowFunctionRecordFilter
	err = util.DecodeMapToStructP(filterKeyMapVal, &fFRR)
	if err != nil {
		return nil, nil, errors.Wrap(err,
			"decode param to struct failed, maybe some filed's spell is wrong")
	}

	if fFRR.FlowIDStr != "" {
		filterInGetPath := filterKeyMapFilterInGetPath["flow_id"]
		if filterInGetPath != web.FilterInGetPathEq {
			return nil, nil, errors.New("flow_id only suport equal search")
		}
		flowUUID, err := value_object.ParseToUUID(fFRR.FlowIDStr)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parse flow_id to uuid failed")
		}
		filterInGetPath.AddToRepositoryFilter(filter, "flow_id", flowUUID)
	}

	if fFRR.FlowOriginIDStr != "" {
		filterInGetPath := filterKeyMapFilterInGetPath["flow_origin_id"]
		if filterInGetPath != web.FilterInGetPathEq {
			return nil, nil, errors.New("flow_origin_id only suport equal search")
		}
		originUUID, err := value_object.ParseToUUID(fFRR.FlowOriginIDStr)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parse flow_origin_id to uuid failed")
		}
		filterInGetPath.AddToRepositoryFilter(filter, "flow_origin_id", originUUID)
	}

	if !fFRR.TriggerTime.IsZero() {
		filterInGetPath := filterKeyMapFilterInGetPath["trigger_time"]
		filterInGetPath.AddToRepositoryFilter(filter, "trigger_time", fFRR.TriggerTime)
	}

	if !fFRR.StartTime.IsZero() {
		filterInGetPath := filterKeyMapFilterInGetPath["start_time"]
		filterInGetPath.AddToRepositoryFilter(filter, "start_time", fFRR.StartTime)
	}

	if !fFRR.EndTime.IsZero() {
		filterInGetPath := filterKeyMapFilterInGetPath["end_time"]
		filterInGetPath.AddToRepositoryFilter(filter, "end_time", fFRR.EndTime)
	}

	if fFRR.State.IsRunStateValid() {
		filterInGetPath := filterKeyMapFilterInGetPath["status"]
		filterInGetPath.AddToRepositoryFilter(filter, "status", fFRR.State)
	}

	if _, ok := filterKeyMapVal["canceled"]; ok {
		filterInGetPath := filterKeyMapFilterInGetPath["canceled"]
		filterInGetPath.AddToRepositoryFilter(filter, "canceled", fFRR.Canceled)
	}

	if _, ok := filterKeyMapVal["timeout_canceled"]; ok {
		filterInGetPath := filterKeyMapFilterInGetPath["timeout_canceled"]
		filterInGetPath.AddToRepositoryFilter(filter, "timeout_canceled", fFRR.Canceled)
	}

	return filter, filterOp, nil
}
