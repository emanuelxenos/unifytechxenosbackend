package service

import (
	"context"
	"fmt"
	"time"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
)

type CaixaService struct {
	db *database.PostgresDB
}

func NewCaixaService(db *database.PostgresDB) *CaixaService {
	return &CaixaService{db: db}
}

func (s *CaixaService) Status(ctx context.Context, empresaID, usuarioID int) (*models.CaixaStatusResponse, error) {
	var sessao models.SessaoCaixa
	var operadorNome string

	err := s.db.Pool.QueryRow(ctx,
		`SELECT sc.id_sessao, sc.empresa_id, sc.caixa_fisico_id, sc.usuario_id,
		        sc.codigo_sessao, sc.data_abertura, sc.saldo_inicial,
		        sc.total_vendas, sc.total_sangrias, sc.total_suprimentos,
		        sc.status, u.nome
		 FROM sessao_caixa sc
		 JOIN usuario u ON sc.usuario_id = u.id_usuario
		 WHERE sc.empresa_id = $1 AND sc.usuario_id = $2 AND sc.status = 'aberto'`,
		empresaID, usuarioID,
	).Scan(
		&sessao.ID, &sessao.EmpresaID, &sessao.CaixaFisicoID, &sessao.UsuarioID,
		&sessao.CodigoSessao, &sessao.DataAbertura, &sessao.SaldoInicial,
		&sessao.TotalVendas, &sessao.TotalSangrias, &sessao.TotalSuprimentos,
		&sessao.Status, &operadorNome,
	)

	if err != nil {
		return &models.CaixaStatusResponse{
			SessaoAtiva: false,
		}, nil
	}

	return &models.CaixaStatusResponse{
		SessaoAtiva: true,
		Sessao:      &sessao,
		Operador: &models.OperadorInfo{
			ID:   sessao.UsuarioID,
			Nome: operadorNome,
		},
	}, nil
}

