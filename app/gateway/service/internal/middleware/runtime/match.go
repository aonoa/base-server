package runtime

import (
	"net"
	"net/http"
	"strings"
)

type RouteRule struct {
	Path   string
	Method string
	Host   string
}

func normalizeMethod(method string) string {
	method = strings.TrimSpace(method)
	if method == "" || method == "*" {
		return "*"
	}
	return strings.ToUpper(method)
}

func normalizeHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return ""
	}
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		return parsedHost
	}
	return host
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func NewRouteRule(path, method, host string) RouteRule {
	return RouteRule{
		Path:   normalizePath(path),
		Method: normalizeMethod(method),
		Host:   normalizeHost(host),
	}
}

func RouteMatches(rule RouteRule, req *http.Request) bool {
	if req == nil || req.URL == nil {
		return false
	}
	if rule.Method != "*" && rule.Method != strings.ToUpper(req.Method) {
		return false
	}
	if rule.Host != "" && rule.Host != normalizeHost(req.Host) {
		return false
	}
	return matchPath(rule.Path, req.URL.Path)
}

func matchPath(pattern, path string) bool {
	pattern = normalizePath(pattern)
	path = normalizePath(path)
	if pattern == "*" || pattern == "/*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	return path == pattern
}
