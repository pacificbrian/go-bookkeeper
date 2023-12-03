-- +migrate Up

ALTER TABLE `trades` ADD COLUMN `tainted` tinyint(1) DEFAULT 0;
ALTER TABLE `trade_gains` ADD COLUMN `basis_fifo` decimal(16,4) DEFAULT NULL;

-- +migrate Down

ALTER TABLE `trades` DROP COLUMN `tainted`;
ALTER TABLE `trade_gains` DROP COLUMN `basis_fifo`;
