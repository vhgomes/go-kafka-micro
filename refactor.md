# Code Review — go-kafka-micro (order-service + notification-service)

**Serviços analisados:** `order-service`, `notification-service`
**Padrão:** arquitetura hexagonal (domain/port/usecase/adapter)

---

## Checklist Geral

- [ ] 6 problemas **CRÍTICOS** corrigidos
- [ ] 7 problemas **IMPORTANTES** corrigidos
- [ ] 6 **melhorias** aplicadas

---

## ⚠️ Problemas Encontrados

### order-service

#### [CRÍTICO] — Lógica invertida no method check do handler HTTP
**Localização:** `order-service/internal/adapter/handlers/order_handler.go:19-22`
**Problema:** A condição rejeita exatamente o método que deveria aceitar:
`if r.Method == http.MethodPost { http.Error(...); return }`. Isso bloqueia todo POST e deixa passar GET, PUT, DELETE etc. para o fluxo de criação de pedido.
**Impacto:** A API está funcionalmente quebrada — nenhum client real consegue criar pedido via POST. Qualquer outro verbo (que deveria ser rejeitado) executa a lógica de criação.
**Correção:** Inverta a condição para `r.Method != http.MethodPost`.

- [ ] Corrigir

#### [CRÍTICO] — Handler não retorna após `http.Error`
**Localização:** `order_handler.go:26-27`, `:30-31`, `:36-37`
**Problema:** Em três pontos (decode do body, `SaveOrder`, `Encode` da resposta), `http.Error` é chamado mas a função continua executando em vez de dar `return`.
**Impacto:** Após erro de decode, o código segue chamando `SaveOrder` com `items` vazio/inválido; após erro em `SaveOrder`, tenta `Encode(output)` com `output == nil`; e o handler pode tentar escrever múltiplos headers/bodies na mesma resposta (`http: superfluous response.WriteHeader call`).
**Correção:** Adicione `return` imediatamente após cada chamada de `http.Error`.

- [ ] Corrigir

#### [CRÍTICO] — `float32` para valores monetários
**Localização:** `order-service/internal/domain/order.go:11,17` (`Price`, `TotalAmount`)
**Problema:** Ponto flutuante não representa valores decimais de forma exata.
**Impacto:** Erros de arredondamento acumulados no cálculo de `TotalAmount` (`create_order.go:37`, soma de `Quantity * Price`) geram valores monetários incorretos — problema clássico e caro de corrigir depois que já há dados em produção.
**Correção:** Represente dinheiro como inteiro (centavos, `int64`) ou use um tipo decimal (ex. `shopspring/decimal`). Nunca `float32`/`float64` para moeda.

- [ ] Corrigir

#### [CRÍTICO] — Kafka Writer/Reader nunca fechados, sem graceful shutdown
**Localização:** `order-service/cmd/main.go` (writer nunca fechado) e `notification-service/internal/adapter/events/kakfa_event_consumer.go` (reader nunca fechado)
**Problema:** Nenhum dos dois serviços trata `SIGTERM`/`SIGINT`, nem chama `writer.Close()`/`reader.Close()`. O `http.ListenAndServe` do order-service também não tem shutdown gracioso.
**Impacto:** Em deploy, scale-down ou restart, conexões com o Kafka ficam penduradas, requisições em voo são abortadas sem resposta ao client, e mensagens em processamento no notification-service podem ser perdidas ou reprocessadas de forma inconsistente.
**Correção:** Capture sinais de SO, propague um `context.Context` cancelável até writer/reader/HTTP server, e use `http.Server.Shutdown(ctx)` + `reader.Close()`/`writer.Close()` no encerramento.

- [ ] Corrigir

#### [IMPORTANTE] — `context.Context` nunca propagado nas operações de I/O
**Localização:** `order_publisher_kafka.go:40` (`context.Background()`), `create_order.go` (`SaveOrder` sem `ctx`), `kakfa_event_consumer.go:32` (`context.Background()`)
**Problema:** Toda a cadeia usecase → repo/publisher → Kafka roda sobre `context.Background()`. Nenhuma função aceita `ctx` como primeiro parâmetro.
**Impacto:** Impossível aplicar timeout numa escrita/leitura Kafka lenta ou cancelar operações em andamento no shutdown; `r.Context()` no handler HTTP também nunca é usado.
**Correção:** Adicione `ctx context.Context` como primeiro parâmetro em `SaveOrder`, `Publish` e `Consume`, propagando a partir de `r.Context()` no handler e de um `ctx` derivado de signal handling no consumer.

- [ ] Corrigir

#### [IMPORTANTE] — Erros perdem o chain original
**Localização:** `order_publisher_kafka.go:32,42`, `create_order.go:48,52`
**Problema:** Erros criados com `errors.New("mensagem: " + err.Error())` em vez de `fmt.Errorf("mensagem: %w", err)`.
**Impacto:** Quebra `errors.Is`/`errors.As` na cadeia de chamadas — quem consome esse erro não consegue detectar a causa raiz programaticamente (ex.: diferenciar falha de conexão Kafka de erro de serialização), só via string matching.
**Correção:** Troque toda concatenação de erro por `fmt.Errorf("...: %w", err)`.

