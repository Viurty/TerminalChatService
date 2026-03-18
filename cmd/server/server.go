package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	pb "example.com/myapp/api"
	"example.com/myapp/internal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type user struct {
	StreamUser pb.ChatService_ChatServer
	Warn       int
}

type server struct {
	pb.UnimplementedChatServiceServer
	mu        sync.RWMutex
	clients   map[string]user //хранит все логины как ключи и структуру(stream+warn) как значение
	ban_words []string
	passwords []string
}

// Добавляем пользователя в словарь, чтобы в дальнейшем можно было сделать всем рассылку сообщений
func (s *server) addUser(login string, stream pb.ChatService_ChatServer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[login] = user{StreamUser: stream, Warn: 0}
}

// Удаляем пользователя при любом выходе(чтобы не хранить мусорные подключения)
func (s *server) removeUser(login string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, login)
}

// Отправляет всем пользователям сообщение (кроме отправителя)
func (s *server) printMessage(msg *pb.ChatMessage, sender pb.ChatService_ChatServer) {
	response := &pb.ChatMessage{
		IsServer: false,
		Name:     msg.GetName(),
		Text:     msg.GetText(),
	}
	s.mu.RLock()
	targetStreams := make([]pb.ChatService_ChatServer, 0, len(s.clients))
	for _, u := range s.clients {
		if u.StreamUser == sender {
			continue
		}
		targetStreams = append(targetStreams, u.StreamUser)
	}
	s.mu.RUnlock()

	for _, target := range targetStreams {
		if err := target.Send(response); err != nil {
			log.Printf("Ошибка: %v", err)
		}
	}
}

// Отправляет внутренние сообщения от сервера
func (s *server) printFromServer(text string, client pb.ChatService_ChatServer) {
	response := &pb.ChatMessage{
		IsServer: true,
		Text:     text,
	}
	if err := client.Send(response); err != nil {
		log.Printf("Ошибка: %v", err)
	}

}

func parseBanCommand(text string) (string, bool) {
	fields := strings.Fields(text)
	if len(fields) != 2 || fields[0] != "/ban" {
		return "", false
	}
	return fields[1], true
}

// Вся логика чаттинга
func (s *server) Chat(stream pb.ChatService_ChatServer) error {
	claims := internal.GetClaims(stream.Context())
	if claims == nil || claims.Login == "" {
		return status.Error(codes.Unauthenticated, "missing claims")
	}
	s.addUser(claims.Login, stream)
	defer s.removeUser(claims.Login)

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "Ошибка: %v", err)
		}
		msg.Name = claims.Login
		msg.Role = claims.Role
		if strings.HasPrefix(msg.Text, "/ban") {
			banLogin, ok := parseBanCommand(msg.Text)
			if !ok {
				s.printFromServer("ОШИБКА: формат команды /ban <login>", stream)
				continue
			}
			if msg.Role != "admin" {
				s.printFromServer("ОШИБКА: Ты не администратор", stream)
			} else {
				var bannedUserStream pb.ChatService_ChatServer
				s.mu.Lock()
				if u, exists := s.clients[banLogin]; exists {
					u.Warn = 3
					s.clients[banLogin] = u
					bannedUserStream = u.StreamUser
				}
				s.mu.Unlock()

				if bannedUserStream != nil {
					s.printFromServer("ОШИБКА: Ты больше не можешь писать в этот чат", bannedUserStream)
					log.Printf("User %s был забанен", banLogin) // это для тестов
				} else {
					s.printFromServer("ОШИБКА: Пользователь не найден", stream)
				}
			}
			continue
		}

		s.mu.RLock()
		current, exists := s.clients[msg.Name]
		s.mu.RUnlock()
		if !exists {
			return status.Error(codes.Internal, "user session not found")
		}
		if current.Warn >= 3 {
			s.printFromServer("ОШИБКА: Ты больше не можешь писать в этот чат", stream)
			continue
		} else if isBan(s.ban_words, msg.GetText()) {
			s.printFromServer("ОШИБКА: Нельзя ругаться", stream)
			s.mu.Lock()
			current = s.clients[msg.Name]
			current.Warn++
			s.clients[msg.Name] = current
			s.mu.Unlock()
			continue
		} else {
			s.printMessage(msg, stream)
			continue
		}
	}
}

// Авторизует пользователя
func (s *server) AuthUser(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	isAuth, role := internal.CheckPassword(s.passwords, req.Login, req.Password)
	if !isAuth {
		return nil, status.Error(codes.Unauthenticated, "authorization failed")
	}

	token, err := internal.GenerateJWT(req.Login, role)
	if err != nil {
		log.Printf("Ошибка генерации JWT: %v", err)
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &pb.LoginResponse{Token: token}, nil
}

// Проверка на сдержание запрещенных слов
func isBan(words []string, msg string) bool {
	msgLower := strings.ToLower(msg)
	for _, word := range words {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}
		if strings.Contains(msgLower, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

type wrappedServerStream struct {
	grpc.ServerStream
	WrappedContext context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.WrappedContext
}

// middleware, который сохраняет метаданные в контекст от JTW
func authInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "no metadata")
	}
	values := md["authorization"]
	if len(values) == 0 {
		return status.Error(codes.Unauthenticated, "missing token")
	}

	token := values[0]
	if len(token) > 7 && strings.ToLower(token[:7]) == "bearer " {
		token = token[7:]
	}

	claims, err := internal.ValidateToken(token)
	if err != nil {
		return status.Error(codes.Unauthenticated, "invalid token")
	}

	wrapped := &wrappedServerStream{
		ServerStream:   ss,
		WrappedContext: internal.SaveClaims(ss.Context(), claims),
	}

	return handler(srv, wrapped)
}

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Ожидаемый ввод: %s <address> <file path with ban words> <file path with passwords>\n", os.Args[0])
		os.Exit(1)
	}
	address, ban_filepath, password_filepath := os.Args[1], os.Args[2], os.Args[3]

	data, err := os.ReadFile(password_filepath)
	if err != nil {
		log.Fatalf("Не удалось прочитать файл с паролями: %v", err)
	}
	passwords := strings.Split(string(data), "\n")

	data, err = os.ReadFile(ban_filepath)
	if err != nil {
		log.Fatalf("Не удалось прочитать файл с бан-словами: %v", err)
	}
	words := strings.Split(string(data), "\n")

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Ошибка запуска слушателя: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.StreamInterceptor(authInterceptor))
	pb.RegisterChatServiceServer(grpcServer, &server{clients: make(map[string]user), ban_words: words, passwords: passwords})
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Ошибка работы gRPC-сервера: %v", err)
	}
}
