package service

import (
	"context"
	"fmt"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
	"erp-backend/pkg/utils"
)

type UsuarioService struct {
	db *database.PostgresDB
}

func NewUsuarioService(db *database.PostgresDB) *UsuarioService {
	return &UsuarioService{db: db}
}

func (s *UsuarioService) Listar(ctx context.Context, empresaID int) ([]models.Usuario, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT id_usuario, empresa_id, nome, login, perfil,
		        pode_abrir_caixa, pode_fechar_caixa, pode_dar_desconto,
		        limite_desconto_percentual, pode_cancelar_venda, ativo, data_cadastro
		 FROM usuario WHERE empresa_id = $1 ORDER BY nome`, empresaID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar usuários: %w", err)
	}
	defer rows.Close()

	var usuarios []models.Usuario
	for rows.Next() {
		var u models.Usuario
		rows.Scan(&u.ID, &u.EmpresaID, &u.Nome, &u.Login, &u.Perfil,
			&u.PodeAbrirCaixa, &u.PodeFecharCaixa, &u.PodeDarDesconto,
			&u.LimiteDescontoPercentual, &u.PodeCancelarVenda, &u.Ativo, &u.DataCadastro)
		usuarios = append(usuarios, u)
	}
	return usuarios, nil
}

func (s *UsuarioService) Criar(ctx context.Context, empresaID int, req models.CriarUsuarioRequest) (*models.Usuario, error) {
	hash, err := utils.HashPassword(req.Senha)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar hash: %w", err)
	}

	var u models.Usuario
	err = s.db.Pool.QueryRow(ctx,
		`INSERT INTO usuario (empresa_id, nome, cpf, login, senha_hash, perfil,
		                      pode_abrir_caixa, pode_fechar_caixa, pode_dar_desconto,
		                      limite_desconto_percentual, pode_cancelar_venda)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id_usuario, data_cadastro, ativo`,
		empresaID, req.Nome, req.CPF, req.Login, hash, req.Perfil,
		req.PodeAbrirCaixa, req.PodeFecharCaixa, req.PodeDarDesconto,
		req.LimiteDescontoPercentual, req.PodeCancelarVenda,
	).Scan(&u.ID, &u.DataCadastro, &u.Ativo)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar usuário: %w", err)
	}

	u.EmpresaID = empresaID
	u.Nome = req.Nome
	u.Login = req.Login
	u.Perfil = req.Perfil
	return &u, nil
}
