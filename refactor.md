# Code Review v2 — go-kafka-micro (delta desde a última revisão)

**Base de comparação:** `code-review-go-kafka-micro.md` (v1, também versionado no repo como `refactor.md`)
**Mudanças no repo desde então:** Dockerfiles multi-stage, Makefile, `.env.example`, `.gitignore`, docker-compose reescrito (KRaft, sem Zookeeper), graceful shutdown no order-service, `int64` para dinheiro, context propagation nas assinaturas.

## Resumo da evolução

| | Quantidade |
|---|---|
| ✅ Corrigidos de fato | 15 |
| ⚠️ Parcialmente corrigidos / com regressão | 5 |
| 🆕 Novos problemas | 4 |
| 🔴 Ainda não corrigidos | 1 |

Progresso real, mas **o bug mais importante da v1 — conectividade Kafka em container — continua de pé**, só que agora por um motivo diferente (e mais escondido).

---

## ✅ Corrigidos (confirmado no código novo)

- [x] Lógica invertida do method check (`order_handler.go`) — agora `r.Method != http.MethodPost`
- [x] `return` faltando após `http.Error` — presente nos três pontos
- [x] `float32` para dinheiro — `Price`/`TotalAmount` agora `int64` (centavos)
- [x] `package cmd` no notification-service — corrigido para `package main`
- [x] `log.Fatal` dentro da goroutine do consumer — trocado por `log.Println` + `continue`
- [x] Graceful shutdown do HTTP server no order-service — `SIGTERM`/`SIGINT` + `srv.Shutdown(ctx)`
- [x] Body HTTP sem limite de tamanho — `http.MaxBytesReader(w, r.Body, 1024)`
- [x] `bitnami/kafka:latest` — trocado por `confluentinc/cp-kafka:7.6.0` (pinado)
- [x] Binários compilados versionados no repo — agora há `Dockerfile` multi-stage em cada serviço e `.gitignore` cobrindo os binários
- [x] Inicialização posicional de struct (`CreateOrder`) — agora usa named fields
- [x] Dependência `EventConsumer` morta no `SendNotification` — removida do usecase
- [x] Typo `kakfa_event_consumer.go` — renomeado para `kafka_event_consumer.go`
- [x] `err.Error()` exposto cru na resposta HTTP — agora retorna mensagem genérica + loga o erro real internamente
- [x] Mistura de idiomas nas mensagens de erro do `create_order.go` — padronizado em inglês
- [x] Ausência de healthcheck no Kafka — `docker-compose.yaml` agora tem `healthcheck` no serviço `kafka` com `depends_on: condition: service_healthy`

---

## ⚠️ Parcialmente corrigidos / com regressão

#### [CRÍTICO] — Conectividade Kafka em container: ainda quebrada, causa mudou

**Localização:** `order-service/cmd/main.go:14-16` + `notification-service/cmd/main.go:14-17` + `docker-compose.yaml`
**O que mudou:** O `docker-compose.yaml` agora está correto — `KAFKA_ADVERTISED_LISTENERS: PLAINTEXT_INTERNAL://kafka:29092,PLAINTEXT_EXTERNAL://localhost:9092` é exatamente o padrão certo para expor Kafka tanto para containers (`kafka:29092`) quanto para o host (`localhost:9092`).
**Problema:** Nenhum dos dois serviços usa o listener interno. `order-service` faz `os.Setenv("KAFKA_BROKER", "localhost:9092")` **antes** de ler a variável — ou seja, sobrescreve incondicionalmente qualquer valor injetado pelo ambiente, e ainda lê uma chave (`KAFKA_BROKER`) diferente da que o `docker-compose.yaml` define (`KAFKA_BROKER_ADDRESS`). `notification-service` nem tenta: brokers continuam 100% hardcoded em `[]string{"localhost:9092"}`.
**Impacto:** Rodando via `docker-compose up`, os dois serviços ainda tentam falar com `localhost:9092` de dentro do próprio container, que não é o Kafka. O compose está pronto para funcionar, mas o código nunca vai usar o listener certo.
**Correção:** Em `order-service`, remova o `os.Setenv` e leia `KAFKA_BROKER_ADDRESS` diretamente (com fallback pra `localhost:9092` só se a env var não existir). Em `notification-service`, adicione a mesma leitura de env var — hoje não existe nenhuma.

- [x] Corrigir

#### [IMPORTANTE] — `notification-service` não ganhou o graceful shutdown

