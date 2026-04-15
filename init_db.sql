-- Script simplificado para criação das tabelas essenciais
-- Remove INDEX inline (não suportado em PostgreSQL padrão)

CREATE TABLE IF NOT EXISTS empresa (
    id_empresa SERIAL PRIMARY KEY,
    razao_social VARCHAR(200) NOT NULL,
    nome_fantasia VARCHAR(200) NOT NULL,
    cnpj VARCHAR(18) UNIQUE NOT NULL,
    inscricao_estadual VARCHAR(20),
    inscricao_municipal VARCHAR(20),
    logradouro VARCHAR(200) NOT NULL DEFAULT '',
    numero VARCHAR(10) NOT NULL DEFAULT '',
    complemento VARCHAR(100),
    bairro VARCHAR(100) NOT NULL DEFAULT '',
    cidade VARCHAR(100) NOT NULL DEFAULT '',
    estado CHAR(2) NOT NULL DEFAULT 'SP',
    cep VARCHAR(10) NOT NULL DEFAULT '',
    telefone VARCHAR(20) NOT NULL DEFAULT '',
    telefone2 VARCHAR(20),
    email VARCHAR(100) NOT NULL DEFAULT '',
    site VARCHAR(100),
    regime_tributario VARCHAR(50) NOT NULL DEFAULT 'SIMPLES',
    moeda VARCHAR(10) DEFAULT 'R$',
    casas_decimais INTEGER DEFAULT 2,
    fuso_horario VARCHAR(50) DEFAULT 'America/Sao_Paulo',
    logotipo_url VARCHAR(500),
    cor_primaria VARCHAR(7) DEFAULT '#1976D2',
    cor_secundaria VARCHAR(7) DEFAULT '#43A047',
    data_cadastro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    data_atualizacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ativo BOOLEAN DEFAULT TRUE,
    observacoes TEXT
);

CREATE TABLE IF NOT EXISTS usuario (
    id_usuario SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    nome VARCHAR(150) NOT NULL,
    cpf VARCHAR(14),
    rg VARCHAR(20),
    data_nascimento DATE,
    telefone VARCHAR(20),
    email VARCHAR(100),
    endereco TEXT,
    login VARCHAR(50) NOT NULL,
    senha_hash VARCHAR(255) NOT NULL,
    pin_acesso VARCHAR(6),
    perfil VARCHAR(20) NOT NULL DEFAULT 'caixa',
    pode_abrir_caixa BOOLEAN DEFAULT FALSE,
    pode_fechar_caixa BOOLEAN DEFAULT FALSE,
    pode_dar_desconto BOOLEAN DEFAULT FALSE,
    limite_desconto_percentual DECIMAL(5,2) DEFAULT 10.00,
    pode_cancelar_venda BOOLEAN DEFAULT FALSE,
    pode_alterar_preco BOOLEAN DEFAULT FALSE,
    pode_acessar_relatorios BOOLEAN DEFAULT FALSE,
    pode_gerenciar_produtos BOOLEAN DEFAULT FALSE,
    pode_gerenciar_usuarios BOOLEAN DEFAULT FALSE,
    caixa_padrao VARCHAR(20),
    ativo BOOLEAN DEFAULT TRUE,
    data_cadastro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ultimo_acesso TIMESTAMP,
    observacoes TEXT,
    UNIQUE(empresa_id, login)
);

CREATE TABLE IF NOT EXISTS caixa_fisico (
    id_caixa_fisico SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    codigo VARCHAR(20) NOT NULL,
    nome VARCHAR(100) NOT NULL,
    descricao TEXT,
    localizacao VARCHAR(100),
    setor VARCHAR(50),
    impressora_modelo VARCHAR(50),
    impressora_porta VARCHAR(50),
    balanca_modelo VARCHAR(50),
    balanca_porta VARCHAR(50),
    ativo BOOLEAN DEFAULT TRUE,
    data_cadastro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    data_ultimo_uso TIMESTAMP,
    UNIQUE(empresa_id, codigo)
);

