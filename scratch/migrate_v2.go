package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbPort == "" {
		dbPort = "5432"
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
	log.Printf("Conectando ao banco: %s", dbURL)

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Erro ao parsear config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()
	queries := []string{
		"ALTER TABLE produto ADD COLUMN IF NOT EXISTS localizacao VARCHAR(100);",
		"ALTER TABLE produto ADD COLUMN IF NOT EXISTS data_vencimento DATE;",
	}

	for _, q := range queries {
		log.Printf("Executando: %s", q)
		_, err := pool.Exec(ctx, q)
		if err != nil {
			log.Printf("⚠️ Erro (pode ser esperado se a coluna já existir): %v", err)
		} else {
			log.Printf("✅ Sucesso!")
		}
	}

	log.Println("Migração concluída.")
}
