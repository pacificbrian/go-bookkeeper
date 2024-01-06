-- +migrate Up

ALTER TABLE `securities` ADD COLUMN `import_name` varchar(255) DEFAULT NULL;

-- +migrate Down

ALTER TABLE `securities` DROP COLUMN `import_name`;
