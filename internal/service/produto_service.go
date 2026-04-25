package service

import (
	"context"
	"fmt"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
)

type ProdutoService struct {
	db *database.PostgresDB
}

func NewProdutoService(db *database.PostgresDB) *ProdutoService {
	return &ProdutoService{db: db}
}

func (s *ProdutoService) Listar(ctx context.Context, empresaID, page, limit int, categoriaID *int, search string, apenasBaixoEstoque, apenasVencendo bool) ([]models.Produto, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	// Count total
	countQuery := `SELECT COUNT(*) FROM produto WHERE empresa_id = $1 AND ativo = true`
	args := []interface{}{empresaID}
	argIdx := 2

	if categoriaID != nil {
		countQuery += fmt.Sprintf(` AND categoria_id = $%d`, argIdx)
		args = append(args, *categoriaID)
		argIdx++
	}

	if search != "" {
		countQuery += fmt.Sprintf(` AND (nome ILIKE $%d OR codigo_barras ILIKE $%d)`, argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	if apenasBaixoEstoque {
		countQuery += ` AND controlar_estoque = true AND estoque_atual <= estoque_minimo`
	}

	if apenasVencendo {
		countQuery += ` AND data_vencimento IS NOT NULL AND data_vencimento <= CURRENT_DATE + INTERVAL '15 days'`
	}

	var total int
	err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao contar produtos: %w", err)
	}

	// Query produtos
	query := `SELECT p.id_produto, p.empresa_id, p.categoria_id, p.codigo_barras,
	                 p.codigo_interno, p.nome, p.descricao, p.marca, p.unidade_venda,
	                 p.estoque_atual, p.estoque_minimo, p.controlar_estoque, p.preco_custo,
	                 p.preco_venda, p.preco_promocional, p.data_inicio_promocao,
	                 p.data_fim_promocao, p.margem_lucro, p.ativo, p.destacado,
	                 p.data_cadastro, p.data_ultima_venda, p.localizacao,
	                 (SELECT MIN(data_vencimento) FROM estoque_lote WHERE produto_id = p.id_produto AND status = 'ativo') as data_vencimento,
	                 c.nome as categoria_nome
	          FROM produto p
	          LEFT JOIN categoria c ON p.categoria_id = c.id_categoria
	          WHERE p.empresa_id = $1 AND p.ativo = true`

	queryArgs := []interface{}{empresaID}
	qArgIdx := 2

	if categoriaID != nil {
		query += fmt.Sprintf(` AND p.categoria_id = $%d`, qArgIdx)
		queryArgs = append(queryArgs, *categoriaID)
		qArgIdx++
	}

	if search != "" {
		query += fmt.Sprintf(` AND (p.nome ILIKE $%d OR p.codigo_barras ILIKE $%d)`, qArgIdx, qArgIdx)
		queryArgs = append(queryArgs, "%"+search+"%")
		qArgIdx++
	}

	if apenasBaixoEstoque {
		query += ` AND p.controlar_estoque = true AND p.estoque_atual <= p.estoque_minimo`
	}

	if apenasVencendo {
		query += ` AND EXISTS (SELECT 1 FROM estoque_lote WHERE produto_id = p.id_produto AND status = 'ativo' AND data_vencimento <= CURRENT_DATE + INTERVAL '15 days')`
	}

	query += fmt.Sprintf(` ORDER BY p.nome LIMIT $%d OFFSET $%d`, qArgIdx, qArgIdx+1)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := s.db.Pool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao listar produtos: %w", err)
	}
	defer rows.Close()

	var produtos = []models.Produto{}
	for rows.Next() {
		var p models.Produto
		err := rows.Scan(
			&p.ID, &p.EmpresaID, &p.CategoriaID, &p.CodigoBarras,
			&p.CodigoInterno, &p.Nome, &p.Descricao, &p.Marca, &p.UnidadeVenda,
			&p.EstoqueAtual, &p.EstoqueMinimo, &p.ControlarEstoque, &p.PrecoCusto,
			&p.PrecoVenda, &p.PrecoPromocional, &p.DataInicioPromocao,
			&p.DataFimPromocao, &p.MargemLucro, &p.Ativo, &p.Destacado,
			&p.DataCadastro, &p.DataUltimaVenda, &p.Localizacao, &p.DataVencimento,
			&p.CategoriaNome,
		)
		if err != nil {
			continue
		}
		produtos = append(produtos, p)
	}

	return produtos, total, nil
}

func (s *ProdutoService) Buscar(ctx context.Context, empresaID int, codigo, nome *string) (*models.ProdutoBuscaResponse, error) {
	var p models.ProdutoBuscaResponse

	if codigo != nil && *codigo != "" {
		err := s.db.Pool.QueryRow(ctx,
			`SELECT id_produto, codigo_barras, nome, preco_venda, estoque_atual, unidade_venda
			 FROM produto
			 WHERE empresa_id = $1 AND codigo_barras = $2 AND ativo = true`,
			empresaID, *codigo,
		).Scan(&p.ID, &p.CodigoBarras, &p.Nome, &p.PrecoVenda, &p.EstoqueAtual, &p.UnidadeVenda)
		if err != nil {
			return nil, fmt.Errorf("produto não encontrado")
		}
		return &p, nil
	}

	if nome != nil && *nome != "" {
		err := s.db.Pool.QueryRow(ctx,
			`SELECT id_produto, codigo_barras, nome, preco_venda, estoque_atual, unidade_venda
			 FROM produto
			 WHERE empresa_id = $1 AND LOWER(nome) LIKE LOWER($2) AND ativo = true
			 LIMIT 1`,
			empresaID, "%"+*nome+"%",
		).Scan(&p.ID, &p.CodigoBarras, &p.Nome, &p.PrecoVenda, &p.EstoqueAtual, &p.UnidadeVenda)
		if err != nil {
			return nil, fmt.Errorf("produto não encontrado")
		}
		return &p, nil
	}

	return nil, fmt.Errorf("informe código de barras ou nome para busca")
}

