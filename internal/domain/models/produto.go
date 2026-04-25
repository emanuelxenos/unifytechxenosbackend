package models

import (
	"time"
)


type Produto struct {
	ID                int        `json:"id_produto" db:"id_produto"`
	EmpresaID         int        `json:"empresa_id" db:"empresa_id"`
	CategoriaID       *int       `json:"categoria_id,omitempty" db:"categoria_id"`
	CodigoBarras      *string    `json:"codigo_barras,omitempty" db:"codigo_barras"`
	CodigoInterno     *string    `json:"codigo_interno,omitempty" db:"codigo_interno"`
	Nome              string     `json:"nome" db:"nome"`
	Descricao         *string    `json:"descricao,omitempty" db:"descricao"`
	Marca             *string    `json:"marca,omitempty" db:"marca"`
	UnidadeVenda      string     `json:"unidade_venda" db:"unidade_venda"`
	UnidadeCompra     string     `json:"unidade_compra" db:"unidade_compra"`
	FatorConversao    float64    `json:"fator_conversao" db:"fator_conversao"`
	EstoqueAtual      float64    `json:"estoque_atual" db:"estoque_atual"`
	EstoqueMinimo     float64    `json:"estoque_minimo" db:"estoque_minimo"`
	EstoqueMaximo     *float64   `json:"estoque_maximo,omitempty" db:"estoque_maximo"`
	ControlarEstoque  bool       `json:"controlar_estoque" db:"controlar_estoque"`
	PrecoCusto        float64    `json:"preco_custo" db:"preco_custo"`
	PrecoVenda        float64    `json:"preco_venda" db:"preco_venda"`
	PrecoPromocional  *float64   `json:"preco_promocional,omitempty" db:"preco_promocional"`
	DataInicioPromocao *time.Time `json:"data_inicio_promocao,omitempty" db:"data_inicio_promocao"`
	DataFimPromocao   *time.Time `json:"data_fim_promocao,omitempty" db:"data_fim_promocao"`
	MargemLucro       *float64   `json:"margem_lucro,omitempty" db:"margem_lucro"`
	NCM               *string    `json:"ncm,omitempty" db:"ncm"`
	DataCadastro      time.Time  `json:"data_cadastro" db:"data_cadastro"`
	DataUltimaCompra  *time.Time `json:"data_ultima_compra,omitempty" db:"data_ultima_compra"`
	DataUltimaVenda   *time.Time `json:"data_ultima_venda,omitempty" db:"data_ultima_venda"`
	FotoPrincipalURL  *string    `json:"foto_principal_url,omitempty" db:"foto_principal_url"`
	Ativo             bool       `json:"ativo" db:"ativo"`
	Destacado         bool       `json:"destacado" db:"destacado"`
	Localizacao       *string    `json:"localizacao,omitempty" db:"localizacao"`
	DataVencimento    *time.Time `json:"data_vencimento,omitempty" db:"data_vencimento"`

	// Campos calculados / joins
	CategoriaNome *string `json:"categoria_nome,omitempty" db:"categoria_nome"`
}

type CriarProdutoRequest struct {
	CodigoBarras       *string    `json:"codigo_barras,omitempty"`
	CodigoInterno      *string    `json:"codigo_interno,omitempty"`
	Nome               string     `json:"nome"`
	Descricao          *string    `json:"descricao,omitempty"`
	CategoriaID        *int       `json:"categoria_id,omitempty"`
	UnidadeVenda       string     `json:"unidade_venda"`
	ControlarEstoque   bool       `json:"controlar_estoque"`
	EstoqueMinimo      float64    `json:"estoque_minimo"`
	PrecoCusto         float64    `json:"preco_custo"`
	PrecoVenda         float64    `json:"preco_venda"`
	PrecoPromocional   *float64   `json:"preco_promocional,omitempty"`
	DataInicioPromocao *time.Time `json:"data_inicio_promocao,omitempty"`
	DataFimPromocao     *time.Time `json:"data_fim_promocao,omitempty"`
	MargemLucro        *float64   `json:"margem_lucro,omitempty"`
	Marca              *string    `json:"marca,omitempty"`
	Localizacao        *string    `json:"localizacao,omitempty"`
	DataVencimento     *time.Time `json:"data_vencimento,omitempty"`
	FotoPrincipalURL   *string    `json:"foto_principal_url,omitempty"`
}

type ProdutoBuscaResponse struct {
	ID           int     `json:"id_produto"`
	CodigoBarras *string `json:"codigo_barras"`
	Nome         string  `json:"nome"`
	PrecoVenda   float64 `json:"preco_venda"`
	EstoqueAtual float64 `json:"estoque_atual"`
	UnidadeVenda string  `json:"unidade_venda"`
}
