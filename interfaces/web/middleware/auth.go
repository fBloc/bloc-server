package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/interfaces/web"
	user_cache "github.com/fBloc/bloc-server/services/userid_cache"

	"github.com/julienschmidt/httprouter"
)

var userCache *user_cache.UserCacheService

func InjectUserIDCacheService(s *user_cache.UserCacheService) {
	userCache = s
}

func setUserToContext(r *http.Request, user *aggregate.User) *http.Request {
	ctx := context.WithValue(r.Context(), web.RequestContextUserKey, user)
	r = r.WithContext(ctx)
	return r
}

func getUserFromService(token string) (*aggregate.User, error) {
	if token == "" {
		return nil, errors.New("token for auth in req header is empty")
	}
	userIns, err := userCache.GetUserByIDString(token)
	if err != nil {
		return nil, err
	}
	if userIns.IsZero() {
		return nil, errors.New("token find no record")
	}
	return &userIns, nil
}

// LoginAuth 检测需要登录
func LoginAuth(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		user, err := getUserFromService(r.Header.Get("token"))
		if err != nil {
			web.WriteInternalServerErrorResp(&w, err, "")
			return
		}
		if user == nil {
			web.WriteNeedLogin(&w)
			return
		}

		r = setUserToContext(r, user)
		h(w, r, ps)
	}
}

// SuperuserAuth 检测登录用户需要时super_user
func SuperuserAuth(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		user, err := getUserFromService(r.Header.Get("token"))
		if err != nil {
			web.WriteInternalServerErrorResp(&w, err, "")
			return
		}
		if user == nil {
			web.WriteNeedLogin(&w)
			return
		}
		if !user.IsSuper {
			web.WriteNeedSuperUser(&w)
			return
		}

		r = setUserToContext(r, user)
		h(w, r, ps)
	}
}
