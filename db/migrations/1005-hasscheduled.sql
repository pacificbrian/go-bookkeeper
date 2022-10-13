-- +migrate Up

ALTER TABLE `accounts` ADD COLUMN 'has_scheduled' tinyint(1);

-- +migrate Down

ALTER TABLE `accounts` DROP COLUMN 'has_scheduled';
