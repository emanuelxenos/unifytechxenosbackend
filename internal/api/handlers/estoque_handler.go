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

type EstoqueHandler struct {
	estoqueService *service.EstoqueService
}

func NewEstoqueHandler(db *database.PostgresDB) *EstoqueHandler {
	return &EstoqueHandler{estoqueService: service.NewEstoqueService(db)}
}

func (h *EstoqueHandler) EstoqueBaixo(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	produtos, err := h.estoqueService.EstoqueBaixo(r.Context(), claims.EmpresaID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, produtos)
}

func (h *EstoqueHandler) Ajuste(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	var req models.AjusteEstoqueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	if err := h.estoqueService.Ajuste(r.Context(), claims.EmpresaID, claims.UserID, req); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONMessage(w, http.StatusOK, "Estoque ajustado com sucesso")
}

func (h *EstoqueHandler) CriarInventario(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	var req models.CriarInventarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	inv, err := h.estoqueService.CriarInventario(r.Context(), claims.EmpresaID, claims.UserID, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusCreated, inv)
}

func (h *EstoqueHandler) FinalizarInventario(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req models.FinalizarInventarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	if err := h.estoqueService.FinalizarInventario(r.Context(), claims.EmpresaID, id, claims.UserID, req); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONMessage(w, http.StatusOK, "Inventário finalizado com sucesso")
}

func (h *EstoqueHandler) ListarMovimentacoes(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	
	produtoID, _ := strconv.Atoi(r.URL.Query().Get("produto_id"))
	dataInicio := r.URL.Query().Get("data_inicio")
	dataFim := r.URL.Query().Get("data_fim")

	movs, err := h.estoqueService.ListarMovimentacoes(r.Context(), claims.EmpresaID, produtoID, dataInicio, dataFim)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, movs)
}

func (h *EstoqueHandler) ListarInventarios(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	
	dataInicio := r.URL.Query().Get("inicio")
	dataFim := r.URL.Query().Get("fim")

	invs, err := h.estoqueService.ListarInventarios(r.Context(), claims.EmpresaID, dataInicio, dataFim)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, invs)
}

func (h *EstoqueHandler) BuscarInventarioPorId(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "ID inválido")
		return
	}

	inv, err := h.estoqueService.BuscarInventarioPorId(r.Context(), claims.EmpresaID, id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, inv)
}

func (h *EstoqueHandler) AtualizarItemInventario(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	invID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	prodID, _ := strconv.Atoi(chi.URLParam(r, "prodId"))

	var req struct {
		Quantidade float64 `json:"quantidade"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	err := h.estoqueService.AtualizarItemInventario(r.Context(), claims.EmpresaID, invID, prodID, req.Quantidade, claims.UserID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONMessage(w, http.StatusOK, "Item atualizado com sucesso")
}
