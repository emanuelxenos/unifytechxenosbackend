package service

import (
	"context"
	"fmt"

	"erp-backend/internal/infrastructure/database"
	"time"
)

type RelatorioService struct {
	db *database.PostgresDB
}

func NewRelatorioService(db *database.PostgresDB) *RelatorioService {
	return &RelatorioService{db: db}
}

func (s *RelatorioService) GetDB() *database.PostgresDB {
	return s.db
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
	TotalProdutos      int     `json:"total_produtos"`
	ValorTotalCusto    float64 `json:"valor_total_custo"`
	ValorTotalVenda    float64 `json:"valor_total_venda"`
	ProdutosBaixos     int     `json:"produtos_baixo_estoque"`
	SugestaoCompraTotal float64 `json:"sugestao_compra_total"`
	ProdutosVencendo   int     `json:"produtos_vencendo"`
}

type PerformanceProduto struct {
	Mes     string  `json:"mes"`
	Entrada float64 `json:"entrada"`
	Saida   float64 `json:"saida"`
}

func (s *RelatorioService) EstoqueResumo(ctx context.Context, empresaID int) (*RelatorioEstoque, error) {
	rel := &RelatorioEstoque{}
	err := s.db.Pool.QueryRow(ctx,
		`SELECT 
			COUNT(id_produto),
			COALESCE(SUM(estoque_atual * preco_custo), 0),
			COALESCE(SUM(estoque_atual * preco_venda), 0),
			COUNT(CASE WHEN estoque_atual <= estoque_minimo AND controlar_estoque = TRUE THEN 1 END),
			COALESCE(SUM(CASE WHEN estoque_atual < estoque_minimo THEN (estoque_minimo - estoque_atual) * preco_custo ELSE 0 END), 0),
			COUNT(CASE WHEN data_vencimento <= CURRENT_DATE + INTERVAL '15 days' THEN 1 END)
		 FROM produto WHERE empresa_id = $1 AND ativo = TRUE`,
		empresaID).Scan(&rel.TotalProdutos, &rel.ValorTotalCusto, &rel.ValorTotalVenda, &rel.ProdutosBaixos, &rel.SugestaoCompraTotal, &rel.ProdutosVencendo)
	return rel, err
}

type ProdutoRelatorio struct {
	ID           int     `json:"id"`
	Nome         string  `json:"nome"`
	PrecoCusto   float64 `json:"preco_custo"`
	PrecoVenda   float64 `json:"preco_venda"`
	EstoqueAtual float64 `json:"estoque_atual"`
	EstoqueMin   float64 `json:"estoque_minimo"`
	Categoria    string  `json:"categoria"`
}

func (s *RelatorioService) ListaEstoque(ctx context.Context, empresaID int, search string, catID int, baixoEstoque, vencendo bool) ([]ProdutoRelatorio, error) {
	query := `
		SELECT p.id_produto, p.nome, p.preco_custo, p.preco_venda, p.estoque_atual, p.estoque_minimo, COALESCE(c.nome, 'Sem Categoria')
		FROM produto p
		LEFT JOIN categoria c ON p.categoria_id = c.id_categoria
		WHERE p.empresa_id = $1 AND p.ativo = TRUE
	`
	args := []interface{}{empresaID}
	placeholderID := 2

	if search != "" {
		query += fmt.Sprintf(" AND p.nome ILIKE $%d", placeholderID)
		args = append(args, "%"+search+"%")
		placeholderID++
	}
	if catID > 0 {
		query += fmt.Sprintf(" AND p.categoria_id = $%d", placeholderID)
		args = append(args, catID)
		placeholderID++
	}
	if baixoEstoque {
		query += " AND p.estoque_atual <= p.estoque_minimo AND p.controlar_estoque = TRUE"
	}
	if vencendo {
		query += " AND p.data_vencimento <= CURRENT_DATE + INTERVAL '15 days'"
	}

	query += " ORDER BY p.nome ASC LIMIT 1000"

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ProdutoRelatorio
	for rows.Next() {
		var p ProdutoRelatorio
		rows.Scan(&p.ID, &p.Nome, &p.PrecoCusto, &p.PrecoVenda, &p.EstoqueAtual, &p.EstoqueMin, &p.Categoria)
		list = append(list, p)
	}
	return list, nil
}

