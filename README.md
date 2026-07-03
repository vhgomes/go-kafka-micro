# Go Kafka Microservices

Sistema de exemplo construído com arquitetura de microsserviços em Go, utilizando Apache Kafka para comunicação assíncrona e orientada a eventos.

## 🏗 Arquitetura

O projeto é composto por dois serviços independentes que seguem os princípios da Arquitetura Hexagonal (Ports & Adapters):

```text
┌─────────────────┐        orders.created         ┌───────────────────────┐
│   order-service │ ────────────────────────────▶ │  notification-service │
│  HTTP :8080     │          (Kafka)              │  Kafka Consumer       │
└─────────────────┘                               └───────────────────────┘
        │                                                     │
        ▼                                                     ▼
  InMemoryOrderRepository                            LogNotificationSender

```

* **order-service**: Expõe uma API HTTP para a criação de pedidos. Salva o pedido em memória e publica o evento `orders.created` no Kafka.
* **notification-service**: Atua como um worker em background (Consumer) ouvindo o tópico `orders.created` e processando o envio de notificações (atualmente via stdout/logs).

## 🚀 Tecnologias Utilizadas

| Componente | Tecnologia |
| --- | --- |
| **Linguagem** | Go 1.23+ |
| **Mensageria** | Apache Kafka (modo KRaft, sem Zookeeper) via `segmentio/kafka-go` |
| **Identificadores** | `google/uuid` |
| **Orquestração** | Docker & Docker Compose (Multi-stage builds) |
| **Monitoramento** | Kafdrop (UI para inspeção do Kafka) |
| **Automação** | Makefile |

## ⚙️ Pré-requisitos

* [Go 1.23+](https://go.dev/dl/)
* [Docker](https://www.docker.com/products/docker-desktop) e Docker Compose
* `make` instalado (opcional, mas recomendado)

## 🛠 Como Executar

O projeto possui um `Makefile` completo para facilitar a execução. Você tem duas opções de ambiente:

### Opção 1: 100% via Docker (Recomendado para testes)

Sobe toda a infraestrutura e os dois serviços Go em containers isolados.

```bash
# Constrói as imagens e sobe os containers em background
make docker-build
make docker-up

# Para acompanhar os logs
make docker-logs

# Para desligar e limpar tudo
make docker-down

```

### Opção 2: Híbrido (Ideal para desenvolvimento)

Sobe o Kafka via Docker, mas roda os serviços localmente no seu terminal para facilitar o *hot-reload* e *debug*.

```bash
# 1. Suba apenas a infraestrutura do Kafka e Kafdrop
docker-compose up -d kafka kafdrop

# 2. Em seguida, rode os serviços Go localmente
make run

# 3. Para parar a execução local dos serviços
make stop

```

> **Acessos úteis:**
> * **API Order Service:** `http://localhost:8080`
> * **Kafdrop UI:** `http://localhost:9000`
> 
> 

---

## 📖 Documentação da API (order-service)

### `POST /orders`

Cria um novo pedido e dispara o evento no Kafka.
*Nota: O valor de `Price` e `TotalAmount` é representado em centavos (int64).*

**Request Body (JSON):**

```json
[
  {
    "Name": "Teclado Mecânico",
    "Quantity": 1,
    "Price": 35000
  },
  {
    "Name": "Mouse Gamer",
    "Quantity": 2,
    "Price": 15000
  }
]

```

**Exemplo via cURL:**

```bash
curl -X POST http://localhost:8080/orders \
-H "Content-Type: application/json" \
-d '[{"Name": "Teclado Mecânico", "Quantity": 1, "Price": 35000}]'

```

**Response `201 Created`:**

```json
{
  "ID": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "TotalAmount": 65000,
  "CreatedAt": "2026-07-03T17:00:00.000000Z"
}

```

**Códigos de Erro:**

* `400 Bad Request` — Payload inválido, array vazio ou quantidade nula/negativa.
* `405 Method Not Allowed` — Método HTTP diferente de POST.
* `500 Internal Server Error` — Falha na persistência ou comunicação com o Kafka.

---

## 📦 Evento Kafka (`orders.created`)

Estrutura da mensagem enviada para o tópico `orders.created`:

```json
{
  "UserID": "uuid-do-usuario",
  "OrderID": "uuid-do-pedido",
  "Total": 65000,
  "Items": [
    {
      "ProductID": "string",
      "Quantity": 1,
      "Price": 35000
    }
  ]
}

```

## 📁 Estrutura de Diretórios

```text
.
├── notification-service/
│   ├── cmd/main.go               # Entrypoint do worker
│   ├── internal/                 # Domínio, Casos de uso e Adapters (Kafka Consumer)
│   ├── Dockerfile                # Multi-stage build
│   └── go.mod
├── order-service/
│   ├── cmd/main.go               # Entrypoint da API
│   ├── internal/                 # Domínio, Casos de uso, Adapters (HTTP Handler, Kafka Publisher)
│   ├── Dockerfile                # Multi-stage build
│   └── go.mod
├── docker-compose.yaml           # Configuração KRaft, Kafdrop e redes
├── Makefile                      # Automação de comandos (build, run, lint, docker)
└── README.md

```
