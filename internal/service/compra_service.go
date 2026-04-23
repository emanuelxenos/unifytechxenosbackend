package service

import (
	"context"
	"fmt"
	"log"
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
		
		var vcto *time.Time
		if item.DataVencimento != "" {
			parsed, err := time.Parse("2006-01-02", item.DataVencimento)
			if err == nil {
				vcto = &parsed
			}
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO item_compra (compra_id, produto_id, sequencia, quantidade, preco_unitario, valor_total, localizacao, data_vencimento)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			compra.ID, item.ProdutoID, i+1, item.Quantidade, item.PrecoUnitario, vt, item.Localizacao, vcto,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao inserir item: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao confirmar: %w", err)
	}

	// Criar Conta a Pagar Automaticamente
	// A descrição será: "Compra NF [NF] - [Fornecedor]"
	var fornecedorNome string
	_ = s.db.Pool.QueryRow(ctx, "SELECT razao_social FROM fornecedor WHERE id_fornecedor = $1", req.FornecedorID).Scan(&fornecedorNome)

	descricao := fmt.Sprintf("Compra NF %s - %s", req.NumeroNotaFiscal, fornecedorNome)
	_, _ = s.db.Pool.Exec(ctx,
		`INSERT INTO conta_pagar (empresa_id, fornecedor_id, compra_id, descricao, valor_original, data_vencimento, status, usuario_id)
		 VALUES ($1, $2, $3, $4, $5, CURRENT_DATE + INTERVAL '30 days', 'aberta', $6)`,
		empresaID, req.FornecedorID, compra.ID, descricao, valorTotal, usuarioID,
	)

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

	// Buscar dados da nota fiscal para usar como lote padrão
	var numeroNF string
	_ = tx.QueryRow(ctx, "SELECT COALESCE(numero_nota_fiscal, 'S-NF') FROM compra WHERE id_compra = $1", compraID).Scan(&numeroNF)

	itens := req.ItensRecebidos
	if len(itens) == 0 {
		// Se não informou itens, recebe tudo que foi comprado
		rows, err := tx.Query(ctx, "SELECT produto_id, quantidade FROM item_compra WHERE compra_id = $1", compraID)
		if err != nil {
			return fmt.Errorf("erro ao buscar itens para recebimento automático: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var i models.ItemRecebidoRequest
			if err := rows.Scan(&i.ProdutoID, &i.QuantidadeRecebida); err == nil {
				itens = append(itens, i)
			}
		}
	}

	for _, item := range itens {
		// Buscar dados do item da compra (localizacao e vencimento)
		var localizacao *string
		var dataVencimento *time.Time
		err = tx.QueryRow(ctx, 
			`SELECT localizacao, data_vencimento FROM item_compra 
			 WHERE compra_id = $1 AND produto_id = $2`, 
			compraID, item.ProdutoID).Scan(&localizacao, &dataVencimento)
		if err != nil {
			log.Printf("⚠️ Dados de varejo não encontrados para o item %d na compra %d", item.ProdutoID, compraID)
		}

		// 1. Atualizar o item da compra
		_, err = tx.Exec(ctx,
			`UPDATE item_compra SET quantidade_recebida = $1, data_recebimento = CURRENT_TIMESTAMP
			 WHERE compra_id = $2 AND produto_id = $3`,
			item.QuantidadeRecebida, compraID, item.ProdutoID,
		)
		if err != nil {
			return fmt.Errorf("erro ao atualizar item: %w", err)
		}

		// 2. Buscar saldo atual para o log
		var saldoAnterior float64
		_ = tx.QueryRow(ctx, `SELECT estoque_atual FROM produto WHERE id_produto = $1`, item.ProdutoID).Scan(&saldoAnterior)
		novoSaldo := saldoAnterior + item.QuantidadeRecebida

		// 3. CRIAR LOTE (MUITO IMPORTANTE PARA O FEFO)
		loteFabricante := item.LoteFabricante
		if loteFabricante == "" {
			loteFabricante = numeroNF
		}
		
		loteInterno := fmt.Sprintf("COM-%d-%d", compraID, item.ProdutoID)
		vencimento := time.Now().AddDate(1, 0, 0)
		if dataVencimento != nil {
			vencimento = *dataVencimento
		}

		var loteID int
		err = tx.QueryRow(ctx,
			`INSERT INTO estoque_lote (empresa_id, produto_id, localizacao_id, lote_interno, 
			                           lote_fabricante, quantidade_inicial, quantidade_atual, 
			                           data_vencimento, usuario_id, observacao, status)
			 VALUES ($1, $2, (SELECT id_localizacao FROM estoque_localizacao WHERE empresa_id = $1 AND nome = $3 LIMIT 1), 
			         $4, $5, $6, $6, $7, $8, $9, 'ativo')
			 RETURNING id_lote`,
			empresaID, item.ProdutoID, localizacao, loteInterno,
			loteFabricante, item.QuantidadeRecebida, vencimento, usuarioID, fmt.Sprintf("Recebimento Compra NF %s", numeroNF),
		).Scan(&loteID)
		if err != nil {
			return fmt.Errorf("erro ao criar lote de compra: %w", err)
		}

		// 4. Atualizar produto com novos dados de varejo e estoque global
		_, err = tx.Exec(ctx,
			`UPDATE produto 
			 SET estoque_atual = $1, 
			     data_ultima_compra = CURRENT_DATE,
				 localizacao = COALESCE($2, localizacao),
				 data_vencimento = COALESCE($3, data_vencimento)
			 WHERE id_produto = $4 AND empresa_id = $5`,
			novoSaldo, localizacao, dataVencimento, item.ProdutoID, empresaID,
		)
		if err != nil {
			return fmt.Errorf("erro ao atualizar saldo global do produto: %w", err)
		}

		// 5. Registrar movimentação com VÍNCULO AO LOTE
		_, err = tx.Exec(ctx,
			`INSERT INTO estoque_movimentacao (empresa_id, produto_id, tipo_movimentacao,
			 quantidade, saldo_anterior, saldo_atual, origem_tipo, origem_id, usuario_id, lote_id, observacao)
			 VALUES ($1, $2, 'entrada', $3, $4, $5, 'compra', $6, $7, $8, $9)`,
			empresaID, item.ProdutoID, item.QuantidadeRecebida,
			saldoAnterior, novoSaldo, compraID, usuarioID, loteID, fmt.Sprintf("Entrada NF %s", numeroNF),
		)
		if err != nil {
			return fmt.Errorf("erro ao registrar histórico de compra: %w", err)
		}
	}

	_, err = tx.Exec(ctx, `UPDATE compra SET status = 'recebida' WHERE id_compra = $1 AND empresa_id = $2`, compraID, empresaID)
	if err != nil {
		return fmt.Errorf("erro ao finalizar status da compra: %w", err)
	}

	return tx.Commit(ctx)
}

func (s *CompraService) Listar(ctx context.Context, empresaID int, fornecedorID int, status string, notaFiscal string, dataInicio, dataFim string) ([]models.Compra, error) {
	query := `SELECT c.id_compra, c.empresa_id, c.fornecedor_id, c.usuario_id, c.numero_nota_fiscal,
		        c.data_emissao, c.data_entrada, c.data_cadastro, c.valor_produtos, c.valor_total, c.status,
		        f.razao_social as fornecedor_nome
		 FROM compra c
		 LEFT JOIN fornecedor f ON c.fornecedor_id = f.id_fornecedor
		 WHERE c.empresa_id = $1`
	
	args := []interface{}{empresaID}
	argCount := 1

	if fornecedorID > 0 {
		argCount++
		query += fmt.Sprintf(" AND c.fornecedor_id = $%d", argCount)
		args = append(args, fornecedorID)
	}

	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND c.status = $%d", argCount)
		args = append(args, status)
	}

	if notaFiscal != "" {
		argCount++
		query += fmt.Sprintf(" AND c.numero_nota_fiscal ILIKE $%d", argCount)
		args = append(args, "%"+notaFiscal+"%")
	}

	if dataInicio != "" {
		argCount++
		query += fmt.Sprintf(" AND c.data_entrada >= $%d", argCount)
		args = append(args, dataInicio)
	}

	if dataFim != "" {
		argCount++
		query += fmt.Sprintf(" AND c.data_entrada <= $%d", argCount)
		args = append(args, dataFim)
	}

	query += " ORDER BY c.data_entrada DESC"

	rows, err := s.db.Pool.Query(ctx, query, args...)
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
func (s *CompraService) BuscarPorID(ctx context.Context, empresaID, id int) (*models.Compra, error) {
	var c models.Compra
	err := s.db.Pool.QueryRow(ctx,
		`SELECT c.id_compra, c.empresa_id, c.fornecedor_id, c.usuario_id, c.numero_nota_fiscal,
		        c.data_emissao, c.data_entrada, c.data_cadastro, c.valor_produtos, c.valor_total, c.status,
		        f.razao_social as fornecedor_nome
		 FROM compra c
		 LEFT JOIN fornecedor f ON c.fornecedor_id = f.id_fornecedor
		 WHERE c.id_compra = $1 AND c.empresa_id = $2`,
		id, empresaID,
	).Scan(
		&c.ID, &c.EmpresaID, &c.FornecedorID, &c.UsuarioID, &c.NumeroNotaFiscal,
		&c.DataEmissao, &c.DataEntrada, &c.DataCadastro, &c.ValorProdutos, &c.ValorTotal, &c.Status,
		&c.FornecedorNome,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar compra: %w", err)
	}

	// Buscar itens
	rows, err := s.db.Pool.Query(ctx,
		`SELECT ic.id_item_compra, ic.compra_id, ic.produto_id, ic.sequencia, ic.quantidade,
		        ic.quantidade_recebida, ic.preco_unitario, ic.valor_total, ic.valor_desconto,
		        ic.data_recebimento, ic.localizacao, ic.data_vencimento, p.nome as produto_nome
		 FROM item_compra ic
		 LEFT JOIN produto p ON ic.produto_id = p.id_produto
		 WHERE ic.compra_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar itens da compra: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var i models.ItemCompra
		err := rows.Scan(
			&i.ID, &i.CompraID, &i.ProdutoID, &i.Sequencia, &i.Quantidade,
			&i.QuantidadeRecebida, &i.PrecoUnitario, &i.ValorTotal, &i.ValorDesconto,
			&i.DataRecebimento, &i.Localizacao, &i.DataVencimento, &i.ProdutoNome,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler linha de item: %w", err)
		}
		c.Itens = append(c.Itens, i)
	}

	return &c, nil
}
