package service

import (
	"context"
	"fmt"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
)

type ClienteService struct {
	db *database.PostgresDB
}

func NewClienteService(db *database.PostgresDB) *ClienteService {
	return &ClienteService{db: db}
}

func (s *ClienteService) Listar(ctx context.Context, empresaID int, incluirInativos bool) ([]models.Cliente, error) {
	query := `SELECT id_cliente, empresa_id, nome, tipo_pessoa, cpf_cnpj,
	                 telefone, email, limite_credito, saldo_devedor,
	                 data_cadastro, ativo
	          FROM cliente
	          WHERE empresa_id = $1`
	if incluirInativos {
		query += ` AND ativo = false`
	} else {
		query += ` AND ativo = true`
	}
	query += ` ORDER BY nome`

	rows, err := s.db.Pool.Query(ctx, query, empresaID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar clientes: %w", err)
	}
	defer rows.Close()

	var clientes []models.Cliente
	for rows.Next() {
		var c models.Cliente
		err := rows.Scan(
			&c.ID, &c.EmpresaID, &c.Nome, &c.TipoPessoa, &c.CPFCNPJ,
			&c.Telefone, &c.Email, &c.LimiteCredito, &c.SaldoDevedor,
			&c.DataCadastro, &c.Ativo,
		)
		if err != nil {
			continue
		}
		clientes = append(clientes, c)
	}
	return clientes, nil
}

func (s *ClienteService) Criar(ctx context.Context, empresaID int, req models.CriarClienteRequest) (*models.Cliente, error) {
	tipoPessoa := req.TipoPessoa
	if tipoPessoa == "" {
		tipoPessoa = "F"
	}

	var c models.Cliente
	err := s.db.Pool.QueryRow(ctx,
		`INSERT INTO cliente (empresa_id, nome, tipo_pessoa, cpf_cnpj, telefone, email, limite_credito)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id_cliente, data_cadastro, ativo`,
		empresaID, req.Nome, tipoPessoa, req.CPFCNPJ, req.Telefone, req.Email, req.LimiteCredito,
	).Scan(&c.ID, &c.DataCadastro, &c.Ativo)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente: %w", err)
	}

	c.EmpresaID = empresaID
	c.Nome = req.Nome
	c.TipoPessoa = tipoPessoa

	return &c, nil
}

func (s *ClienteService) Atualizar(ctx context.Context, empresaID, clienteID int, req models.CriarClienteRequest) error {
	_, err := s.db.Pool.Exec(ctx,
		`UPDATE cliente SET
		    nome = $1,
		    tipo_pessoa = COALESCE(NULLIF($2, ''), tipo_pessoa),
		    cpf_cnpj = $3,
		    telefone = $4,
		    email = $5,
		    limite_credito = $6
		 WHERE id_cliente = $7 AND empresa_id = $8`,
		req.Nome, req.TipoPessoa, req.CPFCNPJ, req.Telefone, req.Email,
		req.LimiteCredito, clienteID, empresaID,
	)
	if err != nil {
		return fmt.Errorf("erro ao atualizar cliente: %w", err)
	}
	return nil
}

func (s *ClienteService) Inativar(ctx context.Context, empresaID, clienteID int) error {
	_, err := s.db.Pool.Exec(ctx,
		`UPDATE cliente SET ativo = false
		 WHERE id_cliente = $1 AND empresa_id = $2`,
		clienteID, empresaID,
	)
	if err != nil {
		return fmt.Errorf("erro ao inativar cliente: %w", err)
	}
	return nil
}

