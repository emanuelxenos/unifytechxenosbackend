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

type CaixaHandler struct {
	caixaService *service.CaixaService
}

func NewCaixaHandler(db *database.PostgresDB) *CaixaHandler {
	return &CaixaHandler{caixaService: service.NewCaixaService(db)}
}

func (h *CaixaHandler) Status(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	resp, err := h.caixaService.Status(r.Context(), claims.EmpresaID, claims.UserID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSONRaw(w, http.StatusOK, resp)
}

func (h *CaixaHandler) Abrir(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	var req models.AbrirCaixaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	sessao, err := h.caixaService.Abrir(r.Context(), claims.EmpresaID, claims.UserID, req)
	if err != nil {
		utils.Error(w, http.StatusConflict, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"id_sessao":     sessao.ID,
		"codigo_sessao": sessao.CodigoSessao,
		"status":        sessao.Status,
	})
}

func (h *CaixaHandler) Fechar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	var req models.FecharCaixaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	sessao, err := h.caixaService.Fechar(r.Context(), claims.EmpresaID, claims.UserID, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"id_sessao":    sessao.ID,
		"total_vendas": sessao.TotalVendas,
		"diferenca":    sessao.Diferenca,
	})
}

func (h *CaixaHandler) Sangria(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	var req models.SangriaSuprimentoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	if err := h.caixaService.Sangria(r.Context(), claims.EmpresaID, claims.UserID, req); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONMessage(w, http.StatusOK, "Sangria registrada com sucesso")
}

func (h *CaixaHandler) Suprimento(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	var req models.SangriaSuprimentoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	if err := h.caixaService.Suprimento(r.Context(), claims.EmpresaID, claims.UserID, req); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONMessage(w, http.StatusOK, "Suprimento registrado com sucesso")
}
