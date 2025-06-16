FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o api-gateway .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/api-gateway .

EXPOSE 3000

CMD ["./api-gateway"]