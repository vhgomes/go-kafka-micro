# Go Kafka Microservices

Um exemplo de **microsserviços em Go** com comunicação assíncrona via **Apache Kafka**, seguindo os princípios de **Arquitetura Hexagonal** (Ports & Adapters). O projeto demonstra a separação clara entre domínio, casos de uso e adaptadores, com injeção de dependência e foco em desacoplamento.

---

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────────────────────────────┐
│                         Order Service                          │
│  ┌─────────────┐   ┌────────────────┐   ┌───────────────────┐  │
│  │   Handler   │ → │  CreateOrder   │ → │  OrderRepository  │  │
│  │  (HTTP)     │   │   (Usecase)    │   │   (Port)          │  │
│  └─────────────┘   └────────────────┘   └───────────────────┘  │
│         │                   │                        │          │
│         │                   ▼                        ▼          │
│         │           ┌────────────────┐   ┌───────────────────┐  │
│         │           │ OrderPublisher │   │InMemoryOrderRepo  │  │
│         │           │   (Port)       │   │   (Adapter)       │  │
│         │           └────────────────┘   └───────────────────┘  │
│         │                    │                                  │
│         │                    ▼                                  │
│         │          ┌───────────────────┐                       │
│         └─────────▶│ KafkaPublisher    │                       │
│                    │   (Adapter)       │                       │
│                    └───────────────────┘                       │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            │  (Kafka Topic: orders)
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Notification Service                       │
│  ┌───────────────────┐   ┌──────────────────┐                 │
│  │  Kafka Consumer   │ → │ SendNotification  │                 │
│  │   (Adapter)       │   │   (Usecase)       │                 │
│  └───────────────────┘   └──────────────────┘                 │
│                                    │                           │
│                                    ▼                           │
│                          ┌────────────────────┐                │
│                          │ NotificationSender │                │
│                          │      (Port)        │                │
│                          └────────────────────┘                │
│                                    │                           │
│                                    ▼                           │
│                          ┌────────────────────┐                │
│                          │ LogNotification    │                │
│                          │   (Adapter)        │                │
│                          └────────────────────┘                │
└─────────────────────────────────────────────────────────────────┘
```

- **Order Service**: recebe requisições HTTP para criação de pedidos, persiste em memória (repositório in-memory) e publica o evento no Kafka.
- **Notification Service**: consome eventos do Kafka e envia notificações (atualmente via log, adaptável para e-mail/SMS).
- **Kafka**: atua como barramento de eventos, desacoplando os serviços.

---

## 🚀 Tecnologias

| Camada             | Tecnologia                                                                                            |
|--------------------|-------------------------------------------------------------------------------------------------------|
| **Linguagem**      | Go 1.23.4                                                                                             |
| **Web/HTTP**       | net/http (stdlib) – sem frameworks pesados                                                            |
| **Mensageria**     | [segmentio/kafka-go](https://github.com/segmentio/kafka-go) (cliente puro Go para Kafka)              |
| **Kafka Broker**   | [Bitnami Kafka](https://hub.docker.com/r/bitnami/kafka/) + Zookeeper                                  |
| **Container**      | Docker + Docker Compose                                                                               |
| **Observabilidade**| (futuro) Prometheus, OpenTelemetry                                                                    |
| **Testes**         | (futuro) `testing` + `testify` + testcontainers                                                      |

---

## 📁 Estrutura do Projeto

```
.
├── order-service/                # Serviço de pedidos
│   ├── cmd/
│   │   └── main.go               # Ponto de entrada
│   ├── internal/
│   │   ├── adapter/              # Adaptadores concretos (HTTP handlers, Kafka publisher, repo in-memory)
│   │   ├── domain/               # Entidades e regras de negócio (Order, Item)
│   │   ├── port/                 # Interfaces (repositório, publicador)
│   │   └── usecase/              # Casos de uso (CreateOrder)
│   ├── Dockerfile
│   └── go.mod
│
├── notification-service/         # Serviço de notificações
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── adapter/
│   │   │   ├── events/           # Kafka consumer
│   │   │   └── sender/           # Log notification sender
│   │   ├── domain/               # Notification, OrderEvent
│   │   ├── port/                 # Interfaces (consumer, sender)
│   │   └── usecase/              # SendNotification
│   ├── Dockerfile
│   └── go.mod
│
├── docker-compose.yaml           # Orquestração: Zookeeper, Kafka, Kafdrop, order-service, notification-service
├── .env.example                  # Exemplo de variáveis de ambiente
├── .gitignore
└── README.md
```

---

## 🔧 Pré-requisitos

- [Go 1.23+](https://golang.org/dl/)
- [Docker](https://www.docker.com/products/docker-desktop)
- [Docker Compose](https://docs.docker.com/compose/)

---

## 🐳 Executando com Docker Compose

1. **Clone o repositório:**
   ```bash
   git clone https://github.com/vhgomes/go-kafka-micro.git
   cd go-kafka-micro
   ```

2. **(Opcional)** Configure variáveis de ambiente:
   ```bash
   cp .env.example .env
   # Edite .env se necessário (ex: KAFKA_BROKER)
   ```

3. **Suba todos os serviços:**
   ```bash
   docker-compose up -d
   ```
   Isso iniciará:
   - Zookeeper (porta `2181`)
   - Kafka (porta `9092`)
   - Kafdrop (UI para visualizar tópicos, porta `9000`)
   - Order Service (porta `8080`)
   - Notification Service (porta `8081`)

4. **Acesse o Kafdrop** para ver os tópicos: http://localhost:9000

5. **Crie um pedido** (exemplo com `curl`):
   ```bash
   curl -X POST http://localhost:8080/order \
        -H "Content-Type: application/json" \
        -d '{
             "itens": [
               {"name": "Notebook", "quantity": 1, "price": 2500},
               {"name": "Mouse", "quantity": 2, "price": 50}
             ]
           }'
   ```

6. **Verifique os logs do Notification Service** para ver a notificação recebida:
   ```bash
   docker-compose logs notification-service
   ```

7. **Para derrubar os serviços:**
   ```bash
   docker-compose down -v
   ```

---

## 🛠️ Executando Localmente (sem Docker)

### 1. Suba Kafka e Zookeeper via Docker (ou use um broker existente)

```bash
docker run -d --name zookeeper -p 2181:2181 zookeeper:3.8.1
docker run -d --name kafka -p 9092:9092 \
  -e KAFKA_BROKER_ID=1 \
  -e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  -e KAFKA_LISTENERS=PLAINTEXT://0.0.0.0:9092 \
  bitnami/kafka:3.7
