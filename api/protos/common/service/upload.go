package service

import (
	pb "base-server/api/gen/go/common/service/v1"
	"context"
	"io"
	httpx "net/http"

	"github.com/go-kratos/kratos/v2/errors"
	http "github.com/go-kratos/kratos/v2/transport/http"
)

type UploadServiceHTTPServer interface {
	UploadFile(context.Context, *pb.File) (*pb.UploadResponse, error)
}

func RegisterUploadServiceHTTPServer(s *http.Server, srv UploadServiceHTTPServer) {
	r := s.Route("/")
	r.POST("/common-api/v1/file/upload", _UploadService_HTTP_Handler(srv))
}

func _UploadService_HTTP_Handler(srv UploadServiceHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in pb.File
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}

		req := ctx.Request()
		resourceFile, resourceHeader, err := req.FormFile("file")
		if err != nil {
			if errors.Is(err, httpx.ErrMissingFile) {
				return errors.BadRequest("file", "file不能为空")
			}
			return err
		}
		defer resourceFile.Close()

		in.FileName = req.FormValue("fileName")
		if in.FileName == "" {
			in.FileName = resourceHeader.Filename
		}
		in.FileSize = resourceHeader.Size
		in.File, err = io.ReadAll(resourceFile)
		if err != nil {
			return errors.New(502, "read_file_error", "读取文件错误")
		}
		if in.FileSize > 5*1024*1024 || in.FileSize <= 0 {
			return errors.BadRequest("file", "上传文件的大小超过了限定值")
		}

		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.UploadFile(ctx, req.(*pb.File))
		})

		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out)
	}
}
