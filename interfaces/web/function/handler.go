package function

import (
	"net/http"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/interfaces/web/req_context"

	"github.com/julienschmidt/httprouter"
)

func All(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqUser, suc := req_context.GetReqUserFromContext(r.Context())
	if !suc {
		web.WriteInternalServerErrorResp(&w, nil, "get requser from context failed")
		return
	}

	var err error
	var aggFs []*aggregate.Function
	if reqUser.IsSuper {
		aggFs, err = fService.Function.All()
	} else {
		aggFs, err = fService.Function.UserReadAbleAll(reqUser)
	}
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
