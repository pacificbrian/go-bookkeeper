-- +migrate Up

ALTER TABLE `payees` ADD COLUMN `import_name` varchar(255) DEFAULT NULL;
ALTER TABLE `securities` ADD COLUMN `import_name` varchar(255) DEFAULT NULL;

-- +migrate Down

ALTER TABLE `payees` DROP COLUMN `import_name`;
ALTER TABLE `securities` DROP COLUMN `import_name`;
