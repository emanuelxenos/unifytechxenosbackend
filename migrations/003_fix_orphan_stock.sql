-- Script de Consistência: Criar lotes para produtos com saldo mas sem registros em estoque_lote
-- Isso resolve o problema de produtos aparecendo como "N/A" ou "Sem lote ativo"

DO $$
DECLARE
    r RECORD;
    v_lote_id INTEGER;
BEGIN
    FOR r IN 
        SELECT id_produto, empresa_id, estoque_atual, localizacao, data_vencimento 
        FROM produto 
        WHERE estoque_atual > 0 
          AND id_produto NOT IN (SELECT DISTINCT produto_id FROM estoque_lote WHERE status = 'ativo')
          AND controlar_estoque = true
    LOOP
        -- Criar um lote inicial para o saldo órfão
        INSERT INTO estoque_lote (
            empresa_id, 
            produto_id, 
            lote_interno, 
            lote_fabricante, 
            quantidade_inicial, 
            quantidade_atual, 
            data_vencimento, 
            status, 
            observacao
        ) VALUES (
            r.empresa_id, 
            r.id_produto, 
            'SALDO-INICIAL-' || TO_CHAR(NOW(), 'YYYYMMDD'), 
            'S/L', 
            r.estoque_atual, 
            r.estoque_atual, 
            COALESCE(r.data_vencimento, CURRENT_DATE + INTERVAL '1 year'), 
            'ativo', 
            'Lote gerado automaticamente para regularizar saldo órfão'
        ) RETURNING id_lote INTO v_lote_id;

        -- Registrar a movimentação para manter o histórico íntegro
        INSERT INTO estoque_movimentacao (
            empresa_id, 
            produto_id, 
            tipo_movimentacao, 
            quantidade, 
            saldo_anterior, 
            saldo_atual, 
            origem_tipo, 
            usuario_id, 
            observacao, 
            lote_id
        ) VALUES (
            r.empresa_id, 
            r.id_produto, 
            'ajuste', 
            r.estoque_atual, 
            0, 
            r.estoque_atual, 
            'sistema', 
            NULL, 
            'Regularização de saldo legado para sistema de lotes', 
            v_lote_id
        );
    END LOOP;
END $$;
