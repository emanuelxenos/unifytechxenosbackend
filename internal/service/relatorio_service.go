package service

import (
	"context"
	"fmt"

	"erp-backend/internal/infrastructure/database"
)

type RelatorioService struct {
	db *database.PostgresDB
}

func NewRelatorioService(db *database.PostgresDB) *RelatorioService {
	return &RelatorioService{db: db}
}

type RelatorioVendasDia struct {
	TotalVendas       int                      `json:"total_vendas"`
	ValorTotal        float64                  `json:"valor_total"`
	TicketMedio       float64                  `json:"ticket_medio"`
	PorCaixa          []RelatorioVendasCaixa   `json:"por_caixa"`
	PorFormaPagamento []RelatorioFormaPagamento `json:"por_forma_pagamento"`
}

type RelatorioVendasCaixa struct {
	Caixa      string  `json:"caixa"`
	Total      int     `json:"total"`
	ValorTotal float64 `json:"valor_total"`
}

type RelatorioFormaPagamento struct {
	FormaPagamento string  `json:"forma_pagamento"`
	Total          int     `json:"total"`
	ValorTotal     float64 `json:"valor_total"`
}

type ProdutoMaisVendido struct {
	ID          int     `json:"id_produto"`
	Nome        string  `json:"nome"`
	Quantidade  float64 `json:"quantidade_vendida"`
	ValorTotal  float64 `json:"valor_total"`
}

func (s *RelatorioService) preencherDetalhesVenda(ctx context.Context, rel *RelatorioVendasDia, empresaID int, condicaoData string, args ...interface{}) {
	// Por caixa
	queryCaixa := fmt.Sprintf(`SELECT cf.nome, COUNT(*), COALESCE(SUM(v.valor_total), 0)
		 FROM venda v JOIN caixa_fisico cf ON v.caixa_fisico_id = cf.id_caixa_fisico
		 WHERE v.empresa_id = $1 AND %s AND v.status = 'concluida'
		 GROUP BY cf.nome`, condicaoData)
	
	rows, _ := s.db.Pool.Query(ctx, queryCaixa, args...)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var c RelatorioVendasCaixa
			rows.Scan(&c.Caixa, &c.Total, &c.ValorTotal)
			rel.PorCaixa = append(rel.PorCaixa, c)
		}
	}

	// Por forma de pagamento
	queryFp := fmt.Sprintf(`SELECT fp.nome, COUNT(*), COALESCE(SUM(vp.valor), 0)
		 FROM venda_pagamento vp
		 JOIN venda v ON vp.venda_id = v.id_venda
		 JOIN forma_pagamento fp ON vp.forma_pagamento_id = fp.id_forma_pagamento
		 WHERE v.empresa_id = $1 AND %s AND v.status = 'concluida'
		 GROUP BY fp.nome`, condicaoData)
	rows2, _ := s.db.Pool.Query(ctx, queryFp, args...)
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var f RelatorioFormaPagamento
			rows2.Scan(&f.FormaPagamento, &f.Total, &f.ValorTotal)
			rel.PorFormaPagamento = append(rel.PorFormaPagamento, f)
		}
	}
}

func (s *RelatorioService) VendasDia(ctx context.Context, empresaID int) (*RelatorioVendasDia, error) {
	rel := &RelatorioVendasDia{}

	s.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(SUM(valor_total), 0)
		 FROM venda WHERE empresa_id = $1 AND DATE(data_venda) = CURRENT_DATE AND status = 'concluida'`,
		empresaID).Scan(&rel.TotalVendas, &rel.ValorTotal)

	if rel.TotalVendas > 0 {
		rel.TicketMedio = rel.ValorTotal / float64(rel.TotalVendas)
	}

	s.preencherDetalhesVenda(ctx, rel, empresaID, "DATE(v.data_venda) = CURRENT_DATE", empresaID)

	return rel, nil
}

func (s *RelatorioService) VendasMes(ctx context.Context, empresaID int) (*RelatorioVendasDia, error) {
	rel := &RelatorioVendasDia{}
	s.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(SUM(valor_total), 0)
		 FROM venda WHERE empresa_id = $1
		 AND EXTRACT(MONTH FROM data_venda) = EXTRACT(MONTH FROM CURRENT_DATE)
		 AND EXTRACT(YEAR FROM data_venda) = EXTRACT(YEAR FROM CURRENT_DATE)
		 AND status = 'concluida'`,
		empresaID).Scan(&rel.TotalVendas, &rel.ValorTotal)
		
	if rel.TotalVendas > 0 {
		rel.TicketMedio = rel.ValorTotal / float64(rel.TotalVendas)
	}

	condicao := "EXTRACT(MONTH FROM v.data_venda) = EXTRACT(MONTH FROM CURRENT_DATE) AND EXTRACT(YEAR FROM v.data_venda) = EXTRACT(YEAR FROM CURRENT_DATE)"
	s.preencherDetalhesVenda(ctx, rel, empresaID, condicao, empresaID)

	return rel, nil
}

