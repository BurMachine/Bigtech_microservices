package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/segmentio/kafka-go"
)

type Handler interface {
	Handle(ctx context.Context, msg *kafka.Message) error // msg.Payload — для обработки
}

// пример: "отправка уведомления" как console output
type NotificationHandler struct{}

func (NotificationHandler) Handle(ctx context.Context, msg *kafka.Message) error {
	log.Printf("[NOTIFICATION SENT] Event ID: %s, Topic: %s, Payload: %s",
		ExtractID(msg), msg.Topic, string(msg.Value)) // Считаем payload JSON/string
	return nil
}

// ExtractID извлекает ID из ключа сообщения или из payload
func ExtractID(msg *kafka.Message) string {
	// Приоритет 1: Ключ сообщения (PartitionKey)
	if len(msg.Key) > 0 {
		key := string(msg.Key)
		// Если ключ похож на UUID или ID
		if isValidID(key) {
			return key
		}
	}

	// Приоритет 3: Фолбэк на основе метаданных Kafka
	return generateFallbackID(msg)
}

// isValidID проверяет, похоже ли значение на ID/UUID
func isValidID(value string) bool {
	if value == "" {
		return false
	}

	// Проверяем на UUID
	if len(value) == 36 && strings.Contains(value, "-") {
		// Простая проверка формата UUID
		parts := strings.Split(value, "-")
		if len(parts) == 5 {
			return true
		}
	}

	// Проверяем на числовой ID
	if len(value) < 20 { // разумная длина для ID
		for _, char := range value {
			if (char < '0' || char > '9') && (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') {
				return false
			}
		}
		return true
	}

	return false
}

// extractIDFromPayload пытается извлечь ID из JSON payload
func extractIDFromPayload(payload []byte) string {
	if len(payload) == 0 {
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return ""
	}

	// Ищем ID в различных возможных полях
	possibleIDFields := []string{
		"id", "event_id", "message_id", "correlation_id",
		"eventId", "messageId", "correlationId", "uuid",
	}

	for _, field := range possibleIDFields {
		if idValue, exists := data[field]; exists {
			switch v := idValue.(type) {
			case string:
				if v != "" {
					return v
				}
			case float64: // JSON numbers
				return fmt.Sprintf("%.0f", v)
			}
		}
	}

	return ""
}

func generateFallbackID(msg *kafka.Message) string {
	return fmt.Sprintf("kafka-%s-p%d-o%d", msg.Topic, msg.Partition, msg.Offset)
}
