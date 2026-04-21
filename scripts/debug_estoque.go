package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	connStr := "postgres://postgres:admin123@localhost:5432/mercado_db?sslmode=disable"
	ctx := context.Background()
	
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco: %v", err)
	}
	defer pool.Close()

	// 1. Verificar data atual do banco
	var dbDate time.Time
	err = pool.QueryRow(ctx, "SELECT CURRENT_DATE").Scan(&dbDate)
	fmt.Printf("📅 Data Atual do Banco: %s\n", dbDate.Format("02/01/2006"))

	// 2. Listar produtos e suas empresas
	fmt.Println("\n📊 Produtos no Banco:")
	rows, _ := pool.Query(ctx, "SELECT id_produto, empresa_id, nome, data_vencimento FROM produto WHERE ativo = TRUE")
	
	companyCounts := make(map[int]int)
	vencendoCounts := make(map[int]int)
	
	limitDate := dbDate.AddDate(0, 0, 15)

	for rows.Next() {
		var id, empID int
		var nome string
		var vcto *time.Time
		rows.Scan(&id, &empID, &nome, &vcto)
		
		companyCounts[empID]++
		
		if vcto != nil && (vcto.Before(limitDate) || vcto.Equal(limitDate)) {
			vencendoCounts[empID]++
		}
		
		vStr := "N/A"
		if vcto != nil {
			vStr = vcto.Format("02/01/2006")
		}
		fmt.Printf("[%d] Empresa: %d | Item: %-20s | Vcto: %s\n", id, empID, nome, vStr)
	}

	fmt.Println("\n📈 Resumo por Empresa:")
	for empID, count := range companyCounts {
		fmt.Printf("Empresa %d: Total=%d | Vencendo=%d\n", empID, count, vencendoCounts[empID])
	}
}
