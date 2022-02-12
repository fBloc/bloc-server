package function_run_record

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/internal/util"
	"github.com/fBloc/bloc-server/value_object"

	"github.com/pkg/errors"
)

type FunctionRunRecordFilter struct {
	FunctionID   string    `mapstructure:"function_id"`
	FlowID       string    `mapstructure:"flow_id"`
	FlowOriginID string    `mapstructure:"flow_origin_id"`
	Start        time.Time `mapstructure:"start"`
	End          time.Time `mapstructure:"end"`
	Suc          bool      `mapstructure:"suc"`
	Pass         bool      `mapstructure:"pass"`
	Canceled     bool      `mapstructure:"canceled"`
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
				if strings.EqualFold(val, value_object.Asc.String()) {
					filterOp.SetAsc()
				} else if strings.EqualFold(val, value_object.Desc.String()) {
					filterOp.SetDesc()
				}
			} else {
				filterKeyMapVal[key] = val
				filterKeyMapFilterInGetPath[key] = filterInGetPath
			}
		}
	}

	var fRRF FunctionRunRecordFilter
	err = util.DecodeMapToStructP(filterKeyMapVal, &fRRF)
	if err != nil {
		return nil, nil, errors.Wrap(err,
			"decode param to struct failed, maybe some filed's spell is wrong")
	}

	if fRRF.FunctionID != "" {
		filterInGetPath := filterKeyMapFilterInGetPath["function_id"]
		if filterInGetPath != web.FilterInGetPathEq {
			return nil, nil, errors.New("function_id only suport equal search")
		}
		functionUUID, err := value_object.ParseToUUID(fRRF.FunctionID)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parse function_id to uuid failed")
		}
		filterInGetPath.AddToRepositoryFilter(filter, "function_id", functionUUID)
	}

	if fRRF.FlowID != "" {
		filterInGetPath := filterKeyMapFilterInGetPath["flow_id"]
		if filterInGetPath != web.FilterInGetPathEq {
			return nil, nil, errors.New("flow_id only suport equal search")
		}
		flowUUID, err := value_object.ParseToUUID(fRRF.FlowID)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parse flow_id to uuid failed")
		}
		filterInGetPath.AddToRepositoryFilter(filter, "flow_id", flowUUID)
	}

	if fRRF.FlowOriginID != "" {
		filterInGetPath := filterKeyMapFilterInGetPath["flow_origin_id"]
		if filterInGetPath != web.FilterInGetPathEq {
			return nil, nil, errors.New("flow_origin_id only suport equal search")
		}
		flowOriginUUID, err := value_object.ParseToUUID(fRRF.FlowID)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parse flow_origin_id to uuid failed")
		}
		filterInGetPath.AddToRepositoryFilter(filter, "flow_origin_id", flowOriginUUID)
	}

	if !fRRF.Start.IsZero() {
		filterInGetPath := filterKeyMapFilterInGetPath["start"]
		filterInGetPath.AddToRepositoryFilter(filter, "start", fRRF.Start)
	}

	if !fRRF.End.IsZero() {
		filterInGetPath := filterKeyMapFilterInGetPath["end"]
		filterInGetPath.AddToRepositoryFilter(filter, "end", fRRF.End)
	}

	if _, ok := filterKeyMapVal["suc"]; ok {
		filterInGetPath := filterKeyMapFilterInGetPath["suc"]
		filterInGetPath.AddToRepositoryFilter(filter, "suc", fRRF.Canceled)
	}

	if _, ok := filterKeyMapVal["pass"]; ok {
		filterInGetPath := filterKeyMapFilterInGetPath["pass"]
		filterInGetPath.AddToRepositoryFilter(filter, "pass", fRRF.Canceled)
	}

	if _, ok := filterKeyMapVal["canceled"]; ok {
		filterInGetPath := filterKeyMapFilterInGetPath["canceled"]
		filterInGetPath.AddToRepositoryFilter(filter, "canceled", fRRF.Canceled)
	}

	return filter, filterOp, nil
}
