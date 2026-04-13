package service

import (
	"context"
	"fmt"
	"time"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
)

type CompraService struct {
	db *database.PostgresDB
}

func NewCompraService(db *database.PostgresDB) *CompraService {
	return &CompraService{db: db}
}

func (s *CompraService) Criar(ctx context.Context, empresaID, usuarioID int, req models.CriarCompraRequest) (*models.Compra, error) {
	dataEmissao, _ := time.Parse("2006-01-02", req.DataEmissao)

	var valorTotal float64
	for _, item := range req.Itens {
		valorTotal += item.Quantidade * item.PrecoUnitario
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	var compra models.Compra
	err = tx.QueryRow(ctx,
		`INSERT INTO compra (empresa_id, fornecedor_id, usuario_id, numero_nota_fiscal,
		                     data_emissao, data_entrada, valor_produtos, valor_total, status)
		 VALUES ($1, $2, $3, $4, $5, CURRENT_DATE, $6, $7, 'pendente')
		 RETURNING id_compra, data_cadastro, status`,
		empresaID, req.FornecedorID, usuarioID, req.NumeroNotaFiscal,
		dataEmissao, valorTotal, valorTotal,
	).Scan(&compra.ID, &compra.DataCadastro, &compra.Status)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar compra: %w", err)
	}

	for i, item := range req.Itens {
		vt := item.Quantidade * item.PrecoUnitario
		_, err = tx.Exec(ctx,
			`INSERT INTO item_compra (compra_id, produto_id, sequencia, quantidade, preco_unitario, valor_total)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			compra.ID, item.ProdutoID, i+1, item.Quantidade, item.PrecoUnitario, vt,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao inserir item: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao confirmar: %w", err)
	}

	compra.EmpresaID = empresaID
	compra.ValorTotal = valorTotal
	return &compra, nil
}

func (s *CompraService) Receber(ctx context.Context, empresaID, compraID, usuarioID int, req models.ReceberCompraRequest) error {
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, item := range req.ItensRecebidos {
		_, err = tx.Exec(ctx,
			`UPDATE item_compra SET quantidade_recebida = $1, data_recebimento = CURRENT_TIMESTAMP
			 WHERE compra_id = $2 AND produto_id = $3`,
			item.QuantidadeRecebida, compraID, item.ProdutoID,
		)
		if err != nil {
			return fmt.Errorf("erro ao atualizar item: %w", err)
		}

		// Atualizar estoque
		var saldoAtual float64
		tx.QueryRow(ctx, `SELECT estoque_atual FROM produto WHERE id_produto = $1`, item.ProdutoID).Scan(&saldoAtual)

		novoSaldo := saldoAtual + item.QuantidadeRecebida
		_, _ = tx.Exec(ctx,
			`UPDATE produto SET estoque_atual = $1, data_ultima_compra = CURRENT_DATE WHERE id_produto = $2`,
			novoSaldo, item.ProdutoID,
		)

		_, _ = tx.Exec(ctx,
			`INSERT INTO estoque_movimentacao (empresa_id, produto_id, tipo_movimentacao,
			 quantidade, saldo_anterior, saldo_atual, origem_tipo, origem_id, usuario_id)
			 VALUES ($1, $2, 'entrada', $3, $4, $5, 'compra', $6, $7)`,
			empresaID, item.ProdutoID, item.QuantidadeRecebida,
			saldoAtual, novoSaldo, compraID, usuarioID,
		)
	}

	_, _ = tx.Exec(ctx,
		`UPDATE compra SET status = 'recebida' WHERE id_compra = $1`, compraID)

	return tx.Commit(ctx)
}

func (s *CompraService) Listar(ctx context.Context, empresaID int, fornecedorID int) ([]models.Compra, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT c.id_compra, c.empresa_id, c.fornecedor_id, c.usuario_id, c.numero_nota_fiscal,
		        c.data_emissao, c.data_entrada, c.data_cadastro, c.valor_produtos, c.valor_total, c.status,
		        f.razao_social as fornecedor_nome
		 FROM compra c
		 LEFT JOIN fornecedor f ON c.fornecedor_id = f.id_fornecedor
		 WHERE c.empresa_id = $1 AND (c.fornecedor_id = $2 OR $2 = 0)
		 ORDER BY c.data_entrada DESC`, empresaID, fornecedorID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar compras: %w", err)
	}
	defer rows.Close()

	var compras []models.Compra
	for rows.Next() {
		var c models.Compra
		err := rows.Scan(
			&c.ID, &c.EmpresaID, &c.FornecedorID, &c.UsuarioID, &c.NumeroNotaFiscal,
			&c.DataEmissao, &c.DataEntrada, &c.DataCadastro, &c.ValorProdutos, &c.ValorTotal, &c.Status,
			&c.FornecedorNome,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler linha de compra: %w", err)
		}
		compras = append(compras, c)
	}

	return compras, nil
}
