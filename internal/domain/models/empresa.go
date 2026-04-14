package models

import (
	"time"
)

type Empresa struct {
	ID                 int        `json:"id_empresa" db:"id_empresa"`
	RazaoSocial        string     `json:"razao_social" db:"razao_social"`
	NomeFantasia       string     `json:"nome_fantasia" db:"nome_fantasia"`
	CNPJ               string     `json:"cnpj" db:"cnpj"`
	InscricaoEstadual  *string    `json:"inscricao_estadual,omitempty" db:"inscricao_estadual"`
	InscricaoMunicipal *string    `json:"inscricao_municipal,omitempty" db:"inscricao_municipal"`
	Logradouro         string     `json:"logradouro" db:"logradouro"`
	Numero             string     `json:"numero" db:"numero"`
	Complemento        *string    `json:"complemento,omitempty" db:"complemento"`
	Bairro             string     `json:"bairro" db:"bairro"`
	Cidade             string     `json:"cidade" db:"cidade"`
	Estado             string     `json:"estado" db:"estado"`
	CEP                string     `json:"cep" db:"cep"`
	Telefone           string     `json:"telefone" db:"telefone"`
	Telefone2          *string    `json:"telefone2,omitempty" db:"telefone2"`
	Email              string     `json:"email" db:"email"`
	Site               *string    `json:"site,omitempty" db:"site"`
	RegimeTributario   string     `json:"regime_tributario" db:"regime_tributario"`
	Moeda              string     `json:"moeda" db:"moeda"`
	CasasDecimais      int        `json:"casas_decimais" db:"casas_decimais"`
	FusoHorario        string     `json:"fuso_horario" db:"fuso_horario"`
	LogotipoURL        *string    `json:"logotipo_url,omitempty" db:"logotipo_url"`
	CorPrimaria        string     `json:"cor_primaria" db:"cor_primaria"`
	CorSecundaria      string     `json:"cor_secundaria" db:"cor_secundaria"`
	Ativo              bool       `json:"ativo" db:"ativo"`
	DataCadastro       time.Time  `json:"data_cadastro" db:"data_cadastro"`
	DataAtualizacao    *time.Time `json:"data_atualizacao,omitempty" db:"data_atualizacao"`
	Observacoes        *string    `json:"observacoes,omitempty" db:"observacoes"`
}
