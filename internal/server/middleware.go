package server

import (
	"base-server/internal/conf"
	"context"
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// NewWhiteListMatcher Path 白名单.
func NewWhiteListMatcher() selector.MatchFunc {
	whiteList := make(map[string]struct{})
	whiteList["/api.base_api.v1.Base/Login"] = struct{}{}
	whiteList["/api.base_api.v1.Base/Logout"] = struct{}{}
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		return true
	}
}

// MiddlewareAuth Jwt Auth.
func MiddlewareAuth(ac *conf.Auth, e *casbin.Enforcer) middleware.Middleware {
	return selector.Server(
		jwt.Server(
			func(token *jwtv5.Token) (interface{}, error) {
				return []byte(ac.ApiKey), nil
			},
			jwt.WithSigningMethod(jwtv5.SigningMethodHS256),
			jwt.WithClaims(func() jwtv5.Claims {
				return &jwtv5.MapClaims{}
			}),
		),
		MiddlewareCasbin(e),
	).Match(NewWhiteListMatcher()).Build()
}

func MiddlewareCasbin(e *casbin.Enforcer) middleware.Middleware {
	enforceContext := casbin.EnforceContext{RType: "r", PType: "p2", EType: "e", MType: "m2"}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			uid := ""
			aud := ""
			if claims, ok := jwt.FromContext(ctx); ok {
				uid = (*claims.(*jwtv5.MapClaims))["user_id"].(string)
				aud = (*claims.(*jwtv5.MapClaims))["aud"].(string)
				context.WithValue(ctx, "uid", uid)
			} else {
				return nil, errors.Unauthorized("UNAUTHORIZED", "uid is missing")
			}
			if tr, ok := transport.FromServerContext(ctx); ok {
				// 断言成HTTP的Transport可以拿到特殊信息
				fmt.Println(tr.Operation())
				if (aud == "refresh") && (tr.Operation() != "/api.base_api.v1.Base/RefreshToken") {
					// refreshToken只能用来刷新token
					return nil, errors.Unauthorized("UNAUTHORIZED", "Authentication failed")
				}
				if ht, ok := tr.(*http.Transport); ok {
					//enforceContext := casbin.EnforceContext{RType: "r", PType: "p2", EType: "e", MType: "m1"}
					fmt.Println(uid, ht.Request().Method+":"+ht.Request().RequestURI)
					//ok, err := e.Enforce(enforceContext, uid, ht.Request().Method+":"+ht.Request().RequestURI)
					ok, err := e.Enforce(enforceContext, uid, ht.Request().RequestURI, ht.Request().Method)
					//ok, err := e.Enforce(uid, ht.Request().Method+":"+ht.Request().RequestURI, "dom:default")
					if err != nil || !ok {
						// 拒绝请求，抛出异常
						return nil, errors.Unauthorized("UNAUTHORIZED", "Authentication failed")
					}
				}
			}
			return handler(ctx, req)
		}
	}
}

// MiddlewareDemo is middleware demo.
func MiddlewareDemo() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if _, ok := transport.FromServerContext(ctx); ok {
				// Do something on entering
				defer func() {
					// Do something on exiting
				}()
			}
			return handler(ctx, req)
		}
	}
}
