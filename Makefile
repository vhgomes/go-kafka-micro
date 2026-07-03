```makefile
# =============================================================================
# Makefile para Go Kafka Microservices
# =============================================================================
# Comandos principais:
#   make help           - Mostra esta ajuda
#   make build          - Compila ambos os serviços (order e notification)
#   make run            - Executa ambos os serviços localmente (requer Kafka em execução)
#   make test           - Executa todos os testes (quando implementados)
#   make lint           - Executa golangci-lint em ambos os serviços
#   make docker-build   - Constrói as imagens Docker para ambos os serviços
#   make docker-up      - Sobe todos os containers (Kafka + serviços) via docker-compose
#   make docker-down    - Derruba todos os containers
#   make clean          - Remove binários e arquivos temporários
# =============================================================================

# =============================================================================
# Configurações Gerais
# =============================================================================
GO           := go
GOFLAGS      := -mod=readonly
BIN_DIR      := bin
DOCKER_COMPOSE := docker-compose

# Cores para output (opcional)
COLOR_RESET  := \033[0m
COLOR_GREEN  := \033[32m
COLOR_YELLOW := \033[33m
COLOR_CYAN   := \033[36m

# =============================================================================
# Serviços
# =============================================================================
ORDER_SERVICE_DIR       := order-service
NOTIFICATION_SERVICE_DIR := notification-service

ORDER_BIN       := $(BIN_DIR)/order-service
NOTIFICATION_BIN := $(BIN_DIR)/notification-service

# =============================================================================
# Alvos Padrão
# =============================================================================
.PHONY: all
all: help

.PHONY: help
help:
	@printf "$(COLOR_CYAN)Comandos disponíveis:$(COLOR_RESET)\n"
	@printf "  $(COLOR_GREEN)make build$(COLOR_RESET)          Compila ambos os serviços\n"
	@printf "  $(COLOR_GREEN)make run$(COLOR_RESET)            Executa ambos os serviços localmente (requer Kafka)\n"
	@printf "  $(COLOR_GREEN)make test$(COLOR_RESET)           Executa todos os testes\n"
	@printf "  $(COLOR_GREEN)make lint$(COLOR_RESET)           Executa golangci-lint\n"
	@printf "  $(COLOR_GREEN)make docker-build$(COLOR_RESET)   Constrói imagens Docker\n"
	@printf "  $(COLOR_GREEN)make docker-up$(COLOR_RESET)      Sobe containers via docker-compose\n"
	@printf "  $(COLOR_GREEN)make docker-down$(COLOR_RESET)    Derruba containers\n"
	@printf "  $(COLOR_GREEN)make docker-logs$(COLOR_RESET)    Mostra logs dos containers\n"
	@printf "  $(COLOR_GREEN)make clean$(COLOR_RESET)          Remove binários e arquivos temporários\n"
	@printf "  $(COLOR_GREEN)make deps$(COLOR_RESET)           Baixa dependências de ambos os serviços\n"

# =============================================================================
# Build
# =============================================================================
.PHONY: build build-order build-notification
build: build-order build-notification ## Compila todos os serviços
	@printf "$(COLOR_GREEN)✔ Build concluído$(COLOR_RESET)\n"

build-order:
	@printf "$(COLOR_YELLOW)▶ Compilando order-service...$(COLOR_RESET)\n"
	@mkdir -p $(BIN_DIR)
	cd $(ORDER_SERVICE_DIR) && $(GO) build $(GOFLAGS) -o ../$(ORDER_BIN) ./cmd
	@printf "$(COLOR_GREEN)✔ order-service compilado em $(ORDER_BIN)$(COLOR_RESET)\n"

build-notification:
	@printf "$(COLOR_YELLOW)▶ Compilando notification-service...$(COLOR_RESET)\n"
	@mkdir -p $(BIN_DIR)
	cd $(NOTIFICATION_SERVICE_DIR) && $(GO) build $(GOFLAGS) -o ../$(NOTIFICATION_BIN) ./cmd
	@printf "$(COLOR_GREEN)✔ notification-service compilado em $(NOTIFICATION_BIN)$(COLOR_RESET)\n"

# =============================================================================
# Execução Local (requer Kafka em execução)
# =============================================================================
.PHONY: run run-order run-notification
run: run-order run-notification ## Executa ambos os serviços (em foreground, use Ctrl+C para parar)

run-order: build-order
	@printf "$(COLOR_YELLOW)▶ Iniciando order-service...$(COLOR_RESET)\n"
	@$(ORDER_BIN)

run-notification: build-notification
	@printf "$(COLOR_YELLOW)▶ Iniciando notification-service...$(COLOR_RESET)\n"
	@$(NOTIFICATION_BIN)

# =============================================================================
# Testes
# =============================================================================
.PHONY: test test-order test-notification
test: test-order test-notification ## Executa todos os testes
	@printf "$(COLOR_GREEN)✔ Todos os testes concluídos$(COLOR_RESET)\n"

test-order:
	@printf "$(COLOR_YELLOW)▶ Testando order-service...$(COLOR_RESET)\n"
	cd $(ORDER_SERVICE_DIR) && $(GO) test -cover ./...

test-notification:
	@printf "$(COLOR_YELLOW)▶ Testando notification-service...$(COLOR_RESET)\n"
	cd $(NOTIFICATION_SERVICE_DIR) && $(GO) test -cover ./...

# =============================================================================
# Lint
# =============================================================================
.PHONY: lint lint-order lint-notification
lint: lint-order lint-notification ## Executa golangci-lint (se instalado)
	@printf "$(COLOR_GREEN)✔ Lint concluído$(COLOR_RESET)\n"

lint-order:
	@printf "$(COLOR_YELLOW)▶ Lint order-service...$(COLOR_RESET)\n"
	cd $(ORDER_SERVICE_DIR) && golangci-lint run ./...

lint-notification:
	@printf "$(COLOR_YELLOW)▶ Lint notification-service...$(COLOR_RESET)\n"
	cd $(NOTIFICATION_SERVICE_DIR) && golangci-lint run ./...

# =============================================================================
# Docker
# =============================================================================
.PHONY: docker-build docker-up docker-down docker-logs docker-restart

docker-build: ## Constrói as imagens Docker para ambos os serviços
	@printf "$(COLOR_YELLOW)▶ Construindo imagens Docker...$(COLOR_RESET)\n"
	$(DOCKER_COMPOSE) build

docker-up: ## Sobe todos os containers (Kafka + serviços)
	@printf "$(COLOR_YELLOW)▶ Subindo containers...$(COLOR_RESET)\n"
	$(DOCKER_COMPOSE) up -d
	@printf "$(COLOR_GREEN)✔ Containers em execução. Acesse:$(COLOR_RESET)\n"
	@printf "  Order Service:      http://localhost:8080\n"
	@printf "  Notification:       http://localhost:8081\n"
	@printf "  Kafdrop (UI):       http://localhost:9000\n"

docker-down: ## Derruba todos os containers
	@printf "$(COLOR_YELLOW)▶ Derrubando containers...$(COLOR_RESET)\n"
	$(DOCKER_COMPOSE) down -v
	@printf "$(COLOR_GREEN)✔ Containers removidos$(COLOR_RESET)\n"

docker-logs: ## Mostra logs de todos os containers (follow)
	$(DOCKER_COMPOSE) logs -f

docker-restart: docker-down docker-up ## Reinicia todos os containers

# =============================================================================
# Utilitários
# =============================================================================
.PHONY: deps clean

deps: ## Baixa as dependências de ambos os serviços
	@printf "$(COLOR_YELLOW)▶ Baixando dependências...$(COLOR_RESET)\n"
	cd $(ORDER_SERVICE_DIR) && $(GO) mod download
	cd $(NOTIFICATION_SERVICE_DIR) && $(GO) mod download
	@printf "$(COLOR_GREEN)✔ Dependências baixadas$(COLOR_RESET)\n"

clean: ## Remove binários e arquivos temporários
	@printf "$(COLOR_YELLOW)▶ Limpando...$(COLOR_RESET)\n"
	rm -rf $(BIN_DIR)
	rm -f $(ORDER_SERVICE_DIR)/*.test $(NOTIFICATION_SERVICE_DIR)/*.test
	rm -f $(ORDER_SERVICE_DIR)/coverage.out $(NOTIFICATION_SERVICE_DIR)/coverage.out
	@printf "$(COLOR_GREEN)✔ Limpeza concluída$(COLOR_RESET)\n"
```

## 📝 Como usar

### Comandos básicos

| Comando | Descrição |
|---------|-----------|
| `make help` | Mostra a lista de comandos disponíveis |
| `make build` | Compila ambos os serviços |
| `make run` | Executa ambos os serviços localmente (requer Kafka em execução) |
| `make docker-up` | Sobe todos os containers via Docker Compose |
| `make docker-down` | Derruba todos os containers |
| `make clean` | Remove binários e arquivos temporários |

### Exemplos de uso

```bash
# Compilar os serviços
make build

# Subir a stack completa com Docker
make docker-up

# Ver logs de todos os serviços
make docker-logs

# Testar localmente (com Kafka em execução)
make run

# Limpar arquivos gerados
make clean
```

### Observações

- O alvo `run` pressupõe que o Kafka esteja em execução (localmente ou via Docker).
- Os binários são gerados no diretório `bin/`.
- O alvo `docker-build` utiliza o `docker-compose.yaml` da raiz.
- Os alvos de `lint` exigem o `golangci-lint` instalado; caso contrário, você pode removê-los ou adicionar verificação de disponibilidade.

---

**💡 Dica:** Para adicionar mais serviços no futuro, basta estender as variáveis `ORDER_SERVICE_DIR` e criar novos alvos correspondentes.
