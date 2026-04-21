package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RelatorioEstoque struct {
	TotalProdutos      int     `json:"total_produtos"`
	ValorTotalCusto    float64 `json:"valor_total_custo"`
	ValorTotalVenda    float64 `json:"valor_total_venda"`
	ProdutosBaixos     int     `json:"produtos_baixo_estoque"`
	SugestaoCompraTotal float64 `json:"sugestao_compra_total"`
	ProdutosVencendo   int     `json:"produtos_vencendo"`
}

func main() {
	connStr := "postgres://postgres:admin123@localhost:5432/mercado_db?sslmode=disable"
	ctx := context.Background()
	
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco: %v", err)
	}
	defer pool.Close()

	empresaID := 1
	rel := &RelatorioEstoque{}
	
	err = pool.QueryRow(ctx,
		`SELECT 
			COUNT(id_produto),
			COALESCE(SUM(estoque_atual * preco_custo), 0),
			COALESCE(SUM(estoque_atual * preco_venda), 0),
			COUNT(CASE WHEN estoque_atual <= estoque_minimo AND controlar_estoque = TRUE THEN 1 END),
			COALESCE(SUM(CASE WHEN estoque_atual < estoque_minimo THEN (estoque_minimo - estoque_atual) * preco_custo ELSE 0 END), 0),
			COUNT(CASE WHEN data_vencimento <= CURRENT_DATE + INTERVAL '15 days' THEN 1 END)
		 FROM produto WHERE empresa_id = $1 AND ativo = TRUE`,
		empresaID).Scan(&rel.TotalProdutos, &rel.ValorTotalCusto, &rel.ValorTotalVenda, &rel.ProdutosBaixos, &rel.SugestaoCompraTotal, &rel.ProdutosVencendo)

	if err != nil {
		log.Fatalf("Erro na query: %v", err)
	}

	jsonData, _ := json.MarshalIndent(rel, "", "  ")
	fmt.Printf("📦 JSON que o Backend enviaria:\n%s\n", string(jsonData))
}
