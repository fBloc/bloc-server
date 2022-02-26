package web

import (
	"context"

	"github.com/fBloc/bloc-server/aggregate"
)

func GetReqUserFromContext(ctx context.Context) (*aggregate.User, bool) {
	user, ok := ctx.Value(RequestContextUserKey).(*aggregate.User)
	return user, ok
}