func (s *RelatorioService) VendasPeriodo(ctx context.Context, empresaID int, dataInicio, dataFim string) (*RelatorioVendasDia, error) {
	rel := &RelatorioVendasDia{}
	s.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(SUM(valor_total), 0)
		 FROM venda WHERE empresa_id = $1
		 AND DATE(data_venda) >= $2::date AND DATE(data_venda) <= $3::date
		 AND status = 'concluida'`,
		empresaID, dataInicio, dataFim).Scan(&rel.TotalVendas, &rel.ValorTotal)
		
	if rel.TotalVendas > 0 {
		rel.TicketMedio = rel.ValorTotal / float64(rel.TotalVendas)
	}

	condicao := "DATE(v.data_venda) >= $2::date AND DATE(v.data_venda) <= $3::date"
	s.preencherDetalhesVenda(ctx, rel, empresaID, condicao, empresaID, dataInicio, dataFim)

	return rel, nil
}

func (s *RelatorioService) MaisVendidos(ctx context.Context, empresaID int, periodo string) ([]ProdutoMaisVendido, error) {
	days := "30"
	switch periodo {
	case "7d":
		days = "7"
	case "90d":
		days = "90"
	}

	rows, err := s.db.Pool.Query(ctx,
		fmt.Sprintf(`SELECT p.id_produto, p.nome, COALESCE(SUM(iv.quantidade), 0), COALESCE(SUM(iv.valor_liquido), 0)
		 FROM item_venda iv
		 JOIN produto p ON iv.produto_id = p.id_produto
		 JOIN venda v ON iv.venda_id = v.id_venda
		 WHERE v.empresa_id = $1 AND iv.status = 'vendido'
		 AND iv.data_hora >= CURRENT_DATE - INTERVAL '%s days'
		 GROUP BY p.id_produto, p.nome
		 ORDER BY SUM(iv.quantidade) DESC LIMIT 20`, days),
		empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ProdutoMaisVendido
	for rows.Next() {
		var p ProdutoMaisVendido
		rows.Scan(&p.ID, &p.Nome, &p.Quantidade, &p.ValorTotal)
		list = append(list, p)
	}
	return list, nil
}

type RelatorioEstoque struct {
	TotalProdutos   int     `json:"total_produtos"`
	ValorTotalCusto float64 `json:"valor_total_custo"`
	ValorTotalVenda float64 `json:"valor_total_venda"`
	ProdutosBaixos  int     `json:"produtos_baixo_estoque"`
}

func (s *RelatorioService) EstoqueResumo(ctx context.Context, empresaID int) (*RelatorioEstoque, error) {
	rel := &RelatorioEstoque{}
	err := s.db.Pool.QueryRow(ctx,
		`SELECT 
			COUNT(id_produto),
			COALESCE(SUM(estoque_atual * preco_custo), 0),
			COALESCE(SUM(estoque_atual * preco_venda), 0),
			COUNT(CASE WHEN estoque_atual <= estoque_minimo AND controlar_estoque = TRUE THEN 1 END)
		 FROM produto WHERE empresa_id = $1 AND ativo = TRUE`,
		empresaID).Scan(&rel.TotalProdutos, &rel.ValorTotalCusto, &rel.ValorTotalVenda, &rel.ProdutosBaixos)
	return rel, err
}

type RelatorioFinanceiro struct {
	ContasPagarAbertas int     `json:"contas_pagar_abertas"`
	ValorPagarAberto   float64 `json:"valor_pagar_aberto"`
	ContasReceberAberto int     `json:"contas_receber_abertas"`
	ValorReceberAberto  float64 `json:"valor_receber_aberto"`
	SaldoCaixaDia      float64 `json:"saldo_caixa_dia"`
}

func (s *RelatorioService) FinanceiroResumo(ctx context.Context, empresaID int) (*RelatorioFinanceiro, error) {
	rel := &RelatorioFinanceiro{}
	
	s.db.Pool.QueryRow(ctx, `SELECT COUNT(*), COALESCE(SUM(valor_original - valor_pago), 0) FROM conta_pagar WHERE empresa_id = $1 AND status = 'aberta'`, empresaID).Scan(&rel.ContasPagarAbertas, &rel.ValorPagarAberto)
	s.db.Pool.QueryRow(ctx, `SELECT COUNT(*), COALESCE(SUM(valor_original - valor_recebido), 0) FROM conta_receber WHERE empresa_id = $1 AND status = 'aberta'`, empresaID).Scan(&rel.ContasReceberAberto, &rel.ValorReceberAberto)
	s.db.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(CASE WHEN tipo = 'venda' OR tipo = 'recebimento' OR tipo = 'suprimento' THEN valor 
								 WHEN tipo = 'pagamento' OR tipo = 'sangria' THEN -valor ELSE 0 END), 0)
		FROM vw_fluxo_caixa WHERE empresa_id = $1 AND data = CURRENT_DATE`, empresaID).Scan(&rel.SaldoCaixaDia)
		
	return rel, nil
}