func (s *ProdutoService) BuscarPorID(ctx context.Context, empresaID, produtoID int) (*models.Produto, error) {
	var p models.Produto
	err := s.db.Pool.QueryRow(ctx,
		`SELECT p.id_produto, p.empresa_id, p.categoria_id, p.codigo_barras,
		        p.codigo_interno, p.nome, p.descricao, p.marca,
		        p.unidade_venda, p.estoque_atual, p.estoque_minimo,
		        p.controlar_estoque, p.preco_custo, p.preco_venda,
		        p.preco_promocional, p.data_inicio_promocao, p.data_fim_promocao,
		        p.margem_lucro, p.ativo, p.destacado, p.data_cadastro,
		        p.data_ultima_compra, p.data_ultima_venda, p.localizacao,
		        (SELECT MIN(data_vencimento) FROM estoque_lote WHERE produto_id = p.id_produto AND status = 'ativo') as data_vencimento,
		        c.nome as categoria_nome
		 FROM produto p
		 LEFT JOIN categoria c ON p.categoria_id = c.id_categoria
		 WHERE p.id_produto = $1 AND p.empresa_id = $2`,
		produtoID, empresaID,
	).Scan(
		&p.ID, &p.EmpresaID, &p.CategoriaID, &p.CodigoBarras,
		&p.CodigoInterno, &p.Nome, &p.Descricao, &p.Marca,
		&p.UnidadeVenda, &p.EstoqueAtual, &p.EstoqueMinimo,
		&p.ControlarEstoque, &p.PrecoCusto, &p.PrecoVenda,
		&p.PrecoPromocional, &p.DataInicioPromocao, &p.DataFimPromocao,
		&p.MargemLucro, &p.Ativo, &p.Destacado, &p.DataCadastro,
		&p.DataUltimaCompra, &p.DataUltimaVenda, &p.Localizacao, &p.DataVencimento,
		&p.CategoriaNome,
	)
	if err != nil {
		return nil, fmt.Errorf("produto não encontrado")
	}
	return &p, nil
}

func (s *ProdutoService) Criar(ctx context.Context, empresaID int, req models.CriarProdutoRequest) (*models.Produto, error) {
	unidade := req.UnidadeVenda
	if unidade == "" {
		unidade = "UN"
	}

	var p models.Produto
	err := s.db.Pool.QueryRow(ctx,
		`INSERT INTO produto (empresa_id, codigo_barras, codigo_interno, nome, descricao,
		                      categoria_id, unidade_venda, controlar_estoque,
		                      estoque_minimo, preco_custo, preco_venda,
		                      preco_promocional, data_inicio_promocao, data_fim_promocao,
		                      margem_lucro, marca, localizacao, data_vencimento)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		 RETURNING id_produto, data_cadastro, ativo`,
		empresaID, req.CodigoBarras, req.CodigoInterno, req.Nome, req.Descricao,
		req.CategoriaID, unidade, req.ControlarEstoque,
		req.EstoqueMinimo, req.PrecoCusto, req.PrecoVenda,
		req.PrecoPromocional, req.DataInicioPromocao, req.DataFimPromocao,
		req.MargemLucro, req.Marca, req.Localizacao, req.DataVencimento,
	).Scan(&p.ID, &p.DataCadastro, &p.Ativo)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar produto: %w", err)
	}

	p.EmpresaID = empresaID
	p.Nome = req.Nome
	p.PrecoVenda = req.PrecoVenda
	p.PrecoCusto = req.PrecoCusto
	p.UnidadeVenda = unidade
	p.CodigoBarras = req.CodigoBarras

	return &p, nil
}

func (s *ProdutoService) Atualizar(ctx context.Context, empresaID, produtoID int, req models.CriarProdutoRequest) error {
	_, err := s.db.Pool.Exec(ctx,
		`UPDATE produto SET
		    codigo_barras = COALESCE($1, codigo_barras),
		    codigo_interno = $2,
		    nome = $3,
		    descricao = $4,
		    categoria_id = $5,
		    unidade_venda = COALESCE(NULLIF($6, ''), unidade_venda),
		    controlar_estoque = $7,
		    estoque_minimo = $8,
		    preco_custo = $9,
		    preco_venda = $10,
		    preco_promocional = $11,
		    data_inicio_promocao = $12,
		    data_fim_promocao = $13,
		    margem_lucro = $14,
		    marca = $15,
		    localizacao = $16,
		    data_vencimento = $17
		 WHERE id_produto = $18 AND empresa_id = $19`,
		req.CodigoBarras, req.CodigoInterno, req.Nome, req.Descricao, req.CategoriaID,
		req.UnidadeVenda, req.ControlarEstoque, req.EstoqueMinimo,
		req.PrecoCusto, req.PrecoVenda, req.PrecoPromocional,
		req.DataInicioPromocao, req.DataFimPromocao, req.MargemLucro,
		req.Marca, req.Localizacao, req.DataVencimento,
		produtoID, empresaID,
	)
	if err != nil {
		return fmt.Errorf("erro ao atualizar produto: %w", err)
	}
	return nil
}

func (s *ProdutoService) Inativar(ctx context.Context, empresaID, produtoID int) error {
	_, err := s.db.Pool.Exec(ctx,
		`UPDATE produto SET ativo = false WHERE id_produto = $1 AND empresa_id = $2`,
		produtoID, empresaID,
	)
	if err != nil {
		return fmt.Errorf("erro ao inativar produto: %w", err)
	}
	return nil
}
