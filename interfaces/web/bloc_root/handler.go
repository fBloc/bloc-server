package bloc_root

import (
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/julienschmidt/httprouter"
)

func HelloBloc(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	web.WriteSucResp(&w, "Welcome aboard! May the Bloc be with you ~_~")
}
