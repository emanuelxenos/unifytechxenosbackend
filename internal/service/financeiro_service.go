package service

import (
	"context"
	"fmt"
	"time"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
)

type FinanceiroService struct {
	db *database.PostgresDB
}

func NewFinanceiroService(db *database.PostgresDB) *FinanceiroService {
	return &FinanceiroService{db: db}
}

func (s *FinanceiroService) ListarContasPagar(ctx context.Context, empresaID int, status, vencInicio, vencFim *string) ([]models.ContaPagar, error) {
	query := `SELECT cp.id_conta_pagar, cp.empresa_id, cp.fornecedor_id, cp.descricao,
	                 cp.valor_original, cp.valor_pago, cp.data_vencimento, cp.data_pagamento,
	                 cp.status, cp.categoria, cp.data_cadastro,
	                 f.razao_social as fornecedor_nome
	          FROM conta_pagar cp
	          LEFT JOIN fornecedor f ON cp.fornecedor_id = f.id_fornecedor
	          WHERE cp.empresa_id = $1`
	args := []interface{}{empresaID}
	idx := 2

	if status != nil && *status != "" {
		query += fmt.Sprintf(` AND cp.status = $%d`, idx)
		args = append(args, *status)
		idx++
	}
	if vencInicio != nil && *vencInicio != "" {
		query += fmt.Sprintf(` AND cp.data_vencimento >= $%d::date`, idx)
		args = append(args, *vencInicio)
		idx++
	}
	if vencFim != nil && *vencFim != "" {
		query += fmt.Sprintf(` AND cp.data_vencimento <= $%d::date`, idx)
		args = append(args, *vencFim)
		idx++
	}
	query += ` ORDER BY cp.data_vencimento`

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar contas: %w", err)
	}
	defer rows.Close()

	var contas []models.ContaPagar
	for rows.Next() {
		var c models.ContaPagar
		rows.Scan(&c.ID, &c.EmpresaID, &c.FornecedorID, &c.Descricao,
			&c.ValorOriginal, &c.ValorPago, &c.DataVencimento, &c.DataPagamento,
			&c.Status, &c.Categoria, &c.DataCadastro, &c.FornecedorNome)
		contas = append(contas, c)
	}
	return contas, nil
}

func (s *FinanceiroService) CriarContaPagar(ctx context.Context, empresaID, usuarioID int, req models.CriarContaPagarRequest) (*models.ContaPagar, error) {
	dataVenc, _ := time.Parse("2006-01-02", req.DataVencimento)
	cat := req.Categoria
	if cat == "" {
		cat = "geral"
	}

	var c models.ContaPagar
	err := s.db.Pool.QueryRow(ctx,
		`INSERT INTO conta_pagar (empresa_id, fornecedor_id, descricao, valor_original, data_vencimento, categoria, usuario_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id_conta_pagar, status, data_cadastro`,
		empresaID, req.FornecedorID, req.Descricao, req.ValorOriginal, dataVenc, cat, usuarioID,
	).Scan(&c.ID, &c.Status, &c.DataCadastro)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar conta: %w", err)
	}
	c.EmpresaID = empresaID
	c.Descricao = req.Descricao
	c.ValorOriginal = req.ValorOriginal
	c.DataVencimento = dataVenc
	return &c, nil
}

func (s *FinanceiroService) PagarConta(ctx context.Context, empresaID, contaID int, req models.PagarContaRequest) error {
	dataPag, _ := time.Parse("2006-01-02", req.DataPagamento)
	_, err := s.db.Pool.Exec(ctx,
		`UPDATE conta_pagar SET valor_pago = $1, data_pagamento = $2, status = 'paga'
		 WHERE id_conta_pagar = $3 AND empresa_id = $4`,
		req.ValorPago, dataPag, contaID, empresaID,
	)
	return err
}

func (s *FinanceiroService) ListarContasReceber(ctx context.Context, empresaID int) ([]models.ContaReceber, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT cr.id_conta_receber, cr.empresa_id, cr.cliente_id, cr.venda_id,
		        cr.descricao, cr.valor_original, cr.valor_recebido,
		        cr.data_vencimento, cr.data_recebimento, cr.status, cr.data_cadastro,
		        c.nome as cliente_nome
		 FROM conta_receber cr
		 LEFT JOIN cliente c ON cr.cliente_id = c.id_cliente
		 WHERE cr.empresa_id = $1
		 ORDER BY cr.data_vencimento`, empresaID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar contas: %w", err)
	}
	defer rows.Close()

	var contas []models.ContaReceber
	for rows.Next() {
		var c models.ContaReceber
		rows.Scan(&c.ID, &c.EmpresaID, &c.ClienteID, &c.VendaID,
			&c.Descricao, &c.ValorOriginal, &c.ValorRecebido,
			&c.DataVencimento, &c.DataRecebimento, &c.Status, &c.DataCadastro,
			&c.ClienteNome)
		contas = append(contas, c)
	}
	return contas, nil
}

func (s *FinanceiroService) ReceberConta(ctx context.Context, empresaID, contaID int, req models.ReceberContaRequest) error {
	dataRec, _ := time.Parse("2006-01-02", req.DataRecebimento)
	_, err := s.db.Pool.Exec(ctx,
		`UPDATE conta_receber SET valor_recebido = $1, data_recebimento = $2, status = 'recebida'
		 WHERE id_conta_receber = $3 AND empresa_id = $4`,
		req.ValorRecebido, dataRec, contaID, empresaID,
	)
	return err
}

func (s *FinanceiroService) FluxoCaixa(ctx context.Context, empresaID int, dataInicio, dataFim *string) ([]models.FluxoCaixaItem, error) {
	query := `SELECT data, tipo, valor FROM vw_fluxo_caixa WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if dataInicio != nil && *dataInicio != "" {
		query += fmt.Sprintf(` AND data >= $%d::date`, idx)
		args = append(args, *dataInicio)
		idx++
	}
	if dataFim != nil && *dataFim != "" {
		query += fmt.Sprintf(` AND data <= $%d::date`, idx)
		args = append(args, *dataFim)
		idx++
	}
	query += ` ORDER BY data`

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar fluxo: %w", err)
	}
	defer rows.Close()

	var items []models.FluxoCaixaItem
	for rows.Next() {
		var item models.FluxoCaixaItem
		rows.Scan(&item.Data, &item.Tipo, &item.Valor)
		items = append(items, item)
	}
	return items, nil
}