- [ ] Corrigir

#### [IMPORTANTE] — Body da requisição HTTP sem limite de tamanho
**Localização:** `order_handler.go:24-27` (`json.NewDecoder(r.Body)`)
**Problema:** Nenhum limite é aplicado ao body antes do decode.
**Impacto:** Um client malicioso ou buggy pode enviar payload arbitrariamente grande, consumindo memória do processo até esgotar recursos — DoS trivial num endpoint público.
**Correção:** Envolva `r.Body` com `http.MaxBytesReader(w, r.Body, limite)` antes do `NewDecoder`.

- [ ] Corrigir

#### [MELHORIA] — Endereço do broker Kafka hardcoded
**Localização:** `order-service/cmd/main.go:14-17`
**Problema:** `"localhost:9092"` fixo no código.
**Impacto:** Impossibilita rodar em ambientes diferentes (compose, k8s, staging) sem recompilar — é a causa direta do bug de conectividade descrito na seção de docker-compose abaixo.
**Correção:** Leia de variável de ambiente (`KAFKA_BROKERS`) com fallback para `localhost:9092` em dev.

- [ ] Corrigir

#### [MELHORIA] — Inicialização posicional de struct
**Localização:** `create_order.go:24` — `&CreateOrder{repo, publisher}`
**Problema:** Uber Style Guide recomenda named fields em construtores.
**Impacto:** Baixo agora (2 campos), mas escala mal — reordenar os campos do struct pode quebrar o construtor silenciosamente se os tipos coincidirem, sem erro de compilação.
**Correção:** Use `&CreateOrder{repo: repo, publisher: publisher}`.

- [ ] Corrigir

---

### notification-service

#### [CRÍTICO] — `package cmd` em vez de `package main`
**Localização:** `notification-service/cmd/main.go:1`
**Problema:** O arquivo tem `func main()` mas declara `package cmd`, não `package main`.
**Impacto:** Um pacote `main` é exigido pelo toolchain do Go para gerar executável via `go build`/`go run`. Nesse estado, `go run ./cmd` ou `go build ./cmd` não produzem o binário esperado — o executável `notification-service` presente no repositório provavelmente veio de um build anterior a essa quebra, ou de um caminho de build diferente do documentado no README.
**Correção:** Troque `package cmd` por `package main`.

- [ ] Corrigir

#### [CRÍTICO] — `log.Fatal` dentro da goroutine do consumer
**Localização:** `kakfa_event_consumer.go:34,39`
**Problema:** Qualquer erro de leitura do Kafka (`ReadMessage`) ou de `json.Unmarshal` chama `log.Fatal`, que executa `os.Exit(1)` imediatamente.
**Impacto:** Um blip de rede, rebalance de partição ou uma mensagem malformada derruba o processo inteiro — sem retry, sem DLQ. Você já resolve esse tipo de cenário com padrão de DLQ em outros contextos; aqui não foi aplicado.
**Correção:** Substitua `log.Fatal` por log de erro + `continue` (erro de leitura transitório) ou envio a uma DLQ (mensagem malformada). Nunca `os.Exit` dentro de uma goroutine de longa duração.

- [ ] Corrigir

#### [IMPORTANTE] — Dependência `EventConsumer` injetada mas nunca usada
**Localização:** `notification-service/internal/usecase/send_notification.go:9,14`
**Problema:** `SendNotification` recebe `consumer port.EventConsumer` no construtor e guarda no struct, mas `Execute` nunca chama `sn.consumer.Consume(...)` — quem consome o Kafka é o `main.go` diretamente, fora do usecase.
**Impacto:** Dependência morta que engana quem lê o código — parece que o usecase orquestra o consumo, mas não orquestra. Contradiz a intenção de arquitetura hexagonal que o resto do projeto segue bem.
**Correção:** Remova `consumer` de `SendNotification` (ele só precisa do `NotificationSender`), ou inverta a responsabilidade movendo o loop do `main.go` para dentro do usecase.

- [ ] Corrigir

#### [MELHORIA] — Nome de arquivo com typo
**Localização:** `notification-service/internal/adapter/events/kakfa_event_consumer.go`
**Problema:** "kakfa" em vez de "kafka".
**Impacto:** Nenhum funcional, mas prejudica busca/grep e passa impressão de descuido em review externo.
**Correção:** Renomeie para `kafka_event_consumer.go`.

- [ ] Corrigir

---

### Infraestrutura (docker-compose.yaml)

#### [CRÍTICO] — Kafka anunciado em `localhost` quebra conectividade entre containers
**Localização:** `docker-compose.yaml:22` (`KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092`) combinado com `order-service/cmd/main.go:15` e `notification-service/cmd/main.go:15`, ambos hardcoded para `localhost:9092`.
**Problema:** Dentro da rede do Docker Compose, `localhost` no container `order-service` aponta para o próprio container, não para o container `kafka`. O broker está anunciado como acessível apenas em `localhost`.
**Impacto:** `order-service` e `notification-service`, quando rodados via `docker-compose up`, não conseguem se conectar ao Kafka — a stack não sobe funcional em container, só quando rodada localmente na máquina host.
**Correção:** Anuncie o listener usando o nome do serviço na rede compose (`KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092`) e troque o broker address hardcoded nos `main.go` para vir de env var.

