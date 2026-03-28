package service

import (
	pb "base-server/api/gen/go/common/service/v1"
	commonhttp "base-server/api/protos/common/service"
	"base-server/app/common/service/internal/biz"
	"base-server/app/common/service/internal/conf"
	"base-server/pkg/tools"
	"bytes"
	"context"
	"fmt"
	"io"
	nethttp "net/http"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewCommonService)

type CommonService struct {
	pb.UnimplementedUploadServiceServer
	pb.UnimplementedSSEServiceServer
	pb.UnimplementedCommonServiceServer

	uc         *biz.CommonUsecase
	llm        *conf.Llm
	RestServer *khttp.Server
}

func NewCommonService(uc *biz.CommonUsecase, llm *conf.Llm) *CommonService {
	return &CommonService{uc: uc, llm: llm}
}

func (s *CommonService) GetWalkRoute(ctx context.Context, req *emptypb.Empty) (*pb.GetWalkRouteReply, error) {
	items, err := tools.WalkHTTPRoutes(s.RestServer)
	if err != nil {
		return nil, err
	}
	res := &pb.GetWalkRouteReply{Items: make([]*pb.WalkRouteItem, 0, len(items))}
	for _, item := range tools.SortAndUniqueWalkRoutes(items) {
		res.Items = append(res.Items, &pb.WalkRouteItem{Url: item.URL, Method: item.Method})
	}
	return res, nil
}

func (s *CommonService) UploadFile(ctx context.Context, req *pb.File) (*pb.UploadResponse, error) {
	log.Infof("文件:%s,大小:%d", req.FileName, req.FileSize)
	if req.FileSize <= 0 {
		req.FileSize = int64(len(req.File))
	}
	fileName := req.FileName
	if fileName == "" {
		fileName = "aaa.png"
	}
	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err = io.Copy(file, bytes.NewReader(req.File)); err != nil {
		return nil, err
	}

	return &pb.UploadResponse{
		FileInfoId: "456465",
		FullUrl:    "aaa",
		Url:        "https://q1.qlogo.cn/g?b=qq&nk=190848757&s=640",
	}, nil
}

func (s *CommonService) Copilot(req *pb.Msg, stream grpc.ServerStreamingServer[pb.CopilotReply]) error {
	messages, err := s.buildMessages(req)
	if err != nil {
		return err
	}
	chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		BaseURL: s.llm.Agent.OpenAI.ApiBaseUrl,
		Model:   s.llm.Agent.OpenAI.Model,
		APIKey:  s.llm.Agent.OpenAI.ApiKey,
	})
	if err != nil {
		return err
	}
	reader, err := chatModel.Stream(context.Background(), messages)
	if err != nil {
		return err
	}
	defer reader.Close()
	for {
		message, err := reader.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := stream.Send(&pb.CopilotReply{Content: message.Content}); err != nil {
			return err
		}
	}
}

func (s *CommonService) CopilotHTTP(ctx khttp.Context, req *pb.Msg) error {
	w := ctx.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(nethttp.Flusher)
	if !ok {
		nethttp.Error(w, "Streaming unsupported!", nethttp.StatusInternalServerError)
		return nil
	}

	messages, err := s.buildMessages(req)
	if err != nil {
		return err
	}
	chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		BaseURL: s.llm.Agent.OpenAI.ApiBaseUrl,
		Model:   s.llm.Agent.OpenAI.Model,
		APIKey:  s.llm.Agent.OpenAI.ApiKey,
	})
	if err != nil {
		return err
	}
	reader, err := chatModel.Stream(context.Background(), messages)
	if err != nil {
		return err
	}
	defer reader.Close()

	for {
		message, err := reader.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", message.Content); err != nil {
			return err
		}
		flusher.Flush()
	}
}

func (s *CommonService) buildMessages(req *pb.Msg) ([]*schema.Message, error) {
	if len(req.Items) == 0 {
		return nil, errors.BadRequest("items", "items不能为空")
	}
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个{role}。你需要用{style}的语气回答问题。你的目标是帮助程序员保持积极乐观的心态，提供技术建议的同时也要关注他们的心理健康。回答要尽可能的精简"),
		schema.MessagesPlaceholder("chat_history", true),
		schema.UserMessage("问题: {question}"),
	)

	last := req.Items[len(req.Items)-1]
	return template.Format(context.Background(), map[string]any{
		"role":     "程序员鼓励师",
		"style":    "专业、积极且温暖",
		"question": last.Content,
		"chat_history": func(msg *pb.Msg) []*schema.Message {
			res := make([]*schema.Message, 0, len(msg.Items))
			for _, item := range msg.Items[:len(msg.Items)-1] {
				switch item.Role {
				case "assistant":
					res = append(res, schema.AssistantMessage(item.Content, nil))
				case "user":
					res = append(res, schema.UserMessage(item.Content))
				}
			}
			return res
		}(req),
	})
}

var (
	_ commonhttp.UploadServiceHTTPServer = (*CommonService)(nil)
	_ commonhttp.SSEServiceHTTPServer    = (*CommonService)(nil)
)
