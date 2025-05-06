package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	pb "example.com/myapp/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedChatServiceServer
	mu        sync.RWMutex
	clients   map[pb.ChatService_ChatServer]int //хранит всех пользователей как ключи и количество их нарушений как значение
	ban_words []string
}

func (s *server) addClient(stream pb.ChatService_ChatServer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[stream] = 0
}

func (s *server) removeClient(stream pb.ChatService_ChatServer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, stream)
}

func (s *server) printMessage(msg *pb.ChatMessage, sender pb.ChatService_ChatServer) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	response := &pb.ChatMessage{
		IsServer: false,
		Name:     msg.GetName(),
		Text:     msg.GetText(),
	}
	for client := range s.clients {
		if client == sender {
			continue
		}
		if err := client.Send(response); err != nil {
			log.Printf("Ошибка: %v", err)
		}
	}
}

func (s *server) printWarn(client pb.ChatService_ChatServer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	response := &pb.ChatMessage{
		IsServer: true,
		Name:     "Warn",
		Text:     "ОШИБКА: Нельзя ругаться",
	}
	s.clients[client] += 1
	if err := client.Send(response); err != nil {
		log.Printf("Ошибка: %v", err)
	}

}

func (s *server) printBan(client pb.ChatService_ChatServer) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	response := &pb.ChatMessage{
		IsServer: true,
		Name:     "Ban",
		Text:     "ОШИБКА: Ты больше не можешь писать в этот чат",
	}
	if err := client.Send(response); err != nil {
		log.Printf("Ошибка: %v", err)
	}

}

func (s *server) Chat(stream pb.ChatService_ChatServer) error {
	s.addClient(stream)
	defer s.removeClient(stream)

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "Ошибка: %v", err)
		}

		if s.clients[stream] == 3 {
			s.printBan(stream)
		} else if isBan(s.ban_words, msg.GetText()) {
			s.printWarn(stream)
		} else {
			s.printMessage(msg, stream)
		}
	}
}

func isBan(words []string, msg string) bool {
	for _, word := range words {
		if word == "" {
			continue
		}
		if strings.Contains(msg, word) {
			return true
		}
	}
	return false
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Ожидаемый ввод: %s <address> <file_path>\n", os.Args[0])
		os.Exit(1)
	}
	address := os.Args[1]
	file_path := os.Args[2]

	data, err := os.ReadFile(file_path)
	if err != nil {
		log.Printf("Не удалось прочитать файл: %v", err)
	}
	text := string(data)
	words := strings.Split(text, "\n")

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Ошибка запуска слушателя: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChatServiceServer(grpcServer, &server{clients: make(map[pb.ChatService_ChatServer]int), ban_words: words})
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Ошибка работы gRPC-сервера: %v", err)
	}
}
