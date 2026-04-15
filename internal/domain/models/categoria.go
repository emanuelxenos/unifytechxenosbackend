package models

import (
	"time"
)

type Categoria struct {
	ID             int       `json:"id_categoria" db:"id_categoria"`
	EmpresaID      int       `json:"empresa_id" db:"empresa_id"`
	Nome           string    `json:"nome" db:"nome"`
	Descricao      *string   `json:"descricao,omitempty" db:"descricao"`
	CategoriaPaiID *int      `json:"categoria_pai_id,omitempty" db:"categoria_pai_id"`
	Nivel          int       `json:"nivel" db:"nivel"`
	Ativo          bool      `json:"ativo" db:"ativo"`
	OrdemExibicao  int       `json:"ordem_exibicao" db:"ordem_exibicao"`
	DataCadastro   time.Time `json:"data_cadastro" db:"data_cadastro"`
}

type CriarCategoriaRequest struct {
	Nome           string  `json:"nome"`
	Descricao      *string `json:"descricao,omitempty"`
	CategoriaPaiID *int    `json:"categoria_pai_id,omitempty"`
	OrdemExibicao  int     `json:"ordem_exibicao"`
}
