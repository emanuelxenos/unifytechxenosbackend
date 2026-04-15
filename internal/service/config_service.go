package service

import (
	"context"
	"fmt"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
)

type ConfigService struct {
	db *database.PostgresDB
}

func NewConfigService(db *database.PostgresDB) *ConfigService {
	return &ConfigService{db: db}
}

func (s *ConfigService) Listar(ctx context.Context, empresaID int) ([]models.Configuracao, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT id_config, empresa_id, chave, valor, tipo, categoria, descricao, data_atualizacao
		 FROM configuracao WHERE empresa_id = $1 ORDER BY categoria, chave`, empresaID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar configurações: %w", err)
	}
	defer rows.Close()

	var configs []models.Configuracao
	for rows.Next() {
		var c models.Configuracao
		rows.Scan(&c.ID, &c.EmpresaID, &c.Chave, &c.Valor, &c.Tipo, &c.Categoria, &c.Descricao, &c.DataAtualizacao)
		configs = append(configs, c)
	}
	return configs, nil
}

func (s *ConfigService) Atualizar(ctx context.Context, empresaID int, req models.AtualizarConfigRequest) error {
	for _, item := range req.Configs {
		var exists bool
		err := s.db.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM configuracao WHERE empresa_id = $1 AND chave = $2)", empresaID, item.Chave).Scan(&exists)
		if err != nil {
			return fmt.Errorf("erro ao checar config %s: %w", item.Chave, err)
		}

		if exists {
			_, err = s.db.Pool.Exec(ctx,
				`UPDATE configuracao SET valor = $1, data_atualizacao = CURRENT_TIMESTAMP
				 WHERE empresa_id = $2 AND chave = $3`,
				item.Valor, empresaID, item.Chave,
			)
		} else {
			_, err = s.db.Pool.Exec(ctx,
				`INSERT INTO configuracao (empresa_id, chave, valor, tipo, categoria, data_atualizacao)
				 VALUES ($2, $3, $1, 'texto', 'backup', CURRENT_TIMESTAMP)`,
				item.Valor, empresaID, item.Chave,
			)
		}

		if err != nil {
			return fmt.Errorf("erro ao atualizar/inserir config %s: %w", item.Chave, err)
		}
	}
	return nil
}
