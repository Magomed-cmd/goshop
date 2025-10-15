.PHONY: run test clean-mocks regenerate-mocks fresh-mocks build check verify compile lint

-include .env
export

lint:
	golangci-lint run

run:
	go run cmd/goshop/main.go

test:
	gotestsum --format testdox ./... -v -- -count=1

clean-mocks:
	@find . -type d -name "mocks" -exec rm -rf {} + 2>/dev/null || true

regenerate-mocks: clean-mocks
	@mockery --all

fresh-mocks: clean-mocks regenerate-mocks test

build:
	@go build -v ./...

build-fresh:
	@time go build -a -v ./...

check:
	@go vet ./...
	@go fmt ./...

verify: check build

compile:
	@go build -o bin/goshop ./cmd/goshop

setup-integration:
	@cd tests/integration && python3 -m venv venv
	@cd tests/integration && source venv/bin/activate && pip install -r requirements.txt

test-integration:
	@sleep 3
	@cd tests/integration && source venv/bin/activate && python -m pytest -v

test-integration-dev:
	@cd tests/integration && source venv/bin/activate && python -m pytest -v

test-all: test test-integration

tree:
	@tree -I 'venv|node_modules|.git|*.log|tmp|temp' -a

migrate-up:
	migrate -path ./migrations -database "$$(make print-dsn)" up

migrate-down:
	migrate -path ./migrations -database "$$(make print-dsn)" down

migrate-new:
	@if [ -z "$(name)" ]; then \
		echo "Укажи имя миграции: make migrate-new name=add_users_table!!!!!!!!!"; \
		exit 1; \
	fi; \
	migrate create -ext sql -dir ./migrations -seq $(name)