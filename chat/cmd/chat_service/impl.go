package main

import (
	"context"
	pb "github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) CreateDirectChat(ctx context.Context, request *pb.CreateDirectChatRequest) (*pb.CreateDirectChatResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateDirectChat not implemented")
}

func (s *server) GetChat(ctx context.Context, request *pb.GetChatRequest) (*pb.Chat, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetChat not implemented")
}

func (s *server) ListUserChats(ctx context.Context, request *pb.ListUserChatsRequest) (*pb.ListUserChatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUserChats not implemented")
}

func (s *server) ListChatMembers(ctx context.Context, request *pb.ListChatMembersRequest) (*pb.ListChatMembersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListChatMembers not implemented")
}

func (s *server) SendMessage(ctx context.Context, request *pb.SendMessageRequest) (*pb.Message, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMessage not implemented")
}

func (s *server) ListMessages(ctx context.Context, request *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListMessages not implemented")
}

func (s *server) StreamMessages(request *pb.StreamMessagesRequest, g grpc.ServerStreamingServer[pb.Message]) error {
	return status.Errorf(codes.Unimplemented, "method StreamMessages not implemented")
}
