package handlers

import (
	"fmt"
	"net/http"
	"time"

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
	
	// Cabeçalho Premium
	pdf.SetFont("Arial", "B", 22)
	pdf.SetTextColor(41, 128, 185) // Azul escuro UnifyTech
	pdf.CellFormat(190, 12, "UNIFYTECH XENOS ERP", "", 1, "C", false, 0, "")
	
	pdf.SetFont("Arial", "I", 12)
	pdf.SetTextColor(120, 120, 120)

	var titulo string
	var subtitulo string
	
	dataAtual := time.Now().Format("02/01/2006 às 15:04")
	dataMes := time.Now().Format("01/2006")

	switch tipo {
	case "estoque": 
		titulo = "RELATORIO DE ESTOQUE E PATRIMONIO"
		subtitulo = "Posicao atual consolidada em: " + dataAtual
	case "financeiro": 
		titulo = "RESUMO FINANCEIRO"
		subtitulo = "Fotografia de saldo aberto e movimento de caixa em: " + dataAtual
	default: 
		titulo = "RELATORIO DE VENDAS E FATURAMENTO"
		subtitulo = "Acumulado referente ao mes " + dataMes + " (Gerado em: " + dataAtual + ")"
	}
	pdf.CellFormat(190, 8, titulo, "", 1, "C", false, 0, "")
	
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(190, 6, subtitulo, "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Linha Separadora
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(10)

	pdf.SetTextColor(50, 50, 50)
	pdf.SetDrawColor(220, 220, 220)
    
	// Função utilitária para desenhar linhas de tabela
	drawRow := func(label, value string) {
		pdf.SetFont("Arial", "B", 12)
		pdf.SetFillColor(245, 247, 250)
		pdf.CellFormat(90, 12, "  "+label, "1", 0, "L", true, 0, "")
		
		pdf.SetFont("Arial", "", 12)
		pdf.SetFillColor(255, 255, 255)
		pdf.CellFormat(100, 12, "  "+value, "1", 1, "L", true, 0, "")
	}

	if tipo == "estoque" {
		rel, err := h.relatorioService.EstoqueResumo(r.Context(), claims.EmpresaID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		pdf.Ln(5)
		drawRow("Total de Produtos Registrados", fmt.Sprintf("%d itens cadastrados", rel.TotalProdutos))
		drawRow("Alerta de Estoque Baixo", fmt.Sprintf("%d produtos", rel.ProdutosBaixos))
		drawRow("Valor Estimado (Custo)", fmt.Sprintf("R$ %.2f", rel.ValorTotalCusto))
		drawRow("Valor Estimado (Venda)", fmt.Sprintf("R$ %.2f", rel.ValorTotalVenda))
		
	} else if tipo == "financeiro" {
		rel, err := h.relatorioService.FinanceiroResumo(r.Context(), claims.EmpresaID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		pdf.Ln(5)
		drawRow("Contas a Receber (Aberto)", fmt.Sprintf("R$ %.2f", rel.ValorReceberAberto))
		drawRow("Contas a Pagar (Aberto)", fmt.Sprintf("R$ %.2f", rel.ValorPagarAberto))
		drawRow("Evolucao do Caixa no Dia", fmt.Sprintf("R$ %.2f", rel.SaldoCaixaDia))

	} else {
		rel, err := h.relatorioService.VendasMes(r.Context(), claims.EmpresaID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		pdf.Ln(5)
		drawRow("Volume de Vendas", fmt.Sprintf("%d transacoes", rel.TotalVendas))
		drawRow("Ticket Medio", fmt.Sprintf("R$ %.2f", rel.TicketMedio))
		drawRow("Faturamento Bruto", fmt.Sprintf("R$ %.2f", rel.ValorTotal))
	}

	// Rodapé
	pdf.Ln(30)
	pdf.SetFont("Arial", "I", 10)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(190, 10, "Documento gerado automaticamente pelo sistema UniTech Xenos", "", 1, "C", false, 0, "")

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=relatorio_"+tipo+".pdf")
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
