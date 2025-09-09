# Go Kafka Microservices

Este projeto é um exemplo de **microserviços em Go** integrados via **Apache Kafka**, contendo dois serviços principais:

* `order-service`: responsável pelo gerenciamento de pedidos.
* `notification-service`: responsável pelo envio de notificações baseadas em eventos do Kafka.

Além disso, o projeto utiliza:

* **Kafka** e **Zookeeper** para comunicação de eventos.
* **Kafdrop** para monitoramento do Kafka via interface web.


## 🔹 Pré-requisitos

* [Go 1.24+](https://golang.org/dl/)
* [Docker](https://www.docker.com/products/docker-desktop)
* [Docker Compose](https://docs.docker.com/compose/)

---
## 🔹 Estrutura de Comunicação

* **Order-Service** publica eventos de pedidos no Kafka.
* **Notification-Service** consome esses eventos e processa notificações.
---