# Variables
APP_NAME := grpc-cli-quiz

.PHONY: lint
lint:
	golangci-lint run

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix

.PHONY: mod-tidy
mod-tidy:
	go mod tidy

.PHONY: mod-verify
mod-verify:
	go mod verify

.PHONY: mod-download
mod-download:
	go mod download

proto-generate:
	protoc --go_out=./proto --go-grpc_out=./proto/ proto/quiz.proto

start:
	go run . start-quiz