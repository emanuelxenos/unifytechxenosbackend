package handlers

import (
	"net/http"

	"erp-backend/internal/api/middleware"
	"erp-backend/internal/infrastructure/database"
	"erp-backend/internal/service"
	"erp-backend/pkg/utils"
)

type RelatorioHandler struct {
	relatorioService *service.RelatorioService
}

func NewRelatorioHandler(db *database.PostgresDB) *RelatorioHandler {
	return &RelatorioHandler{relatorioService: service.NewRelatorioService(db)}
}

func (h *RelatorioHandler) VendasDia(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	rel, err := h.relatorioService.VendasDia(r.Context(), claims.EmpresaID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, rel)
}

func (h *RelatorioHandler) VendasMes(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	rel, err := h.relatorioService.VendasMes(r.Context(), claims.EmpresaID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, rel)
}

func (h *RelatorioHandler) VendasPeriodo(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	di := r.URL.Query().Get("data_inicio")
	df := r.URL.Query().Get("data_fim")
	if di == "" || df == "" {
		utils.Error(w, http.StatusBadRequest, "data_inicio e data_fim são obrigatórios")
		return
	}

	rel, err := h.relatorioService.VendasPeriodo(r.Context(), claims.EmpresaID, di, df)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, rel)
}

func (h *RelatorioHandler) MaisVendidos(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	periodo := r.URL.Query().Get("periodo")
	if periodo == "" {
		periodo = "30d"
	}

	produtos, err := h.relatorioService.MaisVendidos(r.Context(), claims.EmpresaID, periodo)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, produtos)
}

func (h *RelatorioHandler) ExportarPDF(w http.ResponseWriter, r *http.Request) {
	utils.JSONMessage(w, http.StatusOK, "Exportação PDF disponível em breve")
}

func (h *RelatorioHandler) ExportarExcel(w http.ResponseWriter, r *http.Request) {
	utils.JSONMessage(w, http.StatusOK, "Exportação Excel disponível em breve")
}
