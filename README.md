# 🏪 ERP Backend - Sistema de Caixa de Mercado

API RESTful desenvolvida em **Go** para gerenciamento completo de um sistema de caixa de mercado.

## 🚀 Tecnologias

| Tecnologia | Uso |
|------------|-----|
| **Go 1.21+** | Linguagem principal |
| **Chi v5** | Router HTTP |
| **pgx v5** | Driver PostgreSQL |
| **JWT** | Autenticação |
| **Bcrypt** | Hash de senhas |
| **Gorilla WebSocket** | Comunicação em tempo real |

## 📁 Estrutura do Projeto

```
app-backend/
├── cmd/api/main.go                    # Ponto de entrada
├── internal/
│   ├── api/
│   │   ├── handlers/                  # Controladores (12 handlers)
│   │   ├── middleware/                # Auth JWT, CORS, IP local, Logging
│   │   └── router.go                 # Configuração de rotas
│   ├── domain/models/                 # Modelos de dados (11 arquivos)
│   ├── service/                       # Regras de negócio (12 services)
│   └── infrastructure/
│       ├── database/                  # Conexão PostgreSQL (pgxpool)
│       └── websocket/                 # Hub WebSocket
├── pkg/
│   ├── config/                        # Variáveis de ambiente
│   └── utils/                         # JWT, Hash, Respostas JSON
├── .env                               # Configurações (não versionar em prod)
├── go.mod / go.sum
└── README.md
```

## ⚙️ Configuração

### Pré-requisitos

- Go 1.21 ou superior
- PostgreSQL 14+
- Banco de dados criado com o script `erp.sql`

### Variáveis de Ambiente (.env)

```env
SERVER_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=postgres
DB_NAME=mercado_db
DB_SSLMODE=disable
JWT_SECRET=sua-chave-secreta-aqui
JWT_EXPIRY_HOURS=8
RESTRICT_LOCAL_NETWORK=false
```

### Executar

```bash
# Instalar dependências
go mod tidy

# Rodar o servidor
go run cmd/api/main.go
```

O servidor inicia em `http://localhost:8080`.

## 📋 Endpoints da API

### Públicos (sem autenticação)

| Método | Rota | Descrição |
|--------|------|-----------|
| `POST` | `/api/login` | Autenticação (retorna JWT) |
| `GET` | `/api/discover` | Descoberta do servidor na rede |
| `GET` | `/health` | Health check |

### Caixa

| Método | Rota | Descrição |
|--------|------|-----------|
| `GET` | `/api/caixa/status` | Status do caixa atual |
| `POST` | `/api/caixa/abrir` | Abrir sessão de caixa |
| `POST` | `/api/caixa/fechar` | Fechar sessão de caixa |
| `POST` | `/api/caixa/sangria` | Retirada de dinheiro |
| `POST` | `/api/caixa/suprimento` | Adição de dinheiro |

### Vendas

| Método | Rota | Descrição |
|--------|------|-----------|
| `POST` | `/api/vendas` | Registrar nova venda |
| `GET` | `/api/vendas/dia` | Vendas do dia |
| `GET` | `/api/vendas/{id}` | Buscar venda por ID |
| `POST` | `/api/vendas/{id}/cancelar` | Cancelar venda |

### Produtos

| Método | Rota | Descrição |
|--------|------|-----------|
| `GET` | `/api/produtos` | Listar (paginado) |
| `GET` | `/api/produtos/busca` | Buscar por código/nome |
| `GET` | `/api/produtos/{id}` | Detalhes do produto |
| `POST` | `/api/produtos` | Criar produto *(gerente+)* |
| `PUT` | `/api/produtos/{id}` | Atualizar *(gerente+)* |
| `DELETE` | `/api/produtos/{id}` | Inativar *(gerente+)* |

### Estoque

| Método | Rota | Descrição |
|--------|------|-----------|
| `GET` | `/api/estoque/baixo` | Produtos com estoque baixo |
| `POST` | `/api/estoque/ajuste` | Ajuste manual *(gerente+)* |
| `POST` | `/api/estoque/inventario` | Criar inventário *(gerente+)* |
| `PUT` | `/api/estoque/inventario/{id}` | Finalizar inventário *(gerente+)* |

### Clientes, Fornecedores, Compras

| Método | Rota | Descrição |
|--------|------|-----------|
| `GET/POST/PUT` | `/api/clientes` | CRUD de clientes |
| `GET/POST` | `/api/fornecedores` | CRUD de fornecedores *(gerente+)* |
| `POST` | `/api/compras` | Registrar compra *(gerente+)* |
| `POST` | `/api/compras/{id}/receber` | Receber mercadoria *(gerente+)* |

### Financeiro

| Método | Rota | Descrição |
|--------|------|-----------|
| `GET/POST` | `/api/financeiro/contas-pagar` | Contas a pagar *(gerente+)* |
| `POST` | `/api/financeiro/contas-pagar/{id}/pagar` | Pagar conta *(gerente+)* |
| `GET` | `/api/financeiro/contas-receber` | Contas a receber *(gerente+)* |
| `POST` | `/api/financeiro/contas-receber/{id}/receber` | Receber conta *(gerente+)* |
| `GET` | `/api/financeiro/fluxo-caixa` | Fluxo de caixa *(gerente+)* |

### Relatórios *(supervisor+)*

| Método | Rota | Descrição |
|--------|------|-----------|
| `GET` | `/api/relatorios/vendas/dia` | Relatório do dia |
| `GET` | `/api/relatorios/vendas/mes` | Relatório do mês |
| `GET` | `/api/relatorios/vendas/periodo` | Relatório por período |
| `GET` | `/api/relatorios/produtos/mais-vendidos` | Top 20 mais vendidos |

### Admin *(admin apenas)*

| Método | Rota | Descrição |
|--------|------|-----------|
| `GET/PUT` | `/api/config` | Configurações do sistema |
| `GET/POST` | `/api/usuarios` | Gerenciar usuários |
| `POST` | `/api/backup` | Criar backup |

### WebSocket

```
ws://localhost:8080/ws?token={jwt}
```

Eventos: `nova_venda`, `alerta_estoque`, `caixa_status`

## 🔐 Autenticação

O sistema usa **JWT (JSON Web Token)** com expiração de 8 horas.

```
Authorization: Bearer {token}
```

### Perfis e Hierarquia

| Perfil | Nível | Acessos |
|--------|-------|---------|
| `caixa` | 1 | Vendas, abrir/fechar caixa |
| `supervisor` | 2 | + Cancelar venda, relatórios |
| `gerente` | 3 | + Produtos, estoque, compras, financeiro |
| `admin` | 4 | + Usuários, configurações, backup |

## 📄 Licença

Projeto privado - UnifyTechXenos © 2026
