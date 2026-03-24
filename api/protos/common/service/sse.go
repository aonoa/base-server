package service

import (
	pb "base-server/api/gen/go/common/service/v1"
	"context"

	http "github.com/go-kratos/kratos/v2/transport/http"
)

const OperationSSEServiceCopilot = "/api.common.service.v1.SSEService/Copilot"

type SSEServiceHTTPServer interface {
	CopilotHTTP(http.Context, *pb.Msg) error
}

func RegisterSSEServiceHTTPServer(s *http.Server, srv SSEServiceHTTPServer) {
	r := s.Route("/")
	r.POST("/common-api/v1/copilot/sse", _SSEService_Copilot_HTTP_Handler(srv))
}

func _SSEService_Copilot_HTTP_Handler(srv SSEServiceHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in pb.Msg
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationSSEServiceCopilot)

		h := ctx.Middleware(func(_ context.Context, req interface{}) (interface{}, error) {
			return nil, srv.CopilotHTTP(ctx, req.(*pb.Msg))
		})
		_, err := h(ctx, &in)
		return err
	}
}
