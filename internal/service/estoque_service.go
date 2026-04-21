package service

import (
	"context"
	"fmt"
	"log"
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
	var dataInicio time.Time
	var err error

	// Tentar vários formatos de data comuns
	layouts := []string{time.RFC3339, "2006-01-02T15:04:05.000Z", "2006-01-02T15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		dataInicio, err = time.Parse(layout, req.DataInicio)
		if err == nil {
			break
		}
	}

	if err != nil {
		// Se falhar tudo, usa agora
		dataInicio = time.Now()
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

	// Criar itens do inventário (filtrado por categoria se fornecido)
	query := `
		INSERT INTO inventario_item (inventario_id, produto_id, quantidade_sistema)
		SELECT $1, id_produto, estoque_atual
		FROM produto
		WHERE empresa_id = $2 AND ativo = true AND controlar_estoque = true
	`
	args := []interface{}{inv.ID, empresaID}
	if req.CategoriaID != nil && *req.CategoriaID > 0 {
		query += " AND categoria_id = $3"
		args = append(args, *req.CategoriaID)
	}

	_, err = s.db.Pool.Exec(ctx, query, args...)

	return &inv, nil
}

func (s *EstoqueService) FinalizarInventario(ctx context.Context, empresaID, inventarioID, usuarioID int, req models.FinalizarInventarioRequest) error {
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	log.Printf("🔹 Finalizando inventário ID %d para empresa %d (Usuário: %d)", inventarioID, empresaID, usuarioID)

	// Buscar itens contados diretamente no banco para reconciliação
	rows, err := tx.Query(ctx,
		`SELECT produto_id, quantidade_fisica, quantidade_sistema FROM inventario_item 
		 WHERE inventario_id = $1 AND contado = true`,
		inventarioID,
	)
	if err != nil {
		log.Printf("❌ Erro ao buscar itens do inventário: %v", err)
		return fmt.Errorf("erro ao buscar itens para finalização: %w", err)
	}
	defer rows.Close()

	type itemContado struct {
		ProdutoID        int
		QuantidadeFisica *float64
		QuantidadeSistema float64
	}

	var itens []itemContado
	for rows.Next() {
		var it itemContado
		if err := rows.Scan(&it.ProdutoID, &it.QuantidadeFisica, &it.QuantidadeSistema); err != nil {
			log.Printf("⚠️ Erro ao escanear item do inventário: %v", err)
			return fmt.Errorf("erro ao processar dados de um item: %w", err)
		}
		itens = append(itens, it)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("erro durante a leitura dos itens: %w", err)
	}

	log.Printf("📦 Processando reconciliação para %d itens", len(itens))

	for _, item := range itens {
		if item.QuantidadeFisica == nil {
			log.Printf("⏭️ Pulando produto %d: Quantidade física nula", item.ProdutoID)
			continue
		}

		// Buscar saldo atual real (pode ter mudado desde o início do inventário)
		var saldoAtual float64
		err = tx.QueryRow(ctx,
			`SELECT estoque_atual FROM produto WHERE id_produto = $1`,
			item.ProdutoID,
		).Scan(&saldoAtual)
		if err != nil {
			log.Printf("⚠️ Erro ao buscar saldo atual do produto %d: %v", item.ProdutoID, err)
			continue
		}

		qtdFisica := *item.QuantidadeFisica

		// Atualizar estoque do produto para a quantidade física apurada
		_, err = tx.Exec(ctx,
			`UPDATE produto SET estoque_atual = $1 WHERE id_produto = $2`,
			qtdFisica, item.ProdutoID,
		)
		if err != nil {
			log.Printf("❌ Erro ao atualizar estoque do produto %d: %v", item.ProdutoID, err)
			return fmt.Errorf("erro ao atualizar estoque real do produto %d: %w", item.ProdutoID, err)
		}

		// Registrar movimentação de ajuste
		diferenca := qtdFisica - saldoAtual
		if diferenca != 0 {
			_, err = tx.Exec(ctx,
				`INSERT INTO estoque_movimentacao (empresa_id, produto_id, tipo_movimentacao,
													quantidade, saldo_anterior, saldo_atual,
													origem_tipo, origem_id, usuario_id, observacao)
				 VALUES ($1, $2, 'inventario', $3, $4, $5, 'inventario', $6, $7, $8)`,
				empresaID, item.ProdutoID, diferenca, saldoAtual, qtdFisica,
				inventarioID, usuarioID, "Reconciliação automática de inventário",
			)
			if err != nil {
				log.Printf("❌ Erro ao registrar movimentação para produto %d: %v", item.ProdutoID, err)
				return fmt.Errorf("erro ao registrar movimentação do produto %d: %w", item.ProdutoID, err)
			}
		}
	}

	// Fechar inventário e salvar observações finais
	_, err = tx.Exec(ctx,
		`UPDATE inventario 
		 SET status = 'fechado', 
		     data_fechamento = CURRENT_TIMESTAMP, 
		     data_fim = CURRENT_DATE,
			 observacoes = $1
		 WHERE id_inventario = $2`,
		req.Observacoes, inventarioID,
	)
	if err != nil {
		log.Printf("❌ Erro ao fechar inventário ID %d: %v", inventarioID, err)
		return fmt.Errorf("erro ao fechar inventário: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("❌ Erro no commit da finalização: %v", err)
		return fmt.Errorf("erro ao salvar reconciliação: %w", err)
	}

	log.Printf("✅ Inventário %d finalizado com sucesso!", inventarioID)
	return nil
}

func (s *EstoqueService) BuscarInventarioPorId(ctx context.Context, empresaID, inventarioID int) (*models.Inventario, error) {
	var i models.Inventario
	err := s.db.Pool.QueryRow(ctx,
		`SELECT id_inventario, empresa_id, codigo, descricao, data_inicio, 
		        data_fim, data_fechamento, status, observacoes, usuario_id
		 FROM inventario
		 WHERE id_inventario = $1 AND empresa_id = $2`,
		inventarioID, empresaID,
	).Scan(&i.ID, &i.EmpresaID, &i.Codigo, &i.Descricao, &i.DataInicio,
		&i.DataFim, &i.DataFechamento, &i.Status, &i.Observacoes, &i.UsuarioID)
	if err != nil {
		return nil, fmt.Errorf("inventário não encontrado")
	}

	// Buscar itens
	rows, err := s.db.Pool.Query(ctx,
		`SELECT ii.id_inventario_item, ii.inventario_id, ii.produto_id, 
		        ii.quantidade_sistema, ii.quantidade_fisica, ii.contado, 
		        ii.data_contagem, ii.usuario_contagem_id, ii.observacao, p.nome as produto_nome
		 FROM inventario_item ii
		 JOIN produto p ON ii.produto_id = p.id_produto
		 WHERE ii.inventario_id = $1`,
		inventarioID,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var it models.InventarioItem
			err := rows.Scan(
				&it.ID, &it.InventarioID, &it.ProdutoID,
				&it.QuantidadeSistema, &it.QuantidadeFisica, &it.Contado,
				&it.DataContagem, &it.UsuarioContagemID, &it.Observacao, &it.ProdutoNome,
			)
			if err != nil {
				continue
			}
			i.Itens = append(i.Itens, it)
		}
	}

	return &i, nil
}

func (s *EstoqueService) AtualizarItemInventario(ctx context.Context, empresaID, inventarioID, produtoID int, quantidade float64, usuarioID int) error {
	// Verificar se inventário pertence à empresa e está aberto
	var status string
	err := s.db.Pool.QueryRow(ctx,
		`SELECT status FROM inventario WHERE id_inventario = $1 AND empresa_id = $2`,
		inventarioID, empresaID,
	).Scan(&status)
	if err != nil {
		return fmt.Errorf("inventário não encontrado")
	}
	if status != "aberto" {
		return fmt.Errorf("não é possível atualizar itens de um inventário fechado")
	}

	_, err = s.db.Pool.Exec(ctx,
		`UPDATE inventario_item
		 SET quantidade_fisica = $1, contado = true, data_contagem = CURRENT_TIMESTAMP, usuario_contagem_id = $2
		 WHERE inventario_id = $3 AND produto_id = $4`,
		quantidade, usuarioID, inventarioID, produtoID,
	)
	return err
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

func (s *EstoqueService) ListarInventarios(ctx context.Context, empresaID int, dataInicio, dataFim string) ([]models.Inventario, error) {
	query := `
		SELECT id_inventario, empresa_id, codigo, descricao, data_inicio, 
		        data_fim, data_fechamento, status, observacoes, usuario_id
		 FROM inventario
		 WHERE empresa_id = $1
	`
	args := []interface{}{empresaID}
	placeholderID := 2

	if dataInicio != "" {
		query += fmt.Sprintf(" AND data_inicio >= $%d", placeholderID)
		args = append(args, dataInicio)
		placeholderID++
	}

	if dataFim != "" {
		query += fmt.Sprintf(" AND data_inicio <= $%d", placeholderID)
		args = append(args, dataFim+" 23:59:59")
		placeholderID++
	}

	query += " ORDER BY data_inicio DESC, id_inventario DESC LIMIT 50"

	rows, err := s.db.Pool.Query(ctx, query, args...)
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
