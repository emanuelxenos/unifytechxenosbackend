package service

import (
	"context"
	"fmt"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
	"erp-backend/pkg/utils"
)

type AuthService struct {
	db *database.PostgresDB
}

func NewAuthService(db *database.PostgresDB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Login(ctx context.Context, req models.UsuarioLoginRequest, jwtExpiryHours int) (*models.UsuarioLoginResponse, error) {
	var user models.Usuario
	err := s.db.Pool.QueryRow(ctx,
		`SELECT id_usuario, empresa_id, nome, login, senha_hash, perfil,
		        pode_abrir_caixa, pode_dar_desconto, limite_desconto_percentual, ativo
		 FROM usuario
		 WHERE login = $1 AND ativo = true`, req.Login).Scan(
		&user.ID, &user.EmpresaID, &user.Nome, &user.Login, &user.SenhaHash, &user.Perfil,
		&user.PodeAbrirCaixa, &user.PodeDarDesconto, &user.LimiteDescontoPercentual, &user.Ativo,
	)
	if err != nil {
		return nil, fmt.Errorf("login ou senha inválidos")
	}

	if !utils.CheckPassword(user.SenhaHash, req.Senha) {
		return nil, fmt.Errorf("login ou senha inválidos")
	}

	token, err := utils.GenerateToken(user.ID, user.EmpresaID, user.Perfil, user.Nome, user.Login, jwtExpiryHours)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar token: %w", err)
	}

	// Atualizar último acesso
	_, _ = s.db.Pool.Exec(ctx,
		`UPDATE usuario SET ultimo_acesso = CURRENT_TIMESTAMP WHERE id_usuario = $1`, user.ID)

	return &models.UsuarioLoginResponse{
		Token: token,
		Usuario: models.UsuarioInfo{
			ID:     user.ID,
			Nome:   user.Nome,
			Perfil: user.Perfil,
			Permissoes: models.UsuarioPermissao{
				PodeAbrirCaixa:  user.PodeAbrirCaixa,
				PodeDarDesconto: user.PodeDarDesconto,
				LimiteDesconto:  user.LimiteDescontoPercentual,
			},
		},
	}, nil
}
