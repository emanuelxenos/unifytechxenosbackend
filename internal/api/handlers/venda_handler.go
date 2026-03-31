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

type VendaHandler struct {
	vendaService *service.VendaService
}

func NewVendaHandler(db *database.PostgresDB) *VendaHandler {
	return &VendaHandler{vendaService: service.NewVendaService(db)}
}

func (h *VendaHandler) Criar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	var req models.CriarVendaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}
	if len(req.Itens) == 0 {
		utils.Error(w, http.StatusBadRequest, "Informe pelo menos um item")
		return
	}

	resp, err := h.vendaService.Criar(r.Context(), claims.EmpresaID, claims.UserID, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONRaw(w, http.StatusCreated, resp)
}

func (h *VendaHandler) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "ID inválido")
		return
	}

	venda, err := h.vendaService.BuscarPorID(r.Context(), claims.EmpresaID, id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, venda)
}

func (h *VendaHandler) Cancelar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req models.CancelarVendaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	if err := h.vendaService.Cancelar(r.Context(), claims.EmpresaID, id, req); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONMessage(w, http.StatusOK, "Venda cancelada com sucesso")
}

func (h *VendaHandler) VendasDia(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	data := r.URL.Query().Get("data")
	var dataPtr *string
	if data != "" {
		dataPtr = &data
	}

	vendas, err := h.vendaService.VendasDia(r.Context(), claims.EmpresaID, dataPtr)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, vendas)
}
