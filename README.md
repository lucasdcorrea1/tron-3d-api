<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/MongoDB-7.0-47A248?style=for-the-badge&logo=mongodb&logoColor=white" />
  <img src="https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white" />
  <img src="https://img.shields.io/badge/JWT-Auth-000000?style=for-the-badge&logo=jsonwebtokens&logoColor=white" />
  <img src="https://img.shields.io/badge/Prometheus-Metrics-E6522C?style=for-the-badge&logo=prometheus&logoColor=white" />
</p>

<h1 align="center">tron-3d-api</h1>

<p align="center">
  <strong>API RESTful para loja de impressos 3D</strong><br/>
  Go puro (stdlib router) + MongoDB + JWT + Prometheus
</p>

<p align="center">
  <a href="#arquitetura">Arquitetura</a> &bull;
  <a href="#quick-start">Quick Start</a> &bull;
  <a href="#endpoints">Endpoints</a> &bull;
  <a href="#modelos">Modelos</a> &bull;
  <a href="#autenticacao">Auth</a> &bull;
  <a href="#variaveis-de-ambiente">Config</a> &bull;
  <a href="#deploy">Deploy</a>
</p>

---

## Arquitetura

```
tron-3d-api/
├── cmd/api/
│   └── main.go                 # Entrypoint
├── internal/
│   ├── config/config.go        # Env + .env loading
│   ├── database/mongo.go       # MongoDB connection + indexes
│   ├── handlers/
│   │   ├── products.go         # GET /products, /products/:slug
│   │   ├── categories.go       # GET /categories
│   │   ├── orders.go           # POST/GET /orders (auth)
│   │   ├── admin_products.go   # CRUD admin products
│   │   ├── admin_categories.go # CRUD admin categories
│   │   ├── admin_orders.go     # Admin order management
│   │   └── helpers.go          # Pagination, slug, JSON utils
│   ├── middleware/
│   │   ├── auth.go             # JWT validation
│   │   ├── admin.go            # Admin role check
│   │   ├── cors.go             # CORS headers
│   │   ├── json.go             # Content-Type
│   │   ├── logger.go           # Structured JSON logging (slog)
│   │   └── metrics.go          # Prometheus counters/gauges
│   ├── models/
│   │   ├── product.go          # Product schema
│   │   ├── category.go         # Category schema
│   │   └── order.go            # Order/OrderItem/ShippingAddress
│   └── router/router.go        # Route definitions
├── docker-compose.yml          # API + MongoDB + Mongo Express
├── Dockerfile                  # Multi-stage alpine build
├── Makefile                    # Dev commands
├── .env.example                # Env template
└── go.mod
```

**Zero frameworks.** Router, middleware, handlers - tudo com `net/http` da stdlib do Go.

---

## Quick Start

### Com Docker (recomendado)

```bash
# 1. Clone
git clone https://github.com/lucasdcorrea1/tron-3d-api.git
cd tron-3d-api

# 2. Configure
cp .env.example .env
# Edite .env com seus valores (JWT_SECRET, ADMIN_USER_IDS)

# 3. Suba tudo
make docker-up
```

Servicos disponiveis:

| Servico | Porta | URL |
|---------|-------|-----|
| API | 8090 | `http://localhost:8090/api/v1/health` |
| MongoDB | 27019 | `mongodb://localhost:27019` |
| Mongo Express | 8083 | `http://localhost:8083` |

### Sem Docker

```bash
# Requisitos: Go 1.25+, MongoDB rodando

cp .env.example .env
# Edite .env

make run
# ou: go run ./cmd/api
```

### Build

```bash
make build
# Gera: bin/api
```

---

## Endpoints

### Publicos

| Metodo | Rota | Descricao |
|--------|------|-----------|
| `GET` | `/api/v1/health` | Health check |
| `GET` | `/api/v1/products` | Listar produtos ativos |
| `GET` | `/api/v1/products/{slug}` | Detalhe do produto por slug |
| `GET` | `/api/v1/categories` | Listar categorias ativas |
| `GET` | `/metrics` | Metricas Prometheus |

#### `GET /api/v1/products` — Query Params

| Param | Tipo | Default | Descricao |
|-------|------|---------|-----------|
| `page` | int | 1 | Pagina |
| `limit` | int | 12 | Itens por pagina (max 50) |
| `category_id` | string | — | Filtrar por categoria |
| `featured` | string | — | `"true"` para destaques |
| `search` | string | — | Busca textual (nome + descricao) |

