package models

import (
	"time"
)

type Cliente struct {
	ID              int        `json:"id_cliente" db:"id_cliente"`
	EmpresaID       int        `json:"empresa_id" db:"empresa_id"`
	Nome            string     `json:"nome" db:"nome"`
	TipoPessoa      string     `json:"tipo_pessoa" db:"tipo_pessoa"`
	CPFCNPJ         *string    `json:"cpf_cnpj,omitempty" db:"cpf_cnpj"`
	RGIE            *string    `json:"rg_ie,omitempty" db:"rg_ie"`
	Telefone        *string    `json:"telefone,omitempty" db:"telefone"`
	Telefone2       *string    `json:"telefone2,omitempty" db:"telefone2"`
	Email           *string    `json:"email,omitempty" db:"email"`
	Logradouro      *string    `json:"logradouro,omitempty" db:"logradouro"`
	Numero          *string    `json:"numero,omitempty" db:"numero"`
	Complemento     *string    `json:"complemento,omitempty" db:"complemento"`
	Bairro          *string    `json:"bairro,omitempty" db:"bairro"`
	Cidade          *string    `json:"cidade,omitempty" db:"cidade"`
	Estado          *string    `json:"estado,omitempty" db:"estado"`
	CEP             *string    `json:"cep,omitempty" db:"cep"`
	DataNascimento  *time.Time `json:"data_nascimento,omitempty" db:"data_nascimento"`
	LimiteCredito   float64    `json:"limite_credito" db:"limite_credito"`
	SaldoDevedor    float64    `json:"saldo_devedor" db:"saldo_devedor"`
	DataCadastro    time.Time  `json:"data_cadastro" db:"data_cadastro"`
	DataUltimaCompra *time.Time `json:"data_ultima_compra,omitempty" db:"data_ultima_compra"`
	TotalCompras    float64    `json:"total_compras" db:"total_compras"`
	Ativo           bool       `json:"ativo" db:"ativo"`
	Observacoes     *string    `json:"observacoes,omitempty" db:"observacoes"`
}

type CriarClienteRequest struct {
	Nome          string  `json:"nome"`
	TipoPessoa    string  `json:"tipo_pessoa"`
	CPFCNPJ       *string `json:"cpf_cnpj,omitempty"`
	Telefone      *string `json:"telefone,omitempty"`
	Email         *string `json:"email,omitempty"`
	Endereco      *string `json:"endereco,omitempty"`
	LimiteCredito float64 `json:"limite_credito"`
}
