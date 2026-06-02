# ╔══════════════════════════════════════════════╗
# ║  Makefile — Go monorepo                      ║
# ║  Замінити: APP_NAME, WEB_DIR, CMD            ║
# ╚══════════════════════════════════════════════╝

# ── Налаштування ────────────────────────────────
APP_NAME          = isitdead           # назва бінарника
TELEGRAM_APP_NAME = isitdead-telegram
BUILD_DIR         = ./bin
WEB_DIR           = ./web           # папка фронтенду
CMD               = ./cmd/server    # entrypoint Go
TELEGRAM_CMD      = ./integration/telegram

VERSION   = $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   = -X main.version=$(VERSION)

# ── Кольори для виводу ──────────────────────────
GREEN  = \033[0;32m
YELLOW = \033[0;33m
RESET  = \033[0m

# ── За замовчуванням ─────────────────────────────
.DEFAULT_GOAL := help
.PHONY: build build-telegram run dev dev-front dev-back watch clean test test-postgres docker-up docker-down lint tidy help migrate

# ── Backend ──────────────────────────────────────

## Зібрати бінарник
build:
	@echo "$(GREEN)▶ Building $(APP_NAME) $(VERSION)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) $(CMD)

## Зібрати Telegram integration binary
build-telegram:
	@echo "$(GREEN)▶ Building $(TELEGRAM_APP_NAME) $(VERSION)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(TELEGRAM_APP_NAME) $(TELEGRAM_CMD)

## Зібрати і запустити
run: build
	@echo "$(GREEN)▶ Running...$(RESET)"
	$(BUILD_DIR)/$(APP_NAME)

## Запустити без збірки (для швидкого тесту)
dev-back:
	go run $(CMD)

## Hot-reload через air (go install github.com/air-verse/air@latest)
watch:
	@which air > /dev/null || (echo "$(YELLOW)air не знайдено. Встанови: go install github.com/air-verse/air@latest$(RESET)" && exit 1)
	air -c .air.toml

# ── Frontend ─────────────────────────────────────

## Зібрати фронтенд
build-front:
	@echo "$(GREEN)▶ Building frontend...$(RESET)"
	cd $(WEB_DIR) && npm ci && npm run build

## Запустити dev-сервер фронтенду
dev-front:
	cd $(WEB_DIR) && npm run dev

## Запустити і бек і фронт одночасно
dev:
	@echo "$(GREEN)▶ Starting dev mode...$(RESET)"
	@make -j2 watch dev-front

# ── Збірка всього ────────────────────────────────

## Зібрати фронт + бек разом (для деплою)
build-all: build-front build build-telegram
	@echo "$(GREEN)✔ Full build done$(RESET)"

# ── Тести ────────────────────────────────────────

## Запустити всі Go-тести
test:
	go test ./cmd/... ./internal/... ./integration/... -v

## Запустити тести проти PostgreSQL у Docker
test-postgres:
	docker compose --profile test run --rm test

## Швидко запустити PostgreSQL + app у Docker
docker-up:
	docker compose up --build

## Зупинити Docker-сервіси
docker-down:
	docker compose down

## Тести з покриттям
coverage:
	go test ./cmd/... ./internal/... ./integration/... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✔ coverage.html готовий$(RESET)"

# ── Якість коду ──────────────────────────────────

## Запустити golangci-lint (go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	@which golangci-lint > /dev/null || (echo "$(YELLOW)golangci-lint не знайдено$(RESET)" && exit 1)
	golangci-lint run ./...

## Підчистити go.mod / go.sum
tidy:
	go mod tidy

## Форматування коду
fmt:
	gofmt -w .

# ── БД (опціонально) ─────────────────────────────

## Застосувати міграції (потрібен goose або migrate)
migrate:
	@echo "$(YELLOW)Налаштуй цю ціль під свій інструмент міграцій$(RESET)"
	# goose -dir ./migrations postgres "$(DATABASE_URL)" up

# ── Прибирання ───────────────────────────────────

## Видалити зібрані файли
clean:
	@echo "$(YELLOW)▶ Cleaning...$(RESET)"
	rm -rf $(BUILD_DIR) $(WEB_DIR)/dist coverage.out coverage.html

# ── Довідка ──────────────────────────────────────

## Показати доступні команди
help:
	@echo ""
	@echo "$(GREEN)$(APP_NAME)$(RESET) — доступні команди:"
	@echo ""
	@grep -E '^##' Makefile | sed 's/## /  /' | while IFS= read -r line; do echo "  $$line"; done
	@echo ""