- [ ] Corrigir

#### [IMPORTANTE] — Imagem `bitnami/kafka:latest`
**Localização:** `docker-compose.yaml:14`
**Problema:** Tag `latest` usada para a imagem do Kafka — contraria a diretriz do seu próprio stack (versão sempre pinada).
**Impacto:** Build não é reproduzível; um `docker-compose pull` futuro pode trazer versão do Kafka com breaking changes de configuração sem aviso prévio.
**Correção:** Pin em versão específica, ex. `bitnami/kafka:3.7`.

- [ ] Corrigir

#### [IMPORTANTE] — Binários compilados versionados no repositório
**Localização:** `order-service/order-service`, `notification-service/notification-service` (presentes na árvore do repo)
**Problema:** Executáveis compilados estão commitados junto com o código-fonte.
**Impacto:** Repositório infla a cada build, diffs de binário poluem o histórico do git, e há risco de rodar um binário desatualizado em relação ao código-fonte — como parece ser o caso do bug `package cmd` acima.
**Correção:** Adicione os binários ao `.gitignore` e faça build via Dockerfile multi-stage em vez de montar volume + rodar binário pré-compilado.

- [ ] Corrigir

#### [MELHORIA] — Sem healthcheck nos serviços Go
**Localização:** `docker-compose.yaml:37-55`
**Problema:** `order-service` e `notification-service` não têm `healthcheck` definido.
**Impacto:** `depends_on: - kafka` só espera o container do Kafka *iniciar*, não estar pronto para aceitar conexões — race condition clássica de startup.
**Correção:** Adicione `healthcheck` no serviço `kafka` e use `depends_on: kafka: condition: service_healthy` nos serviços Go.

- [ ] Corrigir

---

### Cross-cutting

#### [IMPORTANTE] — Nenhum teste automatizado no repositório
**Localização:** repositório inteiro (zero arquivos `*_test.go`)
**Problema:** Não há cobertura de teste para usecases, handlers ou adapters.
**Impacto:** As regras de negócio em `create_order.go` (validação de item, cálculo de total) e o bug do method check invertido no handler não seriam pegos por CI algum — dependem 100% de teste manual.
**Correção:** Priorize testes de unidade para `CreateOrder.SaveOrder` (casos de item vazio/quantidade inválida) e um teste de integração HTTP para `OrderHandler.CreateOrder` — isso já teria capturado o bug crítico #1.

- [ ] Corrigir

#### [MELHORIA] — Mistura de idiomas em mensagens de erro
**Localização:** `create_order.go:29,35` ("você precisa enviar itens") vs. `order_publisher_kafka.go:32,42` (mensagens técnicas em inglês)
**Problema:** Erros de domínio em PT-BR, erros de infra em EN, sem padrão definido.
**Impacto:** Inconsistência em logs agregados e em mensagens potencialmente expostas ao client.
**Correção:** Padronize o idioma das mensagens internas (recomendo EN, dado o restante do código e a stack).

- [ ] Corrigir

#### [MELHORIA] — `err.Error()` exposto diretamente na resposta HTTP
**Localização:** `order_handler.go:26,31,37`
**Problema:** `http.Error(w, err.Error(), ...)` devolve a mensagem de erro interna crua para o client da API.
**Impacto:** Vaza detalhes de implementação (mensagens de erro do repo/publisher) para fora da fronteira do serviço. Não é uma vulnerabilidade grave aqui, mas é um hábito arriscado para levar a produção.
**Correção:** Separe erros de domínio (retornáveis ao client) de erros de infraestrutura (logados internamente, resposta genérica ao client).

- [ ] Corrigir

---

## ✅ Pontos Positivos

- **Arquitetura hexagonal consistente (ports & adapters):** separação clara entre `domain`, `port`, `usecase` e `adapter` nos dois serviços, com interfaces (`OrderRepository`, `OrderPublisher`, `EventConsumer`, `NotificationSender`) desacoplando regra de negócio de infraestrutura. Isso facilita trocar `InMemoryOrderRepository` por Postgres+sqlc sem tocar no usecase.
- **Injeção de dependência via construtores:** nenhum global, tudo passado explicitamente — mesmo padrão que você já aplica bem em outros projetos do portfólio.

---

## Resumo por Severidade

| Severidade | Quantidade |
|---|---|
| 🔴 CRÍTICO | 6 |
| 🟡 IMPORTANTE | 7 |
| 🔵 MELHORIA | 6 |

**Prioridade de correção sugerida:** primeiro os bugs que quebram o happy path — method check invertido, `return` faltando no handler, e `package cmd` no notification-service — o sistema não funciona nem em cenário ideal enquanto esses três existirem. Em seguida, a conectividade Kafka no compose (nada funciona containerizado até corrigir isso). Depois, consolidação: context propagation, error wrapping, testes, antes de escalar o projeto.
