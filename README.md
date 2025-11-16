# gRPC Chat Application

Простой gRPC-чат с авторизацией по JWT, поддержкой ролей, бан-словами и системой предупреждений.

## Требования

- Go 1.20+
- protoc (Protocol Buffers Compiler)
- Плагины для генерации Go-кода:

```bash

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

# gRPC Chat Application
```
myapp/
├── api/
│   └── chatpb.proto
├── cmd/
│   ├── server/
│   │   └── server.go
│   ├── client/
│   │   └── client.go
│   └── adduser/
│       └── adduser.go
├── internal/
│   ├──jwt.go
│   │ 
│   └── hash.go
└── go.mod
```
# Установка и сборка
## Генерация gRPC кода
```bash

protoc --go_out=. --go-grpc_out=. api/chatpb.proto
```
## Сборка компонентов
```bash

# Сервер
go build -o chatserver cmd/server/server.go

# Клиент
go build -o chatclient cmd/client/client.go

# Утилита добавления пользователей
go build -o adduser cmd/adduser/adduser.go
```

# Настройка

## Добавляем пользователей
```bash

./adduser passwords.txt alice user alice123
./adduser passwords.txt bob admin bob123
```
## Создание файла с бан-словами

```bash

touch ban.txt
echo "badword1" > ban.txt
echo "badword2" >> ban.txt
echo "badword3" >> ban.txt
```

# Запуск приложения
## Запуск сервера

```bash

./chatserver localhost:8080 ban.txt passwords.txt
```
## Запуск клиентов
```bash

# Терминал 1
./chatclient localhost:8080 alice alice123

# Терминал 2
./chatclient localhost:8080 bob bob123
```