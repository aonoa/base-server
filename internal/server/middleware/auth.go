package middleware

import (
	"base-server/internal/biz"
	"base-server/internal/conf"
	"context"
	"github.com/casbin/casbin/v2"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	middleware2 "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwt2 "github.com/golang-jwt/jwt/v5"
)

// NewWhiteListMatcher Path 白名单.
func NewWhiteListMatcher() selector.MatchFunc {
	whiteList := make(map[string]struct{})
	whiteList["/api.base_api.v1.Base/Login"] = struct{}{}
	whiteList["/api.base_api.v1.Base/Logout"] = struct{}{}
	whiteList["/basic-api/v1/server/file/upload"] = struct{}{}
	whiteList["/api.base_api.v1.Base/IsMenuNameExists"] = struct{}{}
	whiteList["/api.base_api.v1.Base/IsMenuPathExists"] = struct{}{}
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		return true
	}
}

// MiddlewareAuth Jwt Auth.
func MiddlewareAuth(ac *conf.Auth, e *casbin.Enforcer, logger log.Logger) middleware2.Middleware {
	return selector.Server(
		jwt.Server(
			func(token *jwt2.Token) (interface{}, error) {
				return []byte(ac.ApiKey), nil
			},
			jwt.WithSigningMethod(jwt2.SigningMethodHS256),
			jwt.WithClaims(func() jwt2.Claims {
				return &jwt2.MapClaims{}
			}),
		),
		MiddlewareCasbin(e, logger),
	).Match(NewWhiteListMatcher()).Build()
}

func MiddlewareCasbin(e *casbin.Enforcer, logger log.Logger) middleware2.Middleware {
	log := log.NewHelper(logger)
	//enforceContext := casbin.EnforceContext{RType: "r", PType: "p2", EType: "e", MType: "m2"}
	return func(handler middleware2.Handler) middleware2.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			uid := ""
			aud := ""
			if claims, ok := jwt.FromContext(ctx); ok {
				uid = (*claims.(*jwt2.MapClaims))["user_id"].(string)
				aud = (*claims.(*jwt2.MapClaims))["aud"].(string)
				ctx = context.WithValue(ctx, "uid", uid)
			} else {
				return nil, errors.Unauthorized("UNAUTHORIZED", "uid is missing")
			}
			if tr, ok := transport.FromServerContext(ctx); ok {
				// 断言成HTTP的Transport可以拿到特殊信息
				if (aud == "refresh") && (tr.Operation() != "/api.base_api.v1.Base/RefreshToken") {
					// refreshToken只能用来刷新token
					return nil, errors.Unauthorized("UNAUTHORIZED", "Authentication failed")
				}
				if ht, ok := tr.(*http.Transport); ok {
					//enforceContext := casbin.EnforceContext{RType: "r", PType: "p2", EType: "e", MType: "m1"}
					log.Infof("uid:%s, %s", uid, ht.Request().Method+":"+ht.Request().RequestURI)
					//ok, err := e.Enforce(enforceContext, uid, ht.Request().Method+":"+ht.Request().RequestURI)
					ok, err := e.Enforce(biz.RoleToApiEnforceContext, uid, ht.Request().RequestURI, ht.Request().Method)
					//ok, err := e.Enforce(uid, ht.Request().Method+":"+ht.Request().RequestURI, "dom:default")
					if err != nil || !ok {
						// 拒绝请求，抛出异常
						return nil, errors.Forbidden("Forbidden", "Authentication failed")
					}
				}
			}
			return handler(ctx, req)
		}
	}
}
