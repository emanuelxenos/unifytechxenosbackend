-- Migração para adicionar campos de rastreabilidade na tabela item_compra
-- Esses campos permitem salvar o lote, localização e vencimento informados no lançamento da nota

ALTER TABLE item_compra ADD COLUMN IF NOT EXISTS localizacao VARCHAR(100);
ALTER TABLE item_compra ADD COLUMN IF NOT EXISTS data_vencimento DATE;
ALTER TABLE item_compra ADD COLUMN IF NOT EXISTS lote VARCHAR(100);
