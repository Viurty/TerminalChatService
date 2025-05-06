package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	pb "example.com/myapp/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Ожидаемый ввод: %s <address> <name>\n", os.Args[0])
		os.Exit(1)
	}
	address, name := os.Args[1], os.Args[2]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Не удалось подключиться: %v", err)
	}
	defer conn.Close()

	client := pb.NewChatServiceClient(conn)
	stream, err := client.Chat(ctx)
	if err != nil {
		log.Fatalf("Ошибка открытия стрима: %v", err)
	}

	// Отправка сообщений
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			text, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("Ошибка чтения: %v", err)
				return
			}
			text = strings.TrimSpace(text)

			if text == "/exit" {
				stream.CloseSend()
				cancel()
				return
			}

			if err := stream.Send(&pb.ChatMessage{Name: name, Text: text}); err != nil {
				log.Printf("Ошибка Send: %v", err)
				return
			}
		}
	}()

	// Приём сообщений
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("Сервер завершил чат.")
			break
		}
		if err != nil {
			log.Printf("Recv error: %v", err)
			break
		}
		if msg.GetIsServer() {
			fmt.Printf("%s\n", msg.GetText())
		} else {
			fmt.Printf("[%s]: %s\n", msg.GetName(), msg.GetText())
		}
	}

	fmt.Println("Клиент завершил работу.")
}
