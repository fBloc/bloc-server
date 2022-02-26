package function

import (
	"net/http"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/interfaces/web"

	"github.com/julienschmidt/httprouter"
)

func All(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "filter function"

	reqUser, suc := web.GetReqUserFromContext(r.Context())
	if !suc {
		fService.Logger.Errorf(logTags, "failed to get user from context which should be setted by middleware!")
		web.WriteInternalServerErrorResp(&w, r, nil, "get requser from context failed")
		return
	}
	logTags["user_name"] = reqUser.Name

	var err error
	var aggFs []*aggregate.Function
	if reqUser.IsSuper {
		aggFs, err = fService.Function.All()
	} else {
		aggFs, err = fService.Function.UserReadAbleAll(reqUser)
	}
	if err != nil {
		fService.Logger.Errorf(logTags, "get repository failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "visit repository failed")
		return
	}

	groupedFuncs, err := newGroupedFunctionsFromAggFunctions(aggFs)
	if err != nil {
		fService.Logger.Errorf(logTags, "trans to resp failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "trans to resp failed")
		return
	}

	fService.Logger.Infof(logTags, "finished with amount: %d", len(aggFs))
	web.WriteSucResp(&w, r, groupedFuncs)
}
