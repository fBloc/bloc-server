package req_context

import (
	"context"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/interfaces/web"
)

func GetReqUserFromContext(ctx context.Context) (*aggregate.User, bool) {
	user, ok := ctx.Value(web.RequestContextUserKey).(*aggregate.User)
	return user, ok
}
