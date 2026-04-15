package repository

import (
	"context"
	"fmt"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
)

type BackupRepository struct {
	db *database.PostgresDB
}

func NewBackupRepository(db *database.PostgresDB) *BackupRepository {
	repo := &BackupRepository{db: db}
	repo.ensureTableExists()
	return repo
}

func (r *BackupRepository) ensureTableExists() {
	query := `
	CREATE TABLE IF NOT EXISTS backup (
		id_backup SERIAL PRIMARY KEY,
		empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
		nome_arquivo VARCHAR(200) NOT NULL,
		caminho VARCHAR(500) NOT NULL,
		tamanho BIGINT,
		data_backup TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		tipo VARCHAR(20) DEFAULT 'automatico',
		status VARCHAR(20) DEFAULT 'sucesso',
		observacoes TEXT,
		usuario_id INTEGER REFERENCES usuario(id_usuario)
	);`
	_, _ = r.db.Pool.Exec(context.Background(), query)
}

func (r *BackupRepository) SalvarBackup(ctx context.Context, b models.Backup) error {
	query := `
		INSERT INTO backup (
			empresa_id, nome_arquivo, caminho, tamanho, tipo, status, observacoes, usuario_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id_backup, data_backup`

	err := r.db.Pool.QueryRow(ctx, query,
		b.EmpresaID, b.NomeArquivo, b.Caminho, b.Tamanho, b.Tipo, b.Status, b.Observacoes, b.UsuarioID,
	).Scan(&b.ID, &b.DataBackup)

	if err != nil {
		return fmt.Errorf("erro ao salvar log de backup: %w", err)
	}
	return nil
}

func (r *BackupRepository) ListarBackups(ctx context.Context, empresaID int) ([]models.Backup, error) {
	query := `
		SELECT id_backup, empresa_id, nome_arquivo, caminho, tamanho, data_backup, tipo, status, observacoes, usuario_id
		FROM backup 
		WHERE empresa_id = $1 
		ORDER BY data_backup DESC
		LIMIT 50`

	rows, err := r.db.Pool.Query(ctx, query, empresaID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar backups: %w", err)
	}
	defer rows.Close()

	var backups []models.Backup
	for rows.Next() {
		var b models.Backup
		err := rows.Scan(
			&b.ID, &b.EmpresaID, &b.NomeArquivo, &b.Caminho, &b.Tamanho, 
			&b.DataBackup, &b.Tipo, &b.Status, &b.Observacoes, &b.UsuarioID,
		)
		if err != nil {
			return nil, err
		}
		backups = append(backups, b)
	}

	return backups, nil
}
