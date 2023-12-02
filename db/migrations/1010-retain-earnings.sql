-- +migrate Up

ALTER TABLE `securities` ADD COLUMN `accumulated_basis` decimal(16,4) DEFAULT 0.0000;
ALTER TABLE `securities` ADD COLUMN `retained_earnings` decimal(16,4) DEFAULT 0.0000;

-- +migrate Down

ALTER TABLE `securities` DROP COLUMN `accumulated_basis`;
ALTER TABLE `securities` DROP COLUMN `retained_earnings`;
