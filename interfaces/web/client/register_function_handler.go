package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/pkg/ipt"
	"github.com/fBloc/bloc-server/pkg/opt"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/julienschmidt/httprouter"
)

func RegisterFunctions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req RegisterFuncReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(&w, err.Error())
		return
	}

	var wg sync.WaitGroup
	toReportAliveFuncIDs := make(chan value_object.UUID, 40) // 控制下并发在40（别太高）
	for groupName, funcs := range req.GroupNameMapFunctions {
		for _, f := range funcs {
			// 检测是否汇报过
			funcNameMapFunc, ok := reported.groupNameMapFuncNameMapFunc[groupName]
			if !ok {
				reported.Lock()
				reported.groupNameMapFuncNameMapFunc[groupName] = make(map[string]*reportFunction)
				reported.Unlock()
			} else {
				reportedFunc, ok := funcNameMapFunc[f.Name]
				if ok { // 汇报过，直接利用信息
					if req.Who != reportedFunc.ProviderName { // 不允许不同来源的provider创建group_name和name完全相同的function
						web.WriteBadRequestDataResp(
							&w, fmt.Sprintf(
								`another function provider %s already created %s-%s function.
								not allowed create same group_name-name function from different consumer source`,
								req.Who, groupName, f.Name))
						return
					}
					f.ID = reportedFunc.ID
					reportedFunc.LastReportTime = time.Now()
					toReportAliveFuncIDs <- f.ID
					continue
				}
			}

			wg.Add(1)
			go func(group string, httpFunc *HttpFunction, wg *sync.WaitGroup) {
				defer wg.Done()
				// 没有汇报过
				iptD := ipt.GenIptDigest(httpFunc.Ipts)
				optD := opt.GenOptDigest(httpFunc.Opts)
				aggFunc, err := fService.Function.GetSameIptOptFunction(
					iptD, optD)
				if err != nil {
					errMsgPrefix := fmt.Sprintf(
						`get function by same core failed. 
					group_name: %s, function_name: %s, ipt_digest: %s, opt_digest: %s`,
						group, httpFunc.Name, iptD, optD)
					fService.Logger.Errorf("%s:%s", errMsgPrefix, err.Error())
					web.WriteInternalServerErrorResp(
						&w, err, errMsgPrefix)
				}

				if aggFunc.IsZero() { // 没汇报过 + 没查询到记录.表示是第一次汇报。需要持久化存储
					aggFunction := aggregate.Function{
						ID:            value_object.NewUUID(),
						Name:          httpFunc.Name,
						GroupName:     group,
						ProviderName:  req.Who,
						Description:   httpFunc.Description,
						Ipts:          httpFunc.Ipts,
						IptDigest:     iptD,
						Opts:          httpFunc.Opts,
						OptDigest:     optD,
						ProcessStages: httpFunc.ProcessStages}
					err = fService.Function.Create(&aggFunction)
					if err != nil {
						msg := fmt.Sprintf("create function to persistence layer failed: %s", err.Error())
						fService.Logger.Errorf(msg)
						httpFunc.ErrorMsg = msg
						return
					}
					httpFunc.ID = aggFunction.ID
					aggFunc = &aggFunction
				} else { // 没汇报过 + 查到了记录
					httpFunc.ID = aggFunc.ID
					if aggFunc.ProviderName == "" || aggFunc.ProviderName != req.Who {
						err := fService.Function.PatchProviderName(httpFunc.ID, req.Who)
						if err != nil {
							panic(err)
						}
					}
				}

				toReportAliveFuncIDs <- httpFunc.ID

				// 加入到本地汇报缓存
				reported.Lock()
				reported.groupNameMapFuncNameMapFunc[group][httpFunc.Name] = &reportFunction{
					ProviderName: req.Who, GroupName: group, Name: httpFunc.Name,
					ID: httpFunc.ID, LastReportTime: time.Now()}
				reported.idMapFunc[httpFunc.ID] = *aggFunc
				reported.Unlock()
				fService.Logger.Infof("registered func: %s - %s", group, httpFunc.Name)
			}(groupName, f, &wg)
		}
	}

	go func() {
		for functionUUID := range toReportAliveFuncIDs {
			err := fService.Function.AliveReport(functionUUID)
			if err != nil {
				fService.Logger.Errorf(
					"function(id: %s) alive report failed: %s",
					functionUUID.String(), err.Error())
			}
		}
	}()

	wg.Wait()
	close(toReportAliveFuncIDs)

	for _, funcs := range req.GroupNameMapFunctions {
		for _, f := range funcs {
			if f.ErrorMsg != "" {
				web.WriteInternalServerErrorResp(&w, nil, f.ErrorMsg)
				return
			}
		}
	}

	web.WriteSucResp(&w, req)
}
