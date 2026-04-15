package handlers

import (
	"encoding/json"
	"net/http"

	"erp-backend/internal/api/middleware"
	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
	"erp-backend/internal/repository"
	"erp-backend/internal/service"
	"erp-backend/pkg/config"
	"erp-backend/pkg/utils"
)

type ConfigHandler struct {
	configService *service.ConfigService
	backupService *service.BackupService
}

func NewConfigHandler(db *database.PostgresDB, cfg *config.Config) *ConfigHandler {
	configSvc := service.NewConfigService(db)
	backupRepo := repository.NewBackupRepository(db)
	backupSvc := service.NewBackupService(cfg, backupRepo, configSvc)
	return &ConfigHandler{
		configService: configSvc,
		backupService: backupSvc,
	}
}

func (h *ConfigHandler) Listar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	configs, err := h.configService.Listar(r.Context(), claims.EmpresaID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, configs)
}

func (h *ConfigHandler) Atualizar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	var req models.AtualizarConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	if err := h.configService.Atualizar(r.Context(), claims.EmpresaID, req); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONMessage(w, http.StatusOK, "Configurações atualizadas com sucesso")
}

func (h *ConfigHandler) Backup(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	
	err := h.backupService.ExecutarBackup(r.Context(), claims.EmpresaID, &claims.UserID, "manual")
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSONMessage(w, http.StatusOK, "Backup iniciado e concluído com sucesso")
}

func (h *ConfigHandler) ListarBackups(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	backups, err := h.backupService.ListarBackups(r.Context(), claims.EmpresaID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Erro ao listar backups")
		return
	}
	utils.JSON(w, http.StatusOK, backups)
}

func (h *ConfigHandler) Restaurar(w http.ResponseWriter, r *http.Request) {
	utils.JSONMessage(w, http.StatusOK, "Restaurar backup será implementado em breve")
}
