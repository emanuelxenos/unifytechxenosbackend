package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"erp-backend/internal/api/middleware"
	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
	"erp-backend/internal/service"
	"erp-backend/pkg/utils"
)

type FornecedorHandler struct {
	fornecedorService *service.FornecedorService
}

func NewFornecedorHandler(db *database.PostgresDB) *FornecedorHandler {
	return &FornecedorHandler{fornecedorService: service.NewFornecedorService(db)}
}

func (h *FornecedorHandler) Listar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	fornecedores, err := h.fornecedorService.Listar(r.Context(), claims.EmpresaID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, fornecedores)
}

func (h *FornecedorHandler) Criar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	var req models.CriarFornecedorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}
	if req.RazaoSocial == "" {
		utils.Error(w, http.StatusBadRequest, "Razão social é obrigatória")
		return
	}

	fornecedor, err := h.fornecedorService.Criar(r.Context(), claims.EmpresaID, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusCreated, fornecedor)
}

func (h *FornecedorHandler) Atualizar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req models.CriarFornecedorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	if err := h.fornecedorService.Atualizar(r.Context(), claims.EmpresaID, id, req); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONMessage(w, http.StatusOK, "Fornecedor atualizado com sucesso")
}

func (h *FornecedorHandler) Inativar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.fornecedorService.Inativar(r.Context(), claims.EmpresaID, id); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONMessage(w, http.StatusOK, "Fornecedor inativado com sucesso")
}
