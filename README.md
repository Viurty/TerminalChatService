# gRPC Chat Server

Учебный чат на `gRPC` со стримингом сообщений, JWT-аутентификацией, ролями пользователей и базовой модерацией.

## Возможности

- двунаправленный чат через `Chat(stream ChatMessage) returns (stream ChatMessage)`
- аутентификация через `AuthUser(LoginRequest) -> LoginResponse`
- роли пользователей (`admin` / `user`)
- команда модерации `admin`: `/ban <login>`
- фильтрация сообщений по списку запрещённых слов
- хранение паролей в `bcrypt`-хеше

## Структура проекта

```
api/
	chatpb.proto          # gRPC контракт
cmd/
	adduser/              # утилита добавления пользователя в файл паролей
	client/               # CLI клиент чата
	server/               # gRPC сервер
internal/
	hash.go               # проверка/запись паролей
	jwt.go                # генерация/валидация JWT
```

## Требования

- `Go 1.23+`

## Подготовка

1. Клонируйте репозиторий и перейдите в папку проекта.
2. Установите секрет JWT (обязательно):

```zsh
export JWT_SECRET='your-strong-secret'
```

> Без `JWT_SECRET` сервер отклоняет вход пользователей (токен не генерируется).

## Формат входных файлов

### Файл пользователей

Формат строки:

```
<login>;<role>;<bcrypt_hash>
```

Пример:

```
alice;admin;$2a$12$...
bob;user;$2a$12$...
```

### Файл запрещённых слов

Одно слово на строку:

```
badword1
badword2
```

## Быстрый старт

### 1) Добавить пользователей

```zsh
go run ./cmd/adduser/adduser.go ./passwords.txt alice admin qwerty123
go run ./cmd/adduser/adduser.go ./passwords.txt bob user 12345678
```

### 2) Подготовить список бан-слов

```zsh
cat > ban_words.txt <<'EOF'
rude
bad
EOF
```

### 3) Запустить сервер

```zsh
export JWT_SECRET='your-strong-secret'
go run ./cmd/server/server.go 127.0.0.1:50051 ./ban_words.txt ./passwords.txt
```

### 4) Запустить клиентов

```zsh
go run ./cmd/client/client.go 127.0.0.1:50051 alice qwerty123
go run ./cmd/client/client.go 127.0.0.1:50051 bob 12345678
```

## Команды в чате

- `/exit` — выйти из чата
- `/ban <login>` — (только `admin`) запретить пользователю отправку сообщений

Если пользователь отправляет сообщение с запрещённым словом, сервер увеличивает число предупреждений. На 3 предупреждении пользователь больше не может писать.

## Протокол

`api/chatpb.proto`:

- `ChatMessage`
	- `role` — роль пользователя
	- `isServer` — системное сообщение сервера
	- `name` — имя отправителя
	- `text` — текст сообщения
- `LoginRequest`
	- `login`
	- `password`
- `LoginResponse`
	- `token`

## Проверка

```zsh
go test ./...
```

## Ограничения

- клиент подключается без TLS (`insecure`), проект рассчитан на локальную/учебную среду
- хранилище пользователей — файл, без БД
- бан-слова проверяются простым `substring`-поиском

## Troubleshooting

- `authorization failed`:
	- неверный логин/пароль или пользователь отсутствует в файле
- `failed to generate token`:
	- не установлен `JWT_SECRET`
- `ОШИБКА: формат команды /ban <login>`:
	- команда передана не в формате `/ban username`
