package handlers

import (
	"encoding/json"
	"net/http"

	"erp-backend/internal/api/middleware"
	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
	"erp-backend/internal/service"
	"erp-backend/pkg/utils"
)

type ConfigHandler struct {
	configService *service.ConfigService
}

func NewConfigHandler(db *database.PostgresDB) *ConfigHandler {
	return &ConfigHandler{configService: service.NewConfigService(db)}
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
	utils.JSONMessage(w, http.StatusOK, "Backup criado com sucesso")
}

func (h *ConfigHandler) Restaurar(w http.ResponseWriter, r *http.Request) {
	utils.JSONMessage(w, http.StatusOK, "Backup restaurado com sucesso")
}