CREATE TABLE IF NOT EXISTS sessao_caixa (
    id_sessao SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    caixa_fisico_id INTEGER NOT NULL REFERENCES caixa_fisico(id_caixa_fisico),
    usuario_id INTEGER NOT NULL REFERENCES usuario(id_usuario),
    codigo_sessao VARCHAR(50) NOT NULL,
    data_abertura TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_fechamento TIMESTAMP,
    data_ultima_venda TIMESTAMP,
    saldo_inicial DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    total_vendas DECIMAL(12,2) DEFAULT 0.00,
    total_vendas_canceladas DECIMAL(12,2) DEFAULT 0.00,
    total_devolucoes DECIMAL(12,2) DEFAULT 0.00,
    total_descontos_concedidos DECIMAL(12,2) DEFAULT 0.00,
    total_acrescimos DECIMAL(12,2) DEFAULT 0.00,
    total_sangrias DECIMAL(12,2) DEFAULT 0.00,
    total_suprimentos DECIMAL(12,2) DEFAULT 0.00,
    total_dinheiro DECIMAL(12,2) DEFAULT 0.00,
    total_cartao_debito DECIMAL(12,2) DEFAULT 0.00,
    total_cartao_credito DECIMAL(12,2) DEFAULT 0.00,
    total_pix DECIMAL(12,2) DEFAULT 0.00,
    total_vale DECIMAL(12,2) DEFAULT 0.00,
    total_outros DECIMAL(12,2) DEFAULT 0.00,
    saldo_final DECIMAL(12,2) DEFAULT 0.00,
    saldo_final_esperado DECIMAL(12,2) DEFAULT 0.00,
    diferenca DECIMAL(12,2) DEFAULT 0.00,
    status VARCHAR(20) NOT NULL DEFAULT 'aberto',
    observacoes_abertura TEXT,
    observacoes_fechamento TEXT,
    motivo_cancelamento TEXT,
    usuario_fechamento_id INTEGER REFERENCES usuario(id_usuario),
    supervisor_autorizacao_id INTEGER REFERENCES usuario(id_usuario)
);

CREATE TABLE IF NOT EXISTS forma_pagamento (
    id_forma_pagamento SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    nome VARCHAR(50) NOT NULL,
    codigo VARCHAR(20) NOT NULL,
    tipo VARCHAR(20) NOT NULL,
    ativo BOOLEAN DEFAULT TRUE,
    exibir_no_caixa BOOLEAN DEFAULT TRUE,
    requer_troco BOOLEAN DEFAULT FALSE,
    taxa_operacao DECIMAL(5,2) DEFAULT 0.00,
    dias_recebimento INTEGER DEFAULT 0,
    ordem_exibicao INTEGER DEFAULT 0,
    bandeira_cartao VARCHAR(50),
    max_parcelas INTEGER DEFAULT 1,
    data_cadastro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(empresa_id, codigo),
    UNIQUE(empresa_id, nome)
);

CREATE TABLE IF NOT EXISTS cliente (
    id_cliente SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    nome VARCHAR(150) NOT NULL,
    tipo_pessoa CHAR(1) DEFAULT 'F',
    cpf_cnpj VARCHAR(18),
    rg_ie VARCHAR(20),
    telefone VARCHAR(20),
    telefone2 VARCHAR(20),
    email VARCHAR(100),
    logradouro VARCHAR(200),
    numero VARCHAR(10),
    complemento VARCHAR(100),
    bairro VARCHAR(100),
    cidade VARCHAR(100),
    estado CHAR(2),
    cep VARCHAR(10),
    data_nascimento DATE,
    limite_credito DECIMAL(12,2) DEFAULT 0.00,
    saldo_devedor DECIMAL(12,2) DEFAULT 0.00,
    data_cadastro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    data_ultima_compra DATE,
    total_compras DECIMAL(12,2) DEFAULT 0.00,
    ativo BOOLEAN DEFAULT TRUE,
    observacoes TEXT
);

CREATE TABLE IF NOT EXISTS categoria (
    id_categoria SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    nome VARCHAR(100) NOT NULL,
    descricao TEXT,
    categoria_pai_id INTEGER REFERENCES categoria(id_categoria),
    nivel INTEGER DEFAULT 1,
    ativo BOOLEAN DEFAULT TRUE,
    ordem_exibicao INTEGER DEFAULT 0,
    data_cadastro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(empresa_id, nome)
);

