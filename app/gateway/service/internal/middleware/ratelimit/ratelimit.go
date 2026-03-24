package ratelimit

import (
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	configv1 "github.com/go-kratos/gateway/api/gateway/config/v1"
	gwmiddleware "github.com/go-kratos/gateway/middleware"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	ratelimitv1 "base-server/app/gateway/service/internal/middleware/ratelimit/v1"
)

func init() {
	gwmiddleware.RegisterV2("ratelimit", New)
}

type middlewareState struct {
	mu       sync.Mutex
	global   *rate.Limiter
	visitors map[string]*visitor
	cfg      config
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type config struct {
	limitPerSecond rate.Limit
	burst          int
	statusCode     int
	keyFunc        func(*http.Request) string
	cleanupAfter   time.Duration
}

func New(cfg *configv1.Middleware) (gwmiddleware.MiddlewareV2, error) {
	options := &ratelimitv1.RateLimit{}
	if cfg.Options != nil {
		if err := anypb.UnmarshalTo(cfg.Options, options, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, err
		}
	}
	if options.RequestsPerSecond <= 0 {
		return nil, errors.New("ratelimit.requests_per_second must be > 0")
	}
	burst := int(options.Burst)
	if burst <= 0 {
		burst = 1
	}
	statusCode := int(options.StatusCode)
	if statusCode <= 0 {
		statusCode = http.StatusTooManyRequests
	}
	cleanupAfter := 10 * time.Minute
	if options.CleanupAfter != nil && options.CleanupAfter.AsDuration() > 0 {
		cleanupAfter = options.CleanupAfter.AsDuration()
	}
	state := &middlewareState{
		visitors: make(map[string]*visitor),
		cfg: config{
			limitPerSecond: rate.Limit(options.RequestsPerSecond),
			burst:          burst,
			statusCode:     statusCode,
			keyFunc:        selectKeyFunc(options.Scope),
			cleanupAfter:   cleanupAfter,
		},
	}
	if options.Scope == ratelimitv1.Scope_SCOPE_GLOBAL {
		state.global = rate.NewLimiter(state.cfg.limitPerSecond, state.cfg.burst)
	}
	return gwmiddleware.NewWithCloser(state.process, state), nil
}

func (m *middlewareState) process(next http.RoundTripper) http.RoundTripper {
	return gwmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		limiter := m.global
		if limiter == nil {
			limiter = m.visitorLimiter(m.cfg.keyFunc(req))
		}
		if !limiter.Allow() {
			return &http.Response{
				StatusCode: m.cfg.statusCode,
				Status:     http.StatusText(m.cfg.statusCode),
				Header:     make(http.Header),
				Body:       http.NoBody,
			}, nil
		}
		return next.RoundTrip(req)
	})
}

func (m *middlewareState) visitorLimiter(key string) *rate.Limiter {
	if key == "" {
		key = "unknown"
	}
	now := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range m.visitors {
		if now.Sub(v.lastSeen) > m.cfg.cleanupAfter {
			delete(m.visitors, k)
		}
	}
	if existing, ok := m.visitors[key]; ok {
		existing.lastSeen = now
		return existing.limiter
	}
	limiter := rate.NewLimiter(m.cfg.limitPerSecond, m.cfg.burst)
	m.visitors[key] = &visitor{limiter: limiter, lastSeen: now}
	return limiter
}

func (m *middlewareState) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.visitors = nil
	return nil
}

func selectKeyFunc(scope ratelimitv1.Scope) func(*http.Request) string {
	switch scope {
	case ratelimitv1.Scope_SCOPE_GLOBAL:
		return func(*http.Request) string { return "global" }
	case ratelimitv1.Scope_SCOPE_IP:
		return func(req *http.Request) string {
			host, _, err := net.SplitHostPort(req.RemoteAddr)
			if err == nil {
				return host
			}
			return req.RemoteAddr
		}
	default:
		return func(*http.Request) string { return "global" }
	}
}
