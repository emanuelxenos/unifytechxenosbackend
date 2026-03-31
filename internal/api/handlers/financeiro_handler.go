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

type FinanceiroHandler struct {
	financeiroService *service.FinanceiroService
}

func NewFinanceiroHandler(db *database.PostgresDB) *FinanceiroHandler {
	return &FinanceiroHandler{financeiroService: service.NewFinanceiroService(db)}
}

func (h *FinanceiroHandler) ContasPagar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	status := r.URL.Query().Get("status")
	vInicio := r.URL.Query().Get("vencimento_inicio")
	vFim := r.URL.Query().Get("vencimento_fim")
	var sP, viP, vfP *string
	if status != "" { sP = &status }
	if vInicio != "" { viP = &vInicio }
	if vFim != "" { vfP = &vFim }

	contas, err := h.financeiroService.ListarContasPagar(r.Context(), claims.EmpresaID, sP, viP, vfP)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, contas)
}

func (h *FinanceiroHandler) CriarContaPagar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	var req models.CriarContaPagarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	conta, err := h.financeiroService.CriarContaPagar(r.Context(), claims.EmpresaID, claims.UserID, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusCreated, conta)
}

func (h *FinanceiroHandler) PagarConta(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req models.PagarContaRequest
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.financeiroService.PagarConta(r.Context(), claims.EmpresaID, id, req); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONMessage(w, http.StatusOK, "Conta paga com sucesso")
}

func (h *FinanceiroHandler) ContasReceber(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	contas, err := h.financeiroService.ListarContasReceber(r.Context(), claims.EmpresaID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, contas)
}

func (h *FinanceiroHandler) ReceberConta(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req models.ReceberContaRequest
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.financeiroService.ReceberConta(r.Context(), claims.EmpresaID, id, req); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONMessage(w, http.StatusOK, "Conta recebida com sucesso")
}

func (h *FinanceiroHandler) FluxoCaixa(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	di := r.URL.Query().Get("data_inicio")
	df := r.URL.Query().Get("data_fim")
	var diP, dfP *string
	if di != "" { diP = &di }
	if df != "" { dfP = &df }

	items, err := h.financeiroService.FluxoCaixa(r.Context(), claims.EmpresaID, diP, dfP)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, items)
}
