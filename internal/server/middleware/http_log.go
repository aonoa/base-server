package middleware

import (
	"base-server/internal/server/utils"
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	middleware2 "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/mileusna/useragent"
	"io"
	"time"
)

// NewWhiteListMatcherLog Path http操作日志白名单.
func NewWhiteListMatcherLog() selector.MatchFunc {
	whiteList := make(map[string]struct{})
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

// MiddlewareOperateLog 操作日志记录
func MiddlewareOperateLog() middleware2.Middleware {
	return selector.Server(
		MiddlewareHttpLog(),
	).Match(NewWhiteListMatcherLog()).Build()
}

func MiddlewareHttpLog() middleware2.Middleware {
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

						if tr.Operation() == "/api.base_api.v1.Base/Login" {
							name, err := utils.BindLoginRequest(data)
							if err != nil {
								return
							}
							fmt.Printf("登录用户名:%s\n", name)
						}

						fmt.Printf("请求方式:%s, 请求地址:%s, 请求耗时:%s\n", request.Method, request.URL.Path, time.Since(startTime))
						fmt.Printf("访问IP:%s, IP归属地:%s, 响应时间:%s\n", utils.GetClientRealIP(request), utils.ClientIpToLocation(utils.GetClientRealIP(request)), startTime.Format(time.DateTime))
						ua := useragent.Parse(request.UserAgent())
						fmt.Printf("平台:%s %s, 浏览器:%s/%s，访问代理:%s\n", ua.OS, ua.OSVersion, ua.Name, ua.Version, request.UserAgent())
						fmt.Printf("Get参数%v\n", request.URL.Query())

						fmt.Printf("Post参数%s\n", string(data))

						/////// 返回值
						// 获取错误码和是否成功
						statusCode, reason, success := utils.GetStatusCode(handlerErr)
						fmt.Printf("响应状态码:%d, 原因:%s, 请求状态:%v\n", statusCode, reason, success)
						level, stack := extractError(handlerErr)
						fmt.Printf("日志等级:%s, 错误堆栈:%s\n", level, stack)
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

// extractError returns the string of the error
func extractError(err error) (log.Level, string) {
	if err != nil {
		return log.LevelError, fmt.Sprintf("%+v", err)
	}
	return log.LevelInfo, ""
}
