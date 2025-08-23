package middleware

import (
	pb "base-server/api/gen/go/base_api/v1"
	"base-server/internal/biz"
	"base-server/internal/conf"
	"base-server/internal/data/ent"
	"base-server/internal/server/utils"
	"base-server/internal/tools"
	"context"
	"github.com/go-kratos/kratos/v2/log"
	middleware2 "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/mileusna/useragent"
	"io"
	"strings"
	"time"
)

// NewWhiteListMatcherLog Path http操作日志白名单.
func NewWhiteListMatcherLog() selector.MatchFunc {
	whiteList := make(map[string]struct{})
	whiteList["/basic-api/v1/server/file/upload"] = struct{}{}
	whiteList["/api.base_api.v1.Base/GetWalkRoute"] = struct{}{}
	whiteList["/api.base_api.v1.Base/IsMenuNameExists"] = struct{}{}
	whiteList["/api.base_api.v1.Base/IsMenuPathExists"] = struct{}{}
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		return true
	}
}

// MiddlewareOperateLog 操作日志记录
func MiddlewareOperateLog(repo biz.BaseRepo, ac *conf.Auth) middleware2.Middleware {
	return selector.Server(
		MiddlewareHttpLog(repo, ac),
	).Match(NewWhiteListMatcherLog()).Build()
}

func MiddlewareHttpLog(repo biz.BaseRepo, ac *conf.Auth) middleware2.Middleware {
	return func(handler middleware2.Handler) middleware2.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, handlerErr error) {
			startTime := time.Now()
			reply, handlerErr = handler(ctx, req)
			if tr, ok := transport.FromServerContext(ctx); ok {
				defer func() {
					if ht, ok := tr.(*http.Transport); ok {
						request := ht.Request()

						data, err := io.ReadAll(request.Body)
						if err != nil {
							log.Error("日志读取错误")
						}
						defer request.Body.Close()

						ua := useragent.Parse(request.UserAgent())

						statusCode, reason, success := utils.GetStatusCode(handlerErr)
						_, stack := utils.ExtractError(handlerErr)

						logInfo := &ent.SysLogRecord{
							UserID:      "",
							UserName:    "",
							IsLogin:     false,
							SessionID:   "",
							Method:      request.Method,
							Path:        request.URL.Path,
							RequestTime: startTime,
							IPAddress:   utils.GetClientRealIP(request),
							IPLocation:  utils.ClientIpToLocation(utils.GetClientRealIP(request)),
							Latency:     time.Since(startTime).Milliseconds(),
							Os:          ua.OS + " " + ua.OSVersion,
							Browser:     ua.Name + "/" + ua.Version,
							UserAgent:   request.UserAgent(),
							Header:      utils.MarshalToStr(request.Header),
							GetParams:   utils.MarshalToStr(request.URL.Query()),
							PostData:    string(data),
							ResCode:     statusCode,
							Reason:      reason,
							ResStatus:   success,
							Stack:       stack,
						}

						if tr.Operation() == "/api.base_api.v1.Base/Login" {
							name, err := utils.BindLoginRequest(data)
							if err != nil {
								return
							}
							logInfo.IsLogin = true
							logInfo.UserName = name

							if handlerErr == nil {
								reply, err := reply.(*pb.LoginReply)
								if !err {
									return
								}
								logInfo.UserID = reply.UserId
								logInfo.SessionID = tools.MD5(reply.AccessToken) // 取token的md5值
							}
						} else {
							// 解析 jwt
							auths := strings.SplitN(ht.RequestHeader().Get("Authorization"), " ", 2)
							if len(auths) == 2 && strings.EqualFold(auths[0], "Bearer") {
								jwtToken := auths[1]
								var tokenInfo *jwtv5.Token
								tokenInfo, err = jwtv5.Parse(jwtToken, func(token *jwtv5.Token) (interface{}, error) {
									return []byte(ac.ApiKey), nil
								})
								uid := (tokenInfo.Claims.(jwtv5.MapClaims))["user_id"].(string)

								logInfo.UserID = uid
								logInfo.SessionID = tools.MD5(jwtToken)
							}
						}

						// 写入数据库
						err = repo.CreateSysLog(ctx, logInfo)
						if err != nil {
							return
						}
					}
				}()
			}
			return
		}
	}
}

//// extractArgs returns the string of the req
//func extractArgs(req any) string {
//	if redacter, ok := req.(Redacter); ok {
//		return redacter.Redact()
//	}
//	if stringer, ok := req.(fmt.Stringer); ok {
//		return stringer.String()
//	}
//	return fmt.Sprintf("%+v", req)
//}
