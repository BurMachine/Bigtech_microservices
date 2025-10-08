package chat_grpc

import (
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/models"            // Предполагаю path к models (Chat, Message)
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/usecases/chat/dto" // Предполагаю path к DTO
	pb "github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
)

// Requests converters: from pb.Request to dto.DTO

func dtoCreateDirectChatFromCreateDirectChatRequest(r *pb.CreateDirectChatRequest) dto.CreateDirectChatDTO {
	return dto.CreateDirectChatDTO{
		ParticipantID: r.ParticipantId,
	}
}

func dtoGetChatFromGetChatRequest(r *pb.GetChatRequest) dto.GetChatDTO {
	return dto.GetChatDTO{
		ChatID: r.ChatId,
	}
}

func dtoListUserChatsFromListUserChatsRequest(r *pb.ListUserChatsRequest) dto.ListUserChatsDTO {
	return dto.ListUserChatsDTO{
		UserID: r.UserId,
	}
}

func dtoListChatMembersFromListChatMembersRequest(r *pb.ListChatMembersRequest) dto.ListChatMembersDTO {
	return dto.ListChatMembersDTO{
		ChatID: r.ChatId,
	}
}

func dtoSendMessageFromSendMessageRequest(r *pb.SendMessageRequest) dto.SendMessageDTO {
	return dto.SendMessageDTO{
		ChatID: r.ChatId,
		Text:   r.Text,
	}
}

func dtoListMessagesFromListMessagesRequest(r *pb.ListMessagesRequest) dto.ListMessagesDTO {
	cursor := ""
	if r.Cursor != nil {
		cursor = *r.Cursor
	}
	result := dto.ListMessagesDTO{
		ChatID: r.ChatId,
		Limit:  int(r.Limit),
		Cursor: cursor,
	}
	return result
}

func dtoStreamMessagesFromStreamMessagesRequest(r *pb.StreamMessagesRequest) dto.StreamMessagesDTO {
	return dto.StreamMessagesDTO{
		ChatID:      r.ChatId,
		SinceUnixMs: *r.SinceUnixMs,
	}
}

// Responses converters: from models.Entity to pb.Response
// Также вспомогательные: from model to pb (для вложенных, как Chat, Message)

func chatFromModelChat(model *models.Chat) *pb.Chat {
	if model == nil {
		return nil
	}
	return &pb.Chat{
		Id:             model.ID,
		ParticipantIds: model.Participants,
		// CreatedAt: не в entity? Если есть, добавить как unix timestamp: model.CreatedAt.UnixMilli()
	}
}

func messageFromModelMessage(model *models.Message) *pb.Message {
	if model == nil {
		return nil
	}
	return &pb.Message{
		Id:              model.ID,
		ChatId:          model.ChatID,
		SenderId:        model.SenderID,
		Text:            model.Text,
		TimestampUnixMs: model.CreatedAt.UnixMilli(),
	}
}

func createDirectChatResponseFromChatID(chatID string) *pb.CreateDirectChatResponse {
	return &pb.CreateDirectChatResponse{
		ChatId: chatID,
	}
}

func getChatResponseFromModelChat(chat *models.Chat) *pb.Chat {
	return chatFromModelChat(chat)
}

func listMessagesResponseFromModelMessages(messages []*models.Message, nextCursor string) *pb.ListMessagesResponse {
	pbMessages := make([]*pb.Message, len(messages))
	for i, m := range messages {
		pbMessages[i] = messageFromModelMessage(m)
	}
	return &pb.ListMessagesResponse{
		Messages:   pbMessages,
		NextCursor: &nextCursor,
	}
}

func listUserChatsResponseFromModelChats(chats []*models.Chat) *pb.ListUserChatsResponse {
	chatsRes := make([]*pb.Chat, 0, len(chats))
	for _, v := range chats {
		chatsRes = append(chatsRes, chatFromModelChat(v))
	}
	return &pb.ListUserChatsResponse{
		Chats: chatsRes,
	}
}

func listChatMembersResponseFromUserIDs(userIDs []string) *pb.ListChatMembersResponse {
	return &pb.ListChatMembersResponse{
		UserIds: userIDs,
	}
}

func sendMessageResponseFromModelMessage(msg *models.Message) *pb.Message {
	return &pb.Message{
		Id:              msg.ID,
		ChatId:          msg.ChatID,
		SenderId:        msg.SenderID,
		Text:            msg.Text,
		TimestampUnixMs: msg.CreatedAt.Unix(),
	}
}
