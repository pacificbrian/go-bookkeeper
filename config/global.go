/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package config

import flag "github.com/spf13/pflag"

type GlobalConfiguration struct {
	Sessions bool `toml:"sessions"`
	CashFlowLimit int `toml:"cashflow_limit"`
	EnableAutoTaxes bool `toml:"enable_auto_taxes" env-default:"true"`
	EnableImportTradeFixups bool `toml:"enable_import_trade_fixups" env-default:"false"`
}

var DebugFlag bool
var globalConfig *GlobalConfiguration

func GlobalConfig() *GlobalConfiguration {
	return globalConfig
}

func init() {
	flag.BoolVarP(&DebugFlag, "debug", "d", false, "run in debug mode")
	flag.Parse()

	globalConfig = &GetConfig().GlobalConfiguration
}
