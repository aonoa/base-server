package service

import (
	"context"

	pb "base-server/api/gen/go/common/service/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *CommonService) GetMySiteMessageList(ctx context.Context, req *pb.GetMySiteMessageListParams) (*pb.GetMySiteMessageListReply, error) {
	return s.uc.GetMySiteMessageList(ctx, req)
}

func (s *CommonService) GetMySiteMessageUnreadCount(ctx context.Context, req *emptypb.Empty) (*pb.GetMySiteMessageUnreadCountReply, error) {
	return s.uc.GetMySiteMessageUnreadCount(ctx)
}

func (s *CommonService) MarkSiteMessageRead(ctx context.Context, req *pb.MarkSiteMessageReadRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.MarkSiteMessageRead(ctx, req)
}

func (s *CommonService) MarkSiteMessageUnread(ctx context.Context, req *pb.MarkSiteMessageReadRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.MarkSiteMessageUnread(ctx, req)
}

func (s *CommonService) MarkAllSiteMessagesRead(ctx context.Context, req *emptypb.Empty) (*pb.MarkAllSiteMessagesReadReply, error) {
	return s.uc.MarkAllSiteMessagesRead(ctx)
}

func (s *CommonService) GetSiteMessageManageList(ctx context.Context, req *pb.GetSiteMessageManageListParams) (*pb.GetSiteMessageManageListReply, error) {
	return s.uc.GetSiteMessageManageList(ctx, req)
}

func (s *CommonService) CreateSiteMessage(ctx context.Context, req *pb.CreateSiteMessageRequest) (*pb.CreateSiteMessageReply, error) {
	return s.uc.CreateSiteMessage(ctx, req)
}

func (s *CommonService) RecallSiteMessage(ctx context.Context, req *pb.RecallSiteMessageRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.RecallSiteMessage(ctx, req)
}

func (s *CommonService) DeletePendingSiteMessage(ctx context.Context, req *pb.DeletePendingSiteMessageRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DeletePendingSiteMessage(ctx, req)
}