CREATE TABLE IF NOT EXISTS produto (
    id_produto SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    categoria_id INTEGER REFERENCES categoria(id_categoria),
    codigo_barras VARCHAR(50),
    codigo_interno VARCHAR(50),
    nome VARCHAR(200) NOT NULL,
    descricao TEXT,
    marca VARCHAR(100),
    unidade_venda VARCHAR(10) DEFAULT 'UN',
    unidade_compra VARCHAR(10) DEFAULT 'UN',
    fator_conversao DECIMAL(10,4) DEFAULT 1.0,
    estoque_atual DECIMAL(12,4) NOT NULL DEFAULT 0,
    estoque_minimo DECIMAL(12,4) DEFAULT 0,
    estoque_maximo DECIMAL(12,4),
    controlar_estoque BOOLEAN DEFAULT TRUE,
    preco_custo DECIMAL(12,2) DEFAULT 0.00,
    preco_venda DECIMAL(12,2) NOT NULL,
    preco_promocional DECIMAL(12,2),
    data_inicio_promocao DATE,
    data_fim_promocao DATE,
    margem_lucro DECIMAL(5,2),
    ncm VARCHAR(10),
    data_cadastro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    data_ultima_compra DATE,
    data_ultima_venda DATE,
    foto_principal_url VARCHAR(500),
    ativo BOOLEAN DEFAULT TRUE,
    destacado BOOLEAN DEFAULT FALSE,
    observacoes TEXT
);

CREATE TABLE IF NOT EXISTS fornecedor (
    id_fornecedor SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    razao_social VARCHAR(200) NOT NULL,
    nome_fantasia VARCHAR(200),
    cnpj VARCHAR(18),
    inscricao_estadual VARCHAR(20),
    telefone VARCHAR(20),
    telefone2 VARCHAR(20),
    email VARCHAR(100),
    logradouro VARCHAR(200),
    numero VARCHAR(10),
    bairro VARCHAR(100),
    cidade VARCHAR(100),
    estado CHAR(2),
    cep VARCHAR(10),
    nome_contato VARCHAR(100),
    telefone_contato VARCHAR(20),
    prazo_entrega INTEGER DEFAULT 7,
    prazo_pagamento INTEGER DEFAULT 30,
    limite_credito DECIMAL(12,2) DEFAULT 0.00,
    data_cadastro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    data_ultima_compra DATE,
    total_compras DECIMAL(12,2) DEFAULT 0.00,
    ativo BOOLEAN DEFAULT TRUE,
    observacoes TEXT
);

CREATE TABLE IF NOT EXISTS venda (
    id_venda SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    sessao_caixa_id INTEGER NOT NULL REFERENCES sessao_caixa(id_sessao),
    usuario_id INTEGER NOT NULL REFERENCES usuario(id_usuario),
    caixa_fisico_id INTEGER NOT NULL REFERENCES caixa_fisico(id_caixa_fisico),
    numero_venda VARCHAR(50) NOT NULL,
    cliente_id INTEGER REFERENCES cliente(id_cliente),
    cliente_nome VARCHAR(150),
    cliente_documento VARCHAR(20),
    data_venda TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_cancelamento TIMESTAMP,
    valor_total_produtos DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    valor_total_descontos DECIMAL(12,2) DEFAULT 0.00,
    valor_total_acrescimos DECIMAL(12,2) DEFAULT 0.00,
    valor_subtotal DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    valor_frete DECIMAL(12,2) DEFAULT 0.00,
    valor_total DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    valor_pago DECIMAL(12,2) DEFAULT 0.00,
    valor_troco DECIMAL(12,2) DEFAULT 0.00,
    status VARCHAR(20) NOT NULL DEFAULT 'pendente',
    tipo_venda VARCHAR(20) DEFAULT 'venda',
    tipo_emissao VARCHAR(20) DEFAULT 'presencial',
    observacoes TEXT,
    motivo_cancelamento TEXT,
    UNIQUE(empresa_id, numero_venda)
);

