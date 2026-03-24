package server

import (
	"encoding/json"
	"net/http"
	"os"

	gwconfigv1 "github.com/go-kratos/gateway/api/gateway/config/v1"
	"github.com/go-kratos/gateway/client"
	gwmiddleware "github.com/go-kratos/gateway/middleware"
	_ "github.com/go-kratos/gateway/middleware/bbr"
	_ "github.com/go-kratos/gateway/middleware/circuitbreaker"
	_ "github.com/go-kratos/gateway/middleware/cors"
	_ "github.com/go-kratos/gateway/middleware/logging"
	_ "github.com/go-kratos/gateway/middleware/rewrite"
	_ "github.com/go-kratos/gateway/middleware/streamrecorder"
	_ "github.com/go-kratos/gateway/middleware/tracing"
	_ "github.com/go-kratos/gateway/middleware/transcoder"
	"github.com/go-kratos/gateway/proxy"
	"github.com/go-kratos/kratos/v2/log"
	ktransport "github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/protobuf/types/known/durationpb"

	"base-server/app/gateway/service/internal/conf"
)

const (
	healthPath = "/healthz"
)

func NewProxyServer(c *conf.Server, gc *conf.Gateway, logger log.Logger) (ktransport.Server, func(), error) {
	_ = logger

	nativeConfig := buildGatewayConfig(gc)
	clientFactory := client.NewFactory(nil)
	p, err := proxy.New(clientFactory, gwmiddleware.Create)
	if err != nil {
		return nil, nil, err
	}
	if err := p.Update(client.NewBuildContext(nativeConfig), nativeConfig); err != nil {
		return nil, nil, err
	}

	handler := http.NewServeMux()
	handler.Handle(healthPath, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	handler.Handle("/", p)

	addr := ":8000"
	if c != nil && c.Http != nil && c.Http.Addr != "" {
		addr = c.Http.Addr
	}
	return newHTTPProxyServer(handler, addr, c), func() {}, nil
}

func buildGatewayConfig(gc *conf.Gateway) *gwconfigv1.Gateway {
	cfg := &gwconfigv1.Gateway{
		Name:      "gateway",
		Version:   os.Getenv("SERVICE_VERSION"),
		Endpoints: buildEndpoints(gc),
		TlsStore:  buildTLSStore(gc),
	}
	if gc == nil {
		return cfg
	}
	if gc.Name != "" {
		cfg.Name = gc.Name
	}
	if gc.Version != "" {
		cfg.Version = gc.Version
	}
	cfg.Middlewares = buildMiddlewares(gc.Middlewares)
	return cfg
}

func buildTLSStore(gc *conf.Gateway) map[string]*gwconfigv1.TLS {
	if gc == nil || len(gc.TlsStore) == 0 {
		return nil
	}
	out := make(map[string]*gwconfigv1.TLS, len(gc.TlsStore))
	for name, item := range gc.TlsStore {
		if item == nil {
			continue
		}
		out[name] = &gwconfigv1.TLS{
			Insecure:   item.Insecure,
			Cacert:     item.Cacert,
			Cert:       item.Cert,
			Key:        item.Key,
			ServerName: item.ServerName,
		}
	}
	return out
}

func buildEndpoints(gc *conf.Gateway) []*gwconfigv1.Endpoint {
	if gc == nil || len(gc.Endpoints) == 0 {
		return nil
	}
	out := make([]*gwconfigv1.Endpoint, 0, len(gc.Endpoints))
	for _, endpoint := range gc.Endpoints {
		if endpoint == nil {
			continue
		}
		out = append(out, &gwconfigv1.Endpoint{
			Path:        endpoint.Path,
			Method:      endpoint.Method,
			Description: endpoint.Description,
			Protocol:    gwconfigv1.Protocol(endpoint.Protocol),
			Timeout:     cloneDuration(endpoint.Timeout),
			Middlewares: buildMiddlewares(endpoint.Middlewares),
			Backends:    buildBackends(endpoint.Backends),
			Retry:       buildRetry(endpoint.Retry),
			Metadata:    cloneStringMap(endpoint.Metadata),
			Host:        endpoint.Host,
			Stream:      endpoint.Stream,
		})
	}
	return out
}

func buildMiddlewares(items []*conf.Middleware) []*gwconfigv1.Middleware {
	if len(items) == 0 {
		return nil
	}
	out := make([]*gwconfigv1.Middleware, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &gwconfigv1.Middleware{
			Name:     item.Name,
			Options:  item.Options,
			Required: item.Required,
		})
	}
	return out
}

func buildBackends(items []*conf.Backend) []*gwconfigv1.Backend {
	if len(items) == 0 {
		return nil
	}
	out := make([]*gwconfigv1.Backend, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		backend := &gwconfigv1.Backend{
			Target:        item.Target,
			HealthCheck:   buildHealthCheck(item.HealthCheck),
			Tls:           item.Tls,
			TlsConfigName: item.TlsConfigName,
			Metadata:      cloneStringMap(item.Metadata),
		}
		if item.Weight != nil {
			weight := item.GetWeight()
			backend.Weight = &weight
		}
		out = append(out, backend)
	}
	return out
}

func buildHealthCheck(item *conf.HealthCheck) *gwconfigv1.HealthCheck {
	if item == nil {
		return nil
	}
	return &gwconfigv1.HealthCheck{}
}

func buildRetry(item *conf.Retry) *gwconfigv1.Retry {
	if item == nil {
		return nil
	}
	out := &gwconfigv1.Retry{
		Attempts:      item.Attempts,
		PerTryTimeout: cloneDuration(item.PerTryTimeout),
		Priorities:    append([]string(nil), item.Priorities...),
	}
	if len(item.Conditions) > 0 {
		out.Conditions = make([]*gwconfigv1.Condition, 0, len(item.Conditions))
		for _, condition := range item.Conditions {
			if condition == nil {
				continue
			}
			out.Conditions = append(out.Conditions, buildCondition(condition))
		}
	}
	return out
}

func buildCondition(item *conf.Condition) *gwconfigv1.Condition {
	if item == nil {
		return nil
	}
	out := &gwconfigv1.Condition{}
	switch v := item.Condition.(type) {
	case *conf.Condition_ByStatusCode:
		out.Condition = &gwconfigv1.Condition_ByStatusCode{ByStatusCode: v.ByStatusCode}
	case *conf.Condition_ByHeader:
		if v.ByHeader != nil {
			out.Condition = &gwconfigv1.Condition_ByHeader{ByHeader: &gwconfigv1.ConditionHeader{Name: v.ByHeader.Name, Value: v.ByHeader.Value}}
		}
	}
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneDuration(in *durationpb.Duration) *durationpb.Duration {
	if in == nil {
		return nil
	}
	return durationpb.New(in.AsDuration())
}