**Localização:** `notification-service/cmd/main.go`
**O que mudou:** `order-service` ganhou tratamento de `SIGTERM`/`SIGINT` com shutdown gracioso.
**Problema:** `notification-service` continua sem tratar sinais e sem fechar o `reader` do Kafka — o `for event := range eventsCh` roda indefinidamente sem via de saída limpa.
**Impacto:** Em restart/scale-down, o consumer é morto abruptamente; mensagens em processamento podem ser perdidas ou reprocessadas de forma inconsistente no próximo rebalance.
**Correção:** Replique o padrão do `order-service`: capture o sinal, cancele um `context.Context`, e chame `reader.Close()` no encerramento.

- [x] Corrigir

#### [IMPORTANTE] — `context.Context` propagado na assinatura, mas ignorado na prática

**Localização:** `order_publisher_kafka.go` (`Publish` recebe `ctx` mas chama `o.writer.WriteMessages(context.Background(), message)`), `kakfa_event_consumer.go:32` (`ReadMessage(context.Background())`, nunca mudou)
**O que mudou:** `SaveOrder`, `Publish` e `Save` agora aceitam `ctx context.Context` como parâmetro, e o handler passa `r.Context()`.
**Problema:** O `ctx` chega até o publisher mas não é usado na chamada real ao Kafka — o parâmetro existe só de fachada. O consumer nem recebe `ctx` na assinatura.
**Impacto:** Ainda não é possível cancelar ou aplicar timeout numa escrita/leitura Kafka lenta, mesmo com a propagação "pronta" na maior parte da cadeia.
**Correção:** Troque `context.Background()` por `ctx` dentro de `Publish`. Adicione `ctx` como parâmetro de `Consume`/`ReadMessage` também.

- [x] Corrigir

#### [IMPORTANTE] — Erros de infraestrutura perdem a causa raiz por completo

**Localização:** `create_order.go:665,669` — `fmt.Errorf("repository failure: error saving order")`, `fmt.Errorf("publisher failure: error publishing order created message")`
**O que mudou:** A v1 pelo menos concatenava `err.Error()` na mensagem (formato errado, mas a informação sobrevivia). Agora o `err` retornado por `repo.Save`/`publisher.Publish` é descartado inteiramente — nem aparece na string.
**Impacto:** Regressão de debugabilidade: um erro de conexão com Kafka e um erro de serialização geram exatamente a mesma mensagem (`"publisher failure: error publishing order created message"`), sem nenhuma pista de qual foi a causa real.
**Correção:** Use `fmt.Errorf("repository failure: %w", err)` / `fmt.Errorf("publisher failure: %w", err)` — preserva a causa e ainda permite `errors.Is`/`errors.As`.

- [x] Corrigir

#### [IMPORTANTE] — `fmt.Errorf` chamado com string concatenada (sem verbo de formatação)

**Localização:** `order_publisher_kafka.go:509,519` — `fmt.Errorf("KafkaPublisher | PublishOrderCreated error: " + err.Error())`
**O que mudou:** A v1 usava `errors.New(string + err.Error())`, que era funcionalmente "seguro" (só perdia o wrapping). Trocar para `fmt.Errorf` sem usar verbo `%w`/`%s` introduz um problema novo.
**Problema:** `fmt.Errorf` com uma única string concatenada (sem argumentos) interpreta qualquer `%` que aparecer dentro de `err.Error()` como início de verbo de formatação.
**Impacto:** Se a mensagem de erro do Kafka client alguma vez contiver um `%` (não é incomum em mensagens de rede/timeout), a mensagem final fica corrompida com `%!s(MISSING)` ou similar, escondendo o erro real — `go vet` já sinaliza esse padrão.
**Correção:** `fmt.Errorf("KafkaPublisher | PublishOrderCreated error: %w", err)`.

- [x] Corrigir

---

## 🆕 Novos problemas encontrados

#### [MELHORIA] — README desalinhado com a implementação real

**Localização:** `README.md` (seção "Endpoints da API" e exemplos de `curl`)
**Problema:** O README documenta `POST /order` (singular) com body `{"itens": [...]}` e resposta `201 Created`; o código real expõe `POST /orders` (plural, ver `router.go`), espera um array JSON puro (`[{...}]`, sem wrapper), e o handler nunca chama `w.WriteHeader(201)` — a resposta de sucesso sai como `200 OK`.
**Impacto:** Quem seguir o README literalmente (inclusive você, daqui a uns meses) vai receber `404` ou erro de decode ao copiar o exemplo de `curl`.
**Correção:** Atualize o README com o path, formato de payload e status code reais, ou ajuste o handler para casar com o que está documentado — o que fizer mais sentido pro contrato da API.

