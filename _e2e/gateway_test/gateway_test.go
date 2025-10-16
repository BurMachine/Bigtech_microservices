package gateway_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGatewayAPI содержит набор тестов для REST API Gateway.
func TestGatewayAPI(t *testing.T) {
	const gatewayURL = "http://localhost:8078"

	client := &http.Client{Timeout: 10 * time.Second}

	// Тесты для User Service
	t.Run("CreateProfile", func(t *testing.T) {
		reqBody := map[string]string{
			"user_id":  uuid.New().String(),
			"nickname": "testuser",
			"email":    "test@example.com",
		}
		bodyBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp, err := client.Post(gatewayURL+"/v1/profile", "application/json", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var response map[string]interface{}
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &response)
		assert.NoError(t, err)
		assert.Equal(t, reqBody["nickname"], response["nickname"])
	})

	t.Run("GetProfileByID", func(t *testing.T) {
		userID := uuid.New().String()
		resp, err := client.Get(fmt.Sprintf("%s/v1/profile/%s", gatewayURL, userID))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		// Предполагаем, что заглушка возвращает валидный профиль
		var response map[string]interface{}
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &response)
		assert.NoError(t, err)
		assert.Equal(t, userID, response["user_id"])
	})

	// Тесты для Social Service
	t.Run("SendFriendRequest", func(t *testing.T) {
		userID := uuid.New().String()
		resp, err := client.Post(fmt.Sprintf("%s/v1/friends/%s/request", gatewayURL, userID), "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var response map[string]string
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response["request_id"])
	})

	// Тесты для Chat Service
	t.Run("SendMessage", func(t *testing.T) {
		chatID := uuid.New().String()
		reqBody := map[string]string{
			"chat_id":   chatID,
			"sender_id": uuid.New().String(),
			"text":      "Hello",
		}
		bodyBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp, err := client.Post(fmt.Sprintf("%s/v1/chats/%s/messages", gatewayURL, chatID), "application/json", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var response map[string]string
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response["message_id"])
	})

	t.Run("GetChat", func(t *testing.T) {
		chatID := uuid.New().String()
		resp, err := client.Get(fmt.Sprintf("%s/v1/chats/%s", gatewayURL, chatID))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		// Предполагаем, что заглушка возвращает валидный чат
		var response map[string]interface{}
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &response)
		assert.NoError(t, err)
		assert.Equal(t, chatID, response["id"])
	})
}
