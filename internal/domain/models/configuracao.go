package models

import "time"

type Configuracao struct {
	ID              int       `json:"id_config" db:"id_config"`
	EmpresaID       int       `json:"empresa_id" db:"empresa_id"`
	Chave           string    `json:"chave" db:"chave"`
	Valor           *string   `json:"valor" db:"valor"`
	Tipo            string    `json:"tipo" db:"tipo"`
	Categoria       string    `json:"categoria" db:"categoria"`
	Descricao       *string   `json:"descricao,omitempty" db:"descricao"`
	DataAtualizacao time.Time `json:"data_atualizacao" db:"data_atualizacao"`
}

type AtualizarConfigRequest struct {
	Configs []ConfigItem `json:"configs"`
}

type ConfigItem struct {
	Chave string `json:"chave"`
	Valor string `json:"valor"`
}
