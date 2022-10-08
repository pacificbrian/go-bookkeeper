-- +migrate Up

ALTER TABLE `repeat_intervals` ADD COLUMN 'start_day' int(11);

-- +migrate Down

ALTER TABLE `repeat_intervals` DROP COLUMN 'start_day';
