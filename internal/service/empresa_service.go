package service

import (
	"context"
	"fmt"
	"time"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
)

type EmpresaService struct {
	db *database.PostgresDB
}

func NewEmpresaService(db *database.PostgresDB) *EmpresaService {
	return &EmpresaService{db: db}
}

func (s *EmpresaService) Buscar(ctx context.Context, id int) (*models.Empresa, error) {
	var e models.Empresa
	err := s.db.Pool.QueryRow(ctx,
		`SELECT id_empresa, razao_social, nome_fantasia, cnpj,
		        inscricao_estadual, inscricao_municipal, logradouro, numero,
		        complemento, bairro, cidade, estado, cep,
		        telefone, telefone2, email, site, regime_tributario,
		        moeda, casas_decimais, fuso_horario, logotipo_url,
		        cor_primaria, cor_secundaria, ativo, data_cadastro,
		        data_atualizacao, observacoes
		 FROM empresa
		 WHERE id_empresa = $1`, id,
	).Scan(
		&e.ID, &e.RazaoSocial, &e.NomeFantasia, &e.CNPJ,
		&e.InscricaoEstadual, &e.InscricaoMunicipal, &e.Logradouro, &e.Numero,
		&e.Complemento, &e.Bairro, &e.Cidade, &e.Estado, &e.CEP,
		&e.Telefone, &e.Telefone2, &e.Email, &e.Site, &e.RegimeTributario,
		&e.Moeda, &e.CasasDecimais, &e.FusoHorario, &e.LogotipoURL,
		&e.CorPrimaria, &e.CorSecundaria, &e.Ativo, &e.DataCadastro,
		&e.DataAtualizacao, &e.Observacoes,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar empresa: %w", err)
	}
	return &e, nil
}

func (s *EmpresaService) Atualizar(ctx context.Context, e *models.Empresa) error {
	_, err := s.db.Pool.Exec(ctx,
		`UPDATE empresa SET
		    razao_social = $1,
		    nome_fantasia = $2,
		    cnpj = $3,
		    inscricao_estadual = $4,
		    inscricao_municipal = $5,
		    logradouro = $6,
		    numero = $7,
		    complemento = $8,
		    bairro = $9,
		    cidade = $10,
		    estado = $11,
		    cep = $12,
		    telefone = $13,
		    telefone2 = $14,
		    email = $15,
		    site = $16,
		    regime_tributario = $17,
		    moeda = $18,
		    casas_decimais = $19,
		    fuso_horario = $20,
		    logotipo_url = $21,
		    cor_primaria = $22,
		    cor_secundaria = $23,
		    data_atualizacao = $24,
		    observacoes = $25
		 WHERE id_empresa = $26`,
		e.RazaoSocial, e.NomeFantasia, e.CNPJ, e.InscricaoEstadual, e.InscricaoMunicipal,
		e.Logradouro, e.Numero, e.Complemento, e.Bairro, e.Cidade, e.Estado, e.CEP,
		e.Telefone, e.Telefone2, e.Email, e.Site, e.RegimeTributario,
		e.Moeda, e.CasasDecimais, e.FusoHorario, e.LogotipoURL,
		e.CorPrimaria, e.CorSecundaria, time.Now(), e.Observacoes,
		e.ID,
	)
	if err != nil {
		return fmt.Errorf("erro ao atualizar empresa: %w", err)
	}
	return nil
}
