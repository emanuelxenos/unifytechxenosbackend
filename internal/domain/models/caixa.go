package models

import (
	"time"
)

type CaixaFisico struct {
	ID            int        `json:"id_caixa_fisico" db:"id_caixa_fisico"`
	EmpresaID     int        `json:"empresa_id" db:"empresa_id"`
	Codigo        string     `json:"codigo" db:"codigo"`
	Nome          string     `json:"nome" db:"nome"`
	Descricao     *string    `json:"descricao,omitempty" db:"descricao"`
	Localizacao   *string    `json:"localizacao,omitempty" db:"localizacao"`
	Setor         *string    `json:"setor,omitempty" db:"setor"`
	Ativo         bool       `json:"ativo" db:"ativo"`
	DataCadastro  time.Time  `json:"data_cadastro" db:"data_cadastro"`
	DataUltimoUso *time.Time `json:"data_ultimo_uso,omitempty" db:"data_ultimo_uso"`
}

type SessaoCaixa struct {
	ID                      int        `json:"id_sessao" db:"id_sessao"`
	EmpresaID               int        `json:"empresa_id" db:"empresa_id"`
	CaixaFisicoID           int        `json:"caixa_fisico_id" db:"caixa_fisico_id"`
	UsuarioID               int        `json:"usuario_id" db:"usuario_id"`
	CodigoSessao            string     `json:"codigo_sessao" db:"codigo_sessao"`
	DataAbertura            time.Time  `json:"data_abertura" db:"data_abertura"`
	DataFechamento          *time.Time `json:"data_fechamento,omitempty" db:"data_fechamento"`
	DataUltimaVenda         *time.Time `json:"data_ultima_venda,omitempty" db:"data_ultima_venda"`
	SaldoInicial            float64    `json:"saldo_inicial" db:"saldo_inicial"`
	TotalVendas             float64    `json:"total_vendas" db:"total_vendas"`
	TotalVendasCanceladas   float64    `json:"total_vendas_canceladas" db:"total_vendas_canceladas"`
	TotalDescontosConcedidos float64   `json:"total_descontos_concedidos" db:"total_descontos_concedidos"`
	TotalSangrias           float64    `json:"total_sangrias" db:"total_sangrias"`
	TotalSuprimentos        float64    `json:"total_suprimentos" db:"total_suprimentos"`
	TotalDinheiro           float64    `json:"total_dinheiro" db:"total_dinheiro"`
	TotalCartaoDebito       float64    `json:"total_cartao_debito" db:"total_cartao_debito"`
	TotalCartaoCredito      float64    `json:"total_cartao_credito" db:"total_cartao_credito"`
	TotalPix                float64    `json:"total_pix" db:"total_pix"`
	TotalVale               float64    `json:"total_vale" db:"total_vale"`
	TotalOutros             float64    `json:"total_outros" db:"total_outros"`
	SaldoFinal              float64    `json:"saldo_final" db:"saldo_final"`
	SaldoFinalEsperado      float64    `json:"saldo_final_esperado" db:"saldo_final_esperado"`
	Diferenca               float64    `json:"diferenca" db:"diferenca"`
	Status                  string     `json:"status" db:"status"`
	ObservacoesAbertura     *string    `json:"observacoes_abertura,omitempty" db:"observacoes_abertura"`
	ObservacoesFechamento   *string    `json:"observacoes_fechamento,omitempty" db:"observacoes_fechamento"`
}

type CaixaMovimentacao struct {
	ID              int       `json:"id_caixa_movimentacao" db:"id_caixa_movimentacao"`
	EmpresaID       int       `json:"empresa_id" db:"empresa_id"`
	SessaoCaixaID   int       `json:"sessao_caixa_id" db:"sessao_caixa_id"`
	Tipo            string    `json:"tipo" db:"tipo"`
	Valor           float64   `json:"valor" db:"valor"`
	FormaPagamentoID *int     `json:"forma_pagamento_id,omitempty" db:"forma_pagamento_id"`
	Motivo          *string   `json:"motivo,omitempty" db:"motivo"`
	Observacao      *string   `json:"observacao,omitempty" db:"observacao"`
	DataMovimentacao time.Time `json:"data_movimentacao" db:"data_movimentacao"`
	UsuarioID       int       `json:"usuario_id" db:"usuario_id"`
	VendaID         *int      `json:"venda_id,omitempty" db:"venda_id"`
}

type AbrirCaixaRequest struct {
	CaixaFisicoID int     `json:"caixa_fisico_id"`
	SaldoInicial  float64 `json:"saldo_inicial"`
	Observacao    string  `json:"observacao"`
}

type FecharCaixaRequest struct {
	SaldoFinal      float64 `json:"saldo_final"`
	SupervisorSenha string  `json:"supervisor_senha"`
	Observacao      string  `json:"observacao"`
}

type SangriaSuprimentoRequest struct {
	Valor  float64 `json:"valor"`
	Motivo string  `json:"motivo"`
}

type CaixaStatusResponse struct {
	SessaoAtiva bool         `json:"sessao_ativa"`
	Sessao      *SessaoCaixa `json:"sessao,omitempty"`
	Operador    *OperadorInfo `json:"operador,omitempty"`
}

type OperadorInfo struct {
	ID   int    `json:"id"`
	Nome string `json:"nome"`
}
