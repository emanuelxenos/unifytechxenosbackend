package models

import (
	"time"
)

type Usuario struct {
	ID                       int        `json:"id_usuario" db:"id_usuario"`
	EmpresaID                int        `json:"empresa_id" db:"empresa_id"`
	Nome                     string     `json:"nome" db:"nome"`
	CPF                      *string    `json:"cpf,omitempty" db:"cpf"`
	RG                       *string    `json:"rg,omitempty" db:"rg"`
	DataNascimento           *time.Time `json:"data_nascimento,omitempty" db:"data_nascimento"`
	Telefone                 *string    `json:"telefone,omitempty" db:"telefone"`
	Email                    *string    `json:"email,omitempty" db:"email"`
	Endereco                 *string    `json:"endereco,omitempty" db:"endereco"`
	Login                    string     `json:"login" db:"login"`
	SenhaHash                string     `json:"-" db:"senha_hash"`
	PinAcesso                *string    `json:"-" db:"pin_acesso"`
	Perfil                   string     `json:"perfil" db:"perfil"`
	PodeAbrirCaixa           bool       `json:"pode_abrir_caixa" db:"pode_abrir_caixa"`
	PodeFecharCaixa          bool       `json:"pode_fechar_caixa" db:"pode_fechar_caixa"`
	PodeDarDesconto          bool       `json:"pode_dar_desconto" db:"pode_dar_desconto"`
	LimiteDescontoPercentual float64    `json:"limite_desconto_percentual" db:"limite_desconto_percentual"`
	PodeCancelarVenda        bool       `json:"pode_cancelar_venda" db:"pode_cancelar_venda"`
	PodeAlterarPreco         bool       `json:"pode_alterar_preco" db:"pode_alterar_preco"`
	PodeAcessarRelatorios    bool       `json:"pode_acessar_relatorios" db:"pode_acessar_relatorios"`
	PodeGerenciarProdutos    bool       `json:"pode_gerenciar_produtos" db:"pode_gerenciar_produtos"`
	PodeGerenciarUsuarios    bool       `json:"pode_gerenciar_usuarios" db:"pode_gerenciar_usuarios"`
	CaixaPadrao              *string    `json:"caixa_padrao,omitempty" db:"caixa_padrao"`
	Ativo                    bool       `json:"ativo" db:"ativo"`
	DataCadastro             time.Time  `json:"data_cadastro" db:"data_cadastro"`
	UltimoAcesso             *time.Time `json:"ultimo_acesso,omitempty" db:"ultimo_acesso"`
}

type UsuarioLoginRequest struct {
	Login    string `json:"login"`
	Senha    string `json:"senha"`
	Terminal string `json:"terminal"`
}

type UsuarioLoginResponse struct {
	Token   string       `json:"token"`
	Usuario UsuarioInfo  `json:"usuario"`
}

type UsuarioInfo struct {
	ID         int              `json:"id"`
	Nome       string           `json:"nome"`
	Perfil     string           `json:"perfil"`
	Permissoes UsuarioPermissao `json:"permissoes"`
}

type UsuarioPermissao struct {
	PodeAbrirCaixa  bool    `json:"pode_abrir_caixa"`
	PodeDarDesconto bool    `json:"pode_dar_desconto"`
	LimiteDesconto  float64 `json:"limite_desconto"`
}

type CriarUsuarioRequest struct {
	Nome                     string  `json:"nome"`
	Login                    string  `json:"login"`
	Senha                    string  `json:"senha"`
	Perfil                   string  `json:"perfil"`
	CPF                      *string `json:"cpf,omitempty"`
	Telefone                 *string `json:"telefone,omitempty"`
	Email                    *string `json:"email,omitempty"`
	PodeAbrirCaixa           bool    `json:"pode_abrir_caixa"`
	PodeFecharCaixa          bool    `json:"pode_fechar_caixa"`
	PodeDarDesconto          bool    `json:"pode_dar_desconto"`
	LimiteDescontoPercentual float64 `json:"limite_desconto_percentual"`
	PodeCancelarVenda        bool    `json:"pode_cancelar_venda"`
}

type AtualizarUsuarioRequest struct {
	Nome                     string  `json:"nome"`
	Login                    string  `json:"login"`
	Senha                    string  `json:"senha,omitempty"`
	Perfil                   string  `json:"perfil"`
	CPF                      *string `json:"cpf,omitempty"`
	Telefone                 *string `json:"telefone,omitempty"`
	Email                    *string `json:"email,omitempty"`
	PodeAbrirCaixa           bool    `json:"pode_abrir_caixa"`
	PodeFecharCaixa          bool    `json:"pode_fechar_caixa"`
	PodeDarDesconto          bool    `json:"pode_dar_desconto"`
	LimiteDescontoPercentual float64 `json:"limite_desconto_percentual"`
	PodeCancelarVenda        bool    `json:"pode_cancelar_venda"`
}
