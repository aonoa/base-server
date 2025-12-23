package base_api

import (
	pb "base-server/api/gen/go/base_api/v1"
	"context"
	http "github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/protobuf/types/known/emptypb"
)

const OperationSSEServiceCopilot = "/api.base_api.v1.SSEService/Copilot"

type SSEServiceHTTPServer interface {
	Copilot(http.Context, *pb.Msg) (*emptypb.Empty, error)
}

func RegisterSSEServiceHTTPServer(s *http.Server, srv SSEServiceHTTPServer) {
	// 文件上传相关的接口
	r := s.Route("/")
	r.POST("/basic-api/v1/copilot/sse", _SSEService_Copilot0_HTTP_Handler(srv))
}

func _SSEService_Copilot0_HTTP_Handler(srv SSEServiceHTTPServer) func(ctx http.Context) error {
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
		//_, err := srv.DownloadFile(ctx, &in)
		//if err != nil {
		//	return err
		//}
		//return err

		h := ctx.Middleware(func(_ context.Context, req interface{}) (interface{}, error) {
			return srv.Copilot(ctx, req.(*pb.Msg))
		})

		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out)
	}
}
