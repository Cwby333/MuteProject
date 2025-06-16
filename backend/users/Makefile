all: build compose_stop none_images_rmi rm_image users_build compose

build:
	go build -o main ./cmd/main.go
compose_stop:
	docker compose stop
	docker compose rm
rm_image:
	docker rmi users-microservice-users
none_images_rmi:
	docker image prune
users_build:
	docker build -t users:1.0 .
compose:
	docker compose up
cover-html:
	go tool cover -html=.coverage.out -o coverage.html
coverage:
	go test ./... -coverprofile=.coverage.out
update-mocks:
	mockgen -source ./internal/service/userService/service.go -destination ./internal/service/userService/mock_userService/mocks.go -package userServiceMocks && mockgen -source ./internal/adapters/transport/http/userRouter/router.go -destination ./internal/adapters/transport/http/userRouter/userRouterMocks/mocks.go -package userRouterMocks