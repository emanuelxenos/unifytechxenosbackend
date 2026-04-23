package models

import (
	"time"
)

type Compra struct {
	ID               int        `json:"id_compra" db:"id_compra"`
	EmpresaID        int        `json:"empresa_id" db:"empresa_id"`
	FornecedorID     *int       `json:"fornecedor_id,omitempty" db:"fornecedor_id"`
	UsuarioID        int        `json:"usuario_id" db:"usuario_id"`
	NumeroNotaFiscal *string    `json:"numero_nota_fiscal,omitempty" db:"numero_nota_fiscal"`
	SerieNota        *string    `json:"serie_nota,omitempty" db:"serie_nota"`
	ChaveNFe         *string    `json:"chave_nfe,omitempty" db:"chave_nfe"`
	DataEmissao      *time.Time `json:"data_emissao,omitempty" db:"data_emissao"`
	DataEntrada      time.Time  `json:"data_entrada" db:"data_entrada"`
	DataCadastro     time.Time  `json:"data_cadastro" db:"data_cadastro"`
	ValorProdutos    float64    `json:"valor_produtos" db:"valor_produtos"`
	ValorFrete       float64    `json:"valor_frete" db:"valor_frete"`
	ValorDesconto    float64    `json:"valor_desconto" db:"valor_desconto"`
	ValorTotal       float64    `json:"valor_total" db:"valor_total"`
	Status           string     `json:"status" db:"status"`
	Observacoes      *string    `json:"observacoes,omitempty" db:"observacoes"`

	// Joins
	FornecedorNome *string `json:"fornecedor_nome,omitempty" db:"fornecedor_nome"`

	// Relacionamentos
	Itens []ItemCompra `json:"itens,omitempty"`
}

type ItemCompra struct {
	ID                 int        `json:"id_item_compra" db:"id_item_compra"`
	CompraID           int        `json:"compra_id" db:"compra_id"`
	ProdutoID          int        `json:"produto_id" db:"produto_id"`
	Sequencia          int        `json:"sequencia" db:"sequencia"`
	Quantidade         float64    `json:"quantidade" db:"quantidade"`
	QuantidadeRecebida float64    `json:"quantidade_recebida" db:"quantidade_recebida"`
	PrecoUnitario      float64    `json:"preco_unitario" db:"preco_unitario"`
	ValorTotal         float64    `json:"valor_total" db:"valor_total"`
	ValorDesconto      float64    `json:"valor_desconto" db:"valor_desconto"`
	DataRecebimento    *time.Time `json:"data_recebimento,omitempty" db:"data_recebimento"`
	Localizacao        *string    `json:"localizacao,omitempty" db:"localizacao"`
	DataVencimento     *time.Time `json:"data_vencimento,omitempty" db:"data_vencimento"`

	// Join
	ProdutoNome *string `json:"produto_nome,omitempty" db:"produto_nome"`
}

type CriarCompraRequest struct {
	FornecedorID     int                     `json:"fornecedor_id"`
	NumeroNotaFiscal string                  `json:"numero_nota_fiscal"`
	DataEmissao      string                  `json:"data_emissao"`
	Itens            []CriarItemCompraRequest `json:"itens"`
}

type CriarItemCompraRequest struct {
	ProdutoID     int     `json:"produto_id"`
	Quantidade    float64 `json:"quantidade"`
	PrecoUnitario float64 `json:"preco_unitario"`
	Localizacao   string  `json:"localizacao,omitempty"`
	DataVencimento string `json:"data_vencimento,omitempty"`
}

type ReceberCompraRequest struct {
	ItensRecebidos []ItemRecebidoRequest `json:"itens_recebidos"`
}

type ItemRecebidoRequest struct {
	ProdutoID          int     `json:"produto_id"`
	QuantidadeRecebida float64 `json:"quantidade_recebida"`
	LoteFabricante     string  `json:"lote_fabricante,omitempty"`
}
