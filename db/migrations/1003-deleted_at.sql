-- +migrate Up

ALTER TABLE `accounts` ADD COLUMN `deleted_at` datetime;
ALTER TABLE `cash_flows` ADD COLUMN `deleted_at` datetime;
ALTER TABLE `trades` ADD COLUMN `deleted_at` datetime;
ALTER TABLE `users` ADD COLUMN `deleted_at` datetime;

-- +migrate Down

ALTER TABLE `accounts` DROP COLUMN `deleted_at`;
ALTER TABLE `cash_flows` DROP COLUMN `deleted_at`;
ALTER TABLE `trades` DROP COLUMN `deleted_at`;
ALTER TABLE `users` DROP COLUMN `deleted_at`;
