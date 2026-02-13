# 🧠 MoodTracker API

API REST desenvolvida em **Go** para gerenciamento de registros de humor (day logs), tags e geração de relatórios analíticos.

O sistema permite:

- Cadastro e ativação de usuários
- Autenticação com token (JWT)
- Registro de humor diário
- Associação de tags aos registros
- Geração de relatórios por mês, tag e humor

---

## 🚀 Tecnologias

- Go
- PostgreSQL
- Chi Router (go-chi)
- JWT Authentication
- bcrypt
- expvar (metrics)
- Clean Architecture (Handlers → Services → Repositories)

---

## 📁 Arquitetura

O projeto segue separação clara de responsabilidades:

internal/
├── handlers → Camada HTTP
├── services → Regras de negócio
├── repositories → Acesso a dados
├── middleware → Autenticação, CORS, RateLimit
├── models → Entidades e DTOs
└── routers → Definição das rotas


Fluxo da requisição:
Router → Middleware → Handler → Service → Repository → Database


---

## 🔐 Autenticação

A API utiliza autenticação baseada em token JWT.

Fluxo:

1. Criar usuário
2. Ativar usuário
3. Fazer login
4. Receber `authentication_token`
5. Enviar token no header:


---

## 🌐 Base URL


---

# 👤 Usuários

## Criar usuário


### Body

```json
{
  "name": "Luiz",
  "email": "luiz@email.com",
  "phone": "61999999999",
  "password": "12345678"
}

Ativar usuário
POST /v1/users/activate
{
  "cod": 1234,
  "email": "luiz@email.com"
}

POST /v1/auth/login
{
  "cod": 1234,
  "email": "luiz@email.com"
}

POST /v1/auth/login
{
  "email": "luiz@email.com",
  "password": "12345678"
}

{
  "authentication_token": "jwt_token_here"
}

📅 Day Logs

Requer usuário autenticado e ativado.

Base route:
/v1/day_logs

Criar registro
POST /v1/day_logs

{
  "date": "2026-02-01T00:00:00Z",
  "description": "Dia produtivo",
  "mood_label": 3,
  "tags": ["trabalho", "estudo"]
}

Atualizar
PUT /v1/day_logs

Buscar por ID
GET /v1/day_logs/{id}

Buscar por ano
GET /v1/day_logs/year?year=2026

Deletar
DELETE /v1/day_logs/{id}

🏷 Tags

Base route:

/v1/tags

Criar
POST /v1/tags

{
  "name": "trabalho"
}

Listar do usuário (com paginação)
GET /v1/tags/user/{id}?page=1&page_size=20&sort=name

Buscar por ID
GET /v1/tags/{id}

Atualizar
PUT /v1/tags

Deletar
DELETE /v1/tags/{id}

📊 Relatórios

Base route:

/v1/reports


Requer usuário autenticado.

📅 Relatório Mensal
GET /v1/reports/monthly?year=2026&month=2


Retorna:

Distribuição de humor no mês

Tags mais usadas

🏷 Relatório por Tag
GET /v1/reports/tag?tag=trabalho


Retorna:

Distribuição de humor para uma tag específica

Percentual calculado via window function SQL

😀 Relatório por Humor
GET /v1/reports/mood?mood_label=1


Valores possíveis:

Label	Valor
RUIM	1
MEDIO	2
BOM	3

📈 Monitoramento
Métricas
/v1/debug/vars

