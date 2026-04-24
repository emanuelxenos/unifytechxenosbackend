package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"erp-backend/internal/infrastructure/database"
	"erp-backend/internal/service"
	"erp-backend/internal/domain/models"
	"erp-backend/pkg/config"
)

func main() {
	cfg := config.Load()
	db, err := database.NewPostgresDB(cfg.DatabaseURL())
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	
	compraSvc := service.NewCompraService(db)
	empresaID := 1
	usuarioID := 1
	produtoID := 15 // Produto de teste

	fmt.Println("🚀 Iniciando Teste de Fluxo de Compra e Recebimento...")

	// 1. Criar Compra
	reqCompra := models.CriarCompraRequest{
		FornecedorID: 1,
		NumeroNotaFiscal: "TESTE-RATREIO-99",
		DataEmissao: time.Now().Format("2006-01-02"),
		Itens: []models.CriarItemCompraRequest{
			{
				ProdutoID: produtoID,
				Quantidade: 50,
				PrecoUnitario: 10.5,
				Localizacao: "PRATELEIRA-A1",
				Lote: "LOTE-COMPRA-ORIGINAL",
				DataVencimento: "2027-12-31",
			},
		},
	}

	compra, err := compraSvc.Criar(ctx, empresaID, usuarioID, reqCompra)
	if err != nil {
		log.Fatalf("❌ Erro ao criar compra: %v", err)
	}
	fmt.Printf("✅ Compra ID %d criada com sucesso.\n", compra.ID)

	// 2. Receber Compra (Simulando o novo diálogo do Flutter)
	// Vamos mudar a validade no recebimento para testar se ele sobrescreve a do pedido
	validadeRecebimento := time.Now().AddDate(2, 0, 0) // +2 anos
	
	loteFab := "LOTE-CONFIRMADO-PORTARIA"
	vencStr := validadeRecebimento.Format(time.RFC3339)
	
	reqReceber := models.ReceberCompraRequest{
		ItensRecebidos: []models.ItemRecebidoRequest{
			{
				ProdutoID:          produtoID,
				QuantidadeRecebida: 50,
				LoteFabricante:     &loteFab,
				DataVencimento:     &vencStr,
			},
		},
	}

	err = compraSvc.Receber(ctx, empresaID, compra.ID, usuarioID, reqReceber)
	if err != nil {
		log.Fatalf("❌ Erro ao receber compra: %v", err)
	}
	fmt.Println("✅ Recebimento processado com sucesso.")

	// 3. Validação dos Dados
	var loteLote, loteProd string
	var vencLote, vencProd time.Time
	
	err = db.Pool.QueryRow(ctx, 
		`SELECT lote_fabricante, data_vencimento FROM estoque_lote WHERE produto_id = $1 ORDER BY id_lote DESC LIMIT 1`, 
		produtoID).Scan(&loteLote, &vencLote)
	
	_ = db.Pool.QueryRow(ctx, 
		`SELECT COALESCE(localizacao, 'N/A'), data_vencimento FROM produto WHERE id_produto = $1`, 
		produtoID).Scan(&loteProd, &vencProd)

	fmt.Println("\n--- RESULTADOS DA AUDITORIA DO TESTE ---")
	fmt.Printf("Lote no Lote: %s | Vencimento no Lote: %s\n", loteLote, vencLote.Format("02/01/2006"))
	fmt.Printf("Cache no Produto: %s | Vencimento no Produto: %s\n", loteProd, vencProd.Format("02/01/2006"))
	
	if loteLote == "LOTE-CONFIRMADO-PORTARIA" && vencLote.Format("2006-01-02") == validadeRecebimento.Format("2006-01-02") {
		fmt.Println("\n🏆 TESTE APROVADO: O sistema capturou e salvou o lote/validade reais do recebimento!")
	} else {
		fmt.Println("\n❌ TESTE FALHOU: Os dados não conferem.")
	}
	fmt.Println("----------------------------------------")
}
