package service

import (
	"context"
	"fmt"
	"time"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
)

type EstoqueService struct {
	db *database.PostgresDB
}

func NewEstoqueService(db *database.PostgresDB) *EstoqueService {
	return &EstoqueService{db: db}
}

func (s *EstoqueService) EstoqueBaixo(ctx context.Context, empresaID int) ([]models.EstoqueBaixoResponse, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT id_produto, nome, estoque_atual, estoque_minimo
		 FROM produto
		 WHERE empresa_id = $1
		   AND controlar_estoque = true
		   AND estoque_atual <= estoque_minimo
		   AND ativo = true
		 ORDER BY estoque_atual ASC`,
		empresaID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar estoque baixo: %w", err)
	}
	defer rows.Close()

	var produtos []models.EstoqueBaixoResponse
	for rows.Next() {
		var p models.EstoqueBaixoResponse
		err := rows.Scan(&p.IDProduto, &p.Nome, &p.EstoqueAtual, &p.EstoqueMinimo)
		if err != nil {
			continue
		}
		produtos = append(produtos, p)
	}
	return produtos, nil
}

func (s *EstoqueService) Ajuste(ctx context.Context, empresaID, usuarioID int, req models.AjusteEstoqueRequest) error {
	// Buscar saldo atual
	var saldoAtual float64
	err := s.db.Pool.QueryRow(ctx,
		`SELECT estoque_atual FROM produto WHERE id_produto = $1 AND empresa_id = $2`,
		req.ProdutoID, empresaID,
	).Scan(&saldoAtual)
	if err != nil {
		return fmt.Errorf("produto não encontrado")
	}

	var novoSaldo float64
	var tipoMov string
	switch req.Tipo {
	case "entrada":
		novoSaldo = saldoAtual + req.Quantidade
		tipoMov = "entrada"
	case "saida":
		novoSaldo = saldoAtual - req.Quantidade
		tipoMov = "saida"
	default:
		novoSaldo = req.Quantidade
		tipoMov = "ajuste"
	}

	// Atualizar estoque
	_, err = s.db.Pool.Exec(ctx,
		`UPDATE produto SET estoque_atual = $1 WHERE id_produto = $2`,
		novoSaldo, req.ProdutoID,
	)
	if err != nil {
		return fmt.Errorf("erro ao ajustar estoque: %w", err)
	}

	// Registrar movimentação
	_, err = s.db.Pool.Exec(ctx,
		`INSERT INTO estoque_movimentacao (empresa_id, produto_id, tipo_movimentacao,
		                                    quantidade, saldo_anterior, saldo_atual,
		                                    origem_tipo, usuario_id, observacao)
		 VALUES ($1, $2, $3, $4, $5, $6, 'manual', $7, $8)`,
		empresaID, req.ProdutoID, tipoMov, req.Quantidade,
		saldoAtual, novoSaldo, usuarioID, req.Motivo,
	)
	return err
}

func (s *EstoqueService) CriarInventario(ctx context.Context, empresaID, usuarioID int, req models.CriarInventarioRequest) (*models.Inventario, error) {
	dataInicio, err := time.Parse("2006-01-02", req.DataInicio)
	if err != nil {
		return nil, fmt.Errorf("data de início inválida")
	}

	var inv models.Inventario
	err = s.db.Pool.QueryRow(ctx,
		`INSERT INTO inventario (empresa_id, codigo, descricao, data_inicio, usuario_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id_inventario, status`,
		empresaID, req.Codigo, req.Descricao, dataInicio, usuarioID,
	).Scan(&inv.ID, &inv.Status)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar inventário: %w", err)
	}

	inv.EmpresaID = empresaID
	inv.Codigo = req.Codigo
	inv.DataInicio = dataInicio

	// Criar itens do inventário com todos os produtos ativos
	_, err = s.db.Pool.Exec(ctx,
		`INSERT INTO inventario_item (inventario_id, produto_id, quantidade_sistema)
		 SELECT $1, id_produto, estoque_atual
		 FROM produto
		 WHERE empresa_id = $2 AND ativo = true AND controlar_estoque = true`,
		inv.ID, empresaID,
	)

	return &inv, nil
}

func (s *EstoqueService) FinalizarInventario(ctx context.Context, empresaID, inventarioID, usuarioID int, req models.FinalizarInventarioRequest) error {
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, ajuste := range req.Ajustes {
		// Atualizar item do inventário
		_, err = tx.Exec(ctx,
			`UPDATE inventario_item
			 SET quantidade_fisica = $1, contado = true, data_contagem = CURRENT_TIMESTAMP, usuario_contagem_id = $2
			 WHERE inventario_id = $3 AND produto_id = $4`,
			ajuste.QuantidadeFisica, usuarioID, inventarioID, ajuste.ProdutoID,
		)
		if err != nil {
			return fmt.Errorf("erro ao atualizar item inventário: %w", err)
		}

		// Buscar saldo atual
		var saldoAtual float64
		tx.QueryRow(ctx,
			`SELECT estoque_atual FROM produto WHERE id_produto = $1`,
			ajuste.ProdutoID,
		).Scan(&saldoAtual)

		// Atualizar estoque do produto
		_, err = tx.Exec(ctx,
			`UPDATE produto SET estoque_atual = $1 WHERE id_produto = $2`,
			ajuste.QuantidadeFisica, ajuste.ProdutoID,
		)
		if err != nil {
			return fmt.Errorf("erro ao atualizar estoque: %w", err)
		}

		// Registrar movimentação
		_, _ = tx.Exec(ctx,
			`INSERT INTO estoque_movimentacao (empresa_id, produto_id, tipo_movimentacao,
			                                    quantidade, saldo_anterior, saldo_atual,
			                                    origem_tipo, origem_id, usuario_id, observacao)
			 VALUES ($1, $2, 'inventario', $3, $4, $5, 'inventario', $6, $7, 'Ajuste de inventário')`,
			empresaID, ajuste.ProdutoID,
			ajuste.QuantidadeFisica-saldoAtual, saldoAtual, ajuste.QuantidadeFisica,
			inventarioID, usuarioID,
		)
	}

	// Fechar inventário
	_, err = tx.Exec(ctx,
		`UPDATE inventario SET status = 'fechado', data_fechamento = CURRENT_TIMESTAMP, data_fim = CURRENT_DATE
		 WHERE id_inventario = $1`,
		inventarioID,
	)
	if err != nil {
		return fmt.Errorf("erro ao fechar inventário: %w", err)
	}

	return tx.Commit(ctx)
}

func (s *EstoqueService) ListarMovimentacoes(ctx context.Context, empresaID int, produtoID int, dataInicio, dataFim string) ([]models.EstoqueMovimentacao, error) {
	query := `
		SELECT m.id_movimentacao, m.empresa_id, m.produto_id, m.tipo_movimentacao, 
		       m.quantidade, m.saldo_anterior, m.saldo_atual, m.origem_tipo, 
		       m.origem_id, m.data_movimentacao, m.usuario_id, m.observacao, p.nome as produto_nome
		FROM estoque_movimentacao m
		JOIN produto p ON m.produto_id = p.id_produto
		WHERE m.empresa_id = $1
	`
	args := []interface{}{empresaID}
	placeholderID := 2

	if produtoID > 0 {
		query += fmt.Sprintf(" AND m.produto_id = $%d", placeholderID)
		args = append(args, produtoID)
		placeholderID++
	}

	if dataInicio != "" {
		query += fmt.Sprintf(" AND m.data_movimentacao >= $%d", placeholderID)
		args = append(args, dataInicio)
		placeholderID++
	}

	if dataFim != "" {
		query += fmt.Sprintf(" AND m.data_movimentacao <= $%d", placeholderID)
		args = append(args, dataFim+" 23:59:59")
		placeholderID++
	}

	query += " ORDER BY m.data_movimentacao DESC LIMIT 100"

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar movimentações: %w", err)
	}
	defer rows.Close()

	var movs []models.EstoqueMovimentacao
	for rows.Next() {
		var m models.EstoqueMovimentacao
		err := rows.Scan(
			&m.ID, &m.EmpresaID, &m.ProdutoID, &m.TipoMovimentacao,
			&m.Quantidade, &m.SaldoAnterior, &m.SaldoAtual, &m.OrigemTipo,
			&m.OrigemID, &m.DataMovimentacao, &m.UsuarioID, &m.Observacao, &m.ProdutoNome,
		)
		if err != nil {
			continue
		}
		movs = append(movs, m)
	}
	return movs, nil
}

func (s *EstoqueService) ListarInventarios(ctx context.Context, empresaID int) ([]models.Inventario, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT id_inventario, empresa_id, codigo, descricao, data_inicio, 
		        data_fim, data_fechamento, status, observacoes, usuario_id
		 FROM inventario
		 WHERE empresa_id = $1
		 ORDER BY data_inicio DESC`,
		empresaID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar inventários: %w", err)
	}
	defer rows.Close()

	var invs []models.Inventario
	for rows.Next() {
		var i models.Inventario
		err := rows.Scan(
			&i.ID, &i.EmpresaID, &i.Codigo, &i.Descricao, &i.DataInicio,
			&i.DataFim, &i.DataFechamento, &i.Status, &i.Observacoes, &i.UsuarioID,
		)
		if err != nil {
			continue
		}
		invs = append(invs, i)
	}
	return invs, nil
}
