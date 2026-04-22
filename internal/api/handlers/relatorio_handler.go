package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

func (h *RelatorioHandler) RelatorioDRE(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	
	// Default para mês/ano atual se não informado
	agora := time.Now()
	mesStr := r.URL.Query().Get("mes")
	anoStr := r.URL.Query().Get("ano")
	
	mes := int(agora.Month())
	ano := agora.Year()
	
	if mesStr != "" {
		fmt.Sscanf(mesStr, "%d", &mes)
	}
	if anoStr != "" {
		fmt.Sscanf(anoStr, "%d", &ano)
	}

	rel, err := h.relatorioService.DREGerencial(r.Context(), claims.EmpresaID, mes, ano)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, rel)
}

func (h *RelatorioHandler) RelatorioInadimplencia(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	rel, err := h.relatorioService.InadimplenciaResumo(r.Context(), claims.EmpresaID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, rel)
}

func (h *RelatorioHandler) RelatorioCurvaABC(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	rel, err := h.relatorioService.CurvaABC(r.Context(), claims.EmpresaID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, rel)
}

func (h *RelatorioHandler) RelatorioComissoes(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	
	agora := time.Now()
	mesStr := r.URL.Query().Get("mes")
	anoStr := r.URL.Query().Get("ano")
	
	mes := int(agora.Month())
	ano := agora.Year()
	
	if mesStr != "" {
		fmt.Sscanf(mesStr, "%d", &mes)
	}
	if anoStr != "" {
		fmt.Sscanf(anoStr, "%d", &ano)
	}

	rel, err := h.relatorioService.ComissoesOperador(r.Context(), claims.EmpresaID, mes, ano)
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
	case "estoque_lista":
		titulo = "LISTAGEM DETALHADA DE ESTOQUE"
		subtitulo = "Relatorio gerado em: " + dataAtual
	case "financeiro": 
		titulo = "RESUMO FINANCEIRO"
		subtitulo = "Fotografia de saldo aberto e movimento de caixa em: " + dataAtual
	case "dre":
		titulo = "DRE - DEMONSTRATIVO DE RESULTADO"
		subtitulo = "Analise gerencial referente ao mes " + dataMes
	case "inadimplencia":
		titulo = "RELATORIO DE INADIMPLENCIA"
		subtitulo = "Contas a receber vencidas ate " + dataAtual
	case "abc":
		titulo = "CURVA ABC DE PRODUTOS"
		subtitulo = "Classificacao de faturamento (Ultimos 90 dias)"
	case "comissoes":
		titulo = "RELATORIO DE COMISSOES"
		subtitulo = "Resumo por operador - Mes " + dataMes
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

	tr := pdf.UnicodeTranslatorFromDescriptor("")

	if tipo == "estoque" {
		rel, err := h.relatorioService.EstoqueResumo(r.Context(), claims.EmpresaID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		pdf.Ln(5)
		drawRow(tr("Total de Produtos Registrados"), tr(fmt.Sprintf("%d itens cadastrados", rel.TotalProdutos)))
		drawRow(tr("Alerta de Estoque Baixo"), tr(fmt.Sprintf("%d produtos", rel.ProdutosBaixos)))
		drawRow(tr("Valor Estimado (Custo)"), tr(fmt.Sprintf("R$ %.2f", rel.ValorTotalCusto)))
		drawRow(tr("Valor Estimado (Venda)"), tr(fmt.Sprintf("R$ %.2f", rel.ValorTotalVenda)))
		
	} else if tipo == "estoque_lista" {
		search := r.URL.Query().Get("search")
		catID, _ := strconv.Atoi(r.URL.Query().Get("categoria_id"))
		baixoEstoque := r.URL.Query().Get("baixo_estoque") == "true"
		vencendo := r.URL.Query().Get("vencendo") == "true"

		produtos, err := h.relatorioService.ListaEstoque(r.Context(), claims.EmpresaID, search, catID, baixoEstoque, vencendo)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Cabeçalho da Tabela
		pdf.SetFont("Arial", "B", 10)
		pdf.SetFillColor(240, 240, 240)
		pdf.CellFormat(80, 10, tr("Produto"), "1", 0, "L", true, 0, "")
		pdf.CellFormat(40, 10, tr("Categoria"), "1", 0, "L", true, 0, "")
		pdf.CellFormat(30, 10, tr("Qtd"), "1", 0, "C", true, 0, "")
		pdf.CellFormat(40, 10, tr("Preço Venda"), "1", 1, "R", true, 0, "")

		pdf.SetFont("Arial", "", 9)
		for _, p := range produtos {
			pdf.CellFormat(80, 8, " "+tr(p.Nome), "1", 0, "L", false, 0, "")
			pdf.CellFormat(40, 8, " "+tr(p.Categoria), "1", 0, "L", false, 0, "")
			pdf.CellFormat(30, 8, fmt.Sprintf("%.2f", p.EstoqueAtual), "1", 0, "C", false, 0, "")
			pdf.CellFormat(40, 8, fmt.Sprintf("R$ %.2f ", p.PrecoVenda), "1", 1, "R", false, 0, "")
		}
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

	} else if tipo == "dre" {
		rel, err := h.relatorioService.DREGerencial(r.Context(), claims.EmpresaID, int(time.Now().Month()), time.Now().Year())
		if err == nil {
			pdf.Ln(5)
			drawRow("Receita Bruta", fmt.Sprintf("R$ %.2f", rel.ReceitaBruta))
			drawRow("Descontos", fmt.Sprintf("R$ %.2f", rel.Descontos))
			drawRow("Receita Liquida", fmt.Sprintf("R$ %.2f", rel.ReceitaLiquida))
			drawRow("CMV (Custo Mercadoria)", fmt.Sprintf("R$ %.2f", rel.CMV))
			drawRow("Lucro Bruto", fmt.Sprintf("R$ %.2f", rel.LucroBruto))
			drawRow("Despesas Operacionais", fmt.Sprintf("R$ %.2f", rel.Despesas))
			drawRow("LUCRO LIQUIDO", fmt.Sprintf("R$ %.2f", rel.LucroLiquido))
			drawRow("Margem de Lucro (%)", fmt.Sprintf("%.2f%%", rel.MargemPercentual))
		}
	} else if tipo == "inadimplencia" {
		rel, err := h.relatorioService.InadimplenciaResumo(r.Context(), claims.EmpresaID)
		if err == nil {
			pdf.Ln(5)
			drawRow("Total de Titulos Vencidos", fmt.Sprintf("%d faturas", rel.Quantidade))
			drawRow("Valor Total em Atraso", fmt.Sprintf("R$ %.2f", rel.TotalVencido))
		}
	} else if tipo == "abc" {
		rel, err := h.relatorioService.CurvaABC(r.Context(), claims.EmpresaID)
		if err == nil {
			pdf.Ln(5)
			drawRow("Faturamento Periodo (90 dias)", fmt.Sprintf("R$ %.2f", rel.TotalFaturamento))
			drawRow("Total de Itens Analisados", fmt.Sprintf("%d produtos", len(rel.Itens)))
		}
	} else if tipo == "comissoes" {
		rel, err := h.relatorioService.ComissoesOperador(r.Context(), claims.EmpresaID, int(time.Now().Month()), time.Now().Year())
		if err == nil {
			pdf.Ln(5)
			drawRow("Vendas do Mes", fmt.Sprintf("R$ %.2f", rel.TotalGeral))
			drawRow("Total de Comissoes", fmt.Sprintf("R$ %.2f", rel.TotalComissao))
		}
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

func (h *RelatorioHandler) Etiqueta(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	produtoID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if produtoID == 0 {
		utils.Error(w, http.StatusBadRequest, "ID do produto é obrigatório")
		return
	}

	// Buscar produto (usando ProdutoService se disponível ou query direta)
	// Para agilizar, farei query direta
	var nome string
	var preco float64
	err := h.relatorioService.GetDB().Pool.QueryRow(r.Context(),
		`SELECT nome, preco_venda FROM produto WHERE id_produto = $1 AND empresa_id = $2`,
		produtoID, claims.EmpresaID).Scan(&nome, &preco)
	
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Produto não encontrado")
		return
	}

	// Tamanho de etiqueta comum: 40x25mm
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size:    gofpdf.SizeType{Wd: 40, Ht: 25},
	})
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetMargins(2, 2, 2)
	pdf.AddPage()

	// Nome do Produto
	pdf.SetFont("Arial", "B", 7)
	pdf.MultiCell(36, 4, tr(nome), "", "C", false)
	
	// Divisor
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(5, 12, 35, 12)

	// Preço
	pdf.SetY(14)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(36, 6, fmt.Sprintf("R$ %.2f", preco), "", 1, "C", false, 0, "")

	// Rodapé UniTech
	pdf.SetY(20)
	pdf.SetFont("Arial", "I", 5)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(36, 3, "UniTech Xenos ERP", "", 1, "C", false, 0, "")

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=etiqueta.pdf")
	pdf.Output(w)
}

