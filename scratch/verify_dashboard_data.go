package main

import (
	"context"
	"fmt"
	"log"
	"erp-backend/internal/infrastructure/database"
	"erp-backend/pkg/config"
)

func main() {
	cfg := config.Load()
	db, err := database.NewPostgresDB(cfg.DatabaseURL())
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	var totalProdutos int
	var totalCusto, totalVenda float64
	var baixos, vencendo int

	err = db.Pool.QueryRow(ctx, `
		SELECT 
			COUNT(id_produto),
			COALESCE(SUM(CASE WHEN estoque_atual > 0 THEN estoque_atual * preco_custo ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN estoque_atual > 0 THEN estoque_atual * preco_venda ELSE 0 END), 0),
			COUNT(CASE WHEN estoque_atual <= estoque_minimo AND controlar_estoque = TRUE THEN 1 END),
			COUNT(CASE WHEN EXISTS (SELECT 1 FROM estoque_lote WHERE produto_id = id_produto AND status = 'ativo' AND quantidade_atual > 0 AND data_vencimento <= CURRENT_DATE + INTERVAL '15 days') THEN 1 END)
		FROM produto WHERE empresa_id = 1 AND ativo = TRUE
	`).Scan(&totalProdutos, &totalCusto, &totalVenda, &baixos, &vencendo)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n--- 🔍 VERIFICAÇÃO DE DADOS DO DASHBOARD ---")
	fmt.Printf("Total de Produtos: %d\n", totalProdutos)
	fmt.Printf("Valor em Estoque (Custo): R$ %.2f\n", totalCusto)
	fmt.Printf("Potencial de Venda: R$ %.2f\n", totalVenda)
	fmt.Printf("Alerta Estoque Baixo: %d itens\n", baixos)
	fmt.Printf("Alerta Vencendo (15 dias): %d itens\n", vencendo)
	fmt.Println("--------------------------------------------")
}
