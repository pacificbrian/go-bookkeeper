/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package config

import flag "github.com/spf13/pflag"

type GlobalConfiguration struct {
	ServerPort int `toml:"server_port" env:"GOBOOK_SERVER_PORT" env-default:"3000"`
	CashFlowLimit int `toml:"cashflow_limit"`
	LimitImportPayeeNameLength bool
	Sessions bool
	UpdateAccountsOnLogin bool
	// booleans must default to false, see cleanenv #61,#82
	EnableSecurityCharts bool `toml:"enable_security_charts"`
	EnableSecurityFilings bool `toml:"enable_security_filings"`
	DisableAutoTaxes bool `toml:"disable_auto_taxes"`
	DisableImportPayeeNameLength bool `toml:"disable_import_payee_name_length"`
	DisableSessions bool `toml:"disable_sessions"`
	DisableUpdateAccountsOnLogin bool `toml:"disable_update_accounts_on_login"`
	EnableImportTradeFixups bool `toml:"enable_import_trade_fixups"`
}

var DebugFlag bool
var globalConfig *GlobalConfiguration

func GlobalConfig() *GlobalConfiguration {
	if globalConfig == nil {
		globalConfig = &GetConfig().GlobalConfiguration
	}
	globalConfig.LimitImportPayeeNameLength = !globalConfig.DisableImportPayeeNameLength
	globalConfig.Sessions = !globalConfig.DisableSessions
	globalConfig.UpdateAccountsOnLogin = !globalConfig.DisableUpdateAccountsOnLogin
	return globalConfig
}

func init() {
	flag.BoolVarP(&DebugFlag, "debug", "d", false, "run in debug mode")
	flag.Parse()
}
