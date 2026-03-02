# 🧠 MoodTracker API

API REST desenvolvida em **Go** para gerenciamento de registros de humor (Day Logs), Tags e geração de relatórios analíticos.

A aplicação segue arquitetura em camadas com separação clara de responsabilidades:

Router → Middleware → Handler → Service → Repository → Database

---

# 🚀 Tecnologias

- Go
- PostgreSQL
- Chi Router (go-chi)
- JWT Authentication
- bcrypt
- expvar (métricas)
- Arquitetura em camadas (Handlers → Services → Repositories)
- Soft Delete
- Logging estruturado em JSON

---

# 📁 Estrutura do Projeto

```
internal/
├── handlers
├── services
├── repositories
├── middleware
├── models
├── routers
└── utils
```

---

# 🔐 Autenticação

A API utiliza autenticação baseada em **JWT**.

Fluxo:

1. Criar usuário
2. Ativar usuário
3. Realizar login
4. Receber `authentication_token`
5. Enviar no header:

```
Authorization: Bearer {token}
```

---

# 🌐 Base URL

```
http://localhost:4000/v1
```

---

# 👤 Usuários

## Criar usuário

POST `/v1/users/`

```json
{
  "name": "Luiz",
  "email": "luiz@email.com",
  "phone": "61999999999",
  "password": "12345678"
}
```

---

## Ativar usuário

POST `/v1/users/activate`

```json
{
  "cod": 1234,
  "email": "luiz@email.com"
}
```

---

# 🔑 Autenticação

## Login

POST `/v1/auth/login`

```json
{
  "email": "luiz@email.com",
  "password": "12345678"
}
```

### Response

```json
{
  "authentication_token": "jwt_token_here"
}
```

---

# 📅 Day Logs

Requer usuário autenticado e ativado.

Base route: `/v1/day_logs`

## Criar

POST `/v1/day_logs/`

```json
{
  "date": "2026-02-01T00:00:00Z",
  "description": "Dia produtivo",
  "mood_label": "BOM",
  "tags": ["trabalho", "estudo"]
}
```

## Buscar por ID

GET `/v1/day_logs/{id}`

## Buscar por Ano

GET `/v1/day_logs/year?year=2026`

## Atualizar

PUT `/v1/day_logs/`

## Deletar (Soft Delete)

DELETE `/v1/day_logs/{id}`

---

# 🏷 Tags

Requer usuário autenticado e ativado.

Base route: `/v1/tags`

## Criar

POST `/v1/tags/`

```json
{
  "name": "trabalho"
}
```

## Buscar por ID

GET `/v1/tags/{id}`

## Listar por Usuário (com paginação)

GET `/v1/tags/user/{id}?page=1&page_size=20&sort=name`

Sort permitidos:

- id
- name
- -id
- -name

## Atualizar

PUT `/v1/tags/`

## Deletar

DELETE `/v1/tags/{id}`

---

# 📊 Relatórios

Requer usuário autenticado e ativado.

Base route: `/v1/reports`

## 📅 Relatório Mensal

GET `/v1/reports/monthly?year=2026&month=2`

Retorna:

- Distribuição percentual de humor no mês
- Tags mais utilizadas

---

## 🏷 Relatório por Tag

GET `/v1/reports/tag?tag=trabalho`

Retorna:

- Distribuição de humor associada a uma tag específica
- Percentual calculado via Window Functions (PostgreSQL)

---

## 😀 Relatório por Humor

GET `/v1/reports/mood?mood_label=1`

Valores possíveis:

| Label | Valor |
|-------|-------|
| RUIM  | 1     |
| MEDIO | 2     |
| BOM   | 3     |

Retorna:

- Distribuição de tags associadas ao humor selecionado
- Percentual calculado dinamicamente no banco

---

# 📈 Monitoramento

## Métricas

GET `/v1/debug/vars`

Utiliza `expvar` para exposição de métricas internas.

---

# ⚙️ Como Rodar o Projeto

## 1️⃣ Clonar repositório

```
git clone https://github.com/seu-usuario/moodtracker.git
cd moodtracker
```

## 2️⃣ Configurar variáveis de ambiente

Criar `.env`:

```
SERVER_PORT=4000
SERVER_TIMEOUT_READ=3s
SERVER_TIMEOUT_WRITE=5s
SERVER_TIMEOUT_IDLE=5s
SERVER_DEBUG=true

POSTGRES_USER=seu_user
POSTGRES_PASSWORD=sua_senha
POSTGRES_DB=api_db
POSTGRES_PORT=5432

DB_DSN=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_MAX_IDLE_TIME=15m

LIMITER_RPS=2
LIMITER_BURST=4
LIMITER_ENABLED=true

SECRET_KEY=sua_secret
```

## 3️⃣ Rodar aplicação

```
go run ./cmd/api
```

Servidor disponível em:

```
http://localhost:4000
```

---

# 🧪 Melhorias Futuras

- Testes unitários e de integração
- CI/CD
- Documentação Swagger/OpenAPI
- Refresh Token

---

# 👨‍💻 Autor

Desenvolvido por Luiz Henrique.
