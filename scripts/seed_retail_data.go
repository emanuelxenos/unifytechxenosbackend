package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var locations = []string{
	"Corredor A - Gôndola 1",
	"Corredor A - Gôndola 2",
	"Corredor B - Gôndola 1",
	"Corredor B - Gôndola 2",
	"Setor de Limpeza - B1",
	"Setor de Alimentos - C3",
	"Câmara Fria - Carne",
	"Depósito Central",
	"Expositor Frontal",
}

func main() {
	// Configuração do banco (carregada do .env manualmente para esse script)
	// DB_PASS=admin123
	connStr := "postgres://postgres:admin123@localhost:5432/mercado_db?sslmode=disable"
	ctx := context.Background()
	
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco: %v", err)
	}
	defer pool.Close()

	// 1. Buscar IDs de todos os produtos
	rows, err := pool.Query(ctx, "SELECT id_produto, nome FROM produto")
	if err != nil {
		log.Fatalf("Erro ao buscar produtos: %v", err)
	}
	defer rows.Close()

	type Product struct {
		ID   int
		Nome string
	}
	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Nome); err != nil {
			log.Printf("Erro ao ler linha: %v", err)
			continue
		}
		products = append(products, p)
	}

	fmt.Printf("📦 Encontrados %d produtos para atualizar.\n", len(products))

	// 2. Atualizar cada produto com dados aleatórios
	rand.Seed(time.Now().UnixNano())
	now := time.Now()

	for _, p := range products {
		loc := locations[rand.Intn(len(locations))]
		
		// Lógica de vencimento:
		// 20% vencido (passado)
		// 20% alerta (próximos 15 dias)
		// 20% aviso (próximos 30 dias)
		// 40% seguro (futuro distante)
		
		var vencimento time.Time
		r := rand.Intn(100)
		if r < 20 {
			// Vencido: entre 1 e 6 meses atrás
			vencimento = now.AddDate(0, -rand.Intn(6)-1, -rand.Intn(28))
		} else if r < 40 {
			// Alerta: próximos 15 dias
			vencimento = now.AddDate(0, 0, rand.Intn(15))
		} else if r < 60 {
			// Aviso: próximos 30 dias
			vencimento = now.AddDate(0, 0, 15+rand.Intn(15))
		} else {
			// Seguro: 3 a 12 meses no futuro
			vencimento = now.AddDate(0, rand.Intn(9)+3, rand.Intn(28))
		}

		_, err := pool.Exec(ctx, "UPDATE produto SET localizacao = $1, data_vencimento = $2 WHERE id_produto = $3",
			loc, vencimento, p.ID)
		
		if err != nil {
			log.Printf("❌ Erro ao atualizar produto %s: %v", p.Nome, err)
		} else {
			fmt.Printf("✅ Atualizado: %-30s | Local: %-20s | Vcto: %s\n", 
				p.Nome, loc, vencimento.Format("02/01/2006"))
		}
	}

	fmt.Println("\n✨ Atualização concluída com sucesso!")
}
