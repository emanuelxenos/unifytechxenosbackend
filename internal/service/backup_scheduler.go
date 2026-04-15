package service

import (
	"context"
	"log"
	"time"

	"erp-backend/internal/infrastructure/database"
	"erp-backend/internal/repository"
)

type BackupScheduler struct {
	db            *database.PostgresDB
	configService *ConfigService
	backupService *BackupService
	backupRepo    *repository.BackupRepository
}

func NewBackupScheduler(db *database.PostgresDB, configService *ConfigService, backupService *BackupService, backupRepo *repository.BackupRepository) *BackupScheduler {
	return &BackupScheduler{
		db:            db,
		configService: configService,
		backupService: backupService,
		backupRepo:    backupRepo,
	}
}

func (s *BackupScheduler) Start(ctx context.Context) {
	log.Println("🚀 Iniciando agendador de backups automático...")
	
	// Roda a cada 1 minuto para checar necessidades
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				s.verifyAndRunBackups(context.Background())
			}
		}
	}()
}

func (s *BackupScheduler) verifyAndRunBackups(ctx context.Context) {
	// Pega todas as empresas
	// Como não temos um repo genérico de empresa injetado aqui ainda, faremos uma query direta para simplicidade
	rows, err := s.db.Pool.Query(ctx, "SELECT id_empresa FROM empresa WHERE ativo = true")
	if err != nil {
		log.Printf("Erro ao buscar empresas no scheduler de backup: %v", err)
		return
	}
	defer rows.Close()

	var empresas []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err == nil {
			empresas = append(empresas, id)
		}
	}

	for _, empID := range empresas {
		configs, err := s.configService.Listar(ctx, empID)
		if err != nil {
			continue
		}

		interval := "Somente Manual"
		for _, c := range configs {
			if c.Chave == "backup.interval" && c.Valor != nil {
				interval = *c.Valor
			}
		}

		if interval == "Somente Manual" {
			continue
		}

		var duration time.Duration
		switch interval {
		case "A cada 5 minutos":
			duration = 5 * time.Minute
		case "A cada 30 minutos":
			duration = 30 * time.Minute
		case "A cada 1 hora":
			duration = 1 * time.Hour
		case "A cada 6 horas":
			duration = 6 * time.Hour
		case "Diário":
			duration = 24 * time.Hour
		default:
			continue
		}

		// Checa o ultimo backup
		historico, err := s.backupRepo.ListarBackups(ctx, empID)
		var ultimoBackupTime time.Time

		if err == nil && len(historico) > 0 {
			ultimoBackupTime = historico[0].DataBackup
		}

		if ultimoBackupTime.IsZero() || time.Since(ultimoBackupTime) >= duration {
			log.Printf("⏳ Iniciando execução automática de backup para empresa %d (Intervalo: %s)", empID, interval)
			err = s.backupService.ExecutarBackup(ctx, empID, nil, "automatico")
			if err != nil {
				log.Printf("❌ Falha no backup automático da empresa %d: %v", empID, err)
			}
		}
	}
}
