package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	pb "github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr     = flag.String("addr", "localhost:8082", "gRPC server address")
	interval = flag.Duration("interval", 300*time.Millisecond, "interval between messages")
	chatID   = flag.String("chat", "05187671-b98e-4322-b0f0-235167cc0f8e", "chat_id")
)

func main() {
	flag.Parse()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewChatServiceClient(conn)
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	words := []string{"Здарова", "Как дела?", "Го ", "Прив", "Чё по погоде?", "Лол", "хэллоу", "))", "хрш"}

	fmt.Printf("Спамим в %s каждые %v → %s\n", *addr, *interval, *chatID)

	for i := 1; ; i++ {
		<-ticker.C

		text := words[rand.Intn(len(words))]
		if i%17 == 0 { // иногда шлём длинное
			text = fmt.Sprintf("Длинное сообщение номер %d — проверяем как работает с большими текстами", i)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err := client.SendMessage(ctx, &pb.SendMessageRequest{
			ChatId: *chatID,
			Text:   text,
		})
		cancel()

		if err != nil {
			log.Printf("Ошибка #%d: %v", i, err)
		} else {
			fmt.Printf("Отправлено #%d: %s\n", i, text)
		}
	}
}
