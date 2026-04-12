package service

import (
	"context"
	"fmt"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
)

type FornecedorService struct {
	db *database.PostgresDB
}

func NewFornecedorService(db *database.PostgresDB) *FornecedorService {
	return &FornecedorService{db: db}
}

func (s *FornecedorService) Listar(ctx context.Context, empresaID int) ([]models.Fornecedor, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT id_fornecedor, empresa_id, razao_social, nome_fantasia, cnpj,
		        telefone, email, data_cadastro, ativo
		 FROM fornecedor WHERE empresa_id = $1 AND ativo = true
		 ORDER BY razao_social`, empresaID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar fornecedores: %w", err)
	}
	defer rows.Close()

	var list []models.Fornecedor
	for rows.Next() {
		var f models.Fornecedor
		rows.Scan(&f.ID, &f.EmpresaID, &f.RazaoSocial, &f.NomeFantasia, &f.CNPJ,
			&f.Telefone, &f.Email, &f.DataCadastro, &f.Ativo)
		list = append(list, f)
	}
	return list, nil
}

func (s *FornecedorService) Criar(ctx context.Context, empresaID int, req models.CriarFornecedorRequest) (*models.Fornecedor, error) {
	var f models.Fornecedor
	err := s.db.Pool.QueryRow(ctx,
		`INSERT INTO fornecedor (empresa_id, razao_social, cnpj, telefone, email)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id_fornecedor, data_cadastro, ativo`,
		empresaID, req.RazaoSocial, req.CNPJ, req.Telefone, req.Email,
	).Scan(&f.ID, &f.DataCadastro, &f.Ativo)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar fornecedor: %w", err)
	}
	f.EmpresaID = empresaID
	f.RazaoSocial = req.RazaoSocial
	return &f, nil
}

func (s *FornecedorService) Atualizar(ctx context.Context, empresaID, id int, req models.CriarFornecedorRequest) error {
	_, err := s.db.Pool.Exec(ctx,
		`UPDATE fornecedor 
		 SET razao_social = $1, cnpj = $2, telefone = $3, email = $4
		 WHERE id_fornecedor = $5 AND empresa_id = $6`,
		req.RazaoSocial, req.CNPJ, req.Telefone, req.Email, id, empresaID)
	if err != nil {
		return fmt.Errorf("erro ao atualizar fornecedor: %w", err)
	}
	return nil
}

func (s *FornecedorService) Inativar(ctx context.Context, empresaID, id int) error {
	_, err := s.db.Pool.Exec(ctx,
		`UPDATE fornecedor SET ativo = false WHERE id_fornecedor = $1 AND empresa_id = $2`,
		id, empresaID)
	if err != nil {
		return fmt.Errorf("erro ao inativar fornecedor: %w", err)
	}
	return nil
}
