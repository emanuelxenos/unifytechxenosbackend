package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
	Pool *pgxpool.Pool
}

func NewPostgresDB(databaseURL string) (*PostgresDB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear config do banco: %w", err)
	}

	config.MaxConns = 20
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao banco: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("erro ao pingar o banco: %w", err)
	}

	db := &PostgresDB{Pool: pool}

	// Inicializar banco se for a primeira vez
	if err := db.InitializeSchema(ctx); err != nil {
		log.Printf("⚠️ Alerta na inicialização do banco: %v", err)
	}

	return db, nil
}

func (db *PostgresDB) Close() {
	db.Pool.Close()
}

func (db *PostgresDB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.Pool.Ping(ctx)
}

// InitializeSchema verifica se o banco precisa ser instalado e executa o init_db.sql
func (db *PostgresDB) InitializeSchema(ctx context.Context) error {
	lockFile := ".db_installed"
	sqlFile := "init_db.sql"

	// Verifica se o arquivo de trava já existe
	if _, err := os.Stat(lockFile); err == nil {
		return nil // Já instalado, pula
	}

	// Verifica se o arquivo SQL existe
	if _, err := os.Stat(sqlFile); os.IsNotExist(err) {
		return fmt.Errorf("arquivo %s não encontrado para instalação inicial", sqlFile)
	}

	log.Println("📦 Primeira execução detectada. Iniciando instalação do banco de dados...")

	// Lê o script SQL
	content, err := os.ReadFile(sqlFile)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo SQL: %w", err)
	}

	// Executa o script
	_, err = db.Pool.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("erro ao executar script de instalação: %w", err)
	}

	// Cria o arquivo de trava
	err = os.WriteFile(lockFile, []byte(time.Now().Format(time.RFC3339)), 0644)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo de trava %s: %w", lockFile, err)
	}

	log.Println("✅ Banco de dados instalado e configurado com sucesso!")
	return nil
}