func (h *RelatorioHandler) Etiquetas(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	idsStr := r.URL.Query().Get("ids")
	if idsStr == "" {
		utils.Error(w, http.StatusBadRequest, "IDs dos produtos são obrigatórios")
		return
	}

	ids := strings.Split(idsStr, ",")

	// Tamanho de etiqueta comum: 40x25mm
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size:    gofpdf.SizeType{Wd: 40, Ht: 25},
	})
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	for _, idStr := range ids {
		produtoID, _ := strconv.Atoi(idStr)
		if produtoID == 0 {
			continue
		}

		var nome string
		var preco float64
		err := h.relatorioService.GetDB().Pool.QueryRow(r.Context(),
			`SELECT nome, preco_venda FROM produto WHERE id_produto = $1 AND empresa_id = $2`,
			produtoID, claims.EmpresaID).Scan(&nome, &preco)
		
		if err != nil {
			continue
		}

		pdf.SetMargins(2, 2, 2)
		pdf.AddPage()

		// Nome do Produto
		pdf.SetFont("Arial", "B", 7)
		pdf.MultiCell(36, 4, tr(nome), "", "C", false)
		
		// Divisor
		pdf.SetDrawColor(200, 200, 200)
		pdf.Line(5, 12, 35, 12)

		// Preço
		pdf.SetY(14)
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(36, 6, fmt.Sprintf("R$ %.2f", preco), "", 1, "C", false, 0, "")

		// Rodapé UniTech
		pdf.SetY(20)
		pdf.SetFont("Arial", "I", 5)
		pdf.SetTextColor(150, 150, 150)
		pdf.CellFormat(36, 3, "UniTech Xenos ERP", "", 1, "C", false, 0, "")
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=etiquetas_lote.pdf")
	pdf.Output(w)
}

