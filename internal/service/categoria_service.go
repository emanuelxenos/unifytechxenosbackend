package service

import (
	"context"
	"fmt"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
)

type CategoriaService struct {
	db *database.PostgresDB
}

func NewCategoriaService(db *database.PostgresDB) *CategoriaService {
	return &CategoriaService{db: db}
}

func (s *CategoriaService) Listar(ctx context.Context, empresaID, page, limit int, search string) ([]models.Categoria, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	// Count total
	countQuery := `SELECT COUNT(*) FROM categoria WHERE empresa_id = $1 AND ativo = true`
	args := []interface{}{empresaID}
	argIdx := 2

	if search != "" {
		countQuery += fmt.Sprintf(` AND nome ILIKE $%d`, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	var total int
	err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao contar categorias: %w", err)
	}

	// Query categorias
	query := `SELECT c.id_categoria, c.empresa_id, c.nome, c.descricao, c.icone, c.cor_hex, c.categoria_pai_id, c.nivel, c.ativo, c.ordem_exibicao, c.data_cadastro,
	          (SELECT COUNT(*) FROM produto p WHERE p.categoria_id = c.id_categoria AND p.ativo = true) as total_produtos,
	          cp.nome as categoria_pai_nome
	          FROM categoria c
	          LEFT JOIN categoria cp ON c.categoria_pai_id = cp.id_categoria
	          WHERE c.empresa_id = $1 AND c.ativo = true`

	queryArgs := []interface{}{empresaID}
	qArgIdx := 2

	if search != "" {
		query += fmt.Sprintf(` AND c.nome ILIKE $%d`, qArgIdx)
		queryArgs = append(queryArgs, "%"+search+"%")
		qArgIdx++
	}

	query += fmt.Sprintf(` ORDER BY COALESCE(c.categoria_pai_id, c.id_categoria), c.categoria_pai_id NULLS FIRST, c.ordem_exibicao, c.nome LIMIT $%d OFFSET $%d`, qArgIdx, qArgIdx+1)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := s.db.Pool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao listar categorias: %w", err)
	}
	defer rows.Close()

	var categorias []models.Categoria
	for rows.Next() {
		var c models.Categoria
		err := rows.Scan(
			&c.ID, &c.EmpresaID, &c.Nome, &c.Descricao, &c.Icone, &c.CorHex, &c.CategoriaPaiID,
			&c.Nivel, &c.Ativo, &c.OrdemExibicao, &c.DataCadastro, &c.TotalProdutos, &c.CategoriaPaiNome,
		)
		if err != nil {
			continue
		}
		categorias = append(categorias, c)
	}

	return categorias, total, nil
}

func (s *CategoriaService) Criar(ctx context.Context, empresaID int, req models.CriarCategoriaRequest) (*models.Categoria, error) {
	var c models.Categoria
	nivel := 1
	// Logica basica de nivel
	if req.CategoriaPaiID != nil {
		var nivelPai int
		err := s.db.Pool.QueryRow(ctx, "SELECT nivel FROM categoria WHERE id_categoria = $1 AND empresa_id = $2", *req.CategoriaPaiID, empresaID).Scan(&nivelPai)
		if err == nil {
			nivel = nivelPai + 1
		}
	}

	err := s.db.Pool.QueryRow(ctx,
		`INSERT INTO categoria (empresa_id, nome, descricao, icone, cor_hex, categoria_pai_id, nivel, ordem_exibicao)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id_categoria, data_cadastro, ativo`,
		empresaID, req.Nome, req.Descricao, req.Icone, req.CorHex, req.CategoriaPaiID, nivel, req.OrdemExibicao,
	).Scan(&c.ID, &c.DataCadastro, &c.Ativo)
	
	if err != nil {
		return nil, fmt.Errorf("erro ao criar categoria: %w", err)
	}

	c.EmpresaID = empresaID
	c.Nome = req.Nome
	c.Descricao = req.Descricao
	c.Icone = req.Icone
	c.CorHex = req.CorHex
	c.CategoriaPaiID = req.CategoriaPaiID
	c.Nivel = nivel
	c.OrdemExibicao = req.OrdemExibicao

	return &c, nil
}

func (s *CategoriaService) Atualizar(ctx context.Context, empresaID, categoriaID int, req models.CriarCategoriaRequest) error {
	nivel := 1
	if req.CategoriaPaiID != nil {
		var nivelPai int
		err := s.db.Pool.QueryRow(ctx, "SELECT nivel FROM categoria WHERE id_categoria = $1 AND empresa_id = $2", *req.CategoriaPaiID, empresaID).Scan(&nivelPai)
		if err == nil {
			nivel = nivelPai + 1
		}
	}

	_, err := s.db.Pool.Exec(ctx,
		`UPDATE categoria SET
		    nome = $1,
		    descricao = $2,
		    icone = $3,
		    cor_hex = $4,
		    categoria_pai_id = $5,
		    nivel = $6,
		    ordem_exibicao = $7
		 WHERE id_categoria = $8 AND empresa_id = $9`,
		req.Nome, req.Descricao, req.Icone, req.CorHex, req.CategoriaPaiID, nivel, req.OrdemExibicao,
		categoriaID, empresaID,
	)
	if err != nil {
		return fmt.Errorf("erro ao atualizar categoria: %w", err)
	}
	return nil
}

func (s *CategoriaService) Inativar(ctx context.Context, empresaID, categoriaID int) error {
	_, err := s.db.Pool.Exec(ctx,
		`UPDATE categoria SET ativo = false WHERE id_categoria = $1 AND empresa_id = $2`,
		categoriaID, empresaID,
	)
	if err != nil {
		return fmt.Errorf("erro ao inativar categoria: %w", err)
	}
	return nil
}
