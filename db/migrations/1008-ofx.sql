-- +migrate Up

ALTER TABLE `accounts` ADD COLUMN `client_uid` varchar(64) DEFAULT NULL;

CREATE TABLE IF NOT EXISTS `institutions` (
  `id` integer PRIMARY KEY,
  `app_ver` integer DEFAULT NULL,
  `fi_id` integer DEFAULT NULL,
  `app_id` varchar(64) DEFAULT NULL,
  `fi_org` varchar(64) DEFAULT NULL,
  `fi_url` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL
);

INSERT INTO `institutions` (id, name, fi_id, fi_org, fi_url) VALUES
  (1, 'American Express', 3101, 'AMEX', 'https://online.americanexpress.com/myca/ofxdl/desktop/desktopDownload.do?request_type=nl_ofxdownload'),
  (2, 'Citi Credit Card', 24909, 'Citigroup', 'https://mobilesoa.citi.com/CitiOFXInterface'),
  (3, 'Wells Fargo', 3000, 'WF', 'https://ofxdc.wellsfargo.com/ofx/process.ofx');

-- +migrate Down

ALTER TABLE `accounts` DROP COLUMN `client_uid`;
DROP TABLE `institutions`;
