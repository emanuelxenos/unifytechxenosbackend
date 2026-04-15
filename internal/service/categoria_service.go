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

func (s *CategoriaService) Listar(ctx context.Context, empresaID int) ([]models.Categoria, error) {
	query := `SELECT id_categoria, empresa_id, nome, descricao, categoria_pai_id, nivel, ativo, ordem_exibicao, data_cadastro
	          FROM categoria
	          WHERE empresa_id = $1 AND ativo = true
	          ORDER BY ordem_exibicao, nome`

	rows, err := s.db.Pool.Query(ctx, query, empresaID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar categorias: %w", err)
	}
	defer rows.Close()

	var categorias []models.Categoria
	for rows.Next() {
		var c models.Categoria
		err := rows.Scan(
			&c.ID, &c.EmpresaID, &c.Nome, &c.Descricao, &c.CategoriaPaiID,
			&c.Nivel, &c.Ativo, &c.OrdemExibicao, &c.DataCadastro,
		)
		if err != nil {
			continue
		}
		categorias = append(categorias, c)
	}

	return categorias, nil
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
		`INSERT INTO categoria (empresa_id, nome, descricao, categoria_pai_id, nivel, ordem_exibicao)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id_categoria, data_cadastro, ativo`,
		empresaID, req.Nome, req.Descricao, req.CategoriaPaiID, nivel, req.OrdemExibicao,
	).Scan(&c.ID, &c.DataCadastro, &c.Ativo)
	
	if err != nil {
		return nil, fmt.Errorf("erro ao criar categoria: %w", err)
	}

	c.EmpresaID = empresaID
	c.Nome = req.Nome
	c.Descricao = req.Descricao
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
		    categoria_pai_id = $3,
		    nivel = $4,
		    ordem_exibicao = $5
		 WHERE id_categoria = $6 AND empresa_id = $7`,
		req.Nome, req.Descricao, req.CategoriaPaiID, nivel, req.OrdemExibicao,
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
