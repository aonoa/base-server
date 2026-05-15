package authx

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/grpc/metadata"
)

func TestForwardAuthorizationContextForwardsOrganizationID(t *testing.T) {
	ctx := transport.NewServerContext(context.Background(), authTestTransport{
		headers: authTestHeader{
			HeaderAuthorization:  "Bearer token",
			HeaderOrganizationID: "org-a",
		},
	})

	out := ForwardAuthorizationContext(ctx)
	md, ok := metadata.FromOutgoingContext(out)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	if got := firstMetadataValue(md, "authorization"); got != "Bearer token" {
		t.Fatalf("authorization = %q, want Bearer token", got)
	}
	if got := firstMetadataValue(md, HeaderOrganizationID); got != "org-a" {
		t.Fatalf("organization id = %q, want org-a", got)
	}
}

func firstMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

type authTestTransport struct {
	headers authTestHeader
}

func (t authTestTransport) Kind() transport.Kind { return transport.KindHTTP }
func (t authTestTransport) Endpoint() string     { return "" }
func (t authTestTransport) Operation() string    { return "" }
func (t authTestTransport) RequestHeader() transport.Header {
	return t.headers
}
func (t authTestTransport) ReplyHeader() transport.Header {
	return authTestHeader{}
}

type authTestHeader map[string]string

func (h authTestHeader) Get(key string) string {
	return h[key]
}

func (h authTestHeader) Set(key string, value string) {
	h[key] = value
}

func (h authTestHeader) Add(key string, value string) {
	h[key] = value
}

func (h authTestHeader) Keys() []string {
	keys := make([]string, 0, len(h))
	for key := range h {
		keys = append(keys, key)
	}
	return keys
}

func (h authTestHeader) Values(key string) []string {
	value := h.Get(key)
	if value == "" {
		return nil
	}
	return []string{value}
}

var _ transport.Transporter = (*authTestTransport)(nil)
var _ transport.Header = (authTestHeader)(nil)
