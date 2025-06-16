# syntax=docker/dockerfile:1

####################################
# 1) Builder stage (Debian Bullseye)
####################################
FROM golang:1.24-bullseye AS builder

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      git gcc libc6-dev librdkafka-dev pkg-config \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o cmd/main ./cmd/main.go

####################################
# 2) Runtime stage (Debian Bullseye‑Slim)
####################################
FROM debian:bullseye-slim

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      ca-certificates librdkafka1 \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app/cmd
COPY --from=builder /app/cmd/main .


# убедимся, что бинарник точно исполняемый
RUN chmod +x ./main

EXPOSE 8080
ENTRYPOINT ["./main"]

