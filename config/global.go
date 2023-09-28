/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package config

import flag "github.com/spf13/pflag"

type GlobalConfiguration struct {
	ServerPort int `toml:"server_port" env:"GOBOOK_SERVER_PORT" env-default:"3000"`
	Sessions bool `toml:"sessions" env-default:"true"`
	CashFlowLimit int `toml:"cashflow_limit"`
	UpdateAccountsOnLogin bool
	DisableUpdateAccountsOnLogin bool `toml:"disable_update_accounts_on_login"`
	EnableAutoTaxes bool `toml:"enable_auto_taxes" env-default:"true"`
	EnableImportTradeFixups bool `toml:"enable_import_trade_fixups" env-default:"false"`
}

var DebugFlag bool
var globalConfig *GlobalConfiguration

func GlobalConfig() *GlobalConfiguration {
	if globalConfig == nil {
		globalConfig = &GetConfig().GlobalConfiguration
	}
	globalConfig.UpdateAccountsOnLogin = !globalConfig.DisableUpdateAccountsOnLogin
	return globalConfig
}

func init() {
	flag.BoolVarP(&DebugFlag, "debug", "d", false, "run in debug mode")
	flag.Parse()
}
