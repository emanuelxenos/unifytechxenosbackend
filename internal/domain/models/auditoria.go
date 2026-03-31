package models

import "time"

type Auditoria struct {
	ID             int        `json:"id_auditoria" db:"id_auditoria"`
	EmpresaID      int        `json:"empresa_id" db:"empresa_id"`
	Tabela         string     `json:"tabela" db:"tabela"`
	Acao           string     `json:"acao" db:"acao"`
	RegistroID     *int       `json:"registro_id,omitempty" db:"registro_id"`
	ValoresAntigos *string    `json:"valores_antigos,omitempty" db:"valores_antigos"`
	ValoresNovos   *string    `json:"valores_novos,omitempty" db:"valores_novos"`
	DataHora       time.Time  `json:"data_hora" db:"data_hora"`
	UsuarioID      *int       `json:"usuario_id,omitempty" db:"usuario_id"`
	IPAddress      *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent      *string    `json:"user_agent,omitempty" db:"user_agent"`
}
