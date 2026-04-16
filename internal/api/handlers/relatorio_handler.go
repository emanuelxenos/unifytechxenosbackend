package handlers

import (
	"fmt"
	"net/http"

	"erp-backend/internal/api/middleware"
	"erp-backend/internal/infrastructure/database"
	"erp-backend/internal/service"
	"erp-backend/pkg/utils"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
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

func (h *RelatorioHandler) EstoqueResumo(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	rel, err := h.relatorioService.EstoqueResumo(r.Context(), claims.EmpresaID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, rel)
}

func (h *RelatorioHandler) FinanceiroResumo(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	rel, err := h.relatorioService.FinanceiroResumo(r.Context(), claims.EmpresaID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, rel)
}

func (h *RelatorioHandler) ExportarPDF(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	tipo := r.URL.Query().Get("tipo")

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	if tipo == "estoque" {
		rel, err := h.relatorioService.EstoqueResumo(r.Context(), claims.EmpresaID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		pdf.Cell(40, 10, "Relatorio de Estoque")
		pdf.Ln(10)
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(40, 10, fmt.Sprintf("Total de Produtos: %d", rel.TotalProdutos))
		pdf.Ln(10)
		pdf.Cell(40, 10, fmt.Sprintf("Produtos Estoque Baixo: %d", rel.ProdutosBaixos))
		pdf.Ln(10)
		pdf.Cell(40, 10, fmt.Sprintf("Valor Total (Custo): R$ %.2f", rel.ValorTotalCusto))
	} else if tipo == "financeiro" {
		rel, err := h.relatorioService.FinanceiroResumo(r.Context(), claims.EmpresaID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		pdf.Cell(40, 10, "Resumo Financeiro")
		pdf.Ln(10)
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(40, 10, fmt.Sprintf("A Receber (Aberto): R$ %.2f", rel.ValorReceberAberto))
		pdf.Ln(10)
		pdf.Cell(40, 10, fmt.Sprintf("A Pagar (Aberto): R$ %.2f", rel.ValorPagarAberto))
		pdf.Ln(10)
		pdf.Cell(40, 10, fmt.Sprintf("Saldo de Caixa: R$ %.2f", rel.SaldoCaixaDia))
	} else {
		rel, err := h.relatorioService.VendasMes(r.Context(), claims.EmpresaID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		pdf.Cell(40, 10, "Relatorio de Vendas")
		pdf.Ln(10)
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(40, 10, fmt.Sprintf("Total de Vendas: %d", rel.TotalVendas))
		pdf.Ln(10)
		pdf.Cell(40, 10, fmt.Sprintf("Valor Total: R$ %.2f", rel.ValorTotal))
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=relatorio_" + tipo + ".pdf")
	err := pdf.Output(w)
	if err != nil {
		http.Error(w, "Erro ao gerar PDF", http.StatusInternalServerError)
	}
}

func (h *RelatorioHandler) ExportarExcel(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	tipo := r.URL.Query().Get("tipo")

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	if tipo == "estoque" {
		rel, err := h.relatorioService.EstoqueResumo(r.Context(), claims.EmpresaID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		f.SetCellValue("Sheet1", "A1", "Relatório de Estoque")
		f.SetCellValue("Sheet1", "A2", "Total de Produtos:")
		f.SetCellValue("Sheet1", "B2", rel.TotalProdutos)
		f.SetCellValue("Sheet1", "A3", "Produtos Estoque Baixo:")
		f.SetCellValue("Sheet1", "B3", rel.ProdutosBaixos)
		f.SetCellValue("Sheet1", "A4", "Valor Total (Custo):")
		f.SetCellValue("Sheet1", "B4", rel.ValorTotalCusto)
	} else if tipo == "financeiro" {
		rel, err := h.relatorioService.FinanceiroResumo(r.Context(), claims.EmpresaID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		f.SetCellValue("Sheet1", "A1", "Resumo Financeiro")
		f.SetCellValue("Sheet1", "A2", "A Receber (Aberto):")
		f.SetCellValue("Sheet1", "B2", rel.ValorReceberAberto)
		f.SetCellValue("Sheet1", "A3", "A Pagar (Aberto):")
		f.SetCellValue("Sheet1", "B3", rel.ValorPagarAberto)
		f.SetCellValue("Sheet1", "A4", "Saldo de Caixa:")
		f.SetCellValue("Sheet1", "B4", rel.SaldoCaixaDia)
	} else {
		rel, err := h.relatorioService.VendasMes(r.Context(), claims.EmpresaID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		f.SetCellValue("Sheet1", "A1", "Relatório de Vendas")
		f.SetCellValue("Sheet1", "A2", "Total de Vendas:")
		f.SetCellValue("Sheet1", "B2", rel.TotalVendas)
		f.SetCellValue("Sheet1", "A3", "Valor Total:")
		f.SetCellValue("Sheet1", "B3", rel.ValorTotal)
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=relatorio_" + tipo + ".xlsx")
	err := f.Write(w)
	if err != nil {
		http.Error(w, "Erro ao gerar Excel", http.StatusInternalServerError)
	}
}
