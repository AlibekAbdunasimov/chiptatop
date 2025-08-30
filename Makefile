APP_NAME=chiptatop-bot
BIN_DIR=bin

.PHONY: run build tidy deps docker-build docker-run

run:
	go run ./cmd/bot

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/bot

tidy:
	go mod tidy

deps: tidy

docker-build:
	docker build -t $(APP_NAME):dev .

docker-run:
	docker run --rm -e TELEGRAM_BOT_TOKEN -e ENVIRONMENT=production $(APP_NAME):dev