```

### 2. Order Service

```bash
cd order-service
go mod tidy
go run cmd/main.go
```

### 3. Notification Service

```bash
cd notification-service
go mod tidy
go run cmd/main.go
```

### 4. Teste com `curl` (mesmo exemplo acima)

---

## 📦 Endpoints da API

### `POST /order`

Cria um novo pedido e publica um evento no Kafka.

**Request Body:**
```json
{
  "itens": [
    {"name": "string", "quantity": int, "price": int64}
  ]
}
```

**Response (201 Created):**
```json
{
  "id": "uuid",
  "totalAmount": int64,
  "createdAt": "timestamp"
}
```

**Exemplo de erro (400 Bad Request):**
```json
{
  "error": "you must provide at least one item"
}
```

---

## 🧩 Decisões Técnicas

| Decisão                                 | Justificativa                                                                                   |
|-----------------------------------------|-------------------------------------------------------------------------------------------------|
| **Arquitetura Hexagonal**               | Isola a lógica de negócio de detalhes externos (HTTP, Kafka, DB), facilitando testes e evolução.|
| **Injeção de Dependência**              | Permite substituir implementações (ex: repositório em memória → PostgreSQL) sem alterar o usecase. |
| **Kafka como barramento**               | Comunicação assíncrona, desacoplamento, resiliência e escalabilidade horizontal.               |
| **`segmentio/kafka-go`**                | Cliente nativo em Go, performático e com bom suporte a contexts e concorrência.                |
| **In‑memory repository**                | Simplifica o setup inicial e foca no fluxo de eventos; trocável por persistência real.        |
| **Log notification sender**             | Permite observar o fluxo sem depender de serviços externos (e-mail/SMS).                      |

---

## 🧪 Testes (Futuro)

Ainda não há testes automatizados neste projeto. O roadmap inclui:

- Testes unitários para `CreateOrder` (usando mocks dos ports).
- Testes de integração com `testcontainers` para Kafka e PostgreSQL.
- Cobertura com `go test -cover`.

---

## 📝 Observações

- O **Kafka** precisa estar disponível para os serviços se conectarem. Em ambiente Docker Compose, os serviços usam `kafka:9092` (resolvido pelo nome do container).
- O **Kafdrop** é opcional, mas útil para monitorar tópicos e mensagens.
- O repositório contém um arquivo `refactor.md` com uma code review detalhada e recomendações de melhorias – use como guia para evoluir o projeto.

---

## 🤝 Como Contribuir

1. Faça um fork do projeto.
2. Crie uma branch para sua feature: `git checkout -b minha-feature`.
3. Commit suas mudanças: `git commit -m 'Adiciona feature'`.
4. Push para a branch: `git push origin minha-feature`.
5. Abra um Pull Request.

---

## 📄 Licença

Este projeto está sob a licença MIT – sinta-se à vontade para usá-lo e adaptá-lo.

---

## ✍️ Autor

Desenvolvido por [Victor Hugo Gomes](https://github.com/vhgomes)