```json
// Response
{
  "products": [...],
  "total": 42,
  "page": 1,
  "limit": 12
}
```

---

### Autenticados (JWT)

| Metodo | Rota | Descricao |
|--------|------|-----------|
| `POST` | `/api/v1/orders` | Criar pedido |
| `GET` | `/api/v1/orders` | Meus pedidos (paginado) |
| `GET` | `/api/v1/orders/{id}` | Detalhe de um pedido (ownership check) |

#### `POST /api/v1/orders` — Criar Pedido

```json
{
  "items": [
    { "product_id": "6845a...", "quantity": 2 },
    { "product_id": "6845b...", "quantity": 1 }
  ],
  "shipping_address": {
    "name": "Lucas Correa",
    "street": "Rua Exemplo",
    "number": "123",
    "complement": "Apto 4",
    "district": "Centro",
    "city": "Sao Paulo",
    "state": "SP",
    "zip_code": "01001000",
    "phone": "11999999999"
  },
  "notes": "Entregar pela manha"
}
```

**Logica de negocio:**
- Valida estoque disponivel para cada item
- Decrementa estoque atomicamente
- Calcula total a partir dos precos do banco
- Status inicial: `pending` / Payment: `pending`

---

### Admin (JWT + ADMIN_USER_IDS)

#### Produtos

| Metodo | Rota | Descricao |
|--------|------|-----------|
| `GET` | `/api/v1/admin/products` | Listar todos (inclui inativos) |
| `POST` | `/api/v1/admin/products` | Criar produto |
| `PUT` | `/api/v1/admin/products/{id}` | Atualizar produto |
| `DELETE` | `/api/v1/admin/products/{id}` | Deletar produto |

#### Categorias

| Metodo | Rota | Descricao |
|--------|------|-----------|
| `POST` | `/api/v1/admin/categories` | Criar categoria |
| `PUT` | `/api/v1/admin/categories/{id}` | Atualizar categoria |
| `DELETE` | `/api/v1/admin/categories/{id}` | Deletar categoria (409 se tem produtos) |

#### Pedidos

| Metodo | Rota | Descricao |
|--------|------|-----------|
| `GET` | `/api/v1/admin/orders` | Listar todos os pedidos |
| `PUT` | `/api/v1/admin/orders/{id}/status` | Atualizar status/pagamento |

```json
// PUT /api/v1/admin/orders/{id}/status
{
  "status": "printing",
  "payment_status": "paid"
}
```

---

## Modelos

### Product

```
id              ObjectID     auto-generated
name            string       required
slug            string       auto-generated from name
description     string
price           float64      required, > 0
images          []string     URLs das imagens
category_id     ObjectID     required
material        string       ex: "PLA", "PETG", "Resina"
dimensions      string       ex: "10x5x3 cm"
weight          float64      em gramas
print_time_hours float64     tempo estimado de impressao
stock           int          controle de estoque
active          bool         visibilidade na loja
featured        bool         destaque
created_at      datetime
updated_at      datetime
```

### Category

```
id              ObjectID     auto-generated
name            string       required
slug            string       auto-generated
description     string
image           string       URL da imagem
parent_id       ObjectID     categorias hierarquicas (opcional)
sort_order      int          ordenacao
active          bool
created_at      datetime
updated_at      datetime
```

### Order

```
id               ObjectID
user_id          ObjectID     do JWT
items            []OrderItem  snapshot dos produtos
total            float64      calculado no backend
status           string       pending → confirmed → printing → shipped → delivered
payment_status   string       pending | paid | refunded | failed
shipping_address ShippingAddress
notes            string
created_at       datetime
updated_at       datetime
```

### Order Status Flow

```
pending ──→ confirmed ──→ printing ──→ shipped ──→ delivered
   │
   └──→ cancelled (restaura estoque)
```

### Payment Status

```
pending ──→ paid
   │
   ├──→ failed
   └──→ refunded
```

---

## Autenticacao

A API compartilha o `JWT_SECRET` com o backend principal (tron-legacy-api). Tokens emitidos la funcionam aqui.

```
Authorization: Bearer <jwt-token>
```

**Claims esperadas:**
```json
{
  "user_id": "507f1f77bcf86cd799439011",
  "email": "user@example.com",
  "org_id": "...",
  "exp": 1234567890
}
```

**Admin:** definido por `ADMIN_USER_IDS` no `.env` (lista de ObjectIDs hex separados por virgula).

