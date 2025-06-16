# ####################################
# # 1) Builder stage (Alpine + Go 1.24)
# ####################################
# FROM golang:1.24-alpine AS builder

# # git нужен для go modules
# RUN apk add --no-cache git

# WORKDIR /app

# # Копируем .env и конфигурацию на этапе сборки
# COPY .env .env
# COPY config config

# # Кэшируем модули
# COPY go.mod go.sum ./
# RUN go mod download

# # Копируем весь исходник
# COPY . .

# # Собираем статический бинарь, исходник в cmd/main.go
# RUN CGO_ENABLED=0 GOOS=linux \
#     go build -ldflags="-s -w" \
#       -o cmd/main \
#       ./cmd/main.go

# ####################################
# # 2) Runtime stage (Alpine)
# ####################################
# FROM alpine:3.21.3

# # Устанавливаем только сертификаты
# RUN apk add --no-cache ca-certificates

# # Рабочая директория
# WORKDIR /app/cmd

# # Копируем .env и папку config (поддержка config.MustLoad)
# COPY --from=builder /app/.env .env
# COPY --from=builder /app/config ./config

# # Копируем двоичный файл
# COPY --from=builder /app/cmd/main .

# # Делаем бинарь исполняемым
# RUN chmod +x ./main

# EXPOSE 8082
# ENTRYPOINT ["./main"]



####################################
# 1) Builder stage (Alpine + Go 1.24)
####################################
FROM golang:1.24-alpine AS builder

# git нужен для работы с модулями
RUN apk add --no-cache git

WORKDIR /app

# Копируем .env, конфиг и миграции для последующего runtime
COPY .env .env
COPY config config
COPY migrations migrations

# Кэшируем модули
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходник
COPY . .

# Собираем статический бинарь (cmd/main.go)
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" \
      -o cmd/main \
      ./cmd/main.go

####################################
# 2) Runtime stage (Alpine)
####################################
FROM alpine:3.21.3

# Устанавливаем лишь сертификаты
RUN apk add --no-cache ca-certificates

# Переключаемся в директорию бинаря
WORKDIR /app/cmd

# Копируем .env, конфиг, миграции
COPY --from=builder /app/.env .env
COPY --from=builder /app/config ./config
COPY --from=builder /app/migrations ./migrations

# Копируем сам бинарь
COPY --from=builder /app/cmd/main .

# Делаем исполняемым
RUN chmod +x ./main

EXPOSE 8888
ENTRYPOINT ["./main"]
