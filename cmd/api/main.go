package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"erp-backend/internal/api"
	"erp-backend/internal/infrastructure/database"
	ws "erp-backend/internal/infrastructure/websocket"
	"erp-backend/internal/repository"
	"erp-backend/internal/service"
	"erp-backend/pkg/config"
	"erp-backend/pkg/utils"
)

func main() {
	// Carregar configurações
	cfg := config.Load()

	// Configurar JWT
	utils.SetJWTSecret(cfg.JWTSecret)

	// Conectar ao banco
	db, err := database.NewPostgresDB(cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer db.Close()
	log.Println("✅ Conectado ao PostgreSQL")

	// Iniciar WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()
	log.Println("✅ WebSocket Hub iniciado")

	// Configurar e Iniciar Scheduler de Backups Automáticos
	configSvc := service.NewConfigService(db)
	backupRepo := repository.NewBackupRepository(db)
	backupSvc := service.NewBackupService(cfg, backupRepo, configSvc)
	backupScheduler := service.NewBackupScheduler(db, configSvc, backupSvc, backupRepo)
	go backupScheduler.Start(context.Background())

	// Configurar rotas
	router := api.NewRouter(db, cfg, hub)

	// Iniciar servidor
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("🚀 Servidor iniciado em http://localhost%s", addr)
	log.Printf("📋 Ambiente: %s | Versão: %s", cfg.AppEnv, cfg.AppVersion)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
