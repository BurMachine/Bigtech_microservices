package chat

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/models"
)

// TODO getCurrentUserID извлекает ID текущего пользователя из контекста (middleware добавляет ctx.Value("user_id", userID))
func getCurrentUserID(ctx context.Context) (string, error) {
	//userID, ok := ctx.Value("user_id").(string)
	//if !ok || userID == "" {
	//	return "", ErrPermissionDenied
	//}
	return "c3049516-fd64-479c-aca8-976c42df62ce", nil
}

// isUserParticipant проверяет, является ли userID участником чата
func isUserParticipant(chat *models.Chat, userID string) bool {
	for _, p := range chat.Participants {
		if p == userID {
			return true
		}
	}
	return false
}