func (h *RelatorioHandler) SugestaoCompra(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	list, err := h.relatorioService.SugestaoCompra(r.Context(), claims.EmpresaID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, list)
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
	} else if tipo == "estoque_lista" {
		search := r.URL.Query().Get("search")
		catID, _ := strconv.Atoi(r.URL.Query().Get("categoria_id"))
		baixoEstoque := r.URL.Query().Get("baixo_estoque") == "true"
		vencendo := r.URL.Query().Get("vencendo") == "true"

		produtos, err := h.relatorioService.ListaEstoque(r.Context(), claims.EmpresaID, search, catID, baixoEstoque, vencendo)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		f.SetCellValue("Sheet1", "A1", "Listagem Detalhada de Estoque")
		f.SetCellValue("Sheet1", "A2", "Produto")
		f.SetCellValue("Sheet1", "B2", "Categoria")
		f.SetCellValue("Sheet1", "C2", "Estoque")
		f.SetCellValue("Sheet1", "D2", "Preço Venda")

		for i, p := range produtos {
			row := i + 3
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), p.Nome)
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), p.Categoria)
			f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), p.EstoqueAtual)
			f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), p.PrecoVenda)
		}
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
	} else if tipo == "dre" {
		rel, _ := h.relatorioService.DREGerencial(r.Context(), claims.EmpresaID, int(time.Now().Month()), time.Now().Year())
		f.SetCellValue("Sheet1", "A1", "DRE Gerencial")
		f.SetCellValue("Sheet1", "A2", "Receita Líquida:")
		f.SetCellValue("Sheet1", "B2", rel.ReceitaLiquida)
		f.SetCellValue("Sheet1", "A3", "CMV:")
		f.SetCellValue("Sheet1", "B3", rel.CMV)
		f.SetCellValue("Sheet1", "A4", "Lucro Líquido:")
		f.SetCellValue("Sheet1", "B4", rel.LucroLiquido)
	} else if tipo == "abc" {
		rel, _ := h.relatorioService.CurvaABC(r.Context(), claims.EmpresaID)
		f.SetCellValue("Sheet1", "A1", "Curva ABC de Produtos")
		f.SetCellValue("Sheet1", "A2", "Produto")
		f.SetCellValue("Sheet1", "B2", "Faturamento")
		f.SetCellValue("Sheet1", "C2", "Classificação")
		for i, item := range rel.Itens {
			row := i + 3
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), item.Nome)
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), item.Faturamento)
			f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), item.Classificacao)
		}
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

func (h *RelatorioHandler) PerformanceProduto(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	produtoID, _ := strconv.Atoi(r.URL.Query().Get("id"))

	if produtoID == 0 {
		utils.Error(w, http.StatusBadRequest, "ID do produto é obrigatório")
		return
	}

	data, err := h.relatorioService.PerformanceProduto(r.Context(), claims.EmpresaID, produtoID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, data)
}

func (h *RelatorioHandler) AuditoriaMovimentacao(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	produtoID, _ := strconv.Atoi(r.URL.Query().Get("id"))

	if produtoID == 0 {
		utils.Error(w, http.StatusBadRequest, "ID do produto é obrigatório")
		return
	}

	data, err := h.relatorioService.AuditoriaMovimentacao(r.Context(), claims.EmpresaID, produtoID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, data)
}
