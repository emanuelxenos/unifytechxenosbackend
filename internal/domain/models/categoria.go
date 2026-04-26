package models

import (
	"time"
)

type Categoria struct {
	ID             int       `json:"id_categoria" db:"id_categoria"`
	EmpresaID      int       `json:"empresa_id" db:"empresa_id"`
	Nome           string    `json:"nome" db:"nome"`
	Descricao      *string   `json:"descricao,omitempty" db:"descricao"`
	Icone          *string   `json:"icone,omitempty" db:"icone"`
	CorHex         *string   `json:"cor_hex,omitempty" db:"cor_hex"`
	CategoriaPaiID *int      `json:"categoria_pai_id,omitempty" db:"categoria_pai_id"`
	Nivel          int       `json:"nivel" db:"nivel"`
	Ativo          bool      `json:"ativo" db:"ativo"`
	OrdemExibicao  int       `json:"ordem_exibicao" db:"ordem_exibicao"`
	DataCadastro   time.Time `json:"data_cadastro" db:"data_cadastro"`
	TotalProdutos  int       `json:"total_produtos" db:"total_produtos"`
	CategoriaPaiNome *string `json:"categoria_pai_nome,omitempty" db:"categoria_pai_nome"`
}

type CriarCategoriaRequest struct {
	Nome           string  `json:"nome"`
	Descricao      *string `json:"descricao,omitempty"`
	Icone          *string `json:"icone,omitempty"`
	CorHex         *string `json:"cor_hex,omitempty"`
	CategoriaPaiID *int    `json:"categoria_pai_id,omitempty"`
	OrdemExibicao  int     `json:"ordem_exibicao"`
}
