package base_api

import (
	pb "base-server/api/base_api/v1"
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	http "github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/grpc"
	"io"
	httpx "net/http"
)

type UploadFileServiceHTTPServer interface {
	UploadFileHttp(context.Context, *pb.File, ...grpc.CallOption) (*pb.UploadResponse, error)
}

func RegisterFileServiceHTTPServer(s *http.Server, srv UploadFileServiceHTTPServer) {
	// 文件上传相关的接口
	r := s.Route("/")
	r.POST("/basic-api/v1/server/file/upload", _UploadService_SaveSite0_HTTP_Handler(srv))
}

func _UploadService_SaveSite0_HTTP_Handler(srv UploadFileServiceHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in pb.File
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		//MultipartFormReader(ctx)
		req := ctx.Request()
		// 获取resourceFile的文件流
		resourceFile, resourceHeader, err := req.FormFile("file")
		resource1File := req.FormValue("fileName1")
		log.Info(resource1File)
		if err != nil {
			if errors.Is(err, httpx.ErrMissingFile) {
				return errors.BadRequest("resourceFile", "resourceFile不能为空")
			}
			return err
		}
		in.FileSize = resourceHeader.Size

		// io.ReadAll来读取整个文件内容到字节切片
		in.File, err = io.ReadAll(resourceFile)
		if err != nil {
			return errors.New(502, "读取文件错误", "读取文件错误")
		}

		defer resourceFile.Close()
		// 文件大小校验
		if resourceHeader.Size > 5*1024*1024 || resourceHeader.Size <= 0 {
			return errors.BadRequest("resourceFile", "上传文件的大小超过了限定值")
		}

		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.UploadFileHttp(ctx, req.(*pb.File))
		})

		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out)
	}
}

func MultipartFormReader(ctx http.Context) {
	req, _ := ctx.Request().MultipartReader()
	// 循环读取每个part
	for {
		part, err := req.NextPart()
		if err == io.EOF {
			// 没有更多的part
			break
		}
		if err != nil {
			return
		}

		// 获取part的文件名
		fileName := part.FileName()
		if fileName == "" {
			// 如果没有文件名，可能是表单字段而不是文件
			continue
		}

		// 读取part的内容
		var buffer []byte
		partBuffer := make([]byte, 1024)
		for {
			n, err := part.Read(partBuffer)
			if err == io.EOF {
				// 文件读取完毕
				break
			}
			if err != nil {
				return
			}
			buffer = append(buffer, partBuffer[:n]...)
		}

		// 处理读取到的文件内容
		fmt.Printf("Received file: %s with size: %d\n", fileName, len(buffer))

		// 可以选择将文件保存到磁盘或进行其他处理
		// 例如，保存文件到磁盘:
		// file, err := os.Create(fileName)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		// defer file.Close()
		// _, err = file.Write(buffer)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
	}
}
