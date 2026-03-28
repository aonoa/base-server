package httplog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	adminv1 "base-server/api/gen/go/admin/service/v1"
	configv1 "github.com/go-kratos/gateway/api/gateway/config/v1"
	gwmiddleware "github.com/go-kratos/gateway/middleware"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	httplogv1 "base-server/app/gateway/service/internal/middleware/httplog/v1"
	runtimeutil "base-server/app/gateway/service/internal/middleware/runtime"
)

var adminClientProvider func() adminv1.AdminServiceClient

func init() {
	gwmiddleware.Register("httplog", Middleware)
}

func SetAdminClient(provider func() adminv1.AdminServiceClient) {
	adminClientProvider = provider
}

func Middleware(cfg *configv1.Middleware) (gwmiddleware.Middleware, error) {
	options := &httplogv1.HttpLog{}
	if cfg.Options != nil {
		if err := anypb.UnmarshalTo(cfg.Options, options, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, err
		}
	}
	if options.SigningKey == "" {
		return nil, errors.New("httplog.signing_key is required")
	}
	whitelist := make([]runtimeutil.RouteRule, 0, len(options.Whitelist))
	for _, item := range options.Whitelist {
		whitelist = append(whitelist, runtimeutil.NewRouteRule(item.Path, item.Method, item.Host))
	}
	return func(next http.RoundTripper) http.RoundTripper {
		return gwmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			for _, rule := range whitelist {
				if runtimeutil.RouteMatches(rule, req) {
					return next.RoundTrip(req)
				}
			}
			startedAt := time.Now()
			bodyBytes, restoreBodyErr := readAndRestoreBody(req)
			resp, err := next.RoundTrip(req)
			if adminClientProvider != nil && adminClientProvider() != nil {
				writeErr := writeLog(context.Background(), adminClientProvider(), req, bodyBytes, options.SigningKey, startedAt, resp, err)
				if writeErr != nil && err == nil {
					err = writeErr
				}
			}
			if restoreBodyErr != nil && err == nil {
				err = restoreBodyErr
			}
			return resp, err
		})
	}, nil
}

func writeLog(ctx context.Context, client adminv1.AdminServiceClient, req *http.Request, body []byte, signingKey string, startedAt time.Time, resp *http.Response, roundTripErr error) error {
	statusCode := int32(http.StatusOK)
	if resp != nil {
		statusCode = int32(resp.StatusCode)
	}
	reason := ""
	success := statusCode < 400 && roundTripErr == nil
	if roundTripErr != nil {
		reason = roundTripErr.Error()
		success = false
		if statusCode < 400 {
			statusCode = http.StatusBadGateway
		}
	}
	userID, userName, sessionID := parseIdentity(req, signingKey)
	postData := string(body)
	isLogin := req.URL.Path == "/auth-api/v1/login"
	if isLogin {
		loginUser, pass := bindLoginRequest(body)
		userName = loginUser
		masked, marshalErr := json.Marshal(map[string]string{"username": loginUser, "password": strings.Repeat("*", len(pass))})
		if marshalErr == nil {
			postData = string(masked)
		}
	}
	_, err := client.CreateSysLog(ctx, &adminv1.CreateSysLogRequest{
		UserId:      userID,
		UserName:    userName,
		IsLogin:     isLogin,
		SessionId:   sessionID,
		Method:      req.Method,
		Path:        req.URL.Path,
		RequestTime: startedAt.Format(time.DateTime),
		IpAddress:   clientRealIP(req),
		IpLocation:  "",
		Latency:     time.Since(startedAt).Milliseconds(),
		Os:          userAgentOS(req.UserAgent()),
		Browser:     userAgentBrowser(req.UserAgent()),
		UserAgent:   req.UserAgent(),
		Header:      marshalToString(req.Header),
		GetParams:   marshalToString(req.URL.Query()),
		PostData:    postData,
		ResCode:     statusCode,
		Reason:      reason,
		ResStatus:   success,
		Stack:       reason,
	})
	return err
}

func readAndRestoreBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func parseIdentity(req *http.Request, signingKey string) (string, string, string) {
	tokenString := runtimeutil.ParseAuthorizationToken(req.Header.Get("Authorization"))
	if tokenString == "" {
		return "", "", ""
	}
	claims, err := runtimeutil.ParseHS256Token(tokenString, signingKey)
	if err != nil {
		return "", "", ""
	}
	userID, _ := claims["user_id"].(string)
	sessionID, _ := claims["session_id"].(string)
	userName, _ := claims["username"].(string)
	return userID, userName, sessionID
}

func bindLoginRequest(body []byte) (string, string) {
	payload := map[string]any{}
	if err := json.Unmarshal(body, &payload); err == nil {
		username, _ := payload["username"].(string)
		password, _ := payload["password"].(string)
		return username, password
	}
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return "", ""
	}
	return values.Get("username"), values.Get("password")
}

func marshalToString(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

func clientRealIP(req *http.Request) string {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		value := strings.TrimSpace(req.Header.Get(header))
		if value == "" {
			continue
		}
		for _, item := range strings.Split(value, ",") {
			ip := strings.TrimSpace(item)
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(req.RemoteAddr))
	if err == nil && net.ParseIP(host) != nil {
		return host
	}
	if net.ParseIP(req.RemoteAddr) != nil {
		return req.RemoteAddr
	}
	return ""
}

func userAgentOS(value string) string {
	lower := strings.ToLower(value)
	switch {
	case strings.Contains(lower, "windows"):
		return "Windows"
	case strings.Contains(lower, "mac os") || strings.Contains(lower, "macintosh"):
		return "macOS"
	case strings.Contains(lower, "android"):
		return "Android"
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad"):
		return "iOS"
	case strings.Contains(lower, "linux"):
		return "Linux"
	default:
		return ""
	}
}

func userAgentBrowser(value string) string {
	lower := strings.ToLower(value)
	switch {
	case strings.Contains(lower, "edg/"):
		return "Edge"
	case strings.Contains(lower, "chrome/"):
		return "Chrome"
	case strings.Contains(lower, "firefox/"):
		return "Firefox"
	case strings.Contains(lower, "safari/") && !strings.Contains(lower, "chrome/"):
		return "Safari"
	default:
		return ""
	}
}

var _ jwtv5.MapClaims
