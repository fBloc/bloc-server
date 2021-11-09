package function

import (
	"net/http"

	"github.com/fBloc/bloc-backend-go/interfaces/web"

	"github.com/julienschmidt/httprouter"
)

func All(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
