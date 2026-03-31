package service

import (
	"context"
	"fmt"
	"time"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
	"erp-backend/pkg/utils"
)

type VendaService struct {
	db *database.PostgresDB
}

func NewVendaService(db *database.PostgresDB) *VendaService {
	return &VendaService{db: db}
}

func (s *VendaService) Criar(ctx context.Context, empresaID, usuarioID int, req models.CriarVendaRequest) (*models.VendaResponse, error) {
	// Buscar sessão aberta do usuário
	var sessaoID, caixaFisicoID int
	err := s.db.Pool.QueryRow(ctx,
		`SELECT id_sessao, caixa_fisico_id FROM sessao_caixa
		 WHERE empresa_id = $1 AND usuario_id = $2 AND status = 'aberto'`,
		empresaID, usuarioID,
	).Scan(&sessaoID, &caixaFisicoID)
	if err != nil {
		return nil, fmt.Errorf("nenhuma sessão de caixa aberta")
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	// Calcular totais
	var valorTotalProdutos, valorTotalDescontos float64
	for _, item := range req.Itens {
		valorItem := item.Quantidade * item.PrecoUnitario
		valorTotalProdutos += valorItem
		valorTotalDescontos += item.ValorDesconto
	}
	valorTotal := valorTotalProdutos - valorTotalDescontos

	// Calcular valor pago e troco
	var valorPago float64
	for _, pag := range req.Pagamentos {
		valorPago += pag.Valor
	}
	valorTroco := valorPago - valorTotal
	if valorTroco < 0 {
		valorTroco = 0
	}

	// Inserir venda
	var vendaID int
	var numeroVenda string
	err = tx.QueryRow(ctx,
		`INSERT INTO venda (empresa_id, sessao_caixa_id, usuario_id, caixa_fisico_id,
		                    cliente_id, valor_total_produtos, valor_total_descontos,
		                    valor_subtotal, valor_total, valor_pago, valor_troco,
		                    status, tipo_venda, observacoes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'concluida', 'venda', $12)
		 RETURNING id_venda, numero_venda`,
		empresaID, sessaoID, usuarioID, caixaFisicoID,
		req.ClienteID, valorTotalProdutos, valorTotalDescontos,
		valorTotal, valorTotal, valorPago, valorTroco,
		req.Observacoes,
	).Scan(&vendaID, &numeroVenda)
	if err != nil {
		return nil, fmt.Errorf("erro ao registrar venda: %w", err)
	}

	// Inserir itens da venda
	for i, item := range req.Itens {
		valorItem := item.Quantidade * item.PrecoUnitario
		valorLiquido := valorItem - item.ValorDesconto

		_, err = tx.Exec(ctx,
			`INSERT INTO item_venda (venda_id, produto_id, sequencia, quantidade,
			                         preco_unitario, valor_total, valor_desconto,
			                         valor_liquido, status)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'vendido')`,
			vendaID, item.ProdutoID, i+1, item.Quantidade,
			item.PrecoUnitario, valorItem, item.ValorDesconto, valorLiquido,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao inserir item %d: %w", i+1, err)
		}
	}

	// Inserir pagamentos
	for _, pag := range req.Pagamentos {
		parcelas := pag.Parcelas
		if parcelas == 0 {
			parcelas = 1
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO venda_pagamento (venda_id, forma_pagamento_id, valor, parcelas, autorizacao, status)
			 VALUES ($1, $2, $3, $4, $5, 'aprovado')`,
			vendaID, pag.FormaPagamentoID, pag.Valor, parcelas, pag.Autorizacao,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao inserir pagamento: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	return &models.VendaResponse{
		IDVenda:     vendaID,
		NumeroVenda: numeroVenda,
		ValorTotal:  valorTotal,
		ValorTroco:  valorTroco,
	}, nil
}

func (s *VendaService) BuscarPorID(ctx context.Context, empresaID, vendaID int) (*models.Venda, error) {
	var venda models.Venda
	err := s.db.Pool.QueryRow(ctx,
		`SELECT v.id_venda, v.empresa_id, v.sessao_caixa_id, v.usuario_id,
		        v.caixa_fisico_id, v.numero_venda, v.cliente_id,
		        v.data_venda, v.valor_total_produtos, v.valor_total_descontos,
		        v.valor_total, v.valor_pago, v.valor_troco, v.status,
		        u.nome as operador_nome, cf.nome as caixa_nome
		 FROM venda v
		 JOIN usuario u ON v.usuario_id = u.id_usuario
		 JOIN caixa_fisico cf ON v.caixa_fisico_id = cf.id_caixa_fisico
		 WHERE v.id_venda = $1 AND v.empresa_id = $2`,
		vendaID, empresaID,
	).Scan(
		&venda.ID, &venda.EmpresaID, &venda.SessaoCaixaID, &venda.UsuarioID,
		&venda.CaixaFisicoID, &venda.NumeroVenda, &venda.ClienteID,
		&venda.DataVenda, &venda.ValorTotalProdutos, &venda.ValorTotalDescontos,
		&venda.ValorTotal, &venda.ValorPago, &venda.ValorTroco, &venda.Status,
		&venda.OperadorNome, &venda.CaixaNome,
	)
	if err != nil {
		return nil, fmt.Errorf("venda não encontrada")
	}

	// Buscar itens
	rows, err := s.db.Pool.Query(ctx,
		`SELECT iv.id_item_venda, iv.venda_id, iv.produto_id, iv.sequencia,
		        iv.quantidade, iv.preco_unitario, iv.valor_total,
		        iv.valor_desconto, iv.valor_liquido, iv.status,
		        p.nome as produto_nome
		 FROM item_venda iv
		 JOIN produto p ON iv.produto_id = p.id_produto
		 WHERE iv.venda_id = $1
		 ORDER BY iv.sequencia`, vendaID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item models.ItemVenda
			rows.Scan(
				&item.ID, &item.VendaID, &item.ProdutoID, &item.Sequencia,
				&item.Quantidade, &item.PrecoUnitario, &item.ValorTotal,
				&item.ValorDesconto, &item.ValorLiquido, &item.Status,
				&item.ProdutoNome,
			)
			venda.Itens = append(venda.Itens, item)
		}
	}

	// Buscar pagamentos
	rows2, err := s.db.Pool.Query(ctx,
		`SELECT vp.id_venda_pagamento, vp.venda_id, vp.forma_pagamento_id,
		        vp.valor, vp.parcelas, vp.status, fp.nome as forma_pagamento_nome
		 FROM venda_pagamento vp
		 JOIN forma_pagamento fp ON vp.forma_pagamento_id = fp.id_forma_pagamento
		 WHERE vp.venda_id = $1`, vendaID)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var pag models.VendaPagamento
			rows2.Scan(
				&pag.ID, &pag.VendaID, &pag.FormaPagamentoID,
				&pag.Valor, &pag.Parcelas, &pag.Status, &pag.FormaPagamentoNome,
			)
			venda.Pagamentos = append(venda.Pagamentos, pag)
		}
	}

	return &venda, nil
}

func (s *VendaService) Cancelar(ctx context.Context, empresaID, vendaID int, req models.CancelarVendaRequest) error {
	// Validar senha do supervisor
	var senhaHash string
	err := s.db.Pool.QueryRow(ctx,
		`SELECT senha_hash FROM usuario
		 WHERE empresa_id = $1 AND perfil IN ('supervisor', 'gerente', 'admin') AND ativo = true
		 LIMIT 1`,
		empresaID,
	).Scan(&senhaHash)
	if err != nil {
		return fmt.Errorf("supervisor não encontrado")
	}

	if !utils.CheckPassword(senhaHash, req.SenhaSupervisor) {
		return fmt.Errorf("senha do supervisor inválida")
	}

	now := time.Now()
	_, err = s.db.Pool.Exec(ctx,
		`UPDATE venda SET status = 'cancelada', motivo_cancelamento = $1, data_cancelamento = $2
		 WHERE id_venda = $3 AND empresa_id = $4 AND status = 'concluida'`,
		req.Motivo, now, vendaID, empresaID,
	)
	if err != nil {
		return fmt.Errorf("erro ao cancelar venda: %w", err)
	}

	// Cancelar itens da venda
	_, _ = s.db.Pool.Exec(ctx,
		`UPDATE item_venda SET status = 'cancelado' WHERE venda_id = $1`,
		vendaID,
	)

	return nil
}

func (s *VendaService) VendasDia(ctx context.Context, empresaID int, data *string) ([]models.Venda, error) {
	query := `SELECT v.id_venda, v.numero_venda, v.data_venda,
	                 v.valor_total_produtos, v.valor_total_descontos,
	                 v.valor_total, v.valor_pago, v.valor_troco,
	                 v.status, u.nome as operador_nome, cf.nome as caixa_nome
	          FROM venda v
	          JOIN usuario u ON v.usuario_id = u.id_usuario
	          JOIN caixa_fisico cf ON v.caixa_fisico_id = cf.id_caixa_fisico
	          WHERE v.empresa_id = $1`

	var args []interface{}
	args = append(args, empresaID)

	if data != nil && *data != "" {
		query += ` AND DATE(v.data_venda) = $2::date`
		args = append(args, *data)
	} else {
		query += ` AND DATE(v.data_venda) = CURRENT_DATE`
	}
	query += ` ORDER BY v.data_venda DESC`

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar vendas: %w", err)
	}
	defer rows.Close()

	var vendas []models.Venda
	for rows.Next() {
		var v models.Venda
		err := rows.Scan(
			&v.ID, &v.NumeroVenda, &v.DataVenda,
			&v.ValorTotalProdutos, &v.ValorTotalDescontos,
			&v.ValorTotal, &v.ValorPago, &v.ValorTroco,
			&v.Status, &v.OperadorNome, &v.CaixaNome,
		)
		if err != nil {
			continue
		}
		vendas = append(vendas, v)
	}

	return vendas, nil
}
