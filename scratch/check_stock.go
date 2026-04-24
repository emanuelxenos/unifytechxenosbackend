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
	
	query := `
		SELECT 
			p.id_produto, 
			p.nome, 
			p.estoque_atual as saldo_produto,
			COALESCE(SUM(l.quantidade_atual), 0) as saldo_lotes,
			(p.estoque_atual - COALESCE(SUM(l.quantidade_atual), 0)) as diferenca
		FROM produto p
		LEFT JOIN estoque_lote l ON p.id_produto = l.produto_id AND l.status = 'ativo'
		WHERE p.controlar_estoque = true AND p.ativo = true
		GROUP BY p.id_produto, p.nome, p.estoque_atual
		HAVING p.estoque_atual != COALESCE(SUM(l.quantidade_atual), 0)
	`
	
	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	fmt.Println("--- RELATÓRIO DE INCONSISTÊNCIA DE ESTOQUE ---")
	found := false
	for rows.Next() {
		found = true
		var id int
		var nome string
		var saldoProd, saldoLotes, diff float64
		rows.Scan(&id, &nome, &saldoProd, &saldoLotes, &diff)
		fmt.Printf("ID: %d | Produto: %-30s | Produto: %.2f | Lotes: %.2f | DIFF: %.2f\n", id, nome, saldoProd, saldoLotes, diff)
	}
	
	if !found {
		fmt.Println("✅ SUCESSO: Todos os saldos de produtos batem com a soma dos lotes!")
	}
	fmt.Println("----------------------------------------------")
}
