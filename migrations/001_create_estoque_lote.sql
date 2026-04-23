-- Migração para implementar gestão de lotes e localizações
-- Data: 2026-04-23

CREATE TABLE IF NOT EXISTS estoque_localizacao (
    id_localizacao SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    codigo VARCHAR(50) NOT NULL,
    nome VARCHAR(100) NOT NULL,
    descricao TEXT,
    ativo BOOLEAN DEFAULT TRUE,
    data_cadastro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(empresa_id, codigo)
);

CREATE TABLE IF NOT EXISTS estoque_lote (
    id_lote SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    produto_id INTEGER NOT NULL REFERENCES produto(id_produto) ON DELETE CASCADE,
    localizacao_id INTEGER REFERENCES estoque_localizacao(id_localizacao),
    
    -- Identificação
    lote_interno VARCHAR(50) NOT NULL, -- Gerado pelo sistema
    lote_fabricante VARCHAR(50),      -- Código original
    
    -- Quantidades
    quantidade_inicial DECIMAL(12,4) NOT NULL,
    quantidade_atual DECIMAL(12,4) NOT NULL DEFAULT 0,
    
    -- Datas
    data_fabricacao DATE,
    data_vencimento DATE NOT NULL,
    data_recebimento TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Status e Auditoria
    status VARCHAR(20) DEFAULT 'ativo', -- 'ativo', 'bloqueado', 'vencido', 'esgotado'
    observacao TEXT,
    usuario_id INTEGER REFERENCES usuario(id_usuario),
    
    CONSTRAINT check_quantidade_positiva CHECK (quantidade_atual >= 0)
);

CREATE INDEX idx_lote_vencimento ON estoque_lote(data_vencimento);
CREATE INDEX idx_lote_produto_ativo ON estoque_lote(produto_id) WHERE status = 'ativo';

-- Adicionar id_lote em movimentação para rastreabilidade
ALTER TABLE estoque_movimentacao ADD COLUMN IF NOT EXISTS lote_id INTEGER REFERENCES estoque_lote(id_lote);

-- Função para sincronizar estoque_atual do produto
CREATE OR REPLACE FUNCTION sync_estoque_produto_v2() RETURNS TRIGGER AS $$
BEGIN
    UPDATE produto 
    SET estoque_atual = (
        SELECT COALESCE(SUM(quantidade_atual), 0) 
        FROM estoque_lote 
        WHERE produto_id = COALESCE(NEW.produto_id, OLD.produto_id) 
          AND status IN ('ativo', 'bloqueado')
    )
    WHERE id_produto = COALESCE(NEW.produto_id, OLD.produto_id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_sync_lote_estoque ON estoque_lote;
CREATE TRIGGER trg_sync_lote_estoque
AFTER INSERT OR UPDATE OR DELETE ON estoque_lote
FOR EACH ROW EXECUTE FUNCTION sync_estoque_produto_v2();

-- Migração inicial: Converter estoque atual em um lote inicial
INSERT INTO estoque_lote (empresa_id, produto_id, lote_interno, quantidade_inicial, quantidade_atual, data_vencimento, status)
SELECT empresa_id, id_produto, 'LOTE-INICIAL', estoque_atual, estoque_atual, COALESCE(data_vencimento, '2099-12-31'), 'ativo'
FROM produto
WHERE estoque_atual > 0 AND NOT EXISTS (SELECT 1 FROM estoque_lote WHERE produto_id = id_produto);