func (s *RelatorioService) PerformanceProduto(ctx context.Context, empresaID, produtoID int) ([]PerformanceProduto, error) {
	query := `
		WITH meses AS (
			SELECT generate_series(
				CURRENT_DATE - INTERVAL '5 months',
				CURRENT_DATE,
				'1 month'::interval
			)::date as mes_data
		)
		SELECT 
			TO_CHAR(m.mes_data, 'MM/YYYY') as mes,
			COALESCE(SUM(CASE WHEN em.tipo_movimentacao IN ('entrada', 'compra', 'inventario_entrada', 'ajuste_entrada') THEN em.quantidade ELSE 0 END), 0) as entrada,
			COALESCE(SUM(CASE WHEN em.tipo_movimentacao IN ('saida', 'venda', 'venda_item', 'inventario_saida', 'ajuste_saida', 'perda') THEN em.quantidade ELSE 0 END), 0) as saida
		FROM meses m
		LEFT JOIN estoque_movimentacao em ON TO_CHAR(em.data_movimentacao, 'MM/YYYY') = TO_CHAR(m.mes_data, 'MM/YYYY')
			AND em.empresa_id = $1 AND em.produto_id = $2
		GROUP BY m.mes_data
		ORDER BY m.mes_data ASC
	`
	rows, err := s.db.Pool.Query(ctx, query, empresaID, produtoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []PerformanceProduto
	for rows.Next() {
		var p PerformanceProduto
		if err := rows.Scan(&p.Mes, &p.Entrada, &p.Saida); err != nil {
			return nil, err
		}
		list = append(list, p)
	}

	return list, nil
}

type ProdutoSugestao struct {
	ID           int     `json:"id_produto"`
	Nome         string  `json:"nome"`
	EstoqueAtual float64 `json:"estoque_atual"`
	EstoqueMin   float64 `json:"estoque_minimo"`
	SugestaoQtd  float64 `json:"sugestao_quantidade"`
	PrecoCusto   float64 `json:"preco_custo"`
}

func (s *RelatorioService) SugestaoCompra(ctx context.Context, empresaID int) ([]ProdutoSugestao, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT id_produto, nome, estoque_atual, estoque_minimo, preco_custo
		 FROM produto 
		 WHERE empresa_id = $1 AND ativo = TRUE AND controlar_estoque = TRUE AND estoque_atual <= estoque_minimo
		 ORDER BY (estoque_minimo - estoque_atual) DESC`,
		empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ProdutoSugestao
	for rows.Next() {
		var p ProdutoSugestao
		err := rows.Scan(&p.ID, &p.Nome, &p.EstoqueAtual, &p.EstoqueMin, &p.PrecoCusto)
		if err == nil {
			p.SugestaoQtd = (p.EstoqueMin * 2) - p.EstoqueAtual
			if p.SugestaoQtd < 0 {
				p.SugestaoQtd = 0
			}
			list = append(list, p)
		}
	}
	return list, nil
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

type RelatorioDRE struct {
	ReceitaBruta     float64 `json:"receita_bruta"`
	Descontos        float64 `json:"descontos"`
	ReceitaLiquida   float64 `json:"receita_liquida"`
	CMV              float64 `json:"cmv"`
	LucroBruto       float64 `json:"lucro_bruto"`
	Despesas         float64 `json:"despesas"`
	LucroLiquido     float64 `json:"lucro_liquido"`
	MargemPercentual float64 `json:"margem_percentual"`
}

func (s *RelatorioService) DREGerencial(ctx context.Context, empresaID int, mes, ano int) (*RelatorioDRE, error) {
	rel := &RelatorioDRE{}

	// Receitas e Descontos
	s.db.Pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(valor_total), 0), COALESCE(SUM(valor_total_descontos), 0)
		 FROM venda WHERE empresa_id = $1 AND status = 'concluida'
		 AND EXTRACT(MONTH FROM data_venda) = $2 AND EXTRACT(YEAR FROM data_venda) = $3`,
		empresaID, mes, ano).Scan(&rel.ReceitaBruta, &rel.Descontos)

	rel.ReceitaLiquida = rel.ReceitaBruta - rel.Descontos

	// CMV (Custo da Mercadoria Vendida)
	s.db.Pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(iv.quantidade * iv.preco_custo), 0)
		 FROM item_venda iv
		 JOIN venda v ON iv.venda_id = v.id_venda
		 WHERE v.empresa_id = $1 AND v.status = 'concluida'
		 AND EXTRACT(MONTH FROM v.data_venda) = $2 AND EXTRACT(YEAR FROM v.data_venda) = $3`,
		empresaID, mes, ano).Scan(&rel.CMV)

	rel.LucroBruto = rel.ReceitaLiquida - rel.CMV

	// Despesas Operacionais (Contas Pagas no período)
	s.db.Pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(valor_pago), 0)
		 FROM conta_pagar WHERE empresa_id = $1 AND status = 'paga'
		 AND EXTRACT(MONTH FROM data_pagamento) = $2 AND EXTRACT(YEAR FROM data_pagamento) = $3`,
		empresaID, mes, ano).Scan(&rel.Despesas)

	rel.LucroLiquido = rel.LucroBruto - rel.Despesas
	if rel.ReceitaLiquida > 0 {
		rel.MargemPercentual = (rel.LucroLiquido / rel.ReceitaLiquida) * 100
	}

	return rel, nil
}

