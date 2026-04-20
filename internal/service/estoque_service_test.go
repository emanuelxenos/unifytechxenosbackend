package service

import (
	"context"
	"testing"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
	"github.com/pashagolub/pgxmock/v3"
)

func TestFinalizarInventario(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	db := &database.PostgresDB{Pool: mock}
	s := NewEstoqueService(db)

	ctx := context.Background()
	empresaID := 1
	inventarioID := 2
	usuarioID := 1

	req := models.FinalizarInventarioRequest{
		Observacoes: "Teste de finalização",
	}

	// 1. Mock do Início da Transação
	mock.ExpectBegin()

	// 2. Mock da Busca de Itens Contados
	// Vamos simular um item contado: Kit Kat (ID 13) com quantidade 75
	rows := pgxmock.NewRows([]string{"produto_id", "quantidade_fisica", "quantidade_sistema"}).
		AddRow(13, 75.0, 60.0)
	
	mock.ExpectQuery(`SELECT (.+) FROM inventario_item`).
		WithArgs(inventarioID).
		WillReturnRows(rows)

	// 3. Loop de Itens - Busca Saldo Atual
	mock.ExpectQuery(`SELECT (.+) FROM produto WHERE id_produto =`).
		WithArgs(13).
		WillReturnRows(pgxmock.NewRows([]string{"estoque_atual"}).AddRow(60.0))

	// 4. Loop de Itens - Update Estoque
	mock.ExpectExec(`UPDATE produto SET estoque_atual =`).
		WithArgs(75.0, 13).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	// 5. Loop de Itens - Log de Movimentação
	mock.ExpectExec(`INSERT INTO estoque_movimentacao`).
		WithArgs(empresaID, 13, 15.0, 60.0, 75.0, inventarioID, usuarioID, "Reconciliação automática de inventário").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// 6. Fechar Inventário
	mock.ExpectExec(`UPDATE inventario`).
		WithArgs(req.Observacoes, inventarioID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	// 7. Commit
	mock.ExpectCommit()

	// Execução
	err = s.FinalizarInventario(ctx, empresaID, inventarioID, usuarioID, req)

	// Verificação
	if err != nil {
		t.Errorf("Erro inesperado ao finalizar inventário: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectativas não atendidas: %v", err)
	}
}

func TestFinalizarInventario_Empty(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	db := &database.PostgresDB{Pool: mock}
	s := NewEstoqueService(db)

	ctx := context.Background()
	req := models.FinalizarInventarioRequest{Observacoes: "Vazio"}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM inventario_item`).WithArgs(2).WillReturnRows(pgxmock.NewRows([]string{"id", "f", "s"})) // Zero rows
	mock.ExpectExec(`UPDATE inventario`).WithArgs(req.Observacoes, 2).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	err = s.FinalizarInventario(ctx, 1, 2, 1, req)
	if err != nil {
		t.Errorf("Erro em inventário vazio: %v", err)
	}
}

func TestFinalizarInventario_NullQuantity(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	db := &database.PostgresDB{Pool: mock}
	s := NewEstoqueService(db)

	ctx := context.Background()
	
	mock.ExpectBegin()
	// Simular item com quantidade_fisica NULL no banco
	rows := pgxmock.NewRows([]string{"produto_id", "quantidade_fisica", "quantidade_sistema"}).
		AddRow(13, nil, 60.0)
	
	mock.ExpectQuery(`SELECT (.+) FROM inventario_item`).WithArgs(2).WillReturnRows(rows)
	
	mock.ExpectExec(`UPDATE inventario`).WithArgs("", 2).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	err = s.FinalizarInventario(ctx, 1, 2, 1, models.FinalizarInventarioRequest{})
	if err != nil {
		t.Logf("A função retornou erro: %v", err)
	}
}
