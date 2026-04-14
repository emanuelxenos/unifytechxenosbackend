package models

import (
	"time"
)

type ContaPagar struct {
	ID              int        `json:"id_conta_pagar" db:"id_conta_pagar"`
	EmpresaID       int        `json:"empresa_id" db:"empresa_id"`
	FornecedorID    *int       `json:"fornecedor_id,omitempty" db:"fornecedor_id"`
	CompraID        *int       `json:"compra_id,omitempty" db:"compra_id"`
	Descricao       string     `json:"descricao" db:"descricao"`
	Documento       *string    `json:"documento,omitempty" db:"documento"`
	Parcela         *string    `json:"parcela,omitempty" db:"parcela"`
	ValorOriginal   float64    `json:"valor_original" db:"valor_original"`
	ValorPago       float64    `json:"valor_pago" db:"valor_pago"`
	DataVencimento  time.Time  `json:"data_vencimento" db:"data_vencimento"`
	DataPagamento   *time.Time `json:"data_pagamento,omitempty" db:"data_pagamento"`
	Status          string     `json:"status" db:"status"`
	Categoria       string     `json:"categoria" db:"categoria"`
	Observacoes     *string    `json:"observacoes,omitempty" db:"observacoes"`
	DataCadastro    time.Time  `json:"data_cadastro" db:"data_cadastro"`
	UsuarioID       *int       `json:"usuario_id,omitempty" db:"usuario_id"`

	// Join
	FornecedorNome *string `json:"fornecedor_nome,omitempty" db:"fornecedor_nome"`
}

type ContaReceber struct {
	ID              int        `json:"id_conta_receber" db:"id_conta_receber"`
	EmpresaID       int        `json:"empresa_id" db:"empresa_id"`
	ClienteID       *int       `json:"cliente_id,omitempty" db:"cliente_id"`
	VendaID         *int       `json:"venda_id,omitempty" db:"venda_id"`
	Descricao       string     `json:"descricao" db:"descricao"`
	Parcela         *string    `json:"parcela,omitempty" db:"parcela"`
	ValorOriginal   float64    `json:"valor_original" db:"valor_original"`
	ValorRecebido   float64    `json:"valor_recebido" db:"valor_recebido"`
	DataVencimento  time.Time  `json:"data_vencimento" db:"data_vencimento"`
	DataRecebimento *time.Time `json:"data_recebimento,omitempty" db:"data_recebimento"`
	Status          string     `json:"status" db:"status"`
	Observacoes     *string    `json:"observacoes,omitempty" db:"observacoes"`
	DataCadastro    time.Time  `json:"data_cadastro" db:"data_cadastro"`
	UsuarioID       *int       `json:"usuario_id,omitempty" db:"usuario_id"`

	// Join
	ClienteNome *string `json:"cliente_nome,omitempty" db:"cliente_nome"`
}

type CriarContaPagarRequest struct {
	Descricao      string  `json:"descricao"`
	ValorOriginal  float64 `json:"valor_original"`
	DataVencimento string  `json:"data_vencimento"`
	FornecedorID   *int    `json:"fornecedor_id,omitempty"`
	Categoria      string  `json:"categoria"`
}

type PagarContaRequest struct {
	ValorPago     float64 `json:"valor_pago"`
	DataPagamento string  `json:"data_pagamento"`
}

type ReceberContaRequest struct {
	ValorRecebido   float64 `json:"valor_recebido"`
	DataRecebimento string  `json:"data_recebimento"`
}

type FluxoCaixaItem struct {
	Data  time.Time `json:"data" db:"data"`
	Tipo  string    `json:"tipo" db:"tipo"`
	Valor float64   `json:"valor" db:"valor"`
}

type FluxoCaixaResponse struct {
	Items        []FluxoCaixaItem `json:"items"`
	TotalEntrada float64          `json:"total_entrada"`
	TotalSaida   float64          `json:"total_saida"`
	Saldo        float64          `json:"saldo"`
}
