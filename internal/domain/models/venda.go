package models

import (
	"time"
)

type Venda struct {
	ID                  int        `json:"id_venda" db:"id_venda"`
	EmpresaID           int        `json:"empresa_id" db:"empresa_id"`
	SessaoCaixaID       int        `json:"sessao_caixa_id" db:"sessao_caixa_id"`
	UsuarioID           int        `json:"usuario_id" db:"usuario_id"`
	CaixaFisicoID       int        `json:"caixa_fisico_id" db:"caixa_fisico_id"`
	NumeroVenda         string     `json:"numero_venda" db:"numero_venda"`
	ClienteID           *int       `json:"cliente_id,omitempty" db:"cliente_id"`
	ClienteNome         *string    `json:"cliente_nome,omitempty" db:"cliente_nome"`
	ClienteDocumento    *string    `json:"cliente_documento,omitempty" db:"cliente_documento"`
	DataVenda           time.Time  `json:"data_venda" db:"data_venda"`
	DataCancelamento    *time.Time `json:"data_cancelamento,omitempty" db:"data_cancelamento"`
	ValorTotalProdutos  float64    `json:"valor_total_produtos" db:"valor_total_produtos"`
	ValorTotalDescontos float64    `json:"valor_total_descontos" db:"valor_total_descontos"`
	ValorTotalAcrescimos float64   `json:"valor_total_acrescimos" db:"valor_total_acrescimos"`
	ValorSubtotal       float64    `json:"valor_subtotal" db:"valor_subtotal"`
	ValorFrete          float64    `json:"valor_frete" db:"valor_frete"`
	ValorTotal          float64    `json:"valor_total" db:"valor_total"`
	ValorPago           float64    `json:"valor_pago" db:"valor_pago"`
	ValorTroco          float64    `json:"valor_troco" db:"valor_troco"`
	Status              string     `json:"status" db:"status"`
	TipoVenda           string     `json:"tipo_venda" db:"tipo_venda"`
	Observacoes         *string    `json:"observacoes,omitempty" db:"observacoes"`
	MotivoCancelamento  *string    `json:"motivo_cancelamento,omitempty" db:"motivo_cancelamento"`

	// Joins
	OperadorNome *string `json:"operador_nome,omitempty" db:"operador_nome"`
	CaixaNome    *string `json:"caixa_nome,omitempty" db:"caixa_nome"`

	// Relacionamentos carregados
	Itens      []ItemVenda      `json:"itens,omitempty"`
	Pagamentos []VendaPagamento `json:"pagamentos,omitempty"`
}

type ItemVenda struct {
	ID                    int        `json:"id_item_venda" db:"id_item_venda"`
	VendaID               int        `json:"venda_id" db:"venda_id"`
	ProdutoID             int        `json:"produto_id" db:"produto_id"`
	Sequencia             int        `json:"sequencia" db:"sequencia"`
	Quantidade            float64    `json:"quantidade" db:"quantidade"`
	UnidadeVenda          string     `json:"unidade_venda" db:"unidade_venda"`
	PrecoUnitario         float64    `json:"preco_unitario" db:"preco_unitario"`
	PrecoCustoUnitario    *float64   `json:"preco_custo_unitario,omitempty" db:"preco_custo_unitario"`
	ValorTotal            float64    `json:"valor_total" db:"valor_total"`
	ValorDesconto         float64    `json:"valor_desconto" db:"valor_desconto"`
	ValorDescontoPercent  float64    `json:"valor_desconto_percentual" db:"valor_desconto_percentual"`
	ValorAcrescimo        float64    `json:"valor_acrescimo" db:"valor_acrescimo"`
	ValorLiquido          float64    `json:"valor_liquido" db:"valor_liquido"`
	Status                string     `json:"status" db:"status"`
	DataHora              time.Time  `json:"data_hora" db:"data_hora"`

	// Join
	ProdutoNome *string `json:"produto_nome,omitempty" db:"produto_nome"`
}

type VendaPagamento struct {
	ID               int       `json:"id_venda_pagamento" db:"id_venda_pagamento"`
	VendaID          int       `json:"venda_id" db:"venda_id"`
	FormaPagamentoID int       `json:"forma_pagamento_id" db:"forma_pagamento_id"`
	Valor            float64   `json:"valor" db:"valor"`
	TrocoPara        float64   `json:"troco_para" db:"troco_para"`
	Autorizacao      *string   `json:"autorizacao,omitempty" db:"autorizacao"`
	BandeiraCartao   *string   `json:"bandeira_cartao,omitempty" db:"bandeira_cartao"`
	Parcelas         int       `json:"parcelas" db:"parcelas"`
	Status           string    `json:"status" db:"status"`
	DataProcessamento time.Time `json:"data_processamento" db:"data_processamento"`

	// Join
	FormaPagamentoNome *string `json:"forma_pagamento_nome,omitempty" db:"forma_pagamento_nome"`
}

type FormaPagamento struct {
	ID            int     `json:"id_forma_pagamento" db:"id_forma_pagamento"`
	EmpresaID     int     `json:"empresa_id" db:"empresa_id"`
	Nome          string  `json:"nome" db:"nome"`
	Codigo        string  `json:"codigo" db:"codigo"`
	Tipo          string  `json:"tipo" db:"tipo"`
	Ativo         bool    `json:"ativo" db:"ativo"`
	ExibirNoCaixa bool    `json:"exibir_no_caixa" db:"exibir_no_caixa"`
	RequerTroco   bool    `json:"requer_troco" db:"requer_troco"`
	TaxaOperacao  float64 `json:"taxa_operacao" db:"taxa_operacao"`
	OrdemExibicao int     `json:"ordem_exibicao" db:"ordem_exibicao"`
}

type CriarVendaRequest struct {
	ClienteID  *int                        `json:"cliente_id,omitempty"`
	Itens      []CriarItemVendaRequest     `json:"itens"`
	Pagamentos []CriarPagamentoVendaRequest `json:"pagamentos"`
	Observacoes *string                    `json:"observacoes,omitempty"`
}

type CriarItemVendaRequest struct {
	ProdutoID      int     `json:"produto_id"`
	Quantidade     float64 `json:"quantidade"`
	PrecoUnitario  float64 `json:"preco_unitario"`
	ValorDesconto  float64 `json:"valor_desconto"`
}

type CriarPagamentoVendaRequest struct {
	FormaPagamentoID int     `json:"forma_pagamento_id"`
	Valor            float64 `json:"valor"`
	Parcelas         int     `json:"parcelas"`
	Autorizacao      *string `json:"autorizacao,omitempty"`
}

type CancelarVendaRequest struct {
	Motivo          string `json:"motivo"`
	SenhaSupervisor string `json:"senha_supervisor"`
}

type VendaResponse struct {
	IDVenda     int     `json:"id_venda"`
	NumeroVenda string  `json:"numero_venda"`
	ValorTotal  float64 `json:"valor_total"`
	ValorTroco  float64 `json:"valor_troco"`
	Comprovante string  `json:"comprovante"`
}
