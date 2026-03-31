package models

import (
	"time"
)

type Fornecedor struct {
	ID               int        `json:"id_fornecedor" db:"id_fornecedor"`
	EmpresaID        int        `json:"empresa_id" db:"empresa_id"`
	RazaoSocial      string     `json:"razao_social" db:"razao_social"`
	NomeFantasia     *string    `json:"nome_fantasia,omitempty" db:"nome_fantasia"`
	CNPJ             *string    `json:"cnpj,omitempty" db:"cnpj"`
	InscricaoEstadual *string   `json:"inscricao_estadual,omitempty" db:"inscricao_estadual"`
	Telefone         *string    `json:"telefone,omitempty" db:"telefone"`
	Telefone2        *string    `json:"telefone2,omitempty" db:"telefone2"`
	Email            *string    `json:"email,omitempty" db:"email"`
	Logradouro       *string    `json:"logradouro,omitempty" db:"logradouro"`
	Numero           *string    `json:"numero,omitempty" db:"numero"`
	Bairro           *string    `json:"bairro,omitempty" db:"bairro"`
	Cidade           *string    `json:"cidade,omitempty" db:"cidade"`
	Estado           *string    `json:"estado,omitempty" db:"estado"`
	CEP              *string    `json:"cep,omitempty" db:"cep"`
	NomeContato      *string    `json:"nome_contato,omitempty" db:"nome_contato"`
	TelefoneContato  *string    `json:"telefone_contato,omitempty" db:"telefone_contato"`
	PrazoEntrega     int        `json:"prazo_entrega" db:"prazo_entrega"`
	PrazoPagamento   int        `json:"prazo_pagamento" db:"prazo_pagamento"`
	DataCadastro     time.Time  `json:"data_cadastro" db:"data_cadastro"`
	DataUltimaCompra *time.Time `json:"data_ultima_compra,omitempty" db:"data_ultima_compra"`
	TotalCompras     float64    `json:"total_compras" db:"total_compras"`
	Ativo            bool       `json:"ativo" db:"ativo"`
	Observacoes      *string    `json:"observacoes,omitempty" db:"observacoes"`
}

type CriarFornecedorRequest struct {
	RazaoSocial string  `json:"razao_social"`
	CNPJ        *string `json:"cnpj,omitempty"`
	Telefone    *string `json:"telefone,omitempty"`
	Email       *string `json:"email,omitempty"`
	Endereco    *string `json:"endereco,omitempty"`
}
