package biz

import (
	"context"
	"io"
	"testing"

	v1 "base-server/api/gen/go/auth/service/v1"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/go-kratos/kratos/v2/log"
)

func TestApplyApiDeltaUpdatesGroupingPolicyOnPathChange(t *testing.T) {
	m, err := model.NewModelFromString(textModel)
	if err != nil {
		t.Fatalf("new model: %v", err)
	}
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("new enforcer: %v", err)
	}

	uc := &AuthUsecase{
		e:   e,
		log: log.NewHelper(log.NewStdLogger(io.Discard)),
	}

	uc.AddApiToGroup("/auth-api/v1/resources", "data")
	if ok, err := uc.e.HasNamedGroupingPolicy(ApiToGroup, "/auth-api/v1/resources", "api:data"); err != nil || !ok {
		t.Fatalf("expected old grouping policy to exist, ok=%v err=%v", ok, err)
	}

	err = uc.ApplyApiDelta(context.Background(), &v1.ApplyApiDeltaRequest{
		SourceService: "admin",
		Revision:      1,
		Before: &v1.PolicyApi{
			Path:           "/auth-api/v1/resources",
			Method:         "GET",
			ResourcesGroup: "data",
		},
		After: &v1.PolicyApi{
			Path:           "/admin-api/v1/resources",
			Method:         "GET",
			ResourcesGroup: "data",
		},
	})
	if err != nil {
		t.Fatalf("apply api delta: %v", err)
	}

	if ok, err := uc.e.HasNamedGroupingPolicy(ApiToGroup, "/auth-api/v1/resources", "api:data"); err != nil || ok {
		t.Fatalf("expected old grouping policy to be removed, ok=%v err=%v", ok, err)
	}
	if ok, err := uc.e.HasNamedGroupingPolicy(ApiToGroup, "/admin-api/v1/resources", "api:data"); err != nil || !ok {
		t.Fatalf("expected new grouping policy to exist, ok=%v err=%v", ok, err)
	}
}
