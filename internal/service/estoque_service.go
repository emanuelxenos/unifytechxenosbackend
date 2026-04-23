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
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Buscar saldo atual e se controla estoque
	var saldoAtual float64
	var controla bool
	err = tx.QueryRow(ctx,
		`SELECT estoque_atual, controlar_estoque FROM produto WHERE id_produto = $1 AND empresa_id = $2`,
		req.ProdutoID, empresaID,
	).Scan(&saldoAtual, &controla)
	if err != nil {
		return fmt.Errorf("produto não encontrado")
	}

	if !controla {
		return fmt.Errorf("produto não controla estoque")
	}

	var novoSaldo float64
	var tipoMov string
	var loteID *int
	switch req.Tipo {
	case "entrada":
		novoSaldo = saldoAtual + req.Quantidade
		tipoMov = "entrada"

		// Criar novo lote para a entrada
		loteInterno := fmt.Sprintf("L-%s-%d", time.Now().Format("20060102"), req.ProdutoID)
		vencimento := time.Now().AddDate(1, 0, 0) // Default 1 ano se não informado
		
		if req.DataVencimento != nil && *req.DataVencimento != "" {
			parsedDate := time.Time{}
			var parseErr error
			// Tentar vários formatos
			layouts := []string{
				time.RFC3339,
				"2006-01-02T15:04:05.000Z",
				"2006-01-02T15:04:05.000",
				"2006-01-02T15:04:05",
				"2006-01-02",
			}
			
			for _, layout := range layouts {
				parsedDate, parseErr = time.Parse(layout, *req.DataVencimento)
				if parseErr == nil {
					break
				}
			}
			
			if parseErr != nil {
				return fmt.Errorf("formato de data de vencimento inválido: %s", *req.DataVencimento)
			}
			vencimento = parsedDate
		}

		err = tx.QueryRow(ctx,
			`INSERT INTO estoque_lote (empresa_id, produto_id, localizacao_id, lote_interno, 
			                           lote_fabricante, quantidade_inicial, quantidade_atual, 
			                           data_vencimento, usuario_id, observacao)
			 VALUES ($1, $2, $3, $4, $5, $6, $6, $7, $8, $9)
			 RETURNING id_lote`,
			empresaID, req.ProdutoID, req.LocalizacaoID, loteInterno,
			req.LoteFabricante, req.Quantidade, vencimento, usuarioID, req.Motivo,
		).Scan(&loteID)
		if err != nil {
			return fmt.Errorf("erro ao criar lote: %w", err)
		}

		// LOG DE ENTRADA ÚNICO
		_, err = tx.Exec(ctx,
			`INSERT INTO estoque_movimentacao (empresa_id, produto_id, tipo_movimentacao,
											   quantidade, saldo_anterior, saldo_atual,
											   origem_tipo, origem_id, usuario_id, observacao, lote_id)
			 VALUES ($1, $2, 'entrada', $3, $4, $5, 'manual', NULL, $6, $7, $8)`,
			empresaID, req.ProdutoID, req.Quantidade,
			saldoAtual, novoSaldo, usuarioID, req.Motivo, loteID,
		)
		if err != nil {
			return fmt.Errorf("erro ao registrar entrada no histórico: %w", err)
		}

	case "saida", "perda", "ajuste":
		// No caso de 'ajuste' genérico, decidimos se é entrada ou saída baseado na quantidade
		qtdSaida := req.Quantidade
		if req.Tipo == "ajuste" {
			if req.Quantidade > 0 {
				// Tratar ajuste positivo como entrada (recursivo simples)
				req.Tipo = "entrada"
				return s.Ajuste(ctx, empresaID, usuarioID, req)
			}
			qtdSaida = -req.Quantidade // Converter negativo para positivo para o motor de baixa
		} else {
			// Se o tipo já for 'saida' ou 'perda', garantimos que a quantidade seja positiva para a subtração
			if qtdSaida < 0 {
				qtdSaida = -qtdSaida
			}
		}

		novoSaldo = saldoAtual - qtdSaida
		tipoMov = req.Tipo
		
		log.Printf("DEBUG: Iniciando baixa por lotes. Produto: %d, QtdSaida: %.2f, SaldoAtual: %.2f, NovoSaldo: %.2f", req.ProdutoID, qtdSaida, saldoAtual, novoSaldo)

		// 1. O MÉTODO ABAIXO GERA OS LOGS DE SAÍDA (FEFO) E ATUALIZA OS LOTES
		err = s.baixarEstoquePorLotes(ctx, tx, empresaID, req.ProdutoID, qtdSaida, tipoMov, usuarioID, req.Motivo, 0, saldoAtual)
		if err != nil {
			return fmt.Errorf("erro ao baixar estoque por lotes: %w", err)
		}

		// 2. Atualizar saldo global do produto DEPOIS (Para garantir integridade do log)
		_, err = tx.Exec(ctx,
			`UPDATE produto SET estoque_atual = $1 WHERE id_produto = $2 AND empresa_id = $3`,
			novoSaldo, req.ProdutoID, empresaID,
		)
		if err != nil {
			log.Printf("❌ Erro ao atualizar saldo do produto: %v", err)
			return fmt.Errorf("erro ao atualizar saldo do produto: %w", err)
		}

	default:
		return fmt.Errorf("tipo de ajuste '%s' não suportado", req.Tipo)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("❌ Erro no commit da transação: %v", err)
		return fmt.Errorf("erro ao finalizar transação: %w", err)
	}

	log.Printf("✅ Ajuste de estoque concluído com sucesso!")
	return nil
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
		diferenca := qtdFisica - saldoAtual

		if diferenca < 0 {
			// 1. DIVERGÊNCIA NEGATIVA: Usar motor FEFO para baixar a diferença dos lotes
			qtdBaixar := -diferenca
			log.Printf("📉 Inventário: Baixando %.2f do produto %d via FEFO (Diferença negativa)", qtdBaixar, item.ProdutoID)
			
			// Atualizar saldo global PRIMEIRO
			_, err = tx.Exec(ctx, "UPDATE produto SET estoque_atual = $1 WHERE id_produto = $2 AND empresa_id = $3", qtdFisica, item.ProdutoID, empresaID)
			if err != nil {
				return fmt.Errorf("erro ao atualizar saldo global na baixa: %w", err)
			}

			err = s.baixarEstoquePorLotes(ctx, tx, empresaID, item.ProdutoID, qtdBaixar, "inventario", usuarioID, "Reconciliação de Inventário (Falta)", inventarioID, saldoAtual)
			if err != nil {
				return fmt.Errorf("erro ao reconciliar lotes por FEFO: %w", err)
			}
		} else if diferenca > 0 {
			// 2. DIVERGÊNCIA POSITIVA: Criar lote de ajuste para a sobra
			log.Printf("📈 Inventário: Criando lote de ajuste para sobra de %.2f do produto %d", diferenca, item.ProdutoID)
			
			loteInterno := fmt.Sprintf("INV-%d-%d", inventarioID, item.ProdutoID)
			vencimento := time.Now().AddDate(1, 0, 0) // Default 1 ano

			var loteID int
			err = tx.QueryRow(ctx,
				`INSERT INTO estoque_lote (empresa_id, produto_id, lote_interno, 
				                           quantidade_inicial, quantidade_atual, 
				                           data_vencimento, usuario_id, observacao, status)
				 VALUES ($1, $2, $3, $4, $4, $5, $6, $7, 'ativo')
				 RETURNING id_lote`,
				empresaID, item.ProdutoID, loteInterno, diferenca, vencimento, usuarioID, "Reconciliação de Inventário (Sobra)",
			).Scan(&loteID)
			if err != nil {
				return fmt.Errorf("erro ao criar lote de sobra: %w", err)
			}

			// Atualizar saldo global
			_, err = tx.Exec(ctx, "UPDATE produto SET estoque_atual = $1 WHERE id_produto = $2 AND empresa_id = $3", qtdFisica, item.ProdutoID, empresaID)
			if err != nil {
				return fmt.Errorf("erro ao atualizar saldo global na sobra: %w", err)
			}

			// Registrar log de movimentação
			_, err = tx.Exec(ctx,
				`INSERT INTO estoque_movimentacao (empresa_id, produto_id, tipo_movimentacao,
													quantidade, saldo_anterior, saldo_atual,
													origem_tipo, origem_id, usuario_id, observacao, lote_id)
				 VALUES ($1, $2, 'inventario', $3, $4, $5, 'inventario', $6, $7, $8, $9)`,
				empresaID, item.ProdutoID, diferenca, saldoAtual, qtdFisica,
				inventarioID, usuarioID, "Reconciliação de Inventário (Sobra)", loteID,
			)
			if err != nil {
				return fmt.Errorf("erro ao registrar log de sobra: %w", err)
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

func (s *EstoqueService) ListarMovimentacoes(ctx context.Context, empresaID int, produtoID int, dataInicio, dataFim, tipo string) ([]models.EstoqueMovimentacao, error) {
	query := `
		SELECT m.id_movimentacao, m.empresa_id, m.produto_id, m.tipo_movimentacao, 
		       m.quantidade, m.saldo_anterior, m.saldo_atual, m.origem_tipo, 
		       m.origem_id, m.data_movimentacao, m.usuario_id, m.observacao, 
			   p.nome as produto_nome, l.lote_fabricante
		FROM estoque_movimentacao m
		JOIN produto p ON m.produto_id = p.id_produto
		LEFT JOIN estoque_lote l ON m.lote_id = l.id_lote
		WHERE m.empresa_id = $1
	`
	args := []interface{}{empresaID}
	placeholderID := 2

	if produtoID > 0 {
		query += fmt.Sprintf(" AND m.produto_id = $%d", placeholderID)
		args = append(args, produtoID)
		placeholderID++
	}

	if tipo != "" {
		query += fmt.Sprintf(" AND m.tipo_movimentacao = $%d", placeholderID)
		args = append(args, tipo)
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
			&m.OrigemID, &m.DataMovimentacao, &m.UsuarioID, &m.Observacao, 
			&m.ProdutoNome, &m.LoteFabricante,
		)
		if err != nil {
			log.Printf("⚠️ Erro ao escanear movimentação: %v", err)
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

func (s *EstoqueService) ListarLotesPorProduto(ctx context.Context, empresaID, produtoID int) ([]models.EstoqueLote, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT el.id_lote, el.empresa_id, el.produto_id, el.localizacao_id, el.lote_interno, 
		        el.lote_fabricante, el.quantidade_inicial, el.quantidade_atual, el.data_fabricacao, 
		        el.data_vencimento, el.data_recebimento, el.status, el.observacao, el.usuario_id,
		        loc.nome as localizacao_nome
		 FROM estoque_lote el
		 LEFT JOIN estoque_localizacao loc ON el.localizacao_id = loc.id_localizacao
		 WHERE el.empresa_id = $1 AND el.produto_id = $2 AND el.status != 'esgotado'
		 ORDER BY el.data_vencimento ASC`,
		empresaID, produtoID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lotes []models.EstoqueLote
	for rows.Next() {
		var l models.EstoqueLote
		err := rows.Scan(
			&l.ID, &l.EmpresaID, &l.ProdutoID, &l.LocalizacaoID, &l.LoteInterno,
			&l.LoteFabricante, &l.QtdInicial, &l.QtdAtual, &l.DataFabricacao,
			&l.DataVencimento, &l.DataReceb, &l.Status, &l.Observacao, &l.UsuarioID,
			&l.LocalizacaoNome,
		)
		if err != nil {
			log.Printf("Erro ao escanear lote: %v", err)
			continue
		}
		lotes = append(lotes, l)
	}
	return lotes, nil
}

func (s *EstoqueService) ListarLocalizacoes(ctx context.Context, empresaID int) ([]models.EstoqueLocalizacao, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT id_localizacao, empresa_id, codigo, nome, descricao, ativo, data_cadastro
		 FROM estoque_localizacao WHERE empresa_id = $1 AND ativo = true
		 ORDER BY nome ASC`,
		empresaID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locs []models.EstoqueLocalizacao
	for rows.Next() {
		var l models.EstoqueLocalizacao
		err := rows.Scan(&l.ID, &l.EmpresaID, &l.Codigo, &l.Nome, &l.Descricao, &l.Ativo, &l.DataCad)
		if err != nil {
			continue
		}
		locs = append(locs, l)
	}
	return locs, nil
}

func (s *EstoqueService) CriarLocalizacao(ctx context.Context, empresaID int, req models.EstoqueLocalizacao) error {
	_, err := s.db.Pool.Exec(ctx,
		`INSERT INTO estoque_localizacao (empresa_id, codigo, nome, descricao)
		 VALUES ($1, $2, $3, $4)`,
		empresaID, req.Codigo, req.Nome, req.Descricao,
	)
	return err
}

func (s *EstoqueService) BaixarEstoquePorLotes(ctx context.Context, tx database.Tx, empresaID, produtoID int, qtd float64, origemTipo string, origemID, usuarioID int, observacao string, saldoGlobalAtual float64) error {
	return s.baixarEstoquePorLotes(ctx, tx, empresaID, produtoID, qtd, origemTipo, usuarioID, observacao, origemID, saldoGlobalAtual)
}

func (s *EstoqueService) baixarEstoquePorLotes(ctx context.Context, tx database.Tx, empresaID, produtoID int, qtd float64, origemTipo string, usuarioID int, observacao string, origemID int, saldoGlobalAtual float64) error {
	// Buscar lotes ativos ordenados por vencimento (FEFO)
	rows, err := tx.Query(ctx,
		`SELECT id_lote, quantidade_atual, lote_interno 
		 FROM estoque_lote 
		 WHERE empresa_id = $1 AND produto_id = $2 AND status = 'ativo' AND quantidade_atual > 0
		 ORDER BY data_vencimento ASC, data_recebimento ASC`,
		empresaID, produtoID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	type loteItem struct {
		ID   int
		Qtd  float64
		Nome string
	}

	var lotes []loteItem
	for rows.Next() {
		var l loteItem
		if err := rows.Scan(&l.ID, &l.Qtd, &l.Nome); err != nil {
			return err
		}
		lotes = append(lotes, l)
	}

	restante := qtd
	for _, lote := range lotes {
		if restante <= 0 {
			break
		}

		consumir := lote.Qtd
		if consumir > restante {
			consumir = restante
		}

		// Atualizar o lote
		_, err = tx.Exec(ctx,
			`UPDATE estoque_lote SET quantidade_atual = quantidade_atual - $1,
			 status = CASE WHEN (quantidade_atual - $1) <= 0 THEN 'esgotado' ELSE 'ativo' END
			 WHERE id_lote = $2`,
			consumir, lote.ID,
		)
		if err != nil {
			return err
		}

		// Registrar na movimentação para rastreabilidade
		// O saldo anterior é o saldo global que recebemos no início ou o saldo após o consumo do lote anterior
		saldoAnteriorLoop := saldoGlobalAtual
		saldoAtualLoop := saldoGlobalAtual - consumir
		
		obsFinal := fmt.Sprintf("%s - Consumo Lote %s", observacao, lote.Nome)
		_, err = tx.Exec(ctx,
			`INSERT INTO estoque_movimentacao (empresa_id, produto_id, tipo_movimentacao,
											   quantidade, saldo_anterior, saldo_atual,
											   origem_tipo, origem_id, usuario_id, observacao, lote_id)
			 VALUES ($1, $2, 'saida', $3, $4, $5, $6, $7, $8, $9, $10)`,
			empresaID, produtoID, consumir, saldoAnteriorLoop, saldoAtualLoop,
			origemTipo, origemID, usuarioID, obsFinal, lote.ID,
		)
		if err != nil {
			return err
		}

		restante -= consumir
		saldoGlobalAtual = saldoAtualLoop // Atualizar para o próximo lote no loop
	}

	if restante > 0 {
		return fmt.Errorf("estoque insuficiente para completar a baixa por lotes (faltam %.2f)", restante)
	}

	return nil
}
