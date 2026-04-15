package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/repository"
	"erp-backend/pkg/config"
)

type BackupService struct {
	cfg        *config.Config
	backupRepo *repository.BackupRepository
	configSvc  *ConfigService
}

func NewBackupService(cfg *config.Config, backupRepo *repository.BackupRepository, configSvc *ConfigService) *BackupService {
	return &BackupService{
		cfg:        cfg,
		backupRepo: backupRepo,
		configSvc:  configSvc,
	}
}

func (s *BackupService) getPgDumpPath() string {
	path, err := exec.LookPath("pg_dump")
	if err == nil {
		return path
	}

	if runtime.GOOS == "windows" {
		commonPaths := []string{
			`C:\Program Files\PostgreSQL\18\bin\pg_dump.exe`,
			`C:\Program Files\PostgreSQL\17\bin\pg_dump.exe`,
			`C:\Program Files\PostgreSQL\16\bin\pg_dump.exe`,
			`C:\Program Files\PostgreSQL\15\bin\pg_dump.exe`,
			`C:\Program Files\PostgreSQL\14\bin\pg_dump.exe`,
		}
		for _, p := range commonPaths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return "pg_dump"
}

func (s *BackupService) ExecutarBackup(ctx context.Context, empresaID int, usuarioID *int, tipo string) error {
	// 1. Get backup directory from configuracao table or use default
	configs, err := s.configSvc.Listar(ctx, empresaID)
	if err != nil {
		return fmt.Errorf("erro ao buscar configurações: %w", err)
	}

	backupDir := "./backups"
	for _, c := range configs {
		if c.Chave == "backup.dir" && c.Valor != nil && *c.Valor != "" {
			backupDir = *c.Valor
		}
	}

	// 2. Organize in MM-YYYY folder
	now := time.Now()
	mesAno := now.Format("01-2006")
	targetDir := filepath.Join(backupDir, mesAno)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório de backup: %w", err)
	}

	// 3. Format filename
	filename := fmt.Sprintf("backup-%s.sql", now.Format("2006-01-02_150405"))
	filepath := filepath.Join(targetDir, filename)

	// 4. Exec pg_dump
	// Ensure PGPASSWORD is set in environment for pg_dump
	os.Setenv("PGPASSWORD", s.cfg.DBPass)
	defer os.Unsetenv("PGPASSWORD")

	pgDumpPath := s.getPgDumpPath()
	cmd := exec.CommandContext(ctx, pgDumpPath,
		"-h", s.cfg.DBHost,
		"-p", s.cfg.DBPort,
		"-U", s.cfg.DBUser,
		"-d", s.cfg.DBName,
		"-F", "p", // plain text SQL
		"-f", filepath,
	)

	log.Printf("Iniciando backup para: %s", filepath)
	output, err := cmd.CombinedOutput()
	
	status := "sucesso"
	var observacoes *string
	var tamanho *int64

	if err != nil {
		status = "falha"
		obs := fmt.Sprintf("Erro: %v\nOutput: %s", err, string(output))
		observacoes = &obs
		log.Printf("Falha no backup: %s", obs)
	} else {
		// Get file size
		info, errStat := os.Stat(filepath)
		if errStat == nil {
			size := info.Size()
			tamanho = &size
		}
		log.Println("Backup concluído com sucesso")
	}

	// 5. Log in DB
	backupLog := models.Backup{
		EmpresaID:   empresaID,
		NomeArquivo: filename,
		Caminho:     filepath,
		Tamanho:     tamanho,
		Tipo:        tipo,
		Status:      status,
		Observacoes: observacoes,
		UsuarioID:   usuarioID,
	}

	if errLog := s.backupRepo.SalvarBackup(ctx, backupLog); errLog != nil {
		log.Printf("Erro ao salvar log de backup no BD: %v", errLog)
		// We still return the original error if pg_dump failed
	}

	if err != nil {
		return fmt.Errorf("falha ao executar pg_dump: %w", err)
	}
	return nil
}

func (s *BackupService) ListarBackups(ctx context.Context, empresaID int) ([]models.Backup, error) {
	return s.backupRepo.ListarBackups(ctx, empresaID)
}