type InadimplenciaItem struct {
	ID             int     `json:"id"`
	Cliente        string  `json:"cliente"`
	Valor          float64 `json:"valor"`
	DataVencimento string  `json:"data_vencimento"`
	DiasAtraso     int     `json:"dias_atraso"`
}

type RelatorioInadimplencia struct {
	TotalVencido float64             `json:"total_vencido"`
	Quantidade   int                 `json:"quantidade"`
	Itens        []InadimplenciaItem `json:"itens"`
}

func (s *RelatorioService) InadimplenciaResumo(ctx context.Context, empresaID int) (*RelatorioInadimplencia, error) {
	rel := &RelatorioInadimplencia{Itens: []InadimplenciaItem{}}

	rows, err := s.db.Pool.Query(ctx,
		`SELECT cr.id_conta_receber, c.nome, (cr.valor_original - cr.valor_recebido), cr.data_vencimento, 
		        CURRENT_DATE - cr.data_vencimento as dias_atraso
		 FROM conta_receber cr
		 JOIN cliente c ON cr.cliente_id = c.id_cliente
		 WHERE cr.empresa_id = $1 AND cr.status = 'aberta' AND cr.data_vencimento < CURRENT_DATE
		 ORDER BY cr.data_vencimento ASC`,
		empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item InadimplenciaItem
		var dataVenc time.Time
		rows.Scan(&item.ID, &item.Cliente, &item.Valor, &dataVenc, &item.DiasAtraso)
		item.DataVencimento = dataVenc.Format("2006-01-02")
		rel.Itens = append(rel.Itens, item)
		rel.TotalVencido += item.Valor
		rel.Quantidade++
	}

	return rel, nil
}

type ProdutoABC struct {
	ID             int     `json:"id_produto"`
	Nome           string  `json:"nome"`
	Faturamento    float64 `json:"faturamento"`
	Percentual     float64 `json:"percentual"`
	Acumulado      float64 `json:"percentual_acumulado"`
	Classificacao  string  `json:"classificacao"`
}

type RelatorioCurvaABC struct {
	TotalFaturamento float64      `json:"total_faturamento"`
	Itens            []ProdutoABC `json:"itens"`
}