CREATE TABLE IF NOT EXISTS item_venda (
    id_item_venda SERIAL PRIMARY KEY,
    venda_id INTEGER NOT NULL REFERENCES venda(id_venda) ON DELETE CASCADE,
    produto_id INTEGER NOT NULL REFERENCES produto(id_produto),
    sequencia INTEGER NOT NULL,
    quantidade DECIMAL(12,4) NOT NULL,
    unidade_venda VARCHAR(10) DEFAULT 'UN',
    fator_conversao DECIMAL(10,4) DEFAULT 1.0,
    preco_unitario DECIMAL(12,2) NOT NULL,
    preco_custo_unitario DECIMAL(12,2),
    valor_total DECIMAL(12,2) NOT NULL,
    valor_desconto DECIMAL(12,2) DEFAULT 0.00,
    valor_desconto_percentual DECIMAL(5,2) DEFAULT 0.00,
    valor_acrescimo DECIMAL(12,2) DEFAULT 0.00,
    valor_liquido DECIMAL(12,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'vendido',
    data_hora TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    data_cancelamento TIMESTAMP,
    motivo_cancelamento TEXT,
    usuario_cancelamento_id INTEGER REFERENCES usuario(id_usuario),
    UNIQUE(venda_id, sequencia)
);

CREATE TABLE IF NOT EXISTS venda_pagamento (
    id_venda_pagamento SERIAL PRIMARY KEY,
    venda_id INTEGER NOT NULL REFERENCES venda(id_venda) ON DELETE CASCADE,
    forma_pagamento_id INTEGER NOT NULL REFERENCES forma_pagamento(id_forma_pagamento),
    valor DECIMAL(12,2) NOT NULL,
    troco_para DECIMAL(12,2) DEFAULT 0.00,
    autorizacao VARCHAR(100),
    bandeira_cartao VARCHAR(50),
    numero_cartao VARCHAR(4),
    parcelas INTEGER DEFAULT 1,
    valor_parcela DECIMAL(12,2),
    data_processamento TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'aprovado'
);

CREATE TABLE IF NOT EXISTS estoque_movimentacao (
    id_movimentacao SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    produto_id INTEGER NOT NULL REFERENCES produto(id_produto),
    tipo_movimentacao VARCHAR(20) NOT NULL,
    quantidade DECIMAL(12,4) NOT NULL,
    saldo_anterior DECIMAL(12,4) NOT NULL,
    saldo_atual DECIMAL(12,4) NOT NULL,
    origem_tipo VARCHAR(30),
    origem_id INTEGER,
    data_movimentacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    usuario_id INTEGER REFERENCES usuario(id_usuario),
    observacao TEXT,
    lote VARCHAR(50),
    data_validade_lote DATE
);

CREATE TABLE IF NOT EXISTS compra (
    id_compra SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    fornecedor_id INTEGER REFERENCES fornecedor(id_fornecedor),
    usuario_id INTEGER NOT NULL REFERENCES usuario(id_usuario),
    numero_nota_fiscal VARCHAR(50),
    serie_nota VARCHAR(10),
    chave_nfe VARCHAR(44),
    data_emissao DATE,
    data_entrada DATE NOT NULL,
    data_cadastro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    valor_produtos DECIMAL(12,2) DEFAULT 0.00,
    valor_frete DECIMAL(12,2) DEFAULT 0.00,
    valor_desconto DECIMAL(12,2) DEFAULT 0.00,
    valor_total DECIMAL(12,2) DEFAULT 0.00,
    status VARCHAR(20) DEFAULT 'pendente',
    observacoes TEXT
);

CREATE TABLE IF NOT EXISTS item_compra (
    id_item_compra SERIAL PRIMARY KEY,
    compra_id INTEGER NOT NULL REFERENCES compra(id_compra) ON DELETE CASCADE,
    produto_id INTEGER NOT NULL REFERENCES produto(id_produto),
    sequencia INTEGER NOT NULL,
    quantidade DECIMAL(12,4) NOT NULL,
    quantidade_recebida DECIMAL(12,4) DEFAULT 0.00,
    preco_unitario DECIMAL(12,4) NOT NULL,
    valor_total DECIMAL(12,2) NOT NULL,
    valor_desconto DECIMAL(12,2) DEFAULT 0.00,
    data_recebimento TIMESTAMP,
    UNIQUE(compra_id, sequencia)
);

CREATE TABLE IF NOT EXISTS conta_pagar (
    id_conta_pagar SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    fornecedor_id INTEGER REFERENCES fornecedor(id_fornecedor),
    compra_id INTEGER REFERENCES compra(id_compra),
    descricao VARCHAR(200) NOT NULL,
    documento VARCHAR(50),
    parcela VARCHAR(10),
    valor_original DECIMAL(12,2) NOT NULL,
    valor_pago DECIMAL(12,2) DEFAULT 0.00,
    data_vencimento DATE NOT NULL,
    data_pagamento DATE,
    status VARCHAR(20) DEFAULT 'aberta',
    categoria VARCHAR(50) DEFAULT 'fornecedor',
    observacoes TEXT,
    data_cadastro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    usuario_id INTEGER REFERENCES usuario(id_usuario)
);

CREATE TABLE IF NOT EXISTS conta_receber (
    id_conta_receber SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    cliente_id INTEGER REFERENCES cliente(id_cliente),
    venda_id INTEGER REFERENCES venda(id_venda),
    descricao VARCHAR(200) NOT NULL,
    parcela VARCHAR(10),
    valor_original DECIMAL(12,2) NOT NULL,
    valor_recebido DECIMAL(12,2) DEFAULT 0.00,
    data_vencimento DATE NOT NULL,
    data_recebimento DATE,
    status VARCHAR(20) DEFAULT 'aberta',
    observacoes TEXT,
    data_cadastro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    usuario_id INTEGER REFERENCES usuario(id_usuario)
);

CREATE TABLE IF NOT EXISTS caixa_movimentacao (
    id_caixa_movimentacao SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    sessao_caixa_id INTEGER NOT NULL REFERENCES sessao_caixa(id_sessao),
    tipo VARCHAR(20) NOT NULL,
    valor DECIMAL(12,2) NOT NULL,
    forma_pagamento_id INTEGER REFERENCES forma_pagamento(id_forma_pagamento),
    motivo VARCHAR(100),
    observacao TEXT,
    data_movimentacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    usuario_id INTEGER NOT NULL REFERENCES usuario(id_usuario),
    venda_id INTEGER REFERENCES venda(id_venda)
);

CREATE TABLE IF NOT EXISTS inventario (
    id_inventario SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    codigo VARCHAR(20) NOT NULL,
    descricao VARCHAR(200),
    data_inicio DATE NOT NULL,
    data_fim DATE,
    data_fechamento TIMESTAMP,
    status VARCHAR(20) DEFAULT 'aberto',
    observacoes TEXT,
    usuario_id INTEGER REFERENCES usuario(id_usuario),
    UNIQUE(empresa_id, codigo)
);

CREATE TABLE IF NOT EXISTS inventario_item (
    id_inventario_item SERIAL PRIMARY KEY,
    inventario_id INTEGER NOT NULL REFERENCES inventario(id_inventario) ON DELETE CASCADE,
    produto_id INTEGER NOT NULL REFERENCES produto(id_produto),
    quantidade_sistema DECIMAL(12,4) NOT NULL,
    quantidade_fisica DECIMAL(12,4),
    contado BOOLEAN DEFAULT FALSE,
    data_contagem TIMESTAMP,
    usuario_contagem_id INTEGER REFERENCES usuario(id_usuario),
    observacao TEXT,
    UNIQUE(inventario_id, produto_id)
);

CREATE TABLE IF NOT EXISTS configuracao (
    id_config SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    chave VARCHAR(100) NOT NULL,
    valor TEXT,
    tipo VARCHAR(50) DEFAULT 'texto',
    categoria VARCHAR(50) DEFAULT 'geral',
    descricao TEXT,
    data_atualizacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(empresa_id, chave)
);

CREATE TABLE IF NOT EXISTS auditoria (
    id_auditoria SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    tabela VARCHAR(50) NOT NULL,
    acao VARCHAR(20) NOT NULL,
    registro_id INTEGER,
    valores_antigos JSONB,
    valores_novos JSONB,
    data_hora TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    usuario_id INTEGER REFERENCES usuario(id_usuario),
    ip_address INET,
    user_agent TEXT
);

CREATE TABLE IF NOT EXISTS backup (
    id_backup SERIAL PRIMARY KEY,
    empresa_id INTEGER NOT NULL REFERENCES empresa(id_empresa) ON DELETE CASCADE,
    nome_arquivo VARCHAR(200) NOT NULL,
    caminho VARCHAR(500) NOT NULL,
    tamanho BIGINT,
    data_backup TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    tipo VARCHAR(20) DEFAULT 'automatico',
    status VARCHAR(20) DEFAULT 'sucesso',
    observacoes TEXT,
    usuario_id INTEGER REFERENCES usuario(id_usuario)
);

-- Triggers
CREATE OR REPLACE FUNCTION gerar_numero_venda()
RETURNS TRIGGER AS $$
DECLARE seq_num INTEGER; data_str VARCHAR(8);
BEGIN
    data_str := TO_CHAR(CURRENT_DATE, 'YYYYMMDD');
    SELECT COALESCE(MAX(SUBSTRING(numero_venda FROM '-(\d+)$')::INTEGER), 0) + 1
    INTO seq_num FROM venda WHERE empresa_id = NEW.empresa_id AND numero_venda LIKE 'V' || data_str || '-%';
    NEW.numero_venda := 'V' || data_str || '-' || LPAD(seq_num::TEXT, 6, '0');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS tr_gerar_numero_venda ON venda;
CREATE TRIGGER tr_gerar_numero_venda BEFORE INSERT ON venda
FOR EACH ROW WHEN (NEW.numero_venda IS NULL OR NEW.numero_venda = '')
EXECUTE FUNCTION gerar_numero_venda();

CREATE OR REPLACE FUNCTION gerar_codigo_sessao()
RETURNS TRIGGER AS $$
DECLARE caixa_cod VARCHAR(20); seq_num INTEGER;
BEGIN
    SELECT codigo INTO caixa_cod FROM caixa_fisico WHERE id_caixa_fisico = NEW.caixa_fisico_id;
    SELECT COALESCE(COUNT(*), 0) + 1 INTO seq_num FROM sessao_caixa WHERE empresa_id = NEW.empresa_id AND DATE(data_abertura) = CURRENT_DATE;
    NEW.codigo_sessao := 'S' || TO_CHAR(CURRENT_DATE, 'YYYYMMDD') || '-' || caixa_cod || '-' || LPAD(seq_num::TEXT, 3, '0');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS tr_gerar_codigo_sessao ON sessao_caixa;
CREATE TRIGGER tr_gerar_codigo_sessao BEFORE INSERT ON sessao_caixa
FOR EACH ROW WHEN (NEW.codigo_sessao IS NULL OR NEW.codigo_sessao = '')
EXECUTE FUNCTION gerar_codigo_sessao();

-- Dados iniciais
INSERT INTO empresa (razao_social, nome_fantasia, cnpj, logradouro, numero, bairro, cidade, estado, cep, telefone, email)
VALUES ('MERCADO CENTRAL LTDA', 'Mercado Central', '12.345.678/0001-99', 'Rua das Flores', '123', 'Centro', 'São Paulo', 'SP', '01234-567', '(11) 3333-4444', 'contato@mercadocentral.com.br')
ON CONFLICT (cnpj) DO NOTHING;

-- Senha: 123456 (hash bcrypt)
INSERT INTO usuario (empresa_id, nome, login, senha_hash, perfil, pode_abrir_caixa, pode_fechar_caixa, pode_dar_desconto, pode_cancelar_venda)
VALUES
(1, 'Administrador', 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin', TRUE, TRUE, TRUE, TRUE),
(1, 'Gerente', 'gerente', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'gerente', TRUE, TRUE, TRUE, TRUE),
(1, 'Caixa 01', 'caixa01', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'caixa', FALSE, FALSE, FALSE, FALSE),
(1, 'Supervisor', 'supervisor', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'supervisor', TRUE, TRUE, TRUE, TRUE)
ON CONFLICT (empresa_id, login) DO NOTHING;

INSERT INTO caixa_fisico (empresa_id, codigo, nome, localizacao) VALUES
(1, 'CAIXA-01', 'Caixa Principal', 'Entrada'),
(1, 'CAIXA-02', 'Caixa Secundário', 'Fundo')
ON CONFLICT (empresa_id, codigo) DO NOTHING;

INSERT INTO forma_pagamento (empresa_id, nome, codigo, tipo, ordem_exibicao) VALUES
(1, 'Dinheiro', '01', 'dinheiro', 1),
(1, 'Cartão Débito', '02', 'cartao_debito', 2),
(1, 'Cartão Crédito', '03', 'cartao_credito', 3),
(1, 'PIX', '04', 'pix', 4)
ON CONFLICT (empresa_id, codigo) DO NOTHING;

INSERT INTO categoria (empresa_id, nome, descricao, ordem_exibicao) VALUES
(1, 'Alimentos', 'Alimentos em geral', 1),
(1, 'Bebidas', 'Bebidas diversas', 2),
(1, 'Limpeza', 'Produtos de limpeza', 3),
(1, 'Higiene', 'Produtos de higiene pessoal', 4)
ON CONFLICT (empresa_id, nome) DO NOTHING;

INSERT INTO produto (empresa_id, codigo_barras, nome, categoria_id, unidade_venda, estoque_minimo, estoque_atual, preco_custo, preco_venda) VALUES
(1, '7891000315507', 'Arroz Tipo 1 5kg', 1, 'UN', 10, 50, 15.00, 22.90),
(1, '7891910000197', 'Feijão Carioca 1kg', 1, 'UN', 15, 80, 6.50, 8.90),
(1, '7891999010016', 'Açúcar Cristal 1kg', 1, 'UN', 20, 100, 3.20, 4.50),
(1, '7891000053508', 'Óleo de Soja 900ml', 1, 'UN', 30, 150, 4.80, 6.90)
ON CONFLICT DO NOTHING;

INSERT INTO configuracao (empresa_id, chave, valor, descricao) VALUES
(1, 'sistema.versao', '1.0.0', 'Versão do sistema'),
(1, 'caixa.imprimir_comprovante', 'true', 'Imprimir comprovante'),
(1, 'estoque.alerta_minimo', 'true', 'Alertar estoque baixo')
ON CONFLICT (empresa_id, chave) DO NOTHING;

-- Índices
CREATE INDEX IF NOT EXISTS idx_produto_empresa_nome ON produto(empresa_id, nome);
CREATE INDEX IF NOT EXISTS idx_produto_codigo_barras ON produto(codigo_barras);
CREATE INDEX IF NOT EXISTS idx_venda_data ON venda(data_venda);
CREATE INDEX IF NOT EXISTS idx_venda_sessao ON venda(sessao_caixa_id);
CREATE INDEX IF NOT EXISTS idx_item_venda_venda ON item_venda(venda_id);
-- View para Fluxo de Caixa Consolidado
CREATE OR REPLACE VIEW vw_fluxo_caixa AS
-- Entradas de Vendas (Dinheiro/PIX imediato no caixa)
SELECT 
    DATE(vp.data_processamento) as data,
    'venda' as tipo,
    vp.valor as valor,
    v.empresa_id
FROM venda_pagamento vp
JOIN venda v ON vp.venda_id = v.id_venda
WHERE vp.status = 'aprovado'

UNION ALL

-- Entradas de Contas Recebidas
SELECT 
    data_recebimento as data,
    'recebimento' as tipo,
    valor_recebido as valor,
    empresa_id
FROM conta_receber
WHERE status = 'recebida'

UNION ALL

-- Saídas de Contas Pagas (Fornecedores/Despesas)
SELECT 
    data_pagamento as data,
    'pagamento' as tipo,
    -valor_pago as valor,
    empresa_id
FROM conta_pagar
WHERE status = 'paga'

UNION ALL

-- Movimentações Manuais (Sangrias/Suprimentos)
SELECT 
    DATE(data_movimentacao) as data,
    tipo,
    CASE WHEN tipo = 'sangria' THEN -valor ELSE valor END as valor,
    empresa_id
FROM caixa_movimentacao
WHERE tipo IN ('sangria', 'suprimento');

-- Índices adicionais para performance financeira
CREATE INDEX IF NOT EXISTS idx_conta_pagar_vencimento ON conta_pagar(empresa_id, data_vencimento);
CREATE INDEX IF NOT EXISTS idx_conta_receber_vencimento ON conta_receber(empresa_id, data_vencimento);
CREATE INDEX IF NOT EXISTS idx_venda_pagamento_data ON venda_pagamento(data_processamento);
