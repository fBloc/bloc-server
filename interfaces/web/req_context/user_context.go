package req_context

import (
	"context"

	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/interfaces/web"
)

func GetReqUserFromContext(ctx context.Context) (*aggregate.User, bool) {
	user, ok := ctx.Value(web.RequestContextUserKey).(*aggregate.User)
	return user, ok
}
