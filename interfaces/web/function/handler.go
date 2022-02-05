package function

import (
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"

	"github.com/julienschmidt/httprouter"
)

func All(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// TODO 正确处理权限问题，superuser才能获取全部的functions，其他的用户需要受到权限限制
	aggFs, err := fService.Function.All()
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit repository failed")
		return
	}

	groupedFuncs, err := newGroupedFunctionsFromAggFunctions(aggFs)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err,
			"trans to resp failed")
		return
	}

	web.WriteSucResp(&w, groupedFuncs)
}