func (s *RelatorioService) CurvaABC(ctx context.Context, empresaID int) (*RelatorioCurvaABC, error) {
	rel := &RelatorioCurvaABC{Itens: []ProdutoABC{}}

	// 1. Calcular faturamento total nos últimos 90 dias
	s.db.Pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(iv.valor_liquido), 0)
		 FROM item_venda iv
		 JOIN venda v ON iv.venda_id = v.id_venda
		 WHERE v.empresa_id = $1 AND v.status = 'concluida'
		 AND v.data_venda >= CURRENT_DATE - INTERVAL '90 days'`,
		empresaID).Scan(&rel.TotalFaturamento)

	if rel.TotalFaturamento == 0 {
		return rel, nil
	}

	// 2. Buscar faturamento por produto
	rows, err := s.db.Pool.Query(ctx,
		`SELECT p.id_produto, p.nome, SUM(iv.valor_liquido) as faturamento
		 FROM item_venda iv
		 JOIN produto p ON iv.produto_id = p.id_produto
		 JOIN venda v ON iv.venda_id = v.id_venda
		 WHERE v.empresa_id = $1 AND v.status = 'concluida'
		 AND v.data_venda >= CURRENT_DATE - INTERVAL '90 days'
		 GROUP BY p.id_produto, p.nome
		 ORDER BY faturamento DESC`,
		empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	acumulado := 0.0
	for rows.Next() {
		var item ProdutoABC
		rows.Scan(&item.ID, &item.Nome, &item.Faturamento)
		
		item.Percentual = (item.Faturamento / rel.TotalFaturamento) * 100
		acumulado += item.Percentual
		item.Acumulado = acumulado

		if acumulado <= 80.5 { // Uma margem pequena para arredondamentos
			item.Classificacao = "A"
		} else if acumulado <= 95.5 {
			item.Classificacao = "B"
		} else {
			item.Classificacao = "C"
		}
		
		rel.Itens = append(rel.Itens, item)
	}

	return rel, nil
}

type ComissaoOperador struct {
	UsuarioID     int     `json:"usuario_id"`
	Nome          string  `json:"nome"`
	TotalVendas   int     `json:"total_vendas"`
	ValorTotal    float64 `json:"valor_total"`
	TicketMedio   float64 `json:"ticket_medio"`
	Comissao      float64 `json:"comissao"`
}

type RelatorioComissoes struct {
	TotalGeral    float64            `json:"total_geral"`
	TotalComissao float64            `json:"total_comissao"`
	Operadores    []ComissaoOperador `json:"operadores"`
}

func (s *RelatorioService) ComissoesOperador(ctx context.Context, empresaID int, mes, ano int) (*RelatorioComissoes, error) {
	rel := &RelatorioComissoes{Operadores: []ComissaoOperador{}}

	rows, err := s.db.Pool.Query(ctx,
		`SELECT u.id_usuario, u.nome, COUNT(v.id_venda), COALESCE(SUM(v.valor_total), 0)
		 FROM venda v
		 JOIN usuario u ON v.usuario_id = u.id_usuario
		 WHERE v.empresa_id = $1 AND v.status = 'concluida'
		 AND EXTRACT(MONTH FROM v.data_venda) = $2 AND EXTRACT(YEAR FROM v.data_venda) = $3
		 GROUP BY u.id_usuario, u.nome
		 ORDER BY SUM(v.valor_total) DESC`,
		empresaID, mes, ano)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c ComissaoOperador
		rows.Scan(&c.UsuarioID, &c.Nome, &c.TotalVendas, &c.ValorTotal)
		
		if c.TotalVendas > 0 {
			c.TicketMedio = c.ValorTotal / float64(c.TotalVendas)
		}
		c.Comissao = c.ValorTotal * 0.01 // 1% fixo
		
		rel.Operadores = append(rel.Operadores, c)
		rel.TotalGeral += c.ValorTotal
		rel.TotalComissao += c.Comissao
	}

	return rel, nil
}
