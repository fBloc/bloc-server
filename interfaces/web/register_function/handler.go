package register_function

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/interfaces/web"
	"github.com/fBloc/bloc-backend-go/pkg/ipt"
	"github.com/fBloc/bloc-backend-go/pkg/opt"
	"github.com/fBloc/bloc-backend-go/value_object"
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
	for groupName, funcs := range req.GroupNameMapFuncNameMapFunc {
		for _, f := range funcs {
			// 检测是否汇报过
			funcNameMapFunc, ok := groupNameMapFuncNameMapFunc[groupName]
			if !ok {
				groupNameMapFuncNameMapFunc[groupName] = make(map[string]*reportFunction)
			} else {
				reportedFunc, ok := funcNameMapFunc[f.Name]
				if ok { // 汇报过，直接利用信息
					if req.Flag != groupName { // 不允许不同的consumer来源创建group_name和name完全相同的function
						web.WriteBadRequestDataResp(
							&w, fmt.Sprintf(
								`another consumer source %s already created %s-%s function.
								not allowed create same group_name-name function from different consumer source`,
								req.Flag, groupName, f.Name))
						return
					}
					f.ID = reportedFunc.ID
					reportedFunc.LastReportTime = time.Now()
					continue
				}
			}

			wg.Add(1)
			go func(string, *HttpFunction, *sync.WaitGroup) {
				defer wg.Done()
				// 没有汇报过
				iptD, optD := ipt.GenIptDigest(f.Ipts), opt.GenOptDigest(f.Opts)
				aggFunc, err := fService.Function.GetSameIptOptFunction(
					iptD, optD)
				if err != nil {
					errMsgPrefix := fmt.Sprintf(
						`get function by same core failed. 
					group_name: %s, function_name: %s, ipt_digest: %s, opt_digest: %s`,
						groupName, f.Name, iptD, optD)
					fService.Logger.Errorf("%s:%s", errMsgPrefix, err.Error())
					web.WriteInternalServerErrorResp(
						&w, err, errMsgPrefix)
				}

				// 没汇报过 + 没查询到记录.表示是第一次汇报。需要持久化存储
				if aggFunc.IsZero() {
					aggFunction := aggregate.Function{
						ID:            value_object.NewUUID(),
						Name:          f.Name,
						GroupName:     groupName,
						Description:   f.Description,
						Ipts:          f.Ipts,
						IptDigest:     iptD,
						Opts:          f.Opts,
						OptDigest:     optD,
						ProcessStages: f.ProcessStages}
					err = fService.Function.Create(&aggFunction)
					if err != nil {
						msg := fmt.Sprintf("create function to persistence layer failed: %s", err.Error())
						fService.Logger.Errorf(msg)
						f.ErrorMsg = msg
						return
					}
					f.ID = aggFunction.ID
				} else {
					f.ID = aggFunc.ID
				}

				// 加入到本地汇报缓存
				groupNameMapFuncNameMapFunc[groupName][f.Name] = &reportFunction{
					GroupName: groupName, Name: f.Name, ID: f.ID, LastReportTime: time.Now()}
				fService.Logger.Infof("registered func: %s - %s", groupName, f.Name)
			}(groupName, f, &wg)
		}
	}

	wg.Wait()

	for _, funcs := range req.GroupNameMapFuncNameMapFunc {
		for _, f := range funcs {
			if f.ErrorMsg != "" {
				web.WriteInternalServerErrorResp(&w, nil, f.ErrorMsg)
				return
			}
		}
	}

	web.WriteSucResp(&w, req)
}