- [x] Corrigir

#### [MELHORIA] — Porta 8081 do notification-service documentada mas inexistente

**Localização:** `docker-compose.yaml:837` (`ports: "8081:8081"`), `README.md`, `Makefile` (`docker-up`)
**Problema:** `notification-service` é um consumer Kafka puro — não há nenhum `http.ListenAndServe` no código. A porta 8081 mapeada no compose e anunciada no README/Makefile não corresponde a nada rodando de fato.
**Impacto:** Confunde quem for validar o serviço via `curl http://localhost:8081` esperando alguma resposta.
**Correção:** Remova o mapeamento de porta do compose (ou adicione um endpoint de health/metrics real se fizer sentido ter um).

- [x] Corrigir

#### [MELHORIA] — Limite de `MaxBytesReader` pode ser pequeno demais

**Localização:** `order_handler.go:437` — `http.MaxBytesReader(w, r.Body, 1024)`
**Problema:** 1024 bytes comporta ~12-15 itens de pedido, dependendo do tamanho do `Name`. Um carrinho legítimo maior que isso é rejeitado como se fosse payload malicioso.
**Impacto:** Falso positivo de "payload grande demais" em uso normal, sem sinalização clara pro client do porquê.
**Correção:** Extraia o limite pra uma constante nomeada com uma margem mais realista (ex.: 64KB) e documente a decisão.

- [x] Corrigir

#### [MELHORIA] — Typo em `.env.example`

**Localização:** `.env.example:1` — `KAFKA_BROKER=="insira-aqui-o-seu-kafka-broker"`
**Problema:** Sinal de igual duplicado (`==`).
**Impacto:** Nenhum funcional (é só um exemplo), mas quem copiar o arquivo pra `.env` sem reparar herda a sintaxe quebrada.
**Correção:** `KAFKA_BROKER="insira-aqui-o-seu-kafka-broker"`.

- [x] Corrigir

---

## ✅ Pontos Positivos Novos

- **Dockerfiles multi-stage corretos:** builder em `golang:1.23-alpine`, runtime em `alpine:latest`, binário copiado — resolve builds reprodutíveis e imagem final enxuta, no lugar do hack de volume + binário pré-compilado.
- **docker-compose migrado pra KRaft:** eliminar o Zookeeper e usar `confluentinc/cp-kafka` com dual-listener (`PLAINTEXT_INTERNAL`/`PLAINTEXT_EXTERNAL`) é o padrão certo pra esse problema — a configuração em si está correta, só falta o código Go acompanhar.
- **Makefile completo:** `build`, `run`, `test`, `lint`, `docker-*`, `clean` — boa base de DX pro projeto, especialmente já prevendo `golangci-lint` e cobertura de teste.

---

## Checklist Atualizado (o que falta)

- [ ] Corrigir leitura de broker Kafka no `order-service` (parar de sobrescrever a env var, usar `KAFKA_BROKER_ADDRESS`)
- [ ] Adicionar leitura de broker via env var no `notification-service` (hoje inexistente)
- [ ] Graceful shutdown + `reader.Close()` no `notification-service`
- [ ] Usar o `ctx` recebido dentro de `Publish` (trocar `context.Background()`)
- [ ] Propagar `ctx` até `Consume`/`ReadMessage`
- [ ] Voltar a incluir a causa raiz (`%w`) nos erros de `create_order.go`
- [ ] Trocar concatenação por `%w` em `order_publisher_kafka.go` (evita o risco de formatting directive)
- [ ] Alinhar README com a API real (path, payload, status code)
- [ ] Remover ou justificar a porta 8081 do notification-service no compose
- [ ] Revisar limite do `MaxBytesReader`
- [ ] Corrigir `.env.example`
- [ ] Escrever os primeiros testes (`CreateOrder.SaveOrder` + handler HTTP)

## Resumo por Severidade (itens ainda abertos)

| Severidade | Quantidade |
|---|---|
| 🔴 CRÍTICO | 1 |
| 🟡 IMPORTANTE | 4 |
| 🔵 MELHORIA | 4 |

**Prioridade sugerida:** a conectividade Kafka continua sendo o bloqueador nº 1 — sem ela, `docker-compose up` não funciona fim a fim, apesar de todo o resto do compose já estar certo. É a correção mais rápida da lista (é literalmente remover um `os.Setenv` e trocar uma env var em dois lugares) e destrava validar tudo o resto.
