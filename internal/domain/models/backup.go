package models

import "time"

type Backup struct {
	ID           int       `json:"id_backup" db:"id_backup"`
	EmpresaID    int       `json:"empresa_id" db:"empresa_id"`
	NomeArquivo  string    `json:"nome_arquivo" db:"nome_arquivo"`
	Caminho      string    `json:"caminho" db:"caminho"`
	Tamanho      *int64    `json:"tamanho" db:"tamanho"`
	DataBackup   time.Time `json:"data_backup" db:"data_backup"`
	Tipo         string    `json:"tipo" db:"tipo"`
	Status       string    `json:"status" db:"status"`
	Observacoes  *string   `json:"observacoes,omitempty" db:"observacoes"`
	UsuarioID    *int      `json:"usuario_id,omitempty" db:"usuario_id"`
}
