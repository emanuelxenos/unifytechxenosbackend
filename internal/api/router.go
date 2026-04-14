package api

import (
	"github.com/go-chi/chi/v5"

	"erp-backend/internal/api/handlers"
	"erp-backend/internal/api/middleware"
	"erp-backend/internal/infrastructure/database"
	ws "erp-backend/internal/infrastructure/websocket"
	"erp-backend/pkg/config"
)

func NewRouter(db *database.PostgresDB, cfg *config.Config, hub *ws.Hub) *chi.Mux {
	r := chi.NewRouter()

	// Middleware global
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LocalNetworkMiddleware(cfg.RestrictLocalNetwork))

	// Handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	caixaHandler := handlers.NewCaixaHandler(db)
	vendaHandler := handlers.NewVendaHandler(db)
	produtoHandler := handlers.NewProdutoHandler(db)
	estoqueHandler := handlers.NewEstoqueHandler(db)
	clienteHandler := handlers.NewClienteHandler(db)
	fornecedorHandler := handlers.NewFornecedorHandler(db)
	compraHandler := handlers.NewCompraHandler(db)
	financeiroHandler := handlers.NewFinanceiroHandler(db)
	relatorioHandler := handlers.NewRelatorioHandler(db)
	configHandler := handlers.NewConfigHandler(db)
	usuarioHandler := handlers.NewUsuarioHandler(db)
	empresaHandler := handlers.NewEmpresaHandler(db)

	// Rotas públicas
	r.Get("/health", authHandler.Health)
	r.Post("/api/login", authHandler.Login)
	r.Get("/api/discover", authHandler.Discover)

	// WebSocket
	r.Get("/ws", hub.HandleWebSocket)

	// Rotas autenticadas
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		// Caixa
		r.Get("/api/caixa/status", caixaHandler.Status)
		r.Post("/api/caixa/abrir", caixaHandler.Abrir)
		r.Post("/api/caixa/fechar", caixaHandler.Fechar)
		r.Post("/api/caixa/sangria", caixaHandler.Sangria)
		r.Post("/api/caixa/suprimento", caixaHandler.Suprimento)

		// Vendas
		r.Post("/api/vendas", vendaHandler.Criar)
		r.Get("/api/vendas/dia", vendaHandler.VendasDia)
		r.Get("/api/vendas/{id}", vendaHandler.BuscarPorID)
		r.Post("/api/vendas/{id}/cancelar", vendaHandler.Cancelar)

		// Produtos
		r.Get("/api/produtos", produtoHandler.Listar)
		r.Get("/api/produtos/busca", produtoHandler.Buscar)
		r.Get("/api/produtos/{id}", produtoHandler.BuscarPorID)

		// Clientes
		r.Get("/api/clientes", clienteHandler.Listar)
		r.Post("/api/clientes", clienteHandler.Criar)
		r.Put("/api/clientes/{id}", clienteHandler.Atualizar)

		// Fornecedores (leitura)
		r.Get("/api/fornecedores", fornecedorHandler.Listar)

		// Estoque (leitura)
		r.Get("/api/estoque/baixo", estoqueHandler.EstoqueBaixo)

		// Gerente+ (produtos, estoque, compras, financeiro)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireProfile("gerente"))

			// Produtos CRUD
			r.Post("/api/produtos", produtoHandler.Criar)
			r.Put("/api/produtos/{id}", produtoHandler.Atualizar)
			r.Delete("/api/produtos/{id}", produtoHandler.Inativar)

			// Estoque
			r.Post("/api/estoque/ajuste", estoqueHandler.Ajuste)
			r.Post("/api/estoque/inventario", estoqueHandler.CriarInventario)
			r.Put("/api/estoque/inventario/{id}", estoqueHandler.FinalizarInventario)

			// Fornecedores
			r.Post("/api/fornecedores", fornecedorHandler.Criar)
			r.Put("/api/fornecedores/{id}", fornecedorHandler.Atualizar)
			r.Delete("/api/fornecedores/{id}", fornecedorHandler.Inativar)

			// Compras
			r.Get("/api/compras", compraHandler.Listar)
			r.Get("/api/compras/{id}", compraHandler.BuscarPorID)
			r.Post("/api/compras", compraHandler.Criar)
			r.Post("/api/compras/{id}/receber", compraHandler.Receber)

			// Financeiro
			r.Get("/api/financeiro/contas-pagar", financeiroHandler.ContasPagar)
			r.Post("/api/financeiro/contas-pagar", financeiroHandler.CriarContaPagar)
			r.Post("/api/financeiro/contas-pagar/{id}/pagar", financeiroHandler.PagarConta)
			r.Get("/api/financeiro/contas-receber", financeiroHandler.ContasReceber)
			r.Post("/api/financeiro/contas-receber/{id}/receber", financeiroHandler.ReceberConta)
			r.Get("/api/financeiro/fluxo-caixa", financeiroHandler.FluxoCaixa)
		})

		// Supervisor+ (relatórios)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireProfile("supervisor"))

			r.Get("/api/relatorios/vendas/dia", relatorioHandler.VendasDia)
			r.Get("/api/relatorios/vendas/mes", relatorioHandler.VendasMes)
			r.Get("/api/relatorios/vendas/periodo", relatorioHandler.VendasPeriodo)
			r.Get("/api/relatorios/produtos/mais-vendidos", relatorioHandler.MaisVendidos)
			r.Get("/api/relatorios/exportar/pdf", relatorioHandler.ExportarPDF)
			r.Get("/api/relatorios/exportar/excel", relatorioHandler.ExportarExcel)
		})

		// Admin (configurações, usuários, backup)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireProfile("admin"))

			r.Get("/api/config", configHandler.Listar)
			r.Put("/api/config", configHandler.Atualizar)
			r.Get("/api/empresa", empresaHandler.Buscar)
			r.Put("/api/empresa", empresaHandler.Atualizar)
			r.Get("/api/usuarios", usuarioHandler.Listar)
			r.Post("/api/usuarios", usuarioHandler.Criar)
			r.Put("/api/usuarios/{id}", usuarioHandler.Atualizar)
			r.Delete("/api/usuarios/{id}", usuarioHandler.Inativar)
			r.Post("/api/backup", configHandler.Backup)
			r.Post("/api/backup/restaurar", configHandler.Restaurar)
		})
	})

	return r
}