func (s *CaixaService) Abrir(ctx context.Context, empresaID, usuarioID int, req models.AbrirCaixaRequest) (*models.SessaoCaixa, error) {
	// Verificar se já existe sessão aberta para este caixa
	var count int
	err := s.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM sessao_caixa WHERE caixa_fisico_id = $1 AND status = 'aberto'`,
		req.CaixaFisicoID,
	).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar sessão: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("já existe uma sessão aberta para este caixa")
	}

	var sessao models.SessaoCaixa
	obs := req.Observacao
	err = s.db.Pool.QueryRow(ctx,
		`INSERT INTO sessao_caixa (empresa_id, caixa_fisico_id, usuario_id, saldo_inicial, observacoes_abertura)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id_sessao, codigo_sessao, status, data_abertura`,
		empresaID, req.CaixaFisicoID, usuarioID, req.SaldoInicial, obs,
	).Scan(&sessao.ID, &sessao.CodigoSessao, &sessao.Status, &sessao.DataAbertura)

	if err != nil {
		return nil, fmt.Errorf("erro ao abrir caixa: %w", err)
	}

	sessao.EmpresaID = empresaID
	sessao.CaixaFisicoID = req.CaixaFisicoID
	sessao.UsuarioID = usuarioID
	sessao.SaldoInicial = req.SaldoInicial

	// Registrar movimentação de abertura
	_, _ = s.db.Pool.Exec(ctx,
		`INSERT INTO caixa_movimentacao (empresa_id, sessao_caixa_id, tipo, valor, motivo, usuario_id)
		 VALUES ($1, $2, 'abertura', $3, 'Abertura de caixa', $4)`,
		empresaID, sessao.ID, req.SaldoInicial, usuarioID,
	)

	return &sessao, nil
}

func (s *CaixaService) Fechar(ctx context.Context, empresaID, usuarioID int, req models.FecharCaixaRequest) (*models.SessaoCaixa, error) {
	// Buscar sessão aberta do usuário
	var sessao models.SessaoCaixa
	err := s.db.Pool.QueryRow(ctx,
		`SELECT id_sessao, saldo_inicial, total_vendas, total_sangrias, total_suprimentos
		 FROM sessao_caixa
		 WHERE empresa_id = $1 AND usuario_id = $2 AND status = 'aberto'`,
		empresaID, usuarioID,
	).Scan(&sessao.ID, &sessao.SaldoInicial, &sessao.TotalVendas, &sessao.TotalSangrias, &sessao.TotalSuprimentos)

	if err != nil {
		return nil, fmt.Errorf("nenhuma sessão aberta encontrada")
	}

	// Calcular saldo esperado
	saldoEsperado := sessao.SaldoInicial + sessao.TotalVendas + sessao.TotalSuprimentos - sessao.TotalSangrias
	diferenca := req.SaldoFinal - saldoEsperado

	now := time.Now()
	obs := req.Observacao

	_, err = s.db.Pool.Exec(ctx,
		`UPDATE sessao_caixa
		 SET status = 'fechado',
		     data_fechamento = $1,
		     saldo_final = $2,
		     saldo_final_esperado = $3,
		     diferenca = $4,
		     observacoes_fechamento = $5,
		     usuario_fechamento_id = $6
		 WHERE id_sessao = $7`,
		now, req.SaldoFinal, saldoEsperado, diferenca, obs, usuarioID, sessao.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao fechar caixa: %w", err)
	}

	sessao.Status = "fechado"
	sessao.DataFechamento = &now
	sessao.SaldoFinal = req.SaldoFinal
	sessao.SaldoFinalEsperado = saldoEsperado
	sessao.Diferenca = diferenca

	// Registrar movimentação de fechamento
	_, _ = s.db.Pool.Exec(ctx,
		`INSERT INTO caixa_movimentacao (empresa_id, sessao_caixa_id, tipo, valor, motivo, usuario_id)
		 VALUES ($1, $2, 'fechamento', $3, 'Fechamento de caixa', $4)`,
		empresaID, sessao.ID, req.SaldoFinal, usuarioID,
	)

	return &sessao, nil
}

func (s *CaixaService) Sangria(ctx context.Context, empresaID, usuarioID int, req models.SangriaSuprimentoRequest) error {
	// Buscar sessão aberta
	var sessaoID int
	err := s.db.Pool.QueryRow(ctx,
		`SELECT id_sessao FROM sessao_caixa
		 WHERE empresa_id = $1 AND usuario_id = $2 AND status = 'aberto'`,
		empresaID, usuarioID,
	).Scan(&sessaoID)
	if err != nil {
		return fmt.Errorf("nenhuma sessão aberta encontrada")
	}

	// Registrar sangria
	_, err = s.db.Pool.Exec(ctx,
		`INSERT INTO caixa_movimentacao (empresa_id, sessao_caixa_id, tipo, valor, motivo, usuario_id)
		 VALUES ($1, $2, 'sangria', $3, $4, $5)`,
		empresaID, sessaoID, req.Valor, req.Motivo, usuarioID,
	)
	if err != nil {
		return fmt.Errorf("erro ao registrar sangria: %w", err)
	}

	// Atualizar total de sangrias na sessão
	_, err = s.db.Pool.Exec(ctx,
		`UPDATE sessao_caixa SET total_sangrias = total_sangrias + $1 WHERE id_sessao = $2`,
		req.Valor, sessaoID,
	)

	return err
}

func (s *CaixaService) Suprimento(ctx context.Context, empresaID, usuarioID int, req models.SangriaSuprimentoRequest) error {
	// Buscar sessão aberta
	var sessaoID int
	err := s.db.Pool.QueryRow(ctx,
		`SELECT id_sessao FROM sessao_caixa
		 WHERE empresa_id = $1 AND usuario_id = $2 AND status = 'aberto'`,
		empresaID, usuarioID,
	).Scan(&sessaoID)
	if err != nil {
		return fmt.Errorf("nenhuma sessão aberta encontrada")
	}

	// Registrar suprimento
	_, err = s.db.Pool.Exec(ctx,
		`INSERT INTO caixa_movimentacao (empresa_id, sessao_caixa_id, tipo, valor, motivo, usuario_id)
		 VALUES ($1, $2, 'suprimento', $3, $4, $5)`,
		empresaID, sessaoID, req.Valor, req.Motivo, usuarioID,
	)
	if err != nil {
		return fmt.Errorf("erro ao registrar suprimento: %w", err)
	}

	// Atualizar total de suprimentos na sessão
	_, err = s.db.Pool.Exec(ctx,
		`UPDATE sessao_caixa SET total_suprimentos = total_suprimentos + $1 WHERE id_sessao = $2`,
		req.Valor, sessaoID,
	)

	return err
}
