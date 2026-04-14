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

type EmpresaHandler struct {
	empresaService *service.EmpresaService
}

func NewEmpresaHandler(db *database.PostgresDB) *EmpresaHandler {
	return &EmpresaHandler{empresaService: service.NewEmpresaService(db)}
}

func (h *EmpresaHandler) Buscar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		utils.Error(w, http.StatusUnauthorized, "Não autenticado")
		return
	}

	empresa, err := h.empresaService.Buscar(r.Context(), claims.EmpresaID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Empresa não encontrada")
		return
	}

	utils.JSON(w, http.StatusOK, empresa)
}

func (h *EmpresaHandler) Atualizar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		utils.Error(w, http.StatusUnauthorized, "Não autenticado")
		return
	}

	var empresa models.Empresa
	if err := json.NewDecoder(r.Body).Decode(&empresa); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	// Forçar o ID da empresa vindo do token para segurança
	empresa.ID = claims.EmpresaID

	if err := h.empresaService.Atualizar(r.Context(), &empresa); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONMessage(w, http.StatusOK, "Configurações da empresa atualizadas com sucesso")
}
