-- +migrate Up

CREATE TABLE IF NOT EXISTS `account_types` (
  `id` integer PRIMARY KEY,
  `name` varchar(255) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `accounts` (
  `id` integer PRIMARY KEY,
  `account_type_id` int(11) DEFAULT 1,
  `name` varchar(255) DEFAULT NULL,
  `holder` int(11) DEFAULT NULL,
  `number` varchar(255) DEFAULT NULL,
  `routing` int(11) DEFAULT NULL,
  `cash_balance` decimal(16,4) DEFAULT 0.0000,
  `balance` decimal(16,4) DEFAULT 0.0000,
  `currency_type_id` int(11) DEFAULT 1,
  `payee_length` int(11) DEFAULT NULL,
  `transnum_shift` int(11) DEFAULT 0,
  `taxable` tinyint(1) DEFAULT NULL,
  `hidden` tinyint(1) DEFAULT NULL,
  `watchlist` tinyint(1) DEFAULT NULL,
  `created_at` datetime NOT NULL default current_timestamp,
  `updated_at` datetime NOT NULL default current_timestamp,
  `user_id` int(11) DEFAULT NULL,
  `institution_id` int(11) DEFAULT NULL,
  `ofx_index` int(11) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `cash_flow_types` (
  `id` integer PRIMARY KEY,
  `name` varchar(255) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `cash_flows` (
  `id` integer PRIMARY KEY,
  `date` date DEFAULT NULL,
  `amount` decimal(16,4) DEFAULT NULL,
  `report_amount` decimal(16,4) DEFAULT NULL,
  `account_id` int(11) DEFAULT NULL,
  `payee_id` int(11) DEFAULT NULL,
  `category_id` int(11) DEFAULT NULL,
  `split_from` int(11) DEFAULT 0,
  `split` tinyint(1) DEFAULT 0,
  `transfer` tinyint(1) DEFAULT 0,
  `transnum` varchar(255) DEFAULT NULL,
  `memo` varchar(255) DEFAULT NULL,
  `created_at` datetime NOT NULL default current_timestamp,
  `updated_at` datetime NOT NULL default current_timestamp,
  `repeat_interval_id` int(11) DEFAULT NULL,
  `type` varchar(255) DEFAULT NULL,
  `import_id` int(11) DEFAULT NULL,
  `needs_review` tinyint(1) DEFAULT NULL,
  `tax_year` int(11) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `categories` (
  `id` integer PRIMARY KEY,
  `name` varchar(255) DEFAULT NULL,
  `category_type_id` int(11) DEFAULT NULL,
  `omit_from_pie` tinyint(1) DEFAULT NULL,
  `user_id` int(11) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `category_types` (
  `id` integer PRIMARY KEY,
  `name` varchar(255) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `companies` (
  `id` integer PRIMARY KEY,
  `symbol` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `currency_types` (
  `id` integer PRIMARY KEY,
  `name` varchar(255) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `imports` (
  `id` integer PRIMARY KEY,
  `account_id` int(11) DEFAULT NULL,
  `created_on` datetime DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `institutions` (
  `id` integer PRIMARY KEY,
  `name` varchar(255) DEFAULT NULL,
  `client_uid` varchar(32) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `ofx_accounts` (
  `id` integer PRIMARY KEY,
  `account_id` int(11) DEFAULT NULL,
  `institution_id` int(11) DEFAULT NULL,
  `login` varchar(255) DEFAULT NULL,
  `password` varchar(255) DEFAULT NULL,
  `payee_length` int(11) DEFAULT NULL,
  `transnum_shift` int(11) DEFAULT 0,
  `client_uid` varchar(32) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `payees` (
  `id` integer PRIMARY KEY,
  `name` varchar(255) DEFAULT NULL,
  `address` varchar(255) DEFAULT NULL,
  `category_id` int(11) DEFAULT 1,
  `cash_flow_count` int(11) DEFAULT NULL,
  `skip_on_import` tinyint(1) DEFAULT NULL,
  `user_id` int(11) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `repeat_interval_types` (
  `id` integer PRIMARY KEY,
  `name` varchar(255) DEFAULT NULL,
  `days` int(11) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `repeat_intervals` (
  `id` integer PRIMARY KEY,
  `cash_flow_id` int(11) DEFAULT NULL,
  `repeat_interval_type_id` int(11) DEFAULT NULL,
  `repeats_left` int(11) DEFAULT NULL,
  `rate` decimal(12,4) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `securities` (
  `id` integer PRIMARY KEY,
  `account_id` int(11) DEFAULT NULL,
  `security_type_id` int(11) DEFAULT '1',
  `security_basis_type_id` int(11) DEFAULT '1',
  `company_id` int(11) DEFAULT NULL,
  `shares` decimal(14,4) DEFAULT '0.0000',
  `basis` decimal(16,4) DEFAULT '0.0000',
  `value` decimal(16,4) DEFAULT '0.0000',
  `last_quote_update` date DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `security_basis_types` (
  `id` integer PRIMARY KEY,
  `name` varchar(255) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `security_types` (
  `id` integer PRIMARY KEY,
  `name` varchar(255) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `tax_cash_flows` (
  `id` integer PRIMARY KEY,
  `tax_id` int(11) DEFAULT NULL,
  `cash_flow_id` int(11) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `tax_categories` (
  `id` integer PRIMARY KEY,
  `tax_item_id` int(11) DEFAULT NULL,
  `category_id` int(11) DEFAULT NULL,
  `trade_type_id` int(11) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `tax_constants` (
  `id` integer PRIMARY KEY,
  `tax_form_id` int(11) DEFAULT NULL,
  `interest_allowed` int(11) DEFAULT NULL,
  `dividend_allowed` int(11) DEFAULT NULL,
  `standard_dependent_base` int(11) DEFAULT NULL,
  `standard_dependent_extra` int(11) DEFAULT NULL,
  `exemption_hi_agi_s` int(11) DEFAULT NULL,
  `exemption_hi_agi_mfs` int(11) DEFAULT NULL,
  `exemption_hi_agi_mfj` int(11) DEFAULT NULL,
  `exemption_hi_agi_hh` int(11) DEFAULT NULL,
  `exemption_mid_amount_s` int(11) DEFAULT NULL,
  `exemption_mid_amount_mfs` int(11) DEFAULT NULL,
  `exemption_mid_amount_mfj` int(11) DEFAULT NULL,
  `exemption_mid_amount_hh` int(11) DEFAULT NULL,
  `exemption_mid_rate` decimal(12,4) DEFAULT NULL,
  `capgain_collectible_rate` decimal(12,4) DEFAULT NULL,
  `capgain_unrecaptured_rate` decimal(12,4) DEFAULT NULL,
  `caploss_limit_s` int(11) DEFAULT NULL,
  `caploss_limit_mfs` int(11) DEFAULT NULL,
  `caploss_limit_mfj` int(11) DEFAULT NULL,
  `caploss_limit_hh` int(11) DEFAULT NULL,
  `amt_mid_limit_s` int(11) DEFAULT NULL,
  `amt_mid_limit_mfs` int(11) DEFAULT NULL,
  `amt_mid_limit_mfj` int(11) DEFAULT NULL,
  `amt_mid_limit_hh` int(11) DEFAULT NULL,
  `amt_high_limit_s` int(11) DEFAULT NULL,
  `amt_high_limit_mfs` int(11) DEFAULT NULL,
  `amt_high_limit_mfj` int(11) DEFAULT NULL,
  `amt_high_limit_hh` int(11) DEFAULT NULL,
  `item_medical_rate` decimal(12,4) DEFAULT NULL,
  `item_jobmisc_rate` decimal(12,4) DEFAULT NULL,
  `item_casualty_theft_min` int(11) DEFAULT NULL,
  `item_casualty_theft_rate` decimal(12,4) DEFAULT NULL,
  `item_limit_rate` decimal(12,4) DEFAULT NULL,
  `item_limit_upper_rate` decimal(12,4) DEFAULT NULL,
  `amt_medical_rate` decimal(12,4) DEFAULT NULL,
  `amt_low_rate` decimal(12,4) DEFAULT NULL,
  `amt_mid_rate` decimal(12,4) DEFAULT NULL,
  `tax_table_10` int(11) DEFAULT NULL,
  `tax_table_25` int(11) DEFAULT NULL,
  `tax_table_50` int(11) DEFAULT NULL,
  `tax_table_max` int(11) DEFAULT NULL,
  `tax_l1_rate` decimal(12,4) DEFAULT NULL,
  `tax_l2_rate` decimal(12,4) DEFAULT NULL,
  `tax_l3_rate` decimal(12,4) DEFAULT NULL,
  `tax_l4_rate` decimal(12,4) DEFAULT NULL,
  `tax_l5_rate` decimal(12,4) DEFAULT NULL,
  `tax_l6_rate` decimal(12,4) DEFAULT NULL,
  `tax_l7_rate` decimal(12,4) DEFAULT NULL,
  `capgain_rate` decimal(12,4) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `tax_filing_status` (
  `id` integer PRIMARY KEY,
  `name` varchar(255) DEFAULT NULL,
  `label` varchar(20) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `tax_items` (
  `id` integer PRIMARY KEY,
  `name` varchar(255) DEFAULT NULL,
  `type` varchar(255) DEFAULT NULL,
  `tax_type_id` int(11) DEFAULT NULL,
  `tax_category_id` int(11) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `tax_regions` (
  `id` integer PRIMARY KEY,
  `name` varchar(255) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `tax_types` (
  `id` integer PRIMARY KEY,
  `name` varchar(255) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `tax_users` (
  `id` integer PRIMARY KEY,
  `user_id` int(11) DEFAULT NULL,
  `tax_region_id` int(11) DEFAULT NULL,
  `year` int(11) DEFAULT NULL,
  `filing_status` int(11) DEFAULT NULL,
  `exemptions` int(11) DEFAULT NULL,
  `income` decimal(16,4) DEFAULT NULL,
  `agi_income` decimal(16,4) DEFAULT NULL,
  `taxable_income` decimal(16,4) DEFAULT NULL,
  `for_agi` decimal(16,4) DEFAULT NULL,
  `from_agi` decimal(16,4) DEFAULT NULL,
  `standard_deduction` decimal(16,4) DEFAULT NULL,
  `itemized_deduction` decimal(16,4) DEFAULT NULL,
  `exemption` decimal(16,4) DEFAULT NULL,
  `credits` decimal(16,4) DEFAULT NULL,
  `payments` decimal(16,4) DEFAULT NULL,
  `base_tax` decimal(16,4) DEFAULT NULL,
  `other_tax` decimal(16,4) DEFAULT NULL,
  `owed_tax` decimal(16,4) DEFAULT NULL,
  `unpaid_tax` decimal(16,4) DEFAULT NULL,
  `long_capgain_income` decimal(16,4) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `tax_years` (
  `id` integer PRIMARY KEY,
  `year` int(11) DEFAULT NULL,
  `tax_form_id` int(11) DEFAULT NULL,
  `standard_deduction_s` int(11) DEFAULT NULL,
  `standard_deduction_mfs` int(11) DEFAULT NULL,
  `standard_deduction_mfj` int(11) DEFAULT NULL,
  `standard_deduction_hh` int(11) DEFAULT NULL,
  `standard_deduction_extra_s` int(11) DEFAULT NULL,
  `standard_deduction_extra_mfs` int(11) DEFAULT NULL,
  `standard_deduction_extra_mfj` int(11) DEFAULT NULL,
  `standard_deduction_extra_hh` int(11) DEFAULT NULL,
  `exemption_amount` int(11) DEFAULT NULL,
  `exemption_hi_amount` int(11) DEFAULT NULL,
  `exemption_mid_rate` decimal(12,4) DEFAULT NULL,
  `exemption_agi_s` int(11) DEFAULT NULL,
  `exemption_agi_mfs` int(11) DEFAULT NULL,
  `exemption_agi_mfj` int(11) DEFAULT NULL,
  `exemption_agi_hh` int(11) DEFAULT NULL,
  `item_limit_s` int(11) DEFAULT NULL,
  `item_limit_mfs` int(11) DEFAULT NULL,
  `item_limit_mfj` int(11) DEFAULT NULL,
  `item_limit_hh` int(11) DEFAULT NULL,
  `item_limit_rate` decimal(12,4) DEFAULT NULL,
  `capgain_rate` decimal(12,4) DEFAULT NULL,
  `capgain_ti_rate` decimal(12,4) DEFAULT NULL,
  `capgain_ti_limit_s` int(11) DEFAULT NULL,
  `capgain_ti_limit_mfs` int(11) DEFAULT NULL,
  `capgain_ti_limit_mfj` int(11) DEFAULT NULL,
  `capgain_ti_limit_hh` int(11) DEFAULT NULL,
  `amt_low_limit_s` int(11) DEFAULT NULL,
  `amt_low_limit_mfs` int(11) DEFAULT NULL,
  `amt_low_limit_mfj` int(11) DEFAULT NULL,
  `amt_low_limit_hh` int(11) DEFAULT NULL,
  `tax_income_l1_s` int(11) DEFAULT NULL,
  `tax_income_l2_s` int(11) DEFAULT NULL,
  `tax_income_l3_s` int(11) DEFAULT NULL,
  `tax_income_l4_s` int(11) DEFAULT NULL,
  `tax_income_l5_s` int(11) DEFAULT NULL,
  `tax_income_l1_mfs` int(11) DEFAULT NULL,
  `tax_income_l2_mfs` int(11) DEFAULT NULL,
  `tax_income_l3_mfs` int(11) DEFAULT NULL,
  `tax_income_l4_mfs` int(11) DEFAULT NULL,
  `tax_income_l5_mfs` int(11) DEFAULT NULL,
  `tax_income_l1_mfj` int(11) DEFAULT NULL,
  `tax_income_l2_mfj` int(11) DEFAULT NULL,
  `tax_income_l3_mfj` int(11) DEFAULT NULL,
  `tax_income_l4_mfj` int(11) DEFAULT NULL,
  `tax_income_l5_mfj` int(11) DEFAULT NULL,
  `tax_income_l1_hh` int(11) DEFAULT NULL,
  `tax_income_l2_hh` int(11) DEFAULT NULL,
  `tax_income_l3_hh` int(11) DEFAULT NULL,
  `tax_income_l4_hh` int(11) DEFAULT NULL,
  `tax_income_l5_hh` int(11) DEFAULT NULL,
  `tax_income_l6_s` int(11) DEFAULT NULL,
  `tax_income_l6_mfs` int(11) DEFAULT NULL,
  `tax_income_l6_mfj` int(11) DEFAULT NULL,
  `tax_income_l6_hh` int(11) DEFAULT NULL,
  `tax_l1_rate` decimal(12,4) DEFAULT NULL,
  `tax_l2_rate` decimal(12,4) DEFAULT NULL,
  `tax_l3_rate` decimal(12,4) DEFAULT NULL,
  `tax_l4_rate` decimal(12,4) DEFAULT NULL,
  `tax_l5_rate` decimal(12,4) DEFAULT NULL,
  `tax_l6_rate` decimal(12,4) DEFAULT NULL,
  `tax_l7_rate` decimal(12,4) DEFAULT NULL,
  `salt_maximum` int(11) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `taxes` (
  `id` integer PRIMARY KEY,
  `year` date DEFAULT NULL,
  `amount` decimal(16,4) DEFAULT NULL,
  `user_id` int(11) DEFAULT NULL,
  `tax_region_id` int(11) DEFAULT NULL,
  `tax_item_id` int(11) DEFAULT NULL,
  `tax_type_id` int(11) DEFAULT NULL,
  `memo` varchar(255) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `trade_gains` (
  `id` integer PRIMARY KEY,
  `sell_id` int(11) DEFAULT NULL,
  `buy_id` int(11) DEFAULT NULL,
  `days_held` int(11) DEFAULT NULL,
  `shares` decimal(14,4) DEFAULT NULL,
  `adjusted_shares` decimal(14,4) DEFAULT NULL,
  `basis` decimal(16,4) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `trade_types` (
  `id` integer PRIMARY KEY,
  `name` varchar(255) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `trades` (
  `id` integer PRIMARY KEY,
  `date` date DEFAULT NULL,
  `account_id` int(11) DEFAULT NULL,
  `security_id` int(11) DEFAULT NULL,
  `trade_type_id` int(11) DEFAULT NULL,
  `shares` decimal(14,4) DEFAULT NULL,
  `adjusted_shares` decimal(14,4) DEFAULT NULL,
  `amount` decimal(16,4) DEFAULT NULL,
  `price` decimal(16,4) DEFAULT NULL,
  `basis` decimal(16,4) DEFAULT NULL,
  `closed` tinyint(1) DEFAULT 0,
  `created_at` datetime NOT NULL default current_timestamp,
  `updated_at` datetime NOT NULL default current_timestamp,
  `import_id` int(11) DEFAULT NULL,
  `needs_review` tinyint(1) DEFAULT NULL,
  `tax_year` int(11) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `user_settings` (
  `id` integer PRIMARY KEY,
  `user_id` int(11) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `value` decimal(12,4) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `users` (
  `id` integer PRIMARY KEY,
  `login` varchar(255) DEFAULT NULL,
  `openid` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `first_name` varchar(80) DEFAULT NULL,
  `last_name` varchar(80) DEFAULT NULL,
  `password_digest` varchar(80) DEFAULT NULL,
  `salt` varchar(80) DEFAULT NULL,
  `remember_token` varchar(40) DEFAULT NULL,
  `remember_token_expires_at` datetime DEFAULT NULL,
  `activation_code` varchar(40) DEFAULT NULL,
  `activated_at` datetime DEFAULT NULL,
  `cashflow_limit` int(11) DEFAULT 200,
  `created_at` datetime NOT NULL default current_timestamp,
  `updated_at` datetime NOT NULL default current_timestamp
);

-- +migrate Down
DROP TABLE `account_types`;
DROP TABLE `accounts`;
DROP TABLE `cash_flow_types`;
DROP TABLE `cash_flows`;
DROP TABLE `categories`;
DROP TABLE `category_types`;
DROP TABLE `companies`;
DROP TABLE `currency_types`;
DROP TABLE `imports`;
DROP TABLE `institutions`;
DROP TABLE `ofx_accounts`;
DROP TABLE `payees`;
DROP TABLE `repeat_interval_types`;
DROP TABLE `repeat_intervals`;
DROP TABLE `securities`;
DROP TABLE `security_basis_types`;
DROP TABLE `security_types`;
DROP TABLE `tax_cash_flows`;
DROP TABLE `tax_categories`;
DROP TABLE `tax_constants`;
DROP TABLE `tax_filing_status`;
DROP TABLE `tax_items`;
DROP TABLE `tax_regions`;
DROP TABLE `tax_types`;
DROP TABLE `tax_users`;
DROP TABLE `tax_years`;
DROP TABLE `taxes`;
DROP TABLE `trade_gains`;
DROP TABLE `trade_types`;
DROP TABLE `trades`;
DROP TABLE `user_settings`;
DROP TABLE `users`;
