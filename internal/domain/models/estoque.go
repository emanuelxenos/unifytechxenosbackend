package models

import (
	"time"
)

type EstoqueMovimentacao struct {
	ID               int       `json:"id_movimentacao" db:"id_movimentacao"`
	EmpresaID        int       `json:"empresa_id" db:"empresa_id"`
	ProdutoID        int       `json:"produto_id" db:"produto_id"`
	TipoMovimentacao string    `json:"tipo_movimentacao" db:"tipo_movimentacao"`
	Quantidade       float64   `json:"quantidade" db:"quantidade"`
	SaldoAnterior    float64   `json:"saldo_anterior" db:"saldo_anterior"`
	SaldoAtual       float64   `json:"saldo_atual" db:"saldo_atual"`
	OrigemTipo       *string   `json:"origem_tipo,omitempty" db:"origem_tipo"`
	OrigemID         *int      `json:"origem_id,omitempty" db:"origem_id"`
	DataMovimentacao time.Time `json:"data_movimentacao" db:"data_movimentacao"`
	UsuarioID        *int      `json:"usuario_id,omitempty" db:"usuario_id"`
	Observacao       *string   `json:"observacao,omitempty" db:"observacao"`

	// Joins
	ProdutoNome *string `json:"produto_nome,omitempty" db:"produto_nome"`
}

type Inventario struct {
	ID             int        `json:"id_inventario" db:"id_inventario"`
	EmpresaID      int        `json:"empresa_id" db:"empresa_id"`
	Codigo         string     `json:"codigo" db:"codigo"`
	Descricao      *string    `json:"descricao,omitempty" db:"descricao"`
	DataInicio     time.Time  `json:"data_inicio" db:"data_inicio"`
	DataFim        *time.Time `json:"data_fim,omitempty" db:"data_fim"`
	DataFechamento *time.Time `json:"data_fechamento,omitempty" db:"data_fechamento"`
	Status         string     `json:"status" db:"status"`
	Observacoes    *string    `json:"observacoes,omitempty" db:"observacoes"`
	UsuarioID      *int       `json:"usuario_id,omitempty" db:"usuario_id"`

	Itens []InventarioItem `json:"itens,omitempty"`
}

type InventarioItem struct {
	ID                 int        `json:"id_inventario_item" db:"id_inventario_item"`
	InventarioID       int        `json:"inventario_id" db:"inventario_id"`
	ProdutoID          int        `json:"produto_id" db:"produto_id"`
	QuantidadeSistema  float64    `json:"quantidade_sistema" db:"quantidade_sistema"`
	QuantidadeFisica   *float64   `json:"quantidade_fisica,omitempty" db:"quantidade_fisica"`
	Diferenca          *float64   `json:"diferenca,omitempty" db:"diferenca"`
	Contado            bool       `json:"contado" db:"contado"`
	DataContagem       *time.Time `json:"data_contagem,omitempty" db:"data_contagem"`
	UsuarioContagemID  *int       `json:"usuario_contagem_id,omitempty" db:"usuario_contagem_id"`
	Observacao         *string    `json:"observacao,omitempty" db:"observacao"`

	// Join
	ProdutoNome *string `json:"produto_nome,omitempty" db:"produto_nome"`
}

type AjusteEstoqueRequest struct {
	ProdutoID  int     `json:"produto_id"`
	Quantidade float64 `json:"quantidade"`
	Tipo       string  `json:"tipo"`
	Motivo     string  `json:"motivo"`
}

type CriarInventarioRequest struct {
	Codigo      string `json:"codigo"`
	Descricao   string `json:"descricao"`
	DataInicio  string `json:"data_inicio"`
	CategoriaID *int   `json:"categoria_id,omitempty"`
}

type FinalizarInventarioRequest struct {
	Ajustes         []AjusteInventarioRequest `json:"ajustes"`
	Observacoes     string                    `json:"observacoes"`
	SupervisorSenha string                    `json:"supervisor_senha"`
}

type AjusteInventarioRequest struct {
	ProdutoID        int     `json:"produto_id"`
	QuantidadeFisica float64 `json:"quantidade_fisica"`
}

type EstoqueBaixoResponse struct {
	IDProduto     int     `json:"id_produto"`
	Nome          string  `json:"nome"`
	EstoqueAtual  float64 `json:"estoque_atual"`
	EstoqueMinimo float64 `json:"estoque_minimo"`
}
