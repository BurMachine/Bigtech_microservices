// spammer.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type MessageRequest struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

var (
	baseURL = flag.String("url", "http://localhost:8078", "Base URL API (например http://localhost:8080)")
	chatID  = flag.String("chat", "05187671-b98e-4322-b0f0-235167cc0f8e", "ID чата")
	rps     = flag.Float64("rps", 5.0, "Примерное количество запросов в секунду")
)

var messages = []string{
	"Привет! Как дела?",
	"Здарова!",
	"Го",
	"Кто-нибудь живой?",
	"Тест нагрузки",
	"хэллоу",
	"тут спам-бот завёлся",
	"хрш хрш",
	"пинг",
	"проверка связи",
	"всем йо!",
	"чё как?",
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	url := fmt.Sprintf("%s/v1/chats/%s/messages", *baseURL, *chatID)

	fmt.Printf("Спамим в чат → %s\n", url)
	fmt.Printf("Цель: ~%.1f RPS | Ctrl+C — остановить\n\n", *rps)

	// Вычисляем интервал между запросами
	interval := time.Second / time.Duration(*rps)
	if *rps <= 0 {
		interval = 100 * time.Millisecond // защита от деления на 0
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	count := 0
	start := time.Now()

	for range ticker.C {
		count++

		text := messages[rand.Intn(len(messages))]
		if count%13 == 0 {
			text = fmt.Sprintf("ДЛИННОЕ СООБЩЕНИЕ #%d — проверяем нагрузку и лимиты", count)
		} else {
			text = fmt.Sprintf("%s [#%d]", text, count)
		}

		payload := MessageRequest{
			ChatID: *chatID,
			Text:   text,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Ошибка marshal JSON: %v", err)
			continue
		}

		resp, err := http.Post(url, "application/json", bytes.NewReader(body))
		if err != nil {
			log.Printf("Ошибка отправки: %v", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Сервер вернул %d %s", resp.StatusCode, resp.Status)
		} else {
			fmt.Printf("Отправлено #%d: %s\n", count, text)
		}

		resp.Body.Close()

		// Каждые 100 запросов — статистика
		if count%100 == 0 {
			elapsed := time.Since(start).Seconds()
			actualRPS := float64(count) / elapsed
			fmt.Printf("\nСтатистика: отправлено %d | реальный RPS: %.1f\n\n", count, actualRPS)
		}
	}
}
