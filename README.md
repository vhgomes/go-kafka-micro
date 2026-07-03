# go-kafka-micro

Microserviços em Go integrados via Apache Kafka, seguindo arquitetura hexagonal (ports & adapters). Dois serviços independentes, comunicação assíncrona por eventos.

## Arquitetura

```text
┌─────────────────┐        orders.created         ┌───────────────────────┐
│   order-service │ ────────────────────────────▶ │  notification-service │
│  HTTP :8080     │          (Kafka)              │  Kafka Consumer       │
└─────────────────┘                               └───────────────────────┘
        │                                                     │
        ▼                                                     ▼
  InMemoryOrderRepository                            LogNotificationSender
```

- **order-service**: expõe HTTP para criação de pedidos, persiste (in-memory) e publica evento `orders.created` no Kafka.
- **notification-service**: consome `orders.created` e dispara notificação (atualmente logada em stdout).

Cada serviço segue a mesma estrutura interna:

```
internal/
  domain/    # entidades de negócio
  port/      # interfaces (contratos)
  usecase/   # regras de aplicação
  adapter/   # implementações concretas (HTTP, Kafka, memória)
```

## Stack

| Componente | Tecnologia |
|---|---|
| Linguagem | Go 1.23+ |
| Mensageria | Apache Kafka (`segmentio/kafka-go`) |
| IDs | `google/uuid` |
| Orquestração local | Docker Compose |
| UI de inspeção Kafka | Kafdrop |

## Pré-requisitos

- [Go 1.23+](https://golang.org/dl/)
- [Docker](https://www.docker.com/products/docker-desktop) + [Docker Compose](https://docs.docker.com/compose/)

## Como rodar (desenvolvimento local)

Suba a infraestrutura (Zookeeper + Kafka + Kafdrop):

```bash
docker-compose up zookeeper kafka kafdrop
```

Em terminais separados, rode cada serviço direto com Go (fora do container, apontando para `localhost:9092`):

```bash
# order-service
cd order-service
go run ./cmd

# notification-service
cd notification-service
go run ./cmd
```

- Kafdrop disponível em `http://localhost:9000` para inspecionar os tópicos.
- order-service disponível em `http://localhost:8080`.

> **Nota:** rodar `order-service` e `notification-service` como containers via `docker-compose up` (todos os serviços) atualmente não conecta ao Kafka corretamente, pois o broker é anunciado em `localhost` e os serviços apontam para `localhost:9092` mesmo dentro da rede do compose. Até esse ajuste ser feito, use o fluxo acima (infra em container, serviços Go rodando localmente).

## API — order-service

### `POST /orders`

Cria um pedido e publica o evento `orders.created`.

**Request body:**

```json
[
  {
    "Name": "Teclado mecânico",
    "Quantity": 1,
    "Price": 350.0
  }
]
```

**Response `201 Created`:**

```json
{
  "ID": "b3f1c2a0-...",
  "TotalAmount": 350.0,
  "CreatedAt": "2026-07-03T12:00:00Z"
}
```

**Erros:**
- `400` — lista de itens vazia, quantidade inválida ou JSON malformado.
- `500` — falha ao persistir ou publicar o evento.

## Evento Kafka — `orders.created`

Publicado pelo `order-service`, consumido pelo `notification-service`.

```json
{
  "UserID": "uuid",
  "OrderID": "string",
  "Total": 0,
  "Items": [
    { "ProductID": "string", "Quantity": 0, "Price": 0 }
  ]
}
```

## Estrutura do repositório

```text
order-service/
  cmd/main.go
  internal/
    adapter/handlers/     # HTTP handler + router
    adapter/repo/         # publisher Kafka + repositório in-memory
    domain/                # entidade Order
    port/                  # interfaces OrderRepository, OrderPublisher
    usecase/                # CreateOrder
notification-service/
  cmd/main.go
  internal/
    adapter/events/         # consumer Kafka
    adapter/sender/         # notificação (log)
    domain/                  # Notification, OrderEvent
    port/                    # interfaces EventConsumer, NotificationSender
    usecase/                  # SendNotification
docker-compose.yaml           # Zookeeper, Kafka, Kafdrop, serviços Go
```
