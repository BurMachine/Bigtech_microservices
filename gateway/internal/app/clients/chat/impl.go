package chat_service

import (
	"context"
	"time"

	"github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/models"
)

func (c *Client) CreateDirectChat(ctx context.Context, participantID string) (string, error) {
	req := &chat.CreateDirectChatRequest{ParticipantId: participantID}
	resp, err := c.Client.CreateDirectChat(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.ChatId, nil
}

func (c *Client) GetChat(ctx context.Context, chatID string) (models.Chat, error) {
	req := &chat.GetChatRequest{ChatId: chatID}
	resp, err := c.Client.GetChat(ctx, req)
	if err != nil {
		return models.Chat{}, err
	}
	return models.Chat{
		ID:           resp.Id,
		Participants: resp.ParticipantIds,
		CreatedAt:    time.Now(), // Если нет поля в proto, заглушка; иначе resp.CreatedAt.AsTime() если есть timestamp
	}, nil
}

func (c *Client) ListUserChats(ctx context.Context, userID string) ([]*models.Chat, error) {
	req := &chat.ListUserChatsRequest{UserId: userID}
	resp, err := c.Client.ListUserChats(ctx, req)
	if err != nil {
		return nil, err
	}
	var chats []*models.Chat
	for _, ch := range resp.Chats {
		chats = append(chats, &models.Chat{
			ID:           ch.Id,
			Participants: ch.ParticipantIds,
			CreatedAt:    time.Now(),
		})
	}
	return chats, nil
}

func (c *Client) ListChatMembers(ctx context.Context, chatID string) ([]string, error) {
	req := &chat.ListChatMembersRequest{ChatId: chatID}
	resp, err := c.Client.ListChatMembers(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.UserIds, nil
}

func (c *Client) SendMessage(ctx context.Context, chatID, text string) (models.ChatMessage, error) {
	req := &chat.SendMessageRequest{ChatId: chatID, Text: text}
	resp, err := c.Client.SendMessage(ctx, req)
	if err != nil {
		return models.ChatMessage{}, err
	}
	return models.ChatMessage{
		ID:        resp.Id,
		ChatID:    resp.ChatId,
		SenderID:  resp.SenderId,
		Text:      resp.Text,
		CreatedAt: time.UnixMilli(resp.TimestampUnixMs),
	}, nil
}

func (c *Client) ListMessages(ctx context.Context, chatID string, limit int32, cursor string) (models.ChatListMessagesResponse, error) {
	req := &chat.ListMessagesRequest{ChatId: chatID, Limit: limit, Cursor: &cursor}
	resp, err := c.Client.ListMessages(ctx, req)
	if err != nil {
		return models.ChatListMessagesResponse{}, err
	}
	var messages []*models.ChatMessage
	for _, msg := range resp.Messages {
		messages = append(messages, &models.ChatMessage{
			ID:        msg.Id,
			ChatID:    msg.ChatId,
			SenderID:  msg.SenderId,
			Text:      msg.Text,
			CreatedAt: time.UnixMilli(msg.TimestampUnixMs),
		})
	}
	return models.ChatListMessagesResponse{
		Messages:   messages,
		NextCursor: *resp.NextCursor,
	}, nil
}

func (c *Client) StreamMessages(ctx context.Context, chatID string, sinceUnixMs int64) (<-chan *models.ChatMessage, error) {
	req := &chat.StreamMessagesRequest{ChatId: chatID, SinceUnixMs: &sinceUnixMs}
	stream, err := c.Client.StreamMessages(ctx, req)
	if err != nil {
		return nil, err
	}
	ch := make(chan *models.ChatMessage)
	go func() {
		defer close(ch)
		for {
			msg, err := stream.Recv()
			if err != nil {
				return
			}
			ch <- &models.ChatMessage{
				ID:        msg.Id,
				ChatID:    msg.ChatId,
				SenderID:  msg.SenderId,
				Text:      msg.Text,
				CreatedAt: time.UnixMilli(msg.TimestampUnixMs),
			}
		}
	}()
	return ch, nil
}