---

## Middleware Stack

Requests passam pela seguinte cadeia (de fora pra dentro):

```
Request
  → Logger        Structured JSON logging (slog), skip /health e /metrics
  → Metrics       Prometheus counters (http_requests_total, duration, etc.)
  → CORS          Access-Control headers
  → JSON          Content-Type: application/json
  → [Auth]        JWT validation (rotas protegidas)
  → [Admin]       ADMIN_USER_IDS check (rotas admin)
  → Handler
```

### Metricas Prometheus

| Metrica | Tipo | Descricao |
|---------|------|-----------|
| `http_requests_total` | counter | Total de requests (by method, path, status) |
| `http_active_requests` | gauge | Requests em processamento |
| `app_uptime_seconds` | counter | Tempo de atividade |
| `store3d_orders_created_total` | counter | Pedidos criados |
| `store3d_products_created_total` | counter | Produtos criados |
| `auth_errors_total` | counter | Erros de autenticacao |

---

## Variaveis de Ambiente

| Variavel | Default | Descricao |
|----------|---------|-----------|
| `MONGO_URI` | `mongodb://localhost:27017` | Connection string MongoDB |
| `DB_NAME` | `tron_3d` | Nome do banco |
| `PORT` | `8090` | Porta do servidor |
| `JWT_SECRET` | `change-me-in-production` | Segredo para validar JWTs |
| `FRONTEND_URL` | `https://whodo.com.br` | URL do frontend (referencia) |
| `ADMIN_USER_IDS` | — | IDs dos admins (hex, separados por virgula) |

---

## MongoDB Indexes

Criados automaticamente na inicializacao:

**products:**
- `slug` (unique)
- `category_id`
- `active + featured` (compound)
- `name + description` (text — full-text search)

**categories:**
- `slug` (unique)
- `active + sort_order` (compound)

**orders:**
- `user_id + created_at` (compound)
- `status + created_at` (compound)

---

## Deploy

### Docker Compose (producao)

```bash
# Build e start
docker-compose up -d --build

# Logs
docker-compose logs -f api-3d

# Parar
docker-compose down
```

O Dockerfile usa **multi-stage build** com Alpine para imagem final minima (~15MB):

```dockerfile
# Build: golang:1.25-alpine
# Runtime: alpine:3.21 + ca-certificates
# Binary: statically linked (CGO_ENABLED=0)
```

### Makefile

```bash
make run          # go run ./cmd/api
make build        # go build -o bin/api ./cmd/api
make docker-up    # docker-compose up -d
make docker-down  # docker-compose down
make tidy         # go mod tidy
```

---

## Erros

Todas as respostas de erro seguem o formato:

```json
{
  "message": "descricao do erro"
}
```

| Status | Quando |
|--------|--------|
| `400` | Validacao (campo obrigatorio, preco invalido, estoque insuficiente) |
| `401` | Token ausente ou invalido |
| `403` | Nao e admin |
| `404` | Recurso nao encontrado |
| `409` | Conflito (ex: deletar categoria com produtos associados) |
| `500` | Erro interno |

---

## Features Principais

- **Gestao de Estoque** — Decremento automatico ao criar pedido, restauracao ao cancelar
- **Slug Automatico** — Gerado do nome com normalizacao Unicode, auto-incremento em duplicatas
- **Busca Textual** — Full-text search no MongoDB (nome + descricao dos produtos)
- **Categorias Hierarquicas** — Suporte a `parent_id` para subcategorias
- **Paginacao Consistente** — Todas as listagens com `page`, `limit`, `total`
- **Metricas Prometheus** — Counters, gauges, duracao de requests, metricas de negocio
- **Logging Estruturado** — JSON logs com slog (method, path, status, duration, IP)
- **Multi-stage Docker** — Imagem final Alpine minimalista
- **Zero Frameworks** — 100% stdlib `net/http` router

---

## Autor

<table>
  <tr>
    <td align="center">
      <a href="https://github.com/lucasdcorrea1">
        <img src="https://github.com/lucasdcorrea1.png" width="100px;" alt="Lucas Correa" style="border-radius:50%"/><br />
        <sub><b>Lucas Correa</b></sub>
      </a><br />
      <sub>Criador & Mantenedor</sub>
    </td>
  </tr>
</table>

---

<p align="center">
  <sub>Built with Go + MongoDB by <a href="https://github.com/lucasdcorrea1">@lucasdcorrea1</a></sub>
</p>
